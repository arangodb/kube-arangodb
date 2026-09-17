//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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
	"path"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/gateway"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

func (r *Resources) ensureMemberConfigGatewayConfig(ctx context.Context, cachedStatus inspectorInterface.Inspector, member api.DeploymentStatusMemberElement) (map[string]string, error) {
	if member.Group != api.ServerGroupGateways {
		return nil, nil
	}

	if r.context.GetSpec().Gateway.IsDynamicModePush() {
		// Push mode: Envoy subscribes to CDS/LDS over ADS served by the integration sidecar, instead of
		// watching the mounted ConfigMap directory.
		data, _, _, err := gateway.NodeADSConfig("arangodb", member.Member.ID, utilConstants.EnvoyGatewayADSCluster,
			gateway.ConfigDestinationTargetUnix{
				Path: path.Join(utilConstants.SidecarUnixSocketMountPath, utilConstants.SidecarUnixSocketMountFile),
			})
		if err != nil {
			return nil, err
		}

		return map[string]string{
			utilConstants.GatewayConfigFileName: string(data),
		}, nil
	}

	data, _, _, err := gateway.NodeDynamicConfig("arangodb", member.Member.ID, &gateway.DynamicConfig{
		Path: utilConstants.GatewayCDSVolumeMountDir,
		File: utilConstants.GatewayConfigFileName,
	}, &gateway.DynamicConfig{
		Path: utilConstants.GatewayLDSVolumeMountDir,
		File: utilConstants.GatewayConfigFileName,
	})
	if err != nil {
		return nil, err
	}

	return map[string]string{
		utilConstants.GatewayConfigFileName: string(data),
	}, nil
}
