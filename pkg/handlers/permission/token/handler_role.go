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
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
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

func (h *handler) HandleArangoDBRole(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionToken, st *permissionApi.ArangoPermissionTokenStatus, depl *api.ArangoDeployment, conn sidecarSvcAuthzDefinition.AuthorizationAPIClient) (bool, error) {
	if extension.Spec.Policy != nil && !depl.Status.Conditions.IsTrue(api.ConditionTypeGatewaySidecarEnabled) {
		return false, errors.Errorf("Sidecar is not enabled")
	}

	if extension.Spec.Policy == nil || !depl.GetAcceptedSpec().IsAuthenticated() {
		// Role should be gone
		if st.Role != nil {
			_, err := conn.APIDeleteRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest{
				Name: st.Role.GetName(),
			})
			if err != nil {
				if status.Code(err) != codes.NotFound {
					logger.Err(err).Warn("Failed to delete policy")
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

	if st.User == nil {
		if st.Conditions.UpdateWithHash(permissionApi.ReadyRoleCondition, false, "Role Changed", "Role Changed", "") {
			return true, operator.Reconcile("User gone")
		}

		return false, nil
	}

	if st.Policy == nil {
		if st.Conditions.UpdateWithHash(permissionApi.ReadyRoleCondition, false, "Role Changed", "Role Changed", "") {
			return true, operator.Reconcile("Policy gone")
		}

		return false, nil
	}

	role, err := h.renderRole(st.User.GetName(), st.Policy.GetName())
	if err != nil {
		logger.Err(err).Warn("Failed to render policy")
		return false, err
	}

	name := fmt.Sprintf("managed:operator:%s", extension.GetUID())

	if st.Role == nil {
		// Create the policy
		if _, err := conn.APICreateRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIRoleRequest{
			Name: name,
			Item: role,
		}); err != nil {
			if status.Code(err) != codes.AlreadyExists {
				logger.Err(err).Warn("Failed to create policy")
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
			logger.Err(err).Warn("Failed to get policy")
			return false, err
		}

		st.Role = nil
		return true, operator.Reconcile("Role gone")
	}

	if st.Role.GetChecksum() != role.Hash() || existing.Item.Hash() != role.Hash() {
		if st.Conditions.UpdateWithHash(permissionApi.ReadyRoleCondition, false, "Role Changed", "Role Changed", role.Hash()) {
			return true, operator.Reconcile("Role changed")
		}

		// Create the policy
		if _, err := conn.APIUpdateRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIRoleRequest{
			Name: name,
			Item: role,
		}); err != nil {
			if status.Code(err) != codes.AlreadyExists {
				logger.Err(err).Warn("Failed to create policy")
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

func (h *handler) renderRole(user, policy string) (*sidecarSvcAuthzTypes.Role, error) {
	var r sidecarSvcAuthzTypes.Role

	r.Users = []string{user}
	r.Policies = []string{policy}

	if err := r.Validate(); err != nil {
		return nil, err
	}

	if err := r.Clean(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (h *handler) finalizerRoleRemoval(ctx context.Context, extension *permissionApi.ArangoPermissionToken) error {
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

	conn, err := integration.NewIntegrationConnectionFromDeployment(h.kubeClient, depl, utilToken.WithRelativeDuration(time.Minute))
	if err != nil {
		return err
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
