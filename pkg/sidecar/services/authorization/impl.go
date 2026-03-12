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

package authorization

import (
	"context"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	pbImplAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1"
	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
	sidecarSvcAuthzDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	"github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/pool"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/db"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

type Authorizer interface {
	Plugin() pbImplAuthorizationV1Shared.Plugin

	svc.Handler
}

func NewAuthorizer(client db.Database, authType pbImplAuthorizationV1.ConfigurationType) Authorizer {
	return &implementation{
		policies: pool.NewPooler[*sidecarSvcAuthzTypes.Policy](client.
			CreateCollection("_policies", db.SourceCollectionProps("_users")).
			WithUniqueIndex("policies_unique_sequence_index", "sequence").
			WithTTLIndex("policies_deleted_index", 30*24*time.Hour, "deleted").
			Get(), pool.DefaultPoolerTimeout),
		roles: pool.NewPooler[*sidecarSvcAuthzTypes.Role](client.
			CreateCollection("_roles", db.SourceCollectionProps("_users")).
			WithTTLIndex("roles_deleted_index", 30*24*time.Hour, "deleted").
			Get(), pool.DefaultPoolerTimeout),
		authType: authType,
	}
}

var _ sidecarSvcAuthzDefinition.AuthorizationPoolServiceServer = &implementation{}
var _ sidecarSvcAuthzDefinition.AuthorizationAPIServer = &implementation{}

type implementation struct {
	sidecarSvcAuthzDefinition.UnimplementedAuthorizationPoolServiceServer
	sidecarSvcAuthzDefinition.UnimplementedAuthorizationAPIServer

	policies pool.Pooler[*sidecarSvcAuthzTypes.Policy]
	roles    pool.Pooler[*sidecarSvcAuthzTypes.Role]

	authType pbImplAuthorizationV1.ConfigurationType
}

func (a *implementation) Name() string {
	return "authorization"
}

func (a *implementation) Plugin() pbImplAuthorizationV1Shared.Plugin {
	switch a.authType {
	case pbImplAuthorizationV1.ConfigurationTypeCentral:
		return a
	case pbImplAuthorizationV1.ConfigurationTypeCentralPermissive:
		return pbImplAuthorizationV1Shared.Permissive(a, logger)
	case pbImplAuthorizationV1.ConfigurationTypeAlways:
		return pbImplAuthorizationV1Shared.Permissive(pbImplAuthorizationV1Shared.NewAlwaysPlugin(), logger)
	default:
		return pbImplAuthorizationV1Shared.NewNeverPlugin()
	}
}

func (a *implementation) Health(ctx context.Context) svc.HealthState {
	if !a.roles.Ready() || !a.policies.Ready() {
		return svc.Unhealthy
	}
	return svc.Healthy
}

func (a *implementation) Register(registrar *grpc.Server) {
	sidecarSvcAuthzDefinition.RegisterAuthorizationPoolServiceServer(registrar, a)
	sidecarSvcAuthzDefinition.RegisterAuthorizationAPIServer(registrar, a)
}

func (a *implementation) Gateway(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return sidecarSvcAuthzDefinition.RegisterAuthorizationAPIHandler(ctx, mux, conn)
}

func (a *implementation) Refresh(ctx context.Context) error {
	if err := a.policies.Refresh(ctx); err != nil {
		return err
	}
	if err := a.roles.Refresh(ctx); err != nil {
		return err
	}

	return nil
}
