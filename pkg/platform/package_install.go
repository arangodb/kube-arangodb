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
	"context"
	"io"
	"time"

	"github.com/spf13/cobra"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/executor"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func packageInstall() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "install [flags] ... packages"
	cmd.Short = "Installs the specified setup of the platform"

	if err := cli.RegisterFlags(&cmd, flagPlatformEndpoint, flagPlatformName); err != nil {
		return nil, err
	}

	cmd.RunE = getRunner().With(packageInstallRun).Run

	return &cmd, nil
}

func packageInstallRun(cmd *cobra.Command, args []string) error {
	client, err := getKubernetesClient(cmd)
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

	deployment, err := flagPlatformName.Get(cmd)
	if err != nil {
		return err
	}

	dApi, err := client.Arango().DatabaseV1().ArangoDeployments(ns).Get(cmd.Context(), deployment, meta.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "Unable to find deployment")
	}

	if len(args) < 1 {
		return errors.Errorf("Invalid arguments")
	}

	r, err := getHelmPackages(args...)
	if err != nil {
		return err
	}

	if err := packageInstallRunInstallCharts(cmd, client, hm, ns, r); err != nil {
		return err
	}

	if err := packageInstallRunInstallServices(cmd, client, dApi, r); err != nil {
		return err
	}

	return nil
}

func packageInstallRunInstallServices(cmd *cobra.Command, client kclient.Client, deployment *api.ArangoDeployment, r helm.Package) error {
	return executor.Run(cmd.Context(), logger, 8, func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		for name, releaseSpec := range r.Releases {
			packageInstallRunInstallRelease(cmd, h, client, deployment, name, releaseSpec)
		}

		return nil
	})
}

func packageInstallRunInstallRelease(cmd *cobra.Command, h executor.Handler, client kclient.Client, deployment *api.ArangoDeployment, name string, packageSpec helm.PackageRelease) {
	h.RunAsync(cmd.Context(), func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		chart, err := client.Arango().PlatformV1beta1().ArangoPlatformCharts(deployment.GetNamespace()).Get(ctx, packageSpec.Package, meta.GetOptions{})
		if err != nil {
			return err
		}

		if !chart.Ready() {
			return errors.Errorf("Chart %s is not ready", name)
		}

		if svc, err := client.Arango().PlatformV1beta1().ArangoPlatformServices(deployment.GetNamespace()).Get(ctx, name, meta.GetOptions{}); err != nil {
			if !kerrors.IsNotFound(err) {
				return err
			}

			logger.Debug("Installing Service: %s", name)

			// Prepare Object
			if _, err := client.Arango().PlatformV1beta1().ArangoPlatformServices(deployment.GetNamespace()).Create(ctx, &platformApi.ArangoPlatformService{
				ObjectMeta: meta.ObjectMeta{
					Name:      name,
					Namespace: deployment.GetNamespace(),
				},
				Spec: platformApi.ArangoPlatformServiceSpec{
					Deployment: &sharedApi.Object{
						Name:      deployment.GetName(),
						Namespace: util.NewType(deployment.GetNamespace()),
					},
					Chart: &sharedApi.Object{
						Name: packageSpec.Package,
					},
					Values: sharedApi.Any(packageSpec.Overrides),
				},
			}, meta.CreateOptions{}); err != nil {
				return err
			}

			logger.Info("Installed Service: %s", name)
		} else {
			if svc.Spec.Deployment.GetName() != deployment.GetName() {
				return errors.Errorf("Unable to change Deployment name for %s", name)
			}

			if svc.Spec.Chart.GetName() != chart.GetName() {
				return errors.Errorf("Unable to change Chart name for %s", name)
			}

			if !svc.Spec.Values.Equals(sharedApi.Any(packageSpec.Overrides)) {
				svc.Spec.Values = sharedApi.Any(packageSpec.Overrides)
				_, err := client.Arango().PlatformV1beta1().ArangoPlatformServices(deployment.GetNamespace()).Update(ctx, svc, meta.UpdateOptions{})
				if err != nil {
					return err
				}
				logger.Info("Updated Service: %s", name)
			}
		}

		// Ensure we wait for reconcile
		time.Sleep(time.Second)

		if err := h.Timeout(ctx, t, func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
			svc, err := client.Arango().PlatformV1beta1().ArangoPlatformServices(deployment.GetNamespace()).Get(ctx, name, meta.GetOptions{})
			if err != nil {
				return err
			}

			if svc.Status.ChartInfo == nil {
				return nil
			}

			if svc.Status.ChartInfo.Checksum != chart.Status.Info.GetChecksum() {
				return nil
			}

			if !svc.Status.Conditions.IsTrue(platformApi.ReleaseReadyCondition) {
				return nil
			}

			return io.EOF
		}, 5*time.Minute, time.Second); err != nil {
			if errors.Is(err, io.EOF) {
				return errors.Errorf("Service %s is not ready", name)
			}

			return err
		}

		return nil
	})
}

