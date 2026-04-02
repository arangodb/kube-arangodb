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

package policy

import (
	"context"
	"time"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	permissionApi "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/handlers/permission/shared"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	sidecarSvcAuthzDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/integration"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/patcher"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

type handler struct {
	client     arangoClientSet.Interface
	kubeClient kubernetes.Interface

	eventRecorder event.RecorderInstance

	operator operator.Operator

	provider shared.ClientProvider
}

func (h *handler) Name() string {
	return Kind()
}

func (h *handler) Handle(ctx context.Context, item operation.Item) error {
	// Get Backup object. It also covers NotFound case
	object, err := util.WithKubernetesContextTimeoutP2A2(ctx, h.client.PermissionV1alpha1().ArangoPermissionPolicies(item.Namespace).Get, item.Name, meta.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	if object.GetDeletionTimestamp() != nil {
		// We are deleting the object
		finalizer, err := h.finalizer(ctx, object)
		if err != nil {
			return err
		}

		if finalizer != "" {
			if changed, err := patcher.EnsureFinalizersGone(ctx, h.client.PermissionV1alpha1().ArangoPermissionPolicies(item.Namespace), object, finalizer); err != nil {
				return err
			} else if changed {
				return operator.Reconcile("Finalizers updated")
			}
		}

		return operator.Reconcile("Finalizers pending removal")
	}

	if changed, err := patcher.EnsureFinalizersPresent(ctx, h.client.PermissionV1alpha1().ArangoPermissionPolicies(item.Namespace), object,
		permissionApi.FinalizerArangoPermissionPolicy,
	); err != nil {
		return err
	} else if changed {
		return operator.Reconcile("Finalizers updated")
	}

	status := object.Status.DeepCopy()

	changed, reconcileErr := operator.HandleP3WithStop(ctx, item, object, status, h.handle)
	if reconcileErr != nil && !operator.IsReconcile(reconcileErr) {
		logger.Err(reconcileErr).Warn("Fail for %s %s/%s",
			item.Kind,
			item.Namespace,
			item.Name)

		return reconcileErr
	}

	if !changed {
		return reconcileErr
	}

	logger.Debug("Updating %s %s/%s",
		item.Kind,
		item.Namespace,
		item.Name)

	if _, err := operator.WithArangoPermissionPolicyUpdateStatusInterfaceRetry(context.Background(), h.client.PermissionV1alpha1().ArangoPermissionPolicies(object.GetNamespace()), object, *status, meta.UpdateOptions{}); err != nil {
		return err
	}

	return reconcileErr
}

func (h *handler) CanBeHandled(item operation.Item) bool {
	return item.Group == Group() &&
		utilConstants.Version(Version()).IsCompatible(utilConstants.Version(item.Version)) &&
		item.Kind == Kind()
}

func (h *handler) finalizer(ctx context.Context, extension *permissionApi.ArangoPermissionPolicy) (string, error) {
	for _, finalizer := range extension.GetFinalizers() {
		switch finalizer {
		case permissionApi.FinalizerArangoPermissionPolicy:
			if err := h.finalizerPolicyRemoval(ctx, extension); err != nil {
				return "", err
			}

			return permissionApi.FinalizerArangoPermissionPolicy, nil
		}
	}

	return "", nil
}

func (h *handler) handle(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionPolicy, status *permissionApi.ArangoPermissionPolicyStatus) (bool, error) {
	return operator.HandleP3WithCondition(ctx, &status.Conditions, permissionApi.ReadyCondition, item, extension, status, h.HandleSpecValidity, h.HandleDeployment)
}

func (h *handler) HandleSpecValidity(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionPolicy, status *permissionApi.ArangoPermissionPolicyStatus) (bool, error) {
	if err := extension.Spec.Validate(); err != nil {
		// We have received an error in the spec!

		logger.Err(err).Warn("Invalid Spec on %s", item.String())

		if status.Conditions.Update(permissionApi.SpecValidCondition, false, "Spec is invalid", "Spec is invalid") {
			return true, operator.Stop("Invalid spec")
		}
		return false, operator.Stop("Invalid spec")
	}

	if status.Conditions.Update(permissionApi.SpecValidCondition, true, "Spec is valid", "Spec is valid") {
		logger.WrapObj(item).Debug("Spec is valid")
		return true, nil
	}

	return false, nil
}

func (h *handler) HandleDeployment(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionPolicy, status *permissionApi.ArangoPermissionPolicyStatus) (bool, error) {
	logger := logger.WrapObj(item).Str("deployment", extension.Spec.Deployment.GetName())

	if status.Deployment == nil {
		depl, err := h.client.DatabaseV1().ArangoDeployments(extension.GetNamespace()).Get(ctx, extension.Spec.Deployment.GetName(), meta.GetOptions{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return false, err
			}

			if status.Conditions.Update(permissionApi.DeploymentFoundCondition, false, "Deployment not found", "Deployment not found") {
				logger.Warn("Deployment Not Found")
				return true, operator.Reconcile("Conditions updated")
			}

			return false, operator.Stop("Missing deployment")
		}

		status.Deployment = util.NewType(sharedApi.NewObject(depl))

		logger.Info("Deployment Accepted")

		return true, operator.Reconcile("Deployment Accepted")
	}

	depl, err := h.client.DatabaseV1().ArangoDeployments(extension.GetNamespace()).Get(ctx, extension.Status.Deployment.GetName(), meta.GetOptions{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return false, err
		}

		if status.Conditions.Update(permissionApi.DeploymentFoundCondition, false, "Deployment not found", "Deployment not found") {
			logger.Warn("Deployment Not Found")
			return true, nil
		}

		return false, operator.Stop("Missing deployment, recreate object")
	}

	if !extension.Status.Deployment.Equals(depl) {
		if status.Conditions.Update(permissionApi.DeploymentFoundCondition, false, "Deployment changed", "Deployment changed") {
			logger.Warn("Deployment Changed")
			return true, operator.Reconcile("Conditions updated")
		}

		return false, operator.Stop("Invalid deployment, recreate object")
	}

	if status.Conditions.Update(permissionApi.DeploymentFoundCondition, true, "Deployment found", "Deployment found") {
		logger.Debug("Deployment Found")
		return true, nil
	}

	return operator.HandleP4(ctx, item, extension, status, depl, h.HandleDeploymentSidecarConnection)
}

func (h *handler) HandleDeploymentSidecarConnection(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionPolicy, status *permissionApi.ArangoPermissionPolicyStatus, depl *api.ArangoDeployment) (bool, error) {
	conn, enabled, err := integration.NewIntegrationConnectionFromDeployment(h.kubeClient, depl, utilToken.WithRelativeDuration(time.Minute))
	if err != nil {
		logger.Err(err).Warn("Deployment is not reachable")

		if status.Conditions.Update(permissionApi.SidecarReachableCondition, false, "Deployment sidecar not reachable", "Deployment sidecar not reachable") {
			return true, operator.Reconcile("Conditions updated")
		}

		return false, operator.Stop("Deployment sidecar not reachable")
	}

	if !enabled {
		if status.Conditions.Remove(permissionApi.SidecarReachableCondition) {
			return true, operator.Reconcile("Conditions updated")
		}

		return false, nil
	}

	defer conn.Close()

	if status.Conditions.Update(permissionApi.SidecarReachableCondition, true, "Deployment sidecar reachable", "Deployment sidecar reachable") {
		return true, operator.Reconcile("Conditions updated")
	}

	return operator.HandleP5(ctx, item, extension, status, depl, sidecarSvcAuthzDefinition.NewAuthorizationAPIClient(conn), h.HandleArangoDBPolicy)

}
