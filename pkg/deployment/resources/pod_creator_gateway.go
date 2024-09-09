//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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
	"fmt"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	ArangoGatewayExecutor        = "/usr/local/bin/envoy"
	GatewayVolumeMountDir        = "/etc/gateway/"
	GatewayVolumeName            = "gateway"
	GatewayConfigFileName        = "gateway.yaml"
	GatewayDynamicConfigFileName = "gateway.dynamic.yaml"
	GatewayCDSConfigFileName     = "gateway.dynamic.cds.yaml"
	GatewayLDSConfigFileName     = "gateway.dynamic.lds.yaml"
	GatewayConfigChecksumENV     = "GATEWAY_CONFIG_CHECKSUM"
)

func GetGatewayConfigMapName(name string) string {
	return fmt.Sprintf("%s-gateway", name)
}

func createGatewayVolumes(input pod.Input) pod.Volumes {
	volumes := pod.NewVolumes()

	volumes.AddVolume(k8sutil.CreateVolumeWithConfigMap(GatewayVolumeName, GetGatewayConfigMapName(input.ApiObject.GetName())))
	volumes.AddVolume(k8sutil.CreateVolumeWithConfigMap(MemberConfigVolumeName, input.ArangoMember.GetName()))
	volumes.AddVolumeMount(core.VolumeMount{
		Name:      GatewayVolumeName,
		MountPath: GatewayVolumeMountDir,
		ReadOnly:  true,
	})
	volumes.AddVolumeMount(core.VolumeMount{
		Name:      MemberConfigVolumeName,
		MountPath: MemberConfigVolumeMountDir,
		ReadOnly:  true,
	})

	// TLS
	volumes.Append(pod.TLS(), input)

	// SNI
	volumes.Append(pod.SNIGateway(), input)

	return volumes
}
