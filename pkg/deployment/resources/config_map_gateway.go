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

	"google.golang.org/protobuf/encoding/protojson"
	core "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	pbInventoryV1 "github.com/arangodb/kube-arangodb/integrations/inventory/v1/definition"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1beta1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/gateway"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
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

	cfg.Destinations[utilConstants.EnvoyInventoryConfigDestination] = gateway.ConfigDestination{
		Type:  util.NewType(gateway.ConfigDestinationTypeStatic),
		Match: util.NewType(gateway.ConfigMatchPath),
		AuthExtension: &gateway.ConfigAuthZExtension{
			AuthZExtension: map[string]string{
				pbImplEnvoyAuthV3Shared.AuthConfigAuthRequiredKey: pbImplEnvoyAuthV3Shared.AuthConfigKeywordTrue,
				pbImplEnvoyAuthV3Shared.AuthConfigAuthPassModeKey: string(networkingApi.ArangoRouteSpecAuthenticationPassModeRemove),
			},
		},
		Static: &gateway.ConfigDestinationStatic[*pbInventoryV1.Inventory]{
			Code: util.NewType[uint32](200),
			Response: &pbInventoryV1.Inventory{
				Configuration: &pbInventoryV1.InventoryConfiguration{
					Hash: baseGatewayCfgYamlChecksum,
				},
				Arangodb: pbInventoryV1.NewArangoDBConfiguration(r.context.GetSpec(), r.context.GetStatus()),
			},
			Marshaller: ugrpc.Marshal[*pbInventoryV1.Inventory],
			Options: []util.Mod[protojson.MarshalOptions]{
				func(in *protojson.MarshalOptions) {
					in.EmitDefaultValues = true
				},
			},
		},
	}

	cfg.Destinations[utilConstants.EnvoyIdentityDestination] = gateway.ConfigDestination{
		Type:  util.NewType(gateway.ConfigDestinationTypeHTTP),
		Match: util.NewType(gateway.ConfigMatchPath),
		Path:  util.NewType("/_integration/authn/v1/identity"),
		AuthExtension: &gateway.ConfigAuthZExtension{
			AuthZExtension: map[string]string{
				pbImplEnvoyAuthV3Shared.AuthConfigAuthRequiredKey: pbImplEnvoyAuthV3Shared.AuthConfigKeywordFalse,
				pbImplEnvoyAuthV3Shared.AuthConfigAuthPassModeKey: string(networkingApi.ArangoRouteSpecAuthenticationPassModePass),
			},
		},
		Targets: gateway.ConfigDestinationTargets{
			{
				Host: "127.0.0.1",
				Port: int32(r.context.GetSpec().Integration.GetSidecar().GetHTTPListenPort()),
			},
		},
	}

	cfg.Destinations[utilConstants.EnvoyLoginDestination] = gateway.ConfigDestination{
		Type:  util.NewType(gateway.ConfigDestinationTypeHTTP),
		Match: util.NewType(gateway.ConfigMatchPath),
		Path:  util.NewType("/_integration/authn/v1/login"),
		AuthExtension: &gateway.ConfigAuthZExtension{
			AuthZExtension: map[string]string{
				pbImplEnvoyAuthV3Shared.AuthConfigAuthRequiredKey: pbImplEnvoyAuthV3Shared.AuthConfigKeywordFalse,
				pbImplEnvoyAuthV3Shared.AuthConfigAuthPassModeKey: string(networkingApi.ArangoRouteSpecAuthenticationPassModePass),
			},
		},
		Targets: gateway.ConfigDestinationTargets{
			{
				Host: "127.0.0.1",
				Port: int32(r.context.GetSpec().Integration.GetSidecar().GetHTTPListenPort()),
			},
		},
	}

	cfg.Destinations[utilConstants.EnvoyLogoutDestination] = gateway.ConfigDestination{
		Type:  util.NewType(gateway.ConfigDestinationTypeHTTP),
		Match: util.NewType(gateway.ConfigMatchPath),
		Path:  util.NewType("/_integration/authn/v1/logout"),
		AuthExtension: &gateway.ConfigAuthZExtension{
			AuthZExtension: map[string]string{
				pbImplEnvoyAuthV3Shared.AuthConfigAuthRequiredKey: pbImplEnvoyAuthV3Shared.AuthConfigKeywordFalse,
				pbImplEnvoyAuthV3Shared.AuthConfigAuthPassModeKey: string(networkingApi.ArangoRouteSpecAuthenticationPassModePass),
			},
		},
		Targets: gateway.ConfigDestinationTargets{
			{
				Host: "127.0.0.1",
				Port: int32(r.context.GetSpec().Integration.GetSidecar().GetHTTPListenPort()),
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
		utilConstants.GatewayConfigFileName: string(gatewayCfgYaml),
	}); err != nil {
		return err
	}

	if err := r.ensureGatewayConfigMap(ctx, cachedStatus, configMaps, GetGatewayConfigMapName(r.context.GetAPIObject().GetName(), "cds"), map[string]string{
		utilConstants.GatewayConfigFileName: string(gatewayCfgCDSYaml),
	}); err != nil {
		return err
	}

	if err := r.ensureGatewayConfigMap(ctx, cachedStatus, configMaps, GetGatewayConfigMapName(r.context.GetAPIObject().GetName(), "lds"), map[string]string{
		utilConstants.GatewayConfigFileName: string(gatewayCfgLDSYaml),
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
		if currentSha, expectedSha := util.Optional(cm.Data, utilConstants.ConfigMapChecksumKey, ""), util.Optional(elements, utilConstants.ConfigMapChecksumKey, ""); currentSha != expectedSha || currentSha == "" {
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
			Duration: utilConstants.MaxEnvoyUpstreamTimeout,
		},
	}

	if spec.Gateway.IsDefaultTargetAuthenticationEnabled() {
		cfg.DefaultDestination.AuthExtension.AuthZExtension = map[string]string{
			pbImplEnvoyAuthV3Shared.AuthConfigAuthRequiredKey: pbImplEnvoyAuthV3Shared.AuthConfigKeywordFalse,
			pbImplEnvoyAuthV3Shared.AuthConfigAuthPassModeKey: string(networkingApi.ArangoRouteSpecAuthenticationPassModePass),
		}
	}

	if spec.TLS.IsSecure() {
		// Enabled TLS, add config
		keyPath := filepath.Join(shared.TLSKeyfileVolumeMountDir, utilConstants.SecretTLSKeyfile)
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
				f := path.Join(shared.TLSSNIKeyfileVolumeMountDir, volume, utilConstants.SecretTLSKeyfile)
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
	cfg.Destinations = gateway.ConfigDestinations{}
	if c, err := cachedStatus.ArangoRoute().V1Beta1(); err == nil {
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
				if target.Route.Path == "" {
					log.Warn("ArangoRoute Route Path not defined")
					return nil
				}
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
						pbImplEnvoyAuthV3Shared.AuthConfigAuthRequiredKey: util.BoolSwitch[string](target.Authentication.Type.Get() == networkingApi.ArangoRouteSpecAuthenticationTypeRequired, pbImplEnvoyAuthV3Shared.AuthConfigKeywordTrue, pbImplEnvoyAuthV3Shared.AuthConfigKeywordFalse),
						pbImplEnvoyAuthV3Shared.AuthConfigAuthPassModeKey: string(target.Authentication.PassMode),
					},
				}
				dest.ResponseHeaders = map[string]string{
					utilConstants.EnvoyRouteHeader: at.GetName(),
				}
				cfg.Destinations[target.Route.Path] = dest
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
