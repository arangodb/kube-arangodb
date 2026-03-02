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
	permissionApiPolicy "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1/policy"
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

func (h *handler) HandleArangoDBPolicy(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionToken, st *permissionApi.ArangoPermissionTokenStatus, depl *api.ArangoDeployment, conn sidecarSvcAuthzDefinition.AuthorizationAPIClient) (bool, error) {
	if extension.Spec.Policy != nil && !depl.Status.Conditions.IsTrue(api.ConditionTypeGatewaySidecarEnabled) {
		return false, errors.Errorf("Sidecar is not enabled")
	}

	if extension.Spec.Policy == nil || !depl.GetAcceptedSpec().IsAuthenticated() {
		// Policy should be gone
		if st.Policy != nil {
			_, err := conn.APIDeletePolicy(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest{
				Name: st.Policy.GetName(),
			})
			if err != nil {
				if status.Code(err) != codes.NotFound {
					logger.Err(err).Warn("Failed to delete policy")
					return false, err
				}
			}

			st.Policy = nil

			return true, operator.Reconcile("Policy removed")
		}

		if st.Conditions.Remove(permissionApi.ReadyPolicyCondition) {
			return true, operator.Reconcile("Policy removed")
		}

		return false, nil
	}

	policies, err := h.renderPolicy(extension.Spec.Policy)
	if err != nil {
		logger.Err(err).Warn("Failed to render policy")
		return false, err
	}

	name := fmt.Sprintf("managed:operator:%s", extension.GetUID())

	if st.Policy == nil {
		// Create the policy
		if _, err := conn.APICreatePolicy(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIPolicyRequest{
			Name: name,
			Item: policies,
		}); err != nil {
			if status.Code(err) != codes.AlreadyExists {
				logger.Err(err).Warn("Failed to create policy")
				return false, err
			}
		}
		h.eventRecorder.Normal(extension, "Policy Created", "Policy has been created with hash %s", policies.Hash())

		st.Policy = &sharedApi.Object{
			Name:     name,
			Checksum: util.NewType(policies.Hash()),
		}

		return true, operator.Reconcile("Policy created")
	}

	existing, err := conn.APIGetPolicy(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest{
		Name: st.Policy.GetName(),
	})
	if err != nil {
		if status.Code(err) != codes.NotFound {
			logger.Err(err).Warn("Failed to get policy")
			return false, err
		}

		st.Policy = nil
		return true, operator.Reconcile("Policy gone")
	}

	if st.Policy.GetChecksum() != policies.Hash() || existing.Item.Hash() != policies.Hash() {
		if st.Conditions.UpdateWithHash(permissionApi.ReadyPolicyCondition, false, "Policy Changed", "Policy Changed", policies.Hash()) {
			return true, operator.Reconcile("Policy changed")
		}

		// Create the policy
		if _, err := conn.APIUpdatePolicy(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIPolicyRequest{
			Name: name,
			Item: policies,
		}); err != nil {
			if status.Code(err) != codes.AlreadyExists {
				logger.Err(err).Warn("Failed to create policy")
				return false, err
			}
		}

		h.eventRecorder.Normal(extension, "Policy Updated", "Policy has been updated with hash %s", policies.Hash())

		st.Policy = &sharedApi.Object{
			Name:     name,
			Checksum: util.NewType(policies.Hash()),
		}

		return true, operator.Reconcile("Policy updated")
	}

	if st.Conditions.UpdateWithHash(permissionApi.ReadyPolicyCondition, true, "Policy Ready", "Policy Ready", policies.Hash()) {
		return true, operator.Reconcile("Policy created")
	}

	return false, nil
}

func (h *handler) renderPolicy(in *permissionApiPolicy.Policy) (*sidecarSvcAuthzTypes.Policy, error) {
	var r sidecarSvcAuthzTypes.Policy

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

	if err := r.Validate(); err != nil {
		return nil, err
	}

	if err := r.Clean(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (h *handler) finalizerPolicyRemoval(ctx context.Context, extension *permissionApi.ArangoPermissionToken) error {
	if extension.Status.Deployment == nil || extension.Status.Policy == nil {
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

	if _, err := client.APIDeletePolicy(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest{
		Name: extension.Status.Policy.GetName(),
	}); err != nil {
		if status.Code(err) == codes.NotFound {
			return nil
		}

		return err
	}

	return nil
}
