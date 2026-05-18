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

package role

import (
	"context"
	"fmt"
	"sort"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/permission"
	permissionApi "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	sidecarSvcAuthzDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/integration"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

func (h *handler) HandleArangoDBRole(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionRole, st *permissionApi.ArangoPermissionRoleStatus, depl *api.ArangoDeployment, conn sidecarSvcAuthzDefinition.AuthorizationAPIClient) (bool, error) {
	if !depl.Status.Conditions.IsTrue(api.ConditionTypeGatewaySidecarEnabled) {
		return false, errors.Errorf("Sidecar is not enabled")
	}

	if !depl.GetAcceptedSpec().IsAuthenticated() {
		// Role should be gone
		if st.Role != nil {
			_, err := conn.APIDeleteRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest{
				Name: st.Role.GetName(),
			})
			if err != nil {
				if status.Code(err) != codes.NotFound {
					logger.Err(err).Warn("Failed to delete role")
					return false, err
				}
			}

			st.Role = nil

			return true, operator.Reconcile("Role removed")
		}

		if st.Conditions.Remove(permissionApi.ReadyRoleCondition) {
			return true, operator.Reconcile("Role removed")
		}

		return false, nil
	}

	// List bindings targeting this role by label and collect policy sidecar names
	policies, policyRefs, err := h.collectPoliciesFromBindings(ctx, extension)
	if err != nil {
		logger.Err(err).Warn("Failed to collect policies from bindings")
		return false, err
	}

	// Update Status.Policies if changed
	if !policyRefsEqual(st.Policies, policyRefs) {
		st.Policies = policyRefs
		return true, operator.Reconcile("Policies updated")
	}

	role, err := renderRole(policies)
	if err != nil {
		logger.Err(err).Warn("Failed to render role")
		return false, err
	}

	name := extension.GetName()

	if st.Role == nil {
		// Create the role
		if _, err := conn.APICreateRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIRoleRequest{
			Name: name,
			Item: role,
		}); err != nil {
			if status.Code(err) != codes.AlreadyExists {
				logger.Err(err).Warn("Failed to create role")
				return false, err
			}
		}

		h.eventRecorder.Normal(extension, "Role Created", "Role has been created with hash %s", role.Hash())

		st.Role = &sharedApi.Object{
			Name:     name,
			Checksum: util.NewType(role.Hash()),
		}

		return true, operator.Reconcile("Role created")
	}

	existing, err := conn.APIGetRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest{
		Name: st.Role.GetName(),
	})
	if err != nil {
		if status.Code(err) != codes.NotFound {
			logger.Err(err).Warn("Failed to get role")
			return false, err
		}

		logger.Str("name", st.Role.GetName()).Info("Role gone")

		st.Role = nil
		return true, operator.Reconcile("Role gone")
	}

	if st.Role.GetChecksum() != role.Hash() || existing.Item.Hash() != role.Hash() {
		if st.Conditions.UpdateWithHash(permissionApi.ReadyRoleCondition, false, "Role Changed", "Role Changed", role.Hash()) {
			return true, operator.Reconcile("Role changed")
		}

		// Update the role
		if _, err := conn.APIUpdateRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIRoleRequest{
			Name: name,
			Item: role,
		}); err != nil {
			if status.Code(err) != codes.AlreadyExists {
				logger.Err(err).Warn("Failed to update role")
				return false, err
			}
		}

		h.eventRecorder.Normal(extension, "Role Updated", "Role has been updated with hash %s", role.Hash())

		st.Role = &sharedApi.Object{
			Name:     name,
			Checksum: util.NewType(role.Hash()),
		}

		return true, operator.Reconcile("Role updated")
	}

	if st.Conditions.UpdateWithHash(permissionApi.ReadyRoleCondition, true, "Role Ready", "Role Ready", role.Hash()) {
		return true, operator.Reconcile("Role created")
	}

	return false, nil
}

// collectPoliciesFromBindings lists bindings targeting this role by label and resolves
// their policy CRD references to sidecar policy names.
func (h *handler) collectPoliciesFromBindings(ctx context.Context, extension *permissionApi.ArangoPermissionRole) ([]string, []permissionApi.ArangoPermissionBindingRef, error) {
	bindings, err := h.client.PermissionV1alpha1().ArangoPermissionPolicyRoleBindings(extension.GetNamespace()).List(ctx, meta.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", permission.LabelPolicyRoleBindingRole, extension.GetName()),
	})
	if err != nil {
		return nil, nil, err
	}

	var policies []string
	var policyRefs []permissionApi.ArangoPermissionBindingRef
	seenNames := map[string]struct{}{}
	seenRefs := map[string]struct{}{}

	for _, binding := range bindings.Items {
		if !binding.Status.Conditions.IsTrue(permissionApi.ReadyPolicyCondition) || !binding.Status.Conditions.IsTrue(permissionApi.ReadyRoleCondition) {
			continue
		}

		if binding.Spec.Policy == nil {
			continue
		}

		ref := binding.Spec.Policy
		refKey := ref.Hash()
		if _, ok := seenRefs[refKey]; ok {
			continue
		}
		seenRefs[refKey] = struct{}{}
		policyRefs = append(policyRefs, *ref)

		// Resolve CRD reference to sidecar name
		crdName := ref.GetReference()
		if crdName == "" {
			continue
		}

		policyObj, err := h.client.PermissionV1alpha1().ArangoPermissionPolicies(extension.GetNamespace()).Get(ctx, crdName, meta.GetOptions{})
		if err != nil {
			continue
		}

		if !policyObj.Ready() || policyObj.Status.Policy == nil {
			continue
		}

		sidecarName := policyObj.Status.Policy.GetName()
		if _, ok := seenNames[sidecarName]; ok {
			continue
		}
		seenNames[sidecarName] = struct{}{}
		policies = append(policies, sidecarName)
	}

	sort.Strings(policies)
	sort.Slice(policyRefs, func(i, j int) bool {
		return policyRefs[i].Hash() < policyRefs[j].Hash()
	})

	return policies, policyRefs, nil
}

// policyRefsEqual compares two policy ref lists.
func policyRefsEqual(a, b []permissionApi.ArangoPermissionBindingRef) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Hash() != b[i].Hash() {
			return false
		}
	}
	return true
}

// renderRole builds a sidecar Role with an open scope (allow all) and the given policy names.
func renderRole(policies []string) (*sidecarSvcAuthzTypes.Role, error) {
	var r sidecarSvcAuthzTypes.Role

	r.Policies = policies
	r.Scope = &sidecarSvcAuthzTypes.Policy{
		Statements: []*sidecarSvcAuthzTypes.PolicyStatement{
			{
				Effect:    sidecarSvcAuthzTypes.Effect_Allow,
				Actions:   []string{"*"},
				Resources: []string{"*"},
			},
		},
	}

	if err := r.Validate(); err != nil {
		return nil, err
	}

	if err := r.Clean(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (h *handler) finalizerRoleRemoval(ctx context.Context, extension *permissionApi.ArangoPermissionRole) error {
	if extension.Status.Deployment == nil || extension.Status.Role == nil {
		return nil
	}

	depl, err := h.client.DatabaseV1().ArangoDeployments(extension.GetNamespace()).Get(ctx, extension.Status.Deployment.GetName(), meta.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	if !extension.Status.Deployment.Equals(depl) {
		logger.Warn("Deleting of the user not allowed due to change in UUID")
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

	if _, err := client.APIDeleteRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest{
		Name: extension.Status.Role.GetName(),
	}); err != nil {
		if status.Code(err) == codes.NotFound {
			return nil
		}

		return err
	}

	return nil
}
