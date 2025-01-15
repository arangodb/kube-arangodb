//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	"context"
	"fmt"
	"path"
	"path/filepath"

	core "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	pbImplEnvoyAuthV3 "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1alpha1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/gateway"
	"github.com/arangodb/kube-arangodb/pkg/platform"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/patcher"
)

func (r *Resources) ensureGatewayConfig(ctx context.Context, cachedStatus inspectorInterface.Inspector, configMaps generic.ModClient[*core.ConfigMap]) error {
	cfg, err := r.renderGatewayConfig(cachedStatus)
	if err != nil {
		return errors.WithStack(errors.Wrapf(err, "Failed to generate gateway config"))
	}

	_, baseGatewayCfgYamlChecksum, _, err := cfg.RenderYAML()
	if err != nil {
		return errors.WithStack(errors.Wrapf(err, "Failed to render gateway config"))
	}

	cfg.Destinations[constants.EnvoyInventoryConfigDestination] = gateway.ConfigDestination{
		Type: util.NewType(gateway.ConfigDestinationTypeStatic),
		AuthExtension: &gateway.ConfigAuthZExtension{
			AuthZExtension: map[string]string{
				pbImplEnvoyAuthV3.AuthConfigAuthRequiredKey: pbImplEnvoyAuthV3.AuthConfigKeywordTrue,
				pbImplEnvoyAuthV3.AuthConfigAuthPassModeKey: string(networkingApi.ArangoRouteSpecAuthenticationPassModeRemove),
			},
		},
		Static: &gateway.ConfigDestinationStatic{
			Code: util.NewType[uint32](200),
			Response: &platform.State{
				Configuration: platform.StateConfiguration{
					Hash: baseGatewayCfgYamlChecksum,
				},
			},
		},
	}

	gatewayCfgYaml, _, _, err := cfg.RenderYAML()
	if err != nil {
		return errors.WithStack(errors.Wrapf(err, "Failed to render gateway config"))
	}

	gatewayCfgCDSYaml, _, _, err := cfg.RenderCDSYAML()
	if err != nil {
		return errors.WithStack(errors.Wrapf(err, "Failed to render gateway cds config"))
	}

	gatewayCfgLDSYaml, _, _, err := cfg.RenderLDSYAML()
	if err != nil {
		return errors.WithStack(errors.Wrapf(err, "Failed to render gateway lds config"))
	}

	if err := r.ensureGatewayConfigMap(ctx, cachedStatus, configMaps, GetGatewayConfigMapName(r.context.GetAPIObject().GetName()), map[string]string{
		constants.GatewayConfigFileName: string(gatewayCfgYaml),
	}); err != nil {
		return err
	}

	if err := r.ensureGatewayConfigMap(ctx, cachedStatus, configMaps, GetGatewayConfigMapName(r.context.GetAPIObject().GetName(), "cds"), map[string]string{
		constants.GatewayConfigFileName: string(gatewayCfgCDSYaml),
	}); err != nil {
		return err
	}

	if err := r.ensureGatewayConfigMap(ctx, cachedStatus, configMaps, GetGatewayConfigMapName(r.context.GetAPIObject().GetName(), "lds"), map[string]string{
		constants.GatewayConfigFileName: string(gatewayCfgLDSYaml),
	}); err != nil {
		return err
	}

	return nil
}

func (r *Resources) ensureGatewayConfigMap(ctx context.Context, cachedStatus inspectorInterface.Inspector, configMaps generic.ModClient[*core.ConfigMap], name string, data map[string]string) error {
	log := r.log.Str("section", "gateway-config").Str("name", name)

	elements, err := r.renderConfigMap(data)
	if err != nil {
		return errors.WithStack(errors.Wrapf(err, "Failed to render gateway config"))
	}

	if cm, exists := cachedStatus.ConfigMap().V1().GetSimple(name); !exists {
		// Create
		cm = &core.ConfigMap{
			ObjectMeta: meta.ObjectMeta{
				Name: name,
			},
			Data: elements,
		}

		owner := r.context.GetAPIObject().AsOwner()

		err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return k8sutil.CreateConfigMap(ctxChild, configMaps, cm, &owner)
		})
		if kerrors.IsAlreadyExists(err) {
			// CM added while we tried it also
			return nil
		} else if err != nil {
			// Failed to create
			return errors.WithStack(err)
		}

		return errors.Reconcile()
	} else {
		// CM Exists, checks checksum - if key is not in the map we return empty string
		if currentSha, expectedSha := util.Optional(cm.Data, constants.ConfigMapChecksumKey, ""), util.Optional(elements, constants.ConfigMapChecksumKey, ""); currentSha != expectedSha || currentSha == "" {
			// We need to do the update
			if _, changed, err := patcher.Patcher[*core.ConfigMap](ctx, cachedStatus.ConfigMapsModInterface().V1(), cm, meta.PatchOptions{},
				patcher.PatchConfigMapData(elements)); err != nil {
				log.Err(err).Debug("Failed to patch GatewayConfig ConfigMap")
				return errors.WithStack(err)
			} else if changed {
				log.Str("configmap", cm.GetName()).Str("before", currentSha).Str("after", expectedSha).Info("Updated GatewayConfig")
			}
		}
	}
	return nil
}

