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

package v1

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbImplAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1"
	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func New(ctx context.Context, cfg Configuration) (svc.HandlerInitService, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	plugin, err := cfg.Plugin(ctx)
	if err != nil {
		return nil, err
	}

	return newInternal(cfg, plugin), nil
}

func newInternal(cfg Configuration, plugin pbImplAuthorizationV1Shared.Plugin) *implementation {
	obj := &implementation{
		cfg:    cfg,
		plugin: plugin,
	}

	return obj
}

var _ pbAuthorizationV1.AuthorizationV1Server = &implementation{}
var _ svc.Handler = &implementation{}
var _ svc.HandlerInitService = &implementation{}

type implementation struct {
	pbAuthorizationV1.UnimplementedAuthorizationV1Server

	cfg Configuration

	plugin pbImplAuthorizationV1Shared.Plugin

	auth cache.Object[pbAuthenticationV1.AuthenticationV1Client]
}

func (i *implementation) InitService(svc svc.Service) error {
	i.auth = pbImplAuthenticationV1.ServiceClient(svc)

	return nil
}

func (i *implementation) Name() string {
	return pbAuthorizationV1.Name
}

func (i *implementation) Health(ctx context.Context) svc.HealthState {
	if err := i.plugin.Ready(ctx); err != nil {
		return svc.Degraded
	}
	return svc.Healthy
}

func (i *implementation) Register(registrar *grpc.Server) {
	pbAuthorizationV1.RegisterAuthorizationV1Server(registrar, i)
}

func (i *implementation) Gateway(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return pbAuthorizationV1.RegisterAuthorizationV1Handler(ctx, mux, conn)
}

func (i *implementation) Evaluate(ctx context.Context, request *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
	if err := i.Health(ctx).Require(); err != nil {
		return nil, err
	}

	return i.plugin.Evaluate(ctx, request)
}

func (i *implementation) EvaluateMany(ctx context.Context, request *pbAuthorizationV1.AuthorizationV1PermissionManyRequest) (*pbAuthorizationV1.AuthorizationV1PermissionManyResponse, error) {
	if err := i.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if len(request.Items) == 0 {
		return &pbAuthorizationV1.AuthorizationV1PermissionManyResponse{
			Message: "Missing permission evaluation items",
			Effect:  sidecarSvcAuthzTypes.Effect_Deny,
		}, nil
	}

	var r = make([]*pbAuthorizationV1.AuthorizationV1PermissionResponse, len(request.Items))

	for id, v := range request.Items {
		resp, err := i.Evaluate(ctx, &pbAuthorizationV1.AuthorizationV1PermissionRequest{
			User:     request.User,
			Roles:    request.GetRoles(),
			Action:   v.GetAction(),
			Resource: v.GetResource(),
			Context:  v.GetContext(),
		})
		if err != nil {
			return nil, err
		}

		r[id] = resp
	}

	for _, v := range r {
		if v.GetEffect() == sidecarSvcAuthzTypes.Effect_Deny {
			return &pbAuthorizationV1.AuthorizationV1PermissionManyResponse{
				Message: "One of the requests has been denied",
				Effect:  sidecarSvcAuthzTypes.Effect_Deny,
				Items:   r,
			}, nil
		}
	}

	return &pbAuthorizationV1.AuthorizationV1PermissionManyResponse{
		Message: "Access Granted",
		Effect:  sidecarSvcAuthzTypes.Effect_Allow,
		Items:   r,
	}, nil
}

func (i *implementation) EvaluateToken(ctx context.Context, request *pbAuthorizationV1.AuthorizationV1PermissionTokenRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
	if err := i.Health(ctx).Require(); err != nil {
		return nil, err
	}

	auth, err := i.auth.Get(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "Authentication V1 Plugin not enabled: %v", err)
	}

	resp, err := auth.Validate(ctx, &pbAuthenticationV1.ValidateRequest{Token: request.GetToken()})
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "Unable to validate the token: %v", err)
	}

	if !resp.GetIsValid() {
		return nil, status.Errorf(codes.FailedPrecondition, "JWT Token is invalid")
	}

	return i.Evaluate(ctx, &pbAuthorizationV1.AuthorizationV1PermissionRequest{
		User:     util.BoolSwitch(resp.GetDetails() == nil, nil, resp.GetDetails().User),
		Roles:    resp.GetDetails().GetRoles(),
		Action:   request.GetAction(),
		Resource: request.GetResource(),
		Context:  request.GetContext(),
	})
}

func (i *implementation) EvaluateTokenMany(ctx context.Context, request *pbAuthorizationV1.AuthorizationV1PermissionTokenManyRequest) (*pbAuthorizationV1.AuthorizationV1PermissionManyResponse, error) {
	if err := i.Health(ctx).Require(); err != nil {
		return nil, err
	}

	auth, err := i.auth.Get(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "Authentication V1 Plugin not enabled: %v", err)
	}

	resp, err := auth.Validate(ctx, &pbAuthenticationV1.ValidateRequest{Token: request.GetToken()})
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "Unable to validate the token: %v", err)
	}

	if !resp.GetIsValid() {
		return nil, status.Errorf(codes.FailedPrecondition, "JWT Token is invalid")
	}

	return i.EvaluateMany(ctx, &pbAuthorizationV1.AuthorizationV1PermissionManyRequest{
		User:  util.BoolSwitch(resp.GetDetails() == nil, nil, resp.GetDetails().User),
		Roles: resp.GetDetails().GetRoles(),
		Items: request.GetItems(),
	})
}
