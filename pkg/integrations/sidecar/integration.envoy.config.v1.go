//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package sidecar

import (
	"fmt"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
)

const (
	gatewayPushCDSVolumeName = "gateway-push-cds"
	gatewayPushLDSVolumeName = "gateway-push-lds"
)

// IntegrationEnvoyConfigV1 enables the EnvoyConfigV1 (ADS/xDS) integration on the gateway sidecar. When
// enabled, the sidecar serves the gateway dynamic config to Envoy over ADS (push mode) on the internal
// (unix) listener, and exposes the control channel on the external network listener so the operator can
// push config to it with a superuser JWT.
type IntegrationEnvoyConfigV1 struct {
	Core           *Core
	DeploymentName string
	Spec           api.DeploymentSpec

	// CDSConfigMapName and LDSConfigMapName are the gateway CDS/LDS ConfigMaps mounted read-only into the
	// sidecar so it can seed its initial ADS snapshot from the last-known-good local config on startup.
	CDSConfigMapName string
	LDSConfigMapName string
}

func (i IntegrationEnvoyConfigV1) Name() []string {
	return []string{"ENVOY", "CONFIG", "V1"}
}

func (i IntegrationEnvoyConfigV1) Validate() error {
	return nil
}

func (i IntegrationEnvoyConfigV1) Envs() ([]core.EnvVar, error) {
	t := true

	// Envoy reaches the ADS server over the internal (unix) listener; the operator reaches the control
	// channel over the external network listener. Register this service on both, and expose the external
	// gRPC listener on the sidecar's gRPC container port.
	c := &Core{Internal: &t, External: &t}

	return c.Envs(i,
		core.EnvVar{
			Name:  "INTEGRATION_ENVOY_CONFIG_V1",
			Value: "true",
		},
		core.EnvVar{
			Name:  "SERVICES_EXTERNAL_ENABLED",
			Value: "true",
		},
		core.EnvVar{
			Name:  "SERVICES_EXTERNAL_ADDRESS",
			Value: fmt.Sprintf("0.0.0.0:%d", shared.InternalSidecarContainerPortGRPC),
		},
	), nil
}

func (i IntegrationEnvoyConfigV1) GlobalEnvs() ([]core.EnvVar, error) {
	return nil, nil
}

func (i IntegrationEnvoyConfigV1) Volumes() ([]core.Volume, []core.VolumeMount, error) {
	// Mount the gateway CDS/LDS ConfigMaps read-only so the sidecar can seed its initial ADS snapshot from
	// the last-known-good local config on startup (matching the paths the sidecar reads).
	return []core.Volume{
			{
				Name: gatewayPushCDSVolumeName,
				VolumeSource: core.VolumeSource{
					ConfigMap: &core.ConfigMapVolumeSource{
						LocalObjectReference: core.LocalObjectReference{Name: i.CDSConfigMapName},
					},
				},
			},
			{
				Name: gatewayPushLDSVolumeName,
				VolumeSource: core.VolumeSource{
					ConfigMap: &core.ConfigMapVolumeSource{
						LocalObjectReference: core.LocalObjectReference{Name: i.LDSConfigMapName},
					},
				},
			},
		}, []core.VolumeMount{
			{
				Name:      gatewayPushCDSVolumeName,
				ReadOnly:  true,
				MountPath: utilConstants.GatewayCDSVolumeMountDir,
			},
			{
				Name:      gatewayPushLDSVolumeName,
				ReadOnly:  true,
				MountPath: utilConstants.GatewayLDSVolumeMountDir,
			},
		}, nil
}
