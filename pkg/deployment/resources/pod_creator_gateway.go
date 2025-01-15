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
	"fmt"
	"strings"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func GetGatewayConfigMapName(name string, parts ...string) string {
	if len(parts) == 0 {
		return fmt.Sprintf("%s-gateway", name)
	}

	return fmt.Sprintf("%s-gateway-%s", name, strings.Join(parts, "-"))
}

func createGatewayVolumes(input pod.Input) pod.Volumes {
	volumes := pod.NewVolumes()

	volumes.AddVolume(k8sutil.CreateVolumeWithConfigMap(constants.GatewayVolumeName, GetGatewayConfigMapName(input.ApiObject.GetName())))
	volumes.AddVolume(k8sutil.CreateVolumeWithConfigMap(constants.GatewayCDSVolumeName, GetGatewayConfigMapName(input.ApiObject.GetName(), "cds")))
	volumes.AddVolume(k8sutil.CreateVolumeWithConfigMap(constants.GatewayLDSVolumeName, GetGatewayConfigMapName(input.ApiObject.GetName(), "lds")))
	volumes.AddVolume(k8sutil.CreateVolumeWithConfigMap(constants.MemberConfigVolumeName, input.ArangoMember.GetName()))
	volumes.AddVolumeMount(core.VolumeMount{
		Name:      constants.GatewayVolumeName,
		MountPath: constants.GatewayVolumeMountDir,
		ReadOnly:  true,
	})
	volumes.AddVolumeMount(core.VolumeMount{
		Name:      constants.GatewayCDSVolumeName,
		MountPath: constants.GatewayCDSVolumeMountDir,
		ReadOnly:  true,
	})
	volumes.AddVolumeMount(core.VolumeMount{
		Name:      constants.GatewayLDSVolumeName,
		MountPath: constants.GatewayLDSVolumeMountDir,
		ReadOnly:  true,
	})
	volumes.AddVolumeMount(core.VolumeMount{
		Name:      constants.MemberConfigVolumeName,
		MountPath: constants.MemberConfigVolumeMountDir,
		ReadOnly:  true,
	})

	// TLS
	volumes.Append(pod.TLS(), input)

	// SNI
	volumes.Append(pod.SNIGateway(), input)

	return volumes
}
