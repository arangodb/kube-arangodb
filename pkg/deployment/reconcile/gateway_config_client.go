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

package reconcile

import (
	"context"
	"crypto/tls"
	"fmt"

	pbEnvoyConfigV1 "github.com/arangodb/kube-arangodb/integrations/envoy/config/v1/definition"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// newGatewayConfigClient dials the EnvoyConfigV1 gRPC endpoint on the given gateway member's sidecar
// (the external network listener), authenticated with the provided superuser token. It is shared by the
// push action and the gateway config condition plan builder.
func newGatewayConfigClient(ctx context.Context, apiObject k8sutil.APIObject, spec api.DeploymentSpec, memberID, token string) (pbEnvoyConfigV1.EnvoyConfigV1Client, func(), error) {
	addr := fmt.Sprintf("%s:%d", k8sutil.CreatePodDNSName(apiObject, api.ServerGroupGateways.AsRole(), memberID), shared.InternalSidecarContainerPortGRPC)

	var tlsConfig *tls.Config
	if spec.TLS.IsSecure() {
		// The sidecar presents the deployment TLS certificate; the connection is in-cluster to the Pod.
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	}

	conn, err := ugrpc.NewOptionalTLSGRPCConn(ctx, addr, tlsConfig, ugrpc.TokenAuthInterceptors(token)...)
	if err != nil {
		return nil, nil, err
	}

	return pbEnvoyConfigV1.NewEnvoyConfigV1Client(conn), func() { _ = conn.Close() }, nil
}
