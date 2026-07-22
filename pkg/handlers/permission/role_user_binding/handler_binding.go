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

package role_user_binding

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	permissionApi "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1"
	permissionApiPolicy "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1/policy"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	sidecarSvcAuthzDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/integration"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

func (h *handler) HandleDeploymentSidecarConnection(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionRoleUserBinding, st *permissionApi.ArangoPermissionRoleUserBindingStatus, depl *api.ArangoDeployment) (bool, error) {
	conn, enabled, err := integration.NewIntegrationConnectionFromDeployment(h.kubeClient, depl, utilToken.WithRelativeDuration(time.Minute))
	if err != nil {
		logger.Err(err).Warn("Deployment is not reachable")

		if st.Conditions.Update(permissionApi.SidecarReachableCondition, false, "Deployment sidecar not reachable", "Deployment sidecar not reachable") {
			return true, operator.Reconcile("Conditions updated")
		}

		return false, operator.Stop("Deployment sidecar not reachable")
	}

	if !enabled {
		if st.Conditions.Remove(permissionApi.SidecarReachableCondition) {
			return true, operator.Reconcile("Conditions updated")
		}

		return false, nil
	}

	defer conn.Close()

	if st.Conditions.Update(permissionApi.SidecarReachableCondition, true, "Deployment sidecar reachable", "Deployment sidecar reachable") {
		return true, operator.Reconcile("Conditions updated")
	}

	return operator.HandleP5(ctx, item, extension, st, depl, sidecarSvcAuthzDefinition.NewAuthorizationAPIClient(conn), h.HandleArangoDBBinding)
}

func (h *handler) HandleArangoDBBinding(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionRoleUserBinding, st *permissionApi.ArangoPermissionRoleUserBindingStatus, depl *api.ArangoDeployment, client sidecarSvcAuthzDefinition.AuthorizationAPIClient) (bool, error) {
	// Resolve the referenced ArangoPermissionRole and ensure it is reconciled into the sidecar.
	roleName := extension.Spec.Role.GetReference()

	var sidecarRole string

	if extension.Spec.Role.IsDirect() {
		// A direct sidecar reference targets a role that lives only in the authorization sidecar
		// and has no ArangoPermissionRole CRD; the reference resolves to the sidecar role name.
		if st.Role == nil || st.Role.GetName() != roleName {
			st.Role = &sharedApi.Object{Name: roleName}
			return true, operator.Reconcile("Predefined role reference set")
		}

		sidecarRole = roleName
	} else {
		roleObj, err := h.client.PermissionV1alpha1().ArangoPermissionRoles(extension.GetNamespace()).Get(ctx, roleName, meta.GetOptions{})
		if err != nil {
			if apiErrors.IsNotFound(err) {
				if st.Conditions.Update(permissionApi.ReadyRoleCondition, false, "Role not found", "ArangoPermissionRole not found") {
					return true, operator.Reconcile("Role not found")
				}
				return false, operator.Stop("Role not found")
			}
			return false, err
		}

		if !roleObj.Ready() || roleObj.Status.Role == nil {
			if st.Conditions.Update(permissionApi.ReadyRoleCondition, false, "Role not ready", "ArangoPermissionRole is not ready") {
				return true, operator.Reconcile("Role not ready")
			}
			return false, operator.Stop("Role not ready")
		}

		if st.Role == nil || !st.Role.Equals(roleObj) {
			st.Role = util.NewType(sharedApi.NewObject(roleObj))
			return true, operator.Reconcile("Role reference updated")
		}

		// The sidecar role name matches the ArangoPermissionRole status reference.
		sidecarRole = roleObj.Status.Role.GetName()
	}

	scope, err := renderScope(extension.Spec.Scope)
	if err != nil {
		logger.Err(err).Warn("Failed to render scope policy")
		return false, operator.Stop("Invalid scope")
	}

	hash := extension.Spec.Hash()

	if st.UserRoleBinding == nil {
		if _, err := client.APIAssignUserRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingRequest{
			User:  extension.Spec.UserName,
			Role:  sidecarRole,
			Scope: scope,
		}); err != nil {
			if status.Code(err) != codes.AlreadyExists {
				logger.Err(err).Warn("Failed to assign user role")
				return false, err
			}
		}

		h.eventRecorder.Normal(extension, "Binding Created", "User role binding created for user %s on role %s", extension.Spec.UserName, sidecarRole)

		st.UserRoleBinding = &sharedApi.Object{Name: extension.Spec.UserName, Checksum: util.NewType(hash)}
		return true, operator.Reconcile("User role binding created")
	}

	if st.UserRoleBinding.GetChecksum() != hash {
		if _, err := client.APIReplaceUserRoleScope(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingRequest{
			User:  extension.Spec.UserName,
			Role:  sidecarRole,
			Scope: scope,
		}); err != nil {
			if status.Code(err) == codes.NotFound {
				st.UserRoleBinding = nil
				return true, operator.Reconcile("User role binding gone, recreate")
			}
			logger.Err(err).Warn("Failed to replace user role scope")
			return false, err
		}

		h.eventRecorder.Normal(extension, "Binding Updated", "User role binding scope updated for user %s on role %s", extension.Spec.UserName, sidecarRole)

		st.UserRoleBinding = &sharedApi.Object{Name: extension.Spec.UserName, Checksum: util.NewType(hash)}
		return true, operator.Reconcile("User role binding updated")
	}

	if st.Conditions.Update(permissionApi.ReadyRoleCondition, true, "Binding Ready", "Binding Ready") {
		return true, operator.Reconcile("Binding ready")
	}

	return false, nil
}

