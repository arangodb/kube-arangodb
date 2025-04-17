//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//

package platform

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func packageInstall() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "install [flags] deployment package"
	cmd.Short = "Installs the specified setup of the platform"

	if err := cli.RegisterFlags(&cmd, flagPlatformStage, flagPlatformEndpoint); err != nil {
		return nil, err
	}

	cmd.RunE = getRunner().With(packageInstallRun).Run

	return &cmd, nil
}

func packageInstallRun(cmd *cobra.Command, args []string) error {
	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return errors.Errorf("Unable to get client")
	}

	charts, err := fetchLocallyInstalledCharts(cmd)
	if err != nil {
		return err
	}

	hm, err := getChartManager(cmd)
	if err != nil {
		return err
	}

	ns, err := flagNamespace.Get(cmd)
	if err != nil {
		return err
	}

	hclient, err := getHelmClient(cmd)
	if err != nil {
		return errors.Wrapf(err, "Unable to get helm client")
	}

	if len(args) < 2 {
		return errors.Errorf("Invalid arguments")
	}

	deployment := args[0]
	file := args[1]

	data, err := os.ReadFile(file)
	if err != nil {
		logger.Err(err).Error("Unable to read the file")
		return err
	}

	r, err := util.JsonOrYamlUnmarshal[Package](data)
	if err != nil {
		logger.Err(err).Error("Unable to read the file")
		return err
	}

	for name, version := range r.Packages {
		def, ok := hm.Get(name)
		if !ok {
			return errors.Errorf("Unable to get '%s' chart", name)
		}

		ver, ok := def.Get(version)
		if !ok {
			return errors.Errorf("Unable to get '%s' chart in version `%s`", name, version)
		}

		chart, err := ver.Get(cmd.Context())
		if err != nil {
			return errors.Wrapf(err, "Unable to download chart %s-%s", name, ver.Version())
		}

		logger := logger.Str("chart", name).Str("version", ver.Version())

		if c, ok := charts[name]; !ok {
			logger.Debug("Installing Chart: %s", name)

			_, err := client.Arango().PlatformV1alpha1().ArangoPlatformCharts(ns).Create(cmd.Context(), &platformApi.ArangoPlatformChart{
				ObjectMeta: meta.ObjectMeta{
					Name:      name,
					Namespace: ns,
				},
				Spec: platformApi.ArangoPlatformChartSpec{
					Definition: sharedApi.Data(chart),
				},
			}, meta.CreateOptions{})
			if err != nil {
				return err
			}

			logger.Debug("Installed Chart: %s", name)
		} else {
			if c.Spec.Definition.SHA256() != chart.SHA256SUM() {
				c.Spec.Definition = sharedApi.Data(chart)
				_, err := client.Arango().PlatformV1alpha1().ArangoPlatformCharts(ns).Update(cmd.Context(), c, meta.UpdateOptions{})
				if err != nil {
					return err
				}
				logger.Debug("Updated Chart: %s", name)
			}
		}
	}

	logger.Info("Fetch ArangoDeployment")
	deploymentObject, err := client.Arango().DatabaseV1().ArangoDeployments(ns).Get(cmd.Context(), deployment, meta.GetOptions{})
	if err != nil {
		logger.Err(err).Error("Unable to get deployment")
		return err
	}
	logger.Str("uid", string(deploymentObject.GetUID())).Info("ArangoDeployment Found")

	for name, release := range r.Releases {
		ov, err := release.Overrides.Marshal()
		if err != nil {
			logger.Err(err).Error("Unable to unmarshal values")
			return err
		}

		mergedData, err := helm.NewMergeValues(helm.MergeMaps, map[string]any{
			"arangodb_platform": map[string]any{
				"deployment": map[string]any{
					"name": deployment,
				},
			},
		},

			ov)
		if err != nil {
			return errors.Wrapf(err, "Unable to build helm data")
		}

		logger.Info("Fetch ArangoPlatformChart")

		chartObject, err := waitForChart(cmd.Context(), client, ns, release.Package).Run(cmd.Context(), time.Minute, time.Second)
		if err != nil {
			return err
		}

		logger.Str("uid", string(chartObject.GetUID())).Info("ArangoPlatformChart Found")

		if current, err := hclient.Status(cmd.Context(), name); err != nil {
			return err
		} else if current == nil {
			logger.Info("Service not found, installing")
			if _, err := hclient.Install(cmd.Context(), helm.Chart(chartObject.Spec.Definition), mergedData, func(in *action.Install) {
				in.Labels = map[string]string{
					constants.HelmLabelArangoDBManaged:    "true",
					constants.HelmLabelArangoDBDeployment: deployment,
					constants.HelmLabelArangoDBChart:      release.Package,
					constants.HelmLabelArangoDBType:       "platform",
				}
				in.ReleaseName = name
				in.Namespace = ns
			}); err != nil {
				return err
			}
			logger.Info("Service installed")
		} else {
			logger.Info("Service found, comparing")

			changed := false
			if current.GetChart().GetMetadata().GetVersion() != chartObject.Status.Info.Details.GetVersion() {
				logger.Info("Chart version expected: %s", current.GetChart().GetMetadata().GetVersion())
				changed = true
			}

			if !current.Values.Equals(mergedData) {
				changed = true
				logger.Info("Service values update required")
			} else {
				logger.Info("Service values update not required")
			}

			if changed {
				if _, err := hclient.Upgrade(cmd.Context(), name, helm.Chart(chartObject.Spec.Definition), mergedData, func(in *action.Upgrade) {
					in.Labels = map[string]string{
						constants.HelmLabelArangoDBManaged:    "true",
						constants.HelmLabelArangoDBDeployment: deployment,
						constants.HelmLabelArangoDBChart:      release.Package,
						constants.HelmLabelArangoDBType:       "platform",
					}
					in.Namespace = ns
				}); err != nil {
					return err
				}
				logger.Info("Service updated")
			} else {
				logger.Info("Service up-to-date")
			}
		}
	}

	return nil
}