func packageInstallRunInstallCharts(cmd *cobra.Command, client kclient.Client, hm helm.ChartManager, ns string, r helm.Package) error {
	return executor.Run(cmd.Context(), logger, 8, func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		charts, err := fetchLocallyInstalledCharts(cmd)
		if err != nil {
			return err
		}

		for name, packageSpec := range r.Packages {
			packageInstallRunInstallChart(cmd, h, client, hm, ns, charts, name, packageSpec)
		}

		return nil
	})
}

func packageInstallRunInstallChart(cmd *cobra.Command, h executor.Handler, client kclient.Client, hm helm.ChartManager, ns string, charts map[string]*platformApi.ArangoPlatformChart, name string, packageSpec helm.PackageSpec) {
	h.RunAsync(cmd.Context(), func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		chart, err := packageInstallRunChartExtract(cmd, hm, name, packageSpec)
		if err != nil {
			return err
		}

		logger := logger.Str("chart", name).Str("version", packageSpec.Version)

		if c, ok := charts[name]; !ok {
			logger.Debug("Installing Chart: %s", name)

			_, err := client.Arango().PlatformV1beta1().ArangoPlatformCharts(ns).Create(cmd.Context(), &platformApi.ArangoPlatformChart{
				ObjectMeta: meta.ObjectMeta{
					Name:      name,
					Namespace: ns,
				},
				Spec: platformApi.ArangoPlatformChartSpec{
					Definition: sharedApi.Data(chart),
					Overrides:  sharedApi.Any(packageSpec.Overrides),
				},
			}, meta.CreateOptions{})
			if err != nil {
				return err
			}

			logger.Info("Installed Chart: %s", name)
		} else {
			if c.Spec.Definition.SHA256() != chart.SHA256SUM() || !packageSpec.Overrides.Equals(helm.Values(c.Spec.Overrides)) {
				c.Spec.Definition = sharedApi.Data(chart)
				c.Spec.Overrides = sharedApi.Any(packageSpec.Overrides)
				_, err := client.Arango().PlatformV1beta1().ArangoPlatformCharts(ns).Update(cmd.Context(), c, meta.UpdateOptions{})
				if err != nil {
					return err
				}
				logger.Info("Updated Chart: %s", name)
			}
		}

		if err := h.Timeout(ctx, t, func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
			c, err := client.Arango().PlatformV1beta1().ArangoPlatformCharts(ns).Get(ctx, name, meta.GetOptions{})
			if err != nil {
				return err
			}

			if !c.Ready() {
				return nil
			}

			if c.Status.Info == nil {
				return nil
			}

			return io.EOF
		}, 5*time.Minute, time.Second); err != nil {
			if errors.Is(err, io.EOF) {
				return errors.Errorf("Chart %s is not ready", name)
			}

			return err
		}

		return nil
	})
}

func packageInstallRunChartExtract(cmd *cobra.Command, hm helm.ChartManager, name string, spec helm.PackageSpec) (helm.Chart, error) {
	if !spec.Chart.IsZero() {
		return helm.Chart(spec.Chart), nil
	}
	def, ok := hm.Get(name)
	if !ok {
		return helm.Chart{}, errors.Errorf("Unable to get '%s' chart", name)
	}

	ver, ok := def.Get(spec.Version)
	if !ok {
		return helm.Chart{}, errors.Errorf("Unable to get '%s' chart in version `%s`", name, spec.Version)
	}

	c, err := ver.Get(cmd.Context())
	if err != nil {
		return helm.Chart{}, errors.Wrapf(err, "Unable to download chart %s-%s", name, ver.Version())
	}

	return c, nil
}
