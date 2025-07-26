//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package route

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1beta1"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

var logger = logging.Global().RegisterAndGetLogger("networking-route-operator", logging.Info)

type handler struct {
	client     arangoClientSet.Interface
	kubeClient kubernetes.Interface

	eventRecorder event.RecorderInstance

	operator operator.Operator
}

func (h *handler) Name() string {
	return Kind()
}

func (h *handler) Handle(ctx context.Context, item operation.Item) error {
	// Get Backup object. It also covers NotFound case

	object, err := util.WithKubernetesContextTimeoutP2A2(ctx, h.client.NetworkingV1beta1().ArangoRoutes(item.Namespace).Get, item.Name, meta.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}

		return err
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

	if _, err := operator.WithNetworkingArangoRouteUpdateStatusInterfaceRetry(context.Background(), h.client.NetworkingV1beta1().ArangoRoutes(object.GetNamespace()), object, *status, meta.UpdateOptions{}); err != nil {
		return err
	}

	return reconcileErr
}

func (h *handler) handle(ctx context.Context, item operation.Item, extension *networkingApi.ArangoRoute, status *networkingApi.ArangoRouteStatus) (bool, error) {
	return operator.HandleP3WithCondition(ctx, &status.Conditions, networkingApi.ReadyCondition, item, extension, status, h.HandleSpecValidity, h.HandleArangoDeployment)
}

func (h *handler) HandleSpecValidity(ctx context.Context, item operation.Item, extension *networkingApi.ArangoRoute, status *networkingApi.ArangoRouteStatus) (bool, error) {
	if err := extension.Spec.Validate(); err != nil {
		// We have received an error in the spec!

		logger.Err(err).Warn("Invalid Spec on %s", item.String())

		if status.Conditions.Update(networkingApi.SpecValidCondition, false, "Spec is invalid", err.Error()) {
			return true, operator.Stop("Invalid spec")
		}
		return false, operator.Stop("Invalid spec")
	}

	if status.Conditions.Update(networkingApi.SpecValidCondition, true, "Spec is valid", "Spec is valid") {
		return true, operator.Reconcile("Conditions updated")
	}

	return false, nil
}

func (h *handler) CanBeHandled(item operation.Item) bool {
	return item.Group == Group() &&
		constants.IsCompatible(Version(), item.Version) &&
		item.Kind == Kind()
}
