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
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/pretty"
)

func registryInstall() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "install [flags] [...charts]"
	cmd.Short = "Manages the Chart Installation"

	if err := cli.RegisterFlags(&cmd, flagPlatformStage, flagPlatformEndpoint, flagPlatformName, flagOutput, flagUpgradeVersions, flagAll); err != nil {
		return nil, err
	}

	cmd.RunE = getRunner().With(registryInstallRun).Run

	return &cmd, nil
}

func registryInstallRun(cmd *cobra.Command, args []string) error {
	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return errors.Errorf("Unable to get client")
	}

	ns, err := flagNamespace.Get(cmd)
	if err != nil {
		return err
	}

	all, err := flagAll.Get(cmd)
	if err != nil {
		return err
	}

	update, err := flagUpgradeVersions.Get(cmd)
	if err != nil {
		return err
	}

	hm, err := getChartManager(cmd)
	if err != nil {
		return err
	}

	var toInstall = map[string]helm.ChartManagerRepoVersion{}

	charts, err := fetchLocallyInstalledCharts(cmd)
	if err != nil {
		return err
	}

	if all {
		for _, repo := range hm.Repositories() {
			if shared.ValidateResourceName(repo) != nil {
				continue
			}
			chart, ok := hm.Get(repo)
			if !ok {
				continue
			}

			if _, ok := charts[repo]; !ok || update {
				ver, ok := chart.Latest()
				if !ok {
					return errors.Errorf("Unable to get 'latest' version for: %s", repo)
				}
				toInstall[repo] = ver
			}
		}
	}

	for _, repo := range args {
		p := strings.SplitN(repo, "=", 2)

		chart, ok := hm.Get(p[0])
		if !ok {
			continue
		}

		if len(p) == 1 {
			if _, ok := charts[repo]; !ok || update {
				ver, ok := chart.Latest()
				if !ok {
					return errors.Errorf("Unable to get 'latest' version for: %s", repo)
				}
				toInstall[repo] = ver
			}
		} else {
			ver, ok := chart.Get(p[1])
			if !ok {
				return errors.Errorf("Unable to get '%s' version for: %s", p[1], p[0])
			}
			toInstall[p[0]] = ver
		}
	}

	t := pretty.NewTable[RegistryTable]()

	for name, ver := range toInstall {
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

	for name, ver := range toInstall {
		logger := logger.Str("chart", name).Str("version", ver.Version())

		logger.Debug("Wait For Chart: %s", name)

		if _, err := waitForChart(cmd.Context(), client, ns, name).With(func(in *platformApi.ArangoPlatformChart) error {
			if in.Status.Info.Details.Version != ver.Version() {
				return nil
			}

			return io.EOF
		}).Run(cmd.Context(), time.Minute, time.Second); err != nil {
			return err
		}
	}

	charts, err = fetchLocallyInstalledCharts(cmd)
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
