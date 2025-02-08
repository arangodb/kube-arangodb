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
	"github.com/spf13/cobra"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/pretty"
)

type RegistryTable struct {
	Name             string `table:"Name" table_align:"center" table_header_align:"center"`
	Description      string `table:"Description" table_align:"center" table_header_align:"center"`
	LatestVersion    string `table:"Latest Version" table_align:"center" table_header_align:"center"`
	Installed        bool   `table:"Installed" table_align:"center" table_header_align:"center"`
	Valid            string `table:"Valid" table_align:"center" table_header_align:"center"`
	InstalledVersion string `table:"Installed Version" table_align:"center" table_header_align:"center"`
}

func registryStatus() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "status"
	cmd.Short = "Describes Charts Status"

	if err := cli.RegisterFlags(&cmd, flagPlatformStage, flagPlatformEndpoint, flagPlatformName, flagOutput); err != nil {
		return nil, err
	}

	cmd.RunE = getRunner().With(registryStatusRun).Run

	return &cmd, nil
}

func registryStatusRun(cmd *cobra.Command, args []string) error {
	t := pretty.NewTable[RegistryTable]()

	hm, err := getChartManager(cmd)
	if err != nil {
		return err
	}

	charts, err := fetchLocallyInstalledCharts(cmd)
	if err != nil {
		return err
	}

	for _, name := range hm.Repositories() {
		if shared.ValidateResourceName(name) != nil {
			continue
		}

		repo, ok := hm.Get(name)
		if !ok {
			continue
		}

		version, ok := repo.Latest()

		if !ok {
			continue
		}

		c, ok := charts[name]

		t.Add(RegistryTable{
			Name:          name,
			Description:   version.Chart().Description,
			LatestVersion: version.Chart().Version,
			Installed:     ok,
			Valid: func() string {
				if ok {
					return util.BoolSwitch(c.Status.Conditions.IsTrue(platformApi.ReadyCondition), "true", "false")
				} else {
					return "N/A"
				}
			}(),
			InstalledVersion: func() string {
				if ok {
					if info := c.Status.Info; info != nil {
						if det := info.Details; det != nil {
							return det.Version
						}
					}
				}

				return "N/A"
			}(),
		})
	}

	return renderOutput(cmd, t)
}
