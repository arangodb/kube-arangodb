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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	sidecarSvcAuthzDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	"github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/pool"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

func (a *implementation) APIListRole(ctx context.Context, request *pbSharedV1.OffsetRequest) (*sidecarSvcAuthzDefinition.AuthorizationAPIListResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.auth, "rbac:ListRole", ""); err != nil {
		return nil, err
	}

	page, items := pbSharedV1.Paginate(request, a.roles.Items())

	return &sidecarSvcAuthzDefinition.AuthorizationAPIListResponse{
		Names: items,
		Page:  page,
	}, nil
}

func (a *implementation) APIGetRole(ctx context.Context, request *sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest) (*sidecarSvcAuthzDefinition.AuthorizationAPIRoleResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.auth, "rbac:GetRole", request.GetName()); err != nil {
		return nil, err
	}

	role, index, ok := a.roles.Item(request.GetName())
	if !ok {
		return nil, status.Error(codes.NotFound, "Role not found")
	}

	return &sidecarSvcAuthzDefinition.AuthorizationAPIRoleResponse{
		Name:  request.GetName(),
		Index: index,
		Item:  role,
	}, nil
}

func (a *implementation) APIDeleteRole(ctx context.Context, request *sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest) (*sidecarSvcAuthzDefinition.AuthorizationAPIRoleResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.auth, "rbac:DeleteRole", request.GetName()); err != nil {
		return nil, err
	}

	offset, err := a.roles.Delete(ctx, request.GetName())

	if err != nil {
		if pool.IsPoolNotFound(err) {
			return nil, status.Error(codes.NotFound, "Role not found")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &sidecarSvcAuthzDefinition.AuthorizationAPIRoleResponse{
		Name:  request.GetName(),
		Index: offset,
	}, nil
}

func (a *implementation) APICreateRole(ctx context.Context, request *sidecarSvcAuthzDefinition.AuthorizationAPIRoleRequest) (*sidecarSvcAuthzDefinition.AuthorizationAPIRoleResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.auth, "rbac:CreateRole", request.GetName()); err != nil {
		return nil, err
	}

	if item := request.GetName(); item == "" {
		return nil, status.Error(codes.InvalidArgument, "Name cannot be empty")
	}

	if item := request.GetItem(); item == nil {
		return nil, status.Error(codes.InvalidArgument, "Item cannot be empty")
	}

	res, offset, err := a.roles.Create(ctx, request.GetName(), request.GetItem())

	if err != nil {
		if pool.IsPoolAlreadyExistsError(err) {
			return nil, status.Error(codes.AlreadyExists, "Role already exists")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &sidecarSvcAuthzDefinition.AuthorizationAPIRoleResponse{
		Name:  request.GetName(),
		Index: offset,
		Item:  res,
	}, nil
}

func (a *implementation) APIUpdateRole(ctx context.Context, request *sidecarSvcAuthzDefinition.AuthorizationAPIRoleRequest) (*sidecarSvcAuthzDefinition.AuthorizationAPIRoleResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.auth, "rbac:UpdateRole", request.GetName()); err != nil {
		return nil, err
	}

	if item := request.GetName(); item == "" {
		return nil, status.Error(codes.InvalidArgument, "Name cannot be empty")
	}

	if item := request.GetItem(); item == nil {
		return nil, status.Error(codes.InvalidArgument, "Item cannot be empty")
	}

	res, offset, err := a.roles.Update(ctx, request.GetName(), request.GetItem())

	if err != nil {
		if pool.IsPoolNotFound(err) {
			return nil, status.Error(codes.NotFound, "Role not found")
		}

		if pool.IsPoolNoChangeError(err) {
			return nil, status.Error(codes.AlreadyExists, "Role exists")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &sidecarSvcAuthzDefinition.AuthorizationAPIRoleResponse{
		Name:  request.GetName(),
		Index: offset,
		Item:  res,
	}, nil
}
