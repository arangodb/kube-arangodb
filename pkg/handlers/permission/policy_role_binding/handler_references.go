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

package policy_role_binding

import (
	"context"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/permission"
	permissionApi "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
)

func (h *handler) HandleReferences(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionPolicyRoleBinding, status *permissionApi.ArangoPermissionPolicyRoleBindingStatus, depl *api.ArangoDeployment) (bool, error) {
	changed, err := h.handlePolicyReference(ctx, extension, status)
	if err != nil || changed {
		return changed, err
	}

	return h.handleRoleReference(ctx, extension, status)
}

func (h *handler) handlePolicyReference(ctx context.Context, extension *permissionApi.ArangoPermissionPolicyRoleBinding, status *permissionApi.ArangoPermissionPolicyRoleBindingStatus) (bool, error) {
	// CRD reference — resolve the ArangoPermissionPolicy
	policyObj, err := h.client.PermissionV1alpha1().ArangoPermissionPolicies(extension.GetNamespace()).Get(ctx, extension.Spec.Policy.GetReference(), meta.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			if status.Conditions.Update(permissionApi.ReadyPolicyCondition, false, "Policy not found", "ArangoPermissionPolicy not found") {
				return true, operator.Reconcile("Policy not found")
			}
			return false, operator.Stop("Policy not found")
		}
		return false, err
	}

	if !policyObj.Ready() || policyObj.Status.Policy == nil {
		if status.Conditions.Update(permissionApi.ReadyPolicyCondition, false, "Policy not ready", "ArangoPermissionPolicy is not ready") {
			return true, operator.Reconcile("Policy not ready")
		}
		return false, operator.Stop("Policy not ready")
	}

	policyRef := sharedApi.NewObject(policyObj)
	if status.Policy == nil || !status.Policy.Equals(policyObj) {
		status.Policy = &policyRef
		return true, operator.Reconcile("Policy reference updated")
	}

	if status.Conditions.Update(permissionApi.ReadyPolicyCondition, true, "Policy Ready", "Policy Ready") {
		return true, operator.Reconcile("Policy ready")
	}

	return false, nil
}

func (h *handler) handleRoleReference(ctx context.Context, extension *permissionApi.ArangoPermissionPolicyRoleBinding, status *permissionApi.ArangoPermissionPolicyRoleBindingStatus) (bool, error) {
	roleName := extension.Spec.Role.GetReference()

	// A direct reference targets a role that lives only in the authorization sidecar and has no
	// ArangoPermissionRole CRD (e.g. an operator-managed predefined role). The `direct` reference
	// field addresses it directly by name so a policy can be attached to it. The deployment
	// reconciler (SyncRBACPermissions) picks up the binding and merges the policy into the sidecar
	// role. No CRD lookup and no role label - the reconciler discovers these bindings by role name.
	if extension.Spec.Role.IsDirect() {
		if status.Role == nil || status.Role.GetName() != roleName {
			status.Role = &sharedApi.Object{Name: roleName}
			return true, operator.Reconcile("Predefined role reference set")
		}

		if status.Conditions.Update(permissionApi.ReadyRoleCondition, true, "Role Ready", "Predefined role") {
			return true, operator.Reconcile("Role ready")
		}

		return false, nil
	}

	// CRD reference — resolve the ArangoPermissionRole
	roleObj, err := h.client.PermissionV1alpha1().ArangoPermissionRoles(extension.GetNamespace()).Get(ctx, roleName, meta.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			if status.Conditions.Update(permissionApi.ReadyRoleCondition, false, "Role not found", "ArangoPermissionRole not found") {
				return true, operator.Reconcile("Role not found")
			}
			return false, operator.Stop("Role not found")
		}
		return false, err
	}

	if !roleObj.Ready() {
		if status.Conditions.Update(permissionApi.ReadyRoleCondition, false, "Role not ready", "ArangoPermissionRole is not ready") {
			return true, operator.Reconcile("Role not ready")
		}
		return false, operator.Stop("Role not ready")
	}

	roleRef := sharedApi.NewObject(roleObj)
	if status.Role == nil || !status.Role.Equals(roleObj) {
		status.Role = &roleRef
		return true, operator.Reconcile("Role reference updated")
	}

	// Ensure the role label is set on this binding
	labels := extension.GetLabels()
	if labels == nil || labels[permission.LabelPolicyRoleBindingRole] != roleName {
		if labels == nil {
			labels = make(map[string]string)
		}
		labels[permission.LabelPolicyRoleBindingRole] = roleName
		extension.SetLabels(labels)

		if _, err := h.client.PermissionV1alpha1().ArangoPermissionPolicyRoleBindings(extension.GetNamespace()).Update(ctx, extension, meta.UpdateOptions{}); err != nil {
			return false, err
		}

		return true, operator.Reconcile("Role label set")
	}

	if status.Conditions.Update(permissionApi.ReadyRoleCondition, true, "Role Ready", "Role Ready") {
		return true, operator.Reconcile("Role ready")
	}

	return false, nil
}