// finalizerBindingRemoval detaches the user role binding from the sidecar when the
// CRD is being deleted.
func (h *handler) finalizerBindingRemoval(ctx context.Context, extension *permissionApi.ArangoPermissionRoleUserBinding) error {
	if extension.Status.Deployment == nil || extension.Status.UserRoleBinding == nil || extension.Status.Role == nil {
		return nil
	}

	depl, err := h.client.DatabaseV1().ArangoDeployments(extension.GetNamespace()).Get(ctx, extension.Status.Deployment.GetName(), meta.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	if !extension.Status.Deployment.Equals(depl) {
		logger.Warn("Removal of the user role binding not allowed due to change in UUID")
		return nil
	}

	conn, enabled, err := integration.NewIntegrationConnectionFromDeployment(h.kubeClient, depl, utilToken.WithRelativeDuration(time.Minute))
	if err != nil {
		return err
	}

	if !enabled {
		return nil
	}

	defer conn.Close()

	client := sidecarSvcAuthzDefinition.NewAuthorizationAPIClient(conn)

	if _, err := client.APIRemoveUserRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleRequest{
		User: extension.Spec.UserName,
		Role: extension.Status.Role.GetName(),
	}); err != nil {
		if status.Code(err) == codes.NotFound {
			return nil
		}

		return err
	}

	return nil
}

// renderScope converts the inline CRD scope policy into a sidecar policy. The scope
// is required on the binding spec; an empty/undefined scope would deny everything.
func renderScope(in *permissionApiPolicy.Policy) (*sidecarSvcAuthzTypes.Policy, error) {
	var r sidecarSvcAuthzTypes.Policy

	if in != nil {
		for _, st := range in.Statements {
			var s sidecarSvcAuthzTypes.PolicyStatement

			s.Effect = util.BoolSwitch(st.Effect == permissionApiPolicy.EffectAllow, sidecarSvcAuthzTypes.Effect_Allow, sidecarSvcAuthzTypes.Effect_Deny)
			s.Resources = util.FormatList(st.Resources, func(a permissionApiPolicy.Resource) string {
				return string(a)
			})
			s.Actions = util.FormatList(st.Actions, func(a permissionApiPolicy.Action) string {
				return string(a)
			})

			r.Statements = append(r.Statements, &s)
		}
	}

	if err := r.Validate(); err != nil {
		return nil, err
	}

	if err := r.Clean(); err != nil {
		return nil, err
	}

	return &r, nil
}