func (r *Resources) renderGatewayConfig(cachedStatus inspectorInterface.Inspector) (gateway.Config, error) {
	deploymentName := r.context.GetAPIObject().GetName()

	log := r.log.Str("section", "gateway-config-render")

	spec := r.context.GetSpec()
	svcServingName := fmt.Sprintf("%s-%s", deploymentName, spec.Mode.Get().ServingGroup().AsRole())

	svc, svcExist := cachedStatus.Service().V1().GetSimple(svcServingName)
	if !svcExist {
		return gateway.Config{}, errors.Errorf("Service %s not found", svcServingName)
	}

	var cfg gateway.Config

	cfg.Options = &gateway.ConfigOptions{
		MergeSlashes: util.NewType(true),
	}

	cfg.IntegrationSidecar = &gateway.ConfigDestinationTarget{
		Host: "127.0.0.1",
		Port: int32(r.context.GetSpec().Integration.GetSidecar().GetListenPort()),
	}

	cfg.DefaultDestination = gateway.ConfigDestination{
		Targets: []gateway.ConfigDestinationTarget{
			{
				Host: svc.Spec.ClusterIP,
				Port: shared.ArangoPort,
			},
		},
		AuthExtension: &gateway.ConfigAuthZExtension{},
		Timeout: &meta.Duration{
			Duration: constants.MaxGatewayTimeout,
		},
	}

	if spec.TLS.IsSecure() {
		// Enabled TLS, add config
		keyPath := filepath.Join(shared.TLSKeyfileVolumeMountDir, constants.SecretTLSKeyfile)
		cfg.DefaultTLS = &gateway.ConfigTLS{
			CertificatePath: keyPath,
			PrivateKeyPath:  keyPath,
		}
		cfg.DefaultDestination.Type = util.NewType(gateway.ConfigDestinationTypeHTTPS)

		// Check SNI
		if sni := spec.TLS.GetSNI().Mapping; len(sni) > 0 {
			for _, volume := range util.SortKeys(sni) {
				servers, ok := sni[volume]
				if !ok {
					continue
				}

				var s gateway.ConfigSNI
				f := path.Join(shared.TLSSNIKeyfileVolumeMountDir, volume, constants.SecretTLSKeyfile)
				s.ConfigTLS = gateway.ConfigTLS{
					CertificatePath: f,
					PrivateKeyPath:  f,
				}
				s.ServerNames = servers
				cfg.SNI = append(cfg.SNI, s)
			}
		}
	}

	// Check ArangoRoutes
	if c, err := cachedStatus.ArangoRoute().V1Alpha1(); err == nil {
		cfg.Destinations = gateway.ConfigDestinations{}
		if err := c.Iterate(func(at *networkingApi.ArangoRoute) error {
			log := log.Str("ArangoRoute", at.GetName())
			if !at.Status.Conditions.IsTrue(networkingApi.ReadyCondition) {
				l := log
				if c, ok := at.Status.Conditions.Get(networkingApi.ReadyCondition); ok {
					l.Str("message", c.Message)
				}
				l.Warn("ArangoRoute is not ready")

				return nil
			}

			if target := at.Status.Target; target != nil {
				var dest gateway.ConfigDestination
				if destinations := target.Destinations; len(destinations) > 0 {
					for _, destination := range destinations {
						var t gateway.ConfigDestinationTarget

						t.Host = destination.Host
						t.Port = destination.Port

						dest.Targets = append(dest.Targets, t)
					}
				}
				if tls := target.TLS; tls != nil {
					dest.Type = util.NewType(gateway.ConfigDestinationTypeHTTPS)
					dest.TLS.Insecure = util.NewType(tls.IsInsecure())
				}
				switch target.Protocol {
				case networkingApi.ArangoRouteDestinationProtocolHTTP1:
					dest.Protocol = util.NewType(gateway.ConfigDestinationProtocolHTTP1)
				case networkingApi.ArangoRouteDestinationProtocolHTTP2:
					dest.Protocol = util.NewType(gateway.ConfigDestinationProtocolHTTP2)
				}
				if opts := target.Options; opts != nil {
					for _, upgrade := range opts.Upgrade {
						dest.UpgradeConfigs = append(dest.UpgradeConfigs, gateway.ConfigDestinationUpgrade{
							Type:    string(upgrade.Type),
							Enabled: util.NewType(util.WithDefault(upgrade.Enabled)),
						})
					}
				}
				dest.Path = util.NewType(target.Path)
				dest.Timeout = target.Timeout.DeepCopy()
				dest.AuthExtension = &gateway.ConfigAuthZExtension{
					AuthZExtension: map[string]string{
						pbImplEnvoyAuthV3.AuthConfigAuthRequiredKey: util.BoolSwitch[string](target.Authentication.Type.Get() == networkingApi.ArangoRouteSpecAuthenticationTypeRequired, pbImplEnvoyAuthV3.AuthConfigKeywordTrue, pbImplEnvoyAuthV3.AuthConfigKeywordFalse),
						pbImplEnvoyAuthV3.AuthConfigAuthPassModeKey: string(target.Authentication.PassMode),
					},
				}
				dest.ResponseHeaders = map[string]string{
					constants.EnvoyRouteHeader: at.GetName(),
				}
				cfg.Destinations[at.Spec.GetRoute().GetPath()] = dest
			}

			return nil

		}, func(at *networkingApi.ArangoRoute) bool {
			return at.Spec.GetDeployment() == deploymentName
		}); err != nil {
			return gateway.Config{}, errors.Wrapf(err, "Unable to iterate over ArangoRoutes")
		}
	}

	return cfg, nil
}
