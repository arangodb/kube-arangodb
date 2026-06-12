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

package token

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	permissionApi "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	sidecarSvcAuthzDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
)

// HandleUserGroupBindings attaches/detaches groups to the user via the sidecar
// UserRoleBinding API. Groups come from resolved spec.roles (external roles)
// and the managed group (from spec.policy). Each group is bound to the user
// with its scope from the token spec.
func (h *handler) HandleUserGroupBindings(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionToken, st *permissionApi.ArangoPermissionTokenStatus, depl *api.ArangoDeployment, conn sidecarSvcAuthzDefinition.AuthorizationAPIClient) (bool, error) {
	if st.User == nil {
		return false, nil
	}

	userName := st.User.GetName()

	// Build desired group set from resolved roles + managed group
	desired := make(map[string]*sidecarSvcAuthzTypes.Policy)

	for _, roleRef := range st.Roles {
		desired[roleRef.GetName()] = nil // scope handled by the group itself
	}

	if st.Role != nil {
		desired[st.Role.GetName()] = nil
	}

	// Get current bindings for this user
	currentBindings, err := conn.APIListUserRoleBindings(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIUserRequest{
		User: userName,
	})
	if err != nil {
		if status.Code(err) != codes.NotFound {
			logger.Err(err).Warn("Failed to list user group bindings")
			return false, err
		}
	}

	current := make(map[string]bool)
	if currentBindings != nil {
		for _, b := range currentBindings.GetBindings() {
			current[b.GetRole()] = true
		}
	}

	changed := false

	// Attach missing groups
	for groupName := range desired {
		if current[groupName] {
			continue
		}

		logger.Str("user", userName).Str("group", groupName).Info("Attaching group to user")

		_, err := conn.APIAssignUserRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingRequest{
			User:  userName,
			Role:  groupName,
			Scope: &sidecarSvcAuthzTypes.Policy{}, // empty scope — the group's own scope is the boundary
		})
		if err != nil {
			if status.Code(err) != codes.AlreadyExists {
				logger.Err(err).Str("group", groupName).Warn("Failed to attach group to user")
				return false, err
			}
		}

		changed = true
	}

	// Detach groups that are no longer desired
	for groupName := range current {
		if _, ok := desired[groupName]; ok {
			continue
		}

		logger.Str("user", userName).Str("group", groupName).Info("Detaching group from user")

		_, err := conn.APIRemoveUserRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleRequest{
			User: userName,
			Role: groupName,
		})
		if err != nil {
			if status.Code(err) != codes.NotFound {
				logger.Err(err).Str("group", groupName).Warn("Failed to detach group from user")
				return false, err
			}
		}

		changed = true
	}

	if changed {
		return true, operator.Reconcile("User group bindings updated")
	}

	return false, nil
}
