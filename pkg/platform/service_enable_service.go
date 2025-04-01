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

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func serviceEnableService() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "enable-service [flags] deployment chart"
	cmd.Short = "Manages Service Installation/Management"

	if err := cli.RegisterFlags(&cmd, flagPlatformStage, flagPlatformEndpoint, flagPlatformName, flagValues); err != nil {
		return nil, err
	}

	cmd.RunE = getRunner().With(serviceEnableServiceRun).Run

	return &cmd, nil
}

func serviceEnableServiceRun(cmd *cobra.Command, args []string) error {
	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return errors.Errorf("Unable to get client")
	}

	hclient, err := getHelmClient(cmd)
	if err != nil {
		return errors.Wrapf(err, "Unable to get helm client")
	}

	if len(args) < 2 {
		return errors.Errorf("Invalid arguments")
	}

	ns, err := flagNamespace.Get(cmd)
	if err != nil {
		return err
	}

	deployment, chart := args[0], args[1]

	logger := logger.Str("deployment", deployment).Str("name", chart).Str("chart", chart)

	files, err := flagValues.Get(cmd)
	if err != nil {
		return err
	}

	datas := make([]any, len(files)+1)

	datas[0] = map[string]any{
		"arangodb_platform": map[string]any{
			"deployment": map[string]any{
				"name": deployment,
			},
		},
	}

	for id, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return errors.Wrapf(err, "Unable to find file: %s", f)
		}

		d, err := util.JsonOrYamlUnmarshal[map[string]any](data)
		if err != nil {
			return errors.Wrapf(err, "Unable to load file: %s", f)
		}

		datas[id+1] = d
	}

	mergedData, err := helm.NewMergeValues(helm.MergeMaps, datas...)
	if err != nil {
		return errors.Wrapf(err, "Unable to build helm data")
	}

	logger.Info("Fetch ArangoDeployment")
	deploymentObject, err := client.Arango().DatabaseV1().ArangoDeployments(ns).Get(cmd.Context(), deployment, meta.GetOptions{})
	if err != nil {
		logger.Err(err).Error("Unable to get deployment")
		return err
	}
	logger.Str("uid", string(deploymentObject.GetUID())).Info("ArangoDeployment Found")

	logger.Info("Fetch ArangoPlatformChart")

	chartObject, err := waitForChart(cmd.Context(), client, ns, chart).Run(cmd.Context(), time.Minute, time.Second)
	if err != nil {
		return err
	}

	logger.Str("uid", string(chartObject.GetUID())).Info("ArangoPlatformChart Found")

	if current, err := hclient.Status(cmd.Context(), chart); err != nil {
		println(err)
	} else if current == nil {
		logger.Info("Service not found, installing")
		if _, err := hclient.Install(cmd.Context(), helm.Chart(chartObject.Spec.Definition), mergedData, func(in *action.Install) {
			in.Labels = map[string]string{
				constants.HelmLabelArangoDBManaged:    "true",
				constants.HelmLabelArangoDBDeployment: deployment,
				constants.HelmLabelArangoDBChart:      chart,
				constants.HelmLabelArangoDBType:       "platform",
			}
			in.ReleaseName = chart
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
			if _, err := hclient.Upgrade(cmd.Context(), chart, helm.Chart(chartObject.Spec.Definition), mergedData, func(in *action.Upgrade) {
				in.Labels = map[string]string{
					constants.HelmLabelArangoDBManaged:    "true",
					constants.HelmLabelArangoDBDeployment: deployment,
					constants.HelmLabelArangoDBChart:      chart,
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

	return nil
}
