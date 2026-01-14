//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/config"
	"github.com/spf13/cobra"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/platform/pack"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
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

	if err := cli.RegisterFlags(&cmd, flagPlatformName, flagLicenseManager, flagRegistry, flagLicenseManagerDiscoverCredentials); err != nil {
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

	ns, err := flagNamespace.Get(cmd)
	if err != nil {
		return err
	}

	deployment, err := flagPlatformName.Get(cmd)
	if err != nil {
		return err
	}

	var hosts map[string]util.ModR[config.Host]

	if newHosts, err := cli.LicenseManagerRegistryHosts(cmd, flagLicenseManager, newDeploymentSecretLicenseProviderWrap(client, ns, deployment, flagLicenseManager)); err != nil {
		logger.Err(err).Warn("Unable to fetch credentials")
	} else {
		hosts = newHosts
	}

	reg, err := flagRegistry.Client(cmd, hosts)
	if err != nil {
		return err
	}

	endpoint, err := flagLicenseManager.Endpoint(cmd)
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

	logger.Info("Chart Update")

	if err := packageInstallRunInstallCharts(cmd, client, reg, ns, endpoint, r); err != nil {
		return err
	}

	logger.Info("Service Update")

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
		log = log.Str("type", "release").Str("name", name)

		log.Info("Calculating installation")

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

			log.Debug("Installing Service")

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

			log.Info("Installed Service")
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
				log.Info("Updated Service: %s", name)
			}
		}

		// Ensure we wait for reconcile
		time.Sleep(time.Second)

		log.Info("Waiting...")

		if err := h.Timeout(ctx, t, func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
			log = log.Str("type", "release").Str("name", name)

			svc, err := client.Arango().PlatformV1beta1().ArangoPlatformServices(deployment.GetNamespace()).Get(ctx, name, meta.GetOptions{})
			if err != nil {
				return err
			}

			if svc.Status.ChartInfo == nil {
				log.Warn("No chart info")
				return nil
			}

			if svc.Status.ChartInfo.Checksum != chart.Status.Info.GetChecksum() {
				log.Warn("Chart not yet updated")
				return nil
			}

			if !svc.Status.Conditions.IsTrue(platformApi.ReleaseReadyCondition) {
				log.Warn("Service not yet ready")
				return nil
			}

			return io.EOF
		}, 5*time.Minute, 15*time.Second); err != nil {
			if errors.Is(err, io.EOF) {
				return errors.Errorf("Service %s is not ready", name)
			}

			return err
		}
		log.Info("Ready Release")

		return nil
	})
}

func packageInstallRunInstallCharts(cmd *cobra.Command, client kclient.Client, reg *regclient.RegClient, ns, endpoint string, r helm.Package) error {
	return executor.Run(cmd.Context(), logger, 8, func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		charts, err := fetchLocallyInstalledCharts(cmd)
		if err != nil {
			return err
		}

		for name, packageSpec := range r.Packages {
			packageInstallRunInstallChart(cmd, h, client, reg, ns, endpoint, charts, name, packageSpec)
		}

		return nil
	})
}

func packageInstallRunInstallChart(cmd *cobra.Command, h executor.Handler, client kclient.Client, reg *regclient.RegClient, ns, endpoint string, charts map[string]*platformApi.ArangoPlatformChart, name string, packageSpec helm.PackageSpec) {
	h.RunAsync(cmd.Context(), func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		log = log.Str("type", "chart").Str("name", name)

		log.Info("Calculating installation")

		chart, err := pack.ResolvePackageSpec(ctx, endpoint, name, packageSpec, reg, nil)
		if err != nil {
			return err
		}

		log = logger.Str("chart", name).Str("version", packageSpec.Version)

		if c, ok := charts[name]; !ok {
			log.Debug("Installing Chart")

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

			log.Info("Installed Chart")
		} else {
			if c.Spec.Definition.SHA256() != chart.SHA256SUM() || !packageSpec.Overrides.Equals(helm.Values(c.Spec.Overrides)) {
				c.Spec.Definition = sharedApi.Data(chart)
				c.Spec.Overrides = sharedApi.Any(packageSpec.Overrides)
				_, err := client.Arango().PlatformV1beta1().ArangoPlatformCharts(ns).Update(cmd.Context(), c, meta.UpdateOptions{})
				if err != nil {
					return err
				}
				log.Info("Updated Chart")
			}
		}

		log.Info("Waiting...")

		if err := h.Timeout(ctx, t, func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
			log = log.Str("type", "chart").Str("name", name)

			c, err := client.Arango().PlatformV1beta1().ArangoPlatformCharts(ns).Get(ctx, name, meta.GetOptions{})
			if err != nil {
				return err
			}

			if !c.Ready() {
				log.Warn("Chart not yet ready")
				return nil
			}

			if c.Status.Info == nil {
				log.Warn("Chart not yet accepted")
				return nil
			}

			return io.EOF
		}, 5*time.Minute, 5*time.Second); err != nil {
			if errors.Is(err, io.EOF) {
				return errors.Errorf("Chart %s is not ready", name)
			}

			return err
		}
		log.Info("Ready Chart")

		return nil
	})
}

func newDeploymentSecretLicenseProviderWrap(client kclient.Client, namespace, name string, parent cli.LicenseManagerAuthProvider) cli.LicenseManagerAuthProvider {
	return cli.LicenseManagerStaticAuthProvider(func(cmd *cobra.Command) (string, string, error) {
		if clientID, clientSecret, err := parent.ClientCredentials(cmd); err == nil {
			return clientID, clientSecret, nil
		} else {
			discover, err := flagLicenseManagerDiscoverCredentials.Get(cmd)
			if err != nil || !discover {
				logger.Debug("Deployment Discovery of credentials disabled")
				return "", "", err
			}
		}

		logger.Info("Fetching external client credentials")

		depl, err := client.Arango().DatabaseV1().ArangoDeployments(namespace).Get(cmd.Context(), name, meta.GetOptions{})
		if err != nil {
			return "", "", err
		}

		accepted := depl.GetAcceptedSpec()

		if !accepted.License.HasSecretName() {
			return "", "", errors.Errorf("License Secret not provided in ArangoDeployment")
		}

		secret, err := client.Kubernetes().CoreV1().Secrets(namespace).Get(cmd.Context(), accepted.License.GetSecretName(), meta.GetOptions{})
		if err != nil {
			return "", "", err
		}

		clientID, ok := secret.Data[utilConstants.SecretKeyLicenseClientID]
		if !ok {
			return "", "", errors.Errorf("Secret %s does not have a client id", secret.Name)
		}

		clientSecret, ok := secret.Data[utilConstants.SecretKeyLicenseClientSecret]
		if !ok {
			return "", "", errors.Errorf("Secret %s does not have a client id", secret.Name)
		}

		logger.Str("ClientID", string(clientID)).Info("Using Client License Secret")

		return string(clientID), string(clientSecret), nil
	})
}
