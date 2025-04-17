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
	"encoding/json"

	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

func packageDump() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "dump [flags] deployment"
	cmd.Short = "Dumps the current setup of the platform"

	if err := cli.RegisterFlags(&cmd); err != nil {
		return nil, err
	}

	cmd.RunE = getRunner().With(packageDumpRun).Run

	return &cmd, nil
}

func packageDumpRun(cmd *cobra.Command, args []string) error {
	charts, err := fetchLocallyInstalledCharts(cmd)
	if err != nil {
		return err
	}

	hclient, err := getHelmClient(cmd)
	if err != nil {
		return errors.Wrapf(err, "Unable to get helm client")
	}

	if len(args) < 1 {
		return errors.Errorf("Invalid arguments")
	}

	deployment := args[0]

	var out Package

	out.Packages = map[string]string{}

	out.Releases = map[string]Release{}

	for name, c := range charts {
		if !c.Status.Conditions.IsTrue(platformApi.ReadyCondition) {
			return errors.Errorf("Chart `%s` is not in ready condition", name)
		}
		if info := c.Status.Info; info != nil {
			if det := info.Details; det != nil {
				out.Packages[name] = c.Status.Info.Details.GetVersion()
			}
		}

		existingReleases, err := hclient.List(cmd.Context(), func(in *action.List) {
			in.Selector = meta.FormatLabelSelector(&meta.LabelSelector{
				MatchLabels: map[string]string{
					constants.HelmLabelArangoDBManaged:    "true",
					constants.HelmLabelArangoDBDeployment: deployment,
					constants.HelmLabelArangoDBChart:      name,
					constants.HelmLabelArangoDBType:       "platform",
				},
			})
		})
		if err != nil {
			logger.Err(err).Error("Unable to list releases")
			return err
		}

		for _, release := range existingReleases {
			var r Release

			r.Package = name

			data, err := release.Values.Marshal()
			if err != nil {
				logger.Err(err).Error("Unable to unmarshal values")
				return err
			}

			delete(data, "arangodb_platform")

			if len(data) != 0 {
				values, err := helm.NewValues(data)
				if err != nil {
					logger.Err(err).Error("Unable to marshal values")
					return err
				}

				r.Overrides = values
			}

			out.Releases[release.Name] = r
		}
	}

	d, err := json.Marshal(out)
	if err != nil {
		return err
	}

	d, err = yaml.JSONToYAML(d)
	if err != nil {
		return err
	}

	return render(cmd, "---\n\n%s", string(d))
}
