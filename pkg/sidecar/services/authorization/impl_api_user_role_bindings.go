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
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

// userRoleBindingKey returns the storage key for a user-role binding.
func userRoleBindingKey(user, role string) string {
	return fmt.Sprintf("%s:%s", user, role)
}

// userRoleBindingPrefix returns the prefix for all bindings of a user.
func userRoleBindingPrefix(user string) string {
	return user + ":"
}

func (a *implementation) APIListUserRoleBindings(ctx context.Context, request *sidecarSvcAuthzDefinition.AuthorizationAPIUserRequest) (*sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingListResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.Plugin(), "rbac:ListUserRoleBinding", request.GetUser()); err != nil {
		return nil, err
	}

	if request.GetUser() == "" {
		return nil, status.Error(codes.InvalidArgument, "User cannot be empty")
	}

	prefix := userRoleBindingPrefix(request.GetUser())
	allItems := a.userRoleBindings.Items()

	var bindings []*sidecarSvcAuthzTypes.UserRoleBinding
	for _, name := range allItems {
		if len(name) > len(prefix) && name[:len(prefix)] == prefix {
			if binding, _, ok := a.userRoleBindings.Item(name); ok && binding != nil {
				bindings = append(bindings, binding)
			}
		}
	}

	return &sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingListResponse{
		Bindings: bindings,
	}, nil
}

func (a *implementation) APIAssignUserRole(ctx context.Context, request *sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingRequest) (*sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.Plugin(), "rbac:AssignUserRole", request.GetUser()); err != nil {
		return nil, err
	}

	if request.GetUser() == "" {
		return nil, status.Error(codes.InvalidArgument, "User cannot be empty")
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

	key := userRoleBindingKey(request.GetUser(), request.GetRole())

	if _, index, err := a.userRoleBindings.Create(ctx, key, binding); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	} else {
		identity := authenticator.GetIdentity(ctx)
		logger.Str("targetUser", request.GetUser()).Str("role", request.GetRole()).Str("user", identity.GetUser()).Info("User role assigned")

		return &sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingResponse{
			User:  request.GetUser(),
			Role:  request.GetRole(),
			Scope: request.GetScope(),
			Index: index,
		}, nil
	}
}

func (a *implementation) APIRemoveUserRole(ctx context.Context, request *sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleRequest) (*sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.Plugin(), "rbac:RemoveUserRole", request.GetUser()); err != nil {
		return nil, err
	}

	if request.GetUser() == "" {
		return nil, status.Error(codes.InvalidArgument, "User cannot be empty")
	}

	if request.GetRole() == "" {
		return nil, status.Error(codes.InvalidArgument, "Role cannot be empty")
	}

	key := userRoleBindingKey(request.GetUser(), request.GetRole())

	if index, err := a.userRoleBindings.Delete(ctx, key); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	} else {
		identity := authenticator.GetIdentity(ctx)
		logger.Str("targetUser", request.GetUser()).Str("role", request.GetRole()).Str("user", identity.GetUser()).Info("User role removed")

		return &sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingResponse{
			User:  request.GetUser(),
			Role:  request.GetRole(),
			Index: index,
		}, nil
	}
}

func (a *implementation) APIReplaceUserRoleScope(ctx context.Context, request *sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingRequest) (*sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.Plugin(), "rbac:ReplaceUserRoleScope", request.GetUser()); err != nil {
		return nil, err
	}

	if request.GetUser() == "" {
		return nil, status.Error(codes.InvalidArgument, "User cannot be empty")
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

	key := userRoleBindingKey(request.GetUser(), request.GetRole())

	if _, index, err := a.userRoleBindings.Update(ctx, key, binding); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	} else {
		identity := authenticator.GetIdentity(ctx)
		logger.Str("targetUser", request.GetUser()).Str("role", request.GetRole()).Str("user", identity.GetUser()).Info("User role scope replaced")

		return &sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingResponse{
			User:  request.GetUser(),
			Role:  request.GetRole(),
			Scope: request.GetScope(),
			Index: index,
		}, nil
	}
}
