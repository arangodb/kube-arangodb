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
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sidecarSvcAuthzDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	"github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/pool"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

// groupRoleBindingKey returns the storage key for a group-role binding.
func groupRoleBindingKey(group, role string) string {
	return fmt.Sprintf("%s:%s", group, role)
}

// groupRoleBindingPrefix returns the prefix for all bindings of a group.
func groupRoleBindingPrefix(group string) string {
	return group + ":"
}

func (a *implementation) APIListGroupRoleBindings(ctx context.Context, request *sidecarSvcAuthzDefinition.AuthorizationAPIGroupRequest) (*sidecarSvcAuthzDefinition.AuthorizationAPIGroupRoleBindingListResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.Plugin(), "rbac:ListGroupRoleBinding", request.GetGroup()); err != nil {
		return nil, err
	}

	if request.GetGroup() == "" {
		return nil, status.Error(codes.InvalidArgument, "Group cannot be empty")
	}

	prefix := groupRoleBindingPrefix(request.GetGroup())
	allItems := a.groupRoleBindings.Items()

	var bindings []*sidecarSvcAuthzTypes.UserRoleBinding
	for _, name := range allItems {
		if len(name) > len(prefix) && name[:len(prefix)] == prefix {
			if binding, _, ok := a.groupRoleBindings.Item(name); ok && binding != nil {
				bindings = append(bindings, binding)
			}
		}
	}

	return &sidecarSvcAuthzDefinition.AuthorizationAPIGroupRoleBindingListResponse{
		Bindings: bindings,
	}, nil
}

func (a *implementation) APIAssignGroupRole(ctx context.Context, request *sidecarSvcAuthzDefinition.AuthorizationAPIGroupRoleBindingRequest) (*sidecarSvcAuthzDefinition.AuthorizationAPIGroupRoleBindingResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.Plugin(), "rbac:AssignGroupRole", request.GetGroup()); err != nil {
		return nil, err
	}

	if request.GetGroup() == "" {
		return nil, status.Error(codes.InvalidArgument, "Group cannot be empty")
	}

	if request.GetRole() == "" {
		return nil, status.Error(codes.InvalidArgument, "Role cannot be empty")
	}

	if request.GetScope() == nil {
		return nil, status.Error(codes.InvalidArgument, "Scope cannot be empty")
	}

	binding := &sidecarSvcAuthzTypes.UserRoleBinding{
		Role:  request.GetRole(),
		Scope: request.GetScope(),
	}

	if err := binding.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	key := groupRoleBindingKey(request.GetGroup(), request.GetRole())

	if _, index, err := a.groupRoleBindings.Create(ctx, key, binding); err != nil {
		if pool.IsPoolAlreadyExistsError(err) {
			return nil, status.Error(codes.AlreadyExists, "Group role binding already exists")
		}

		return nil, status.Error(codes.Internal, err.Error())
	} else {
		identity := authenticator.GetIdentity(ctx)
		logger.Str("targetGroup", request.GetGroup()).Str("role", request.GetRole()).Str("user", identity.GetUser()).Info("Group role assigned")

		return &sidecarSvcAuthzDefinition.AuthorizationAPIGroupRoleBindingResponse{
			Group: request.GetGroup(),
			Role:  request.GetRole(),
			Scope: request.GetScope(),
			Index: index,
		}, nil
	}
}

func (a *implementation) APIRemoveGroupRole(ctx context.Context, request *sidecarSvcAuthzDefinition.AuthorizationAPIGroupRoleRequest) (*sidecarSvcAuthzDefinition.AuthorizationAPIGroupRoleBindingResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.Plugin(), "rbac:RemoveGroupRole", request.GetGroup()); err != nil {
		return nil, err
	}

	if request.GetGroup() == "" {
		return nil, status.Error(codes.InvalidArgument, "Group cannot be empty")
	}

	if request.GetRole() == "" {
		return nil, status.Error(codes.InvalidArgument, "Role cannot be empty")
	}

	key := groupRoleBindingKey(request.GetGroup(), request.GetRole())

	if index, err := a.groupRoleBindings.Delete(ctx, key); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	} else {
		identity := authenticator.GetIdentity(ctx)
		logger.Str("targetGroup", request.GetGroup()).Str("role", request.GetRole()).Str("user", identity.GetUser()).Info("Group role removed")

		return &sidecarSvcAuthzDefinition.AuthorizationAPIGroupRoleBindingResponse{
			Group: request.GetGroup(),
			Role:  request.GetRole(),
			Index: index,
		}, nil
	}
}

func (a *implementation) APIReplaceGroupRoleScope(ctx context.Context, request *sidecarSvcAuthzDefinition.AuthorizationAPIGroupRoleBindingRequest) (*sidecarSvcAuthzDefinition.AuthorizationAPIGroupRoleBindingResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.Plugin(), "rbac:ReplaceGroupRoleScope", request.GetGroup()); err != nil {
		return nil, err
	}

	if request.GetGroup() == "" {
		return nil, status.Error(codes.InvalidArgument, "Group cannot be empty")
	}

	if request.GetRole() == "" {
		return nil, status.Error(codes.InvalidArgument, "Role cannot be empty")
	}

	if request.GetScope() == nil {
		return nil, status.Error(codes.InvalidArgument, "Scope cannot be empty")
	}

	binding := &sidecarSvcAuthzTypes.UserRoleBinding{
		Role:  request.GetRole(),
		Scope: request.GetScope(),
	}

	if err := binding.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	key := groupRoleBindingKey(request.GetGroup(), request.GetRole())

	if _, index, err := a.groupRoleBindings.Update(ctx, key, binding); err != nil {
		if pool.IsPoolNotFound(err) {
			return nil, status.Error(codes.NotFound, "Group role binding not found")
		}

		return nil, status.Error(codes.Internal, err.Error())
	} else {
		identity := authenticator.GetIdentity(ctx)
		logger.Str("targetGroup", request.GetGroup()).Str("role", request.GetRole()).Str("user", identity.GetUser()).Info("Group role scope replaced")

		return &sidecarSvcAuthzDefinition.AuthorizationAPIGroupRoleBindingResponse{
			Group: request.GetGroup(),
			Role:  request.GetRole(),
			Scope: request.GetScope(),
			Index: index,
		}, nil
	}
}
