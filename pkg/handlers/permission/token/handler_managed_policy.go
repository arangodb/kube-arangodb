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
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/integration"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

func (h *handler) HandleManagedPolicy(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionToken, st *permissionApi.ArangoPermissionTokenStatus, depl *api.ArangoDeployment, conn sidecarSvcAuthzDefinition.AuthorizationAPIClient) (bool, error) {
	if extension.Spec.Policy != nil && !depl.Status.Conditions.IsTrue(api.ConditionTypeGatewaySidecarEnabled) {
		return false, errors.Errorf("Sidecar is not enabled")
	}

	if extension.Spec.Policy == nil || !depl.GetAcceptedSpec().IsAuthenticated() {
		// Managed policy should be gone
		if st.ManagedPolicy != nil {
			_, err := conn.APIDeletePolicy(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest{
				Name: st.ManagedPolicy.GetName(),
			})
			if err != nil {
				if status.Code(err) != codes.NotFound {
					logger.Err(err).Warn("Failed to delete managed policy")
					return false, err
				}
			}

			st.ManagedPolicy = nil
			return true, operator.Reconcile("Managed policy removed")
		}

		return false, nil
	}

	policy, err := h.renderPolicy(extension.Spec.Policy)
	if err != nil {
		logger.Err(err).Warn("Failed to render managed policy")
		return false, err
	}

	name := fmt.Sprintf("managed:operator:%s:policy", extension.GetUID())

	if st.ManagedPolicy == nil {
		if _, err := conn.APICreatePolicy(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIPolicyRequest{
			Name: name,
			Item: policy,
		}); err != nil {
			if status.Code(err) != codes.AlreadyExists {
				logger.Err(err).Warn("Failed to create managed policy")
				return false, err
			}
		}

		h.eventRecorder.Normal(extension, "Managed Policy Created", "Managed policy has been created with hash %s", policy.Hash())

		st.ManagedPolicy = &sharedApi.Object{
			Name:     name,
			Checksum: util.NewType(policy.Hash()),
		}

		return true, operator.Reconcile("Managed policy created")
	}

	existing, err := conn.APIGetPolicy(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest{
		Name: st.ManagedPolicy.GetName(),
	})
	if err != nil {
		if status.Code(err) != codes.NotFound {
			logger.Err(err).Warn("Failed to get managed policy")
			return false, err
		}

		st.ManagedPolicy = nil
		return true, operator.Reconcile("Managed policy gone")
	}

	if st.ManagedPolicy.GetChecksum() != policy.Hash() || existing.Item.Hash() != policy.Hash() {
		if _, err := conn.APIUpdatePolicy(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIPolicyRequest{
			Name: name,
			Item: policy,
		}); err != nil {
			if status.Code(err) != codes.AlreadyExists {
				logger.Err(err).Warn("Failed to update managed policy")
				return false, err
			}
		}

		h.eventRecorder.Normal(extension, "Managed Policy Updated", "Managed policy has been updated with hash %s", policy.Hash())

		st.ManagedPolicy = &sharedApi.Object{
			Name:     name,
			Checksum: util.NewType(policy.Hash()),
		}

		return true, operator.Reconcile("Managed policy updated")
	}

	return false, nil
}

func (h *handler) finalizerManagedPolicyRemoval(ctx context.Context, extension *permissionApi.ArangoPermissionToken) error {
	if extension.Status.Deployment == nil || extension.Status.ManagedPolicy == nil {
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

	if _, err := client.APIDeletePolicy(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest{
		Name: extension.Status.ManagedPolicy.GetName(),
	}); err != nil {
		if status.Code(err) == codes.NotFound {
			return nil
		}
		return err
	}

	return nil
}
