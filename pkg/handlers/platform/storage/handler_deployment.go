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

package storage

import (
	"context"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func (h *handler) HandleArangoDeployment(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformStorage, status *platformApi.ArangoPlatformStorageStatus) (bool, error) {
	deployment, err := util.WithKubernetesContextTimeoutP2A2(ctx, h.client.DatabaseV1().ArangoDeployments(item.Namespace).Get, extension.GetName(), meta.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			// Condition for Found should be set to false
			if util.Or(
				status.Conditions.Update(platformApi.DeploymentFoundCondition, false, "ArangoDeployment not found", "ArangoDeployment not found"),
			) {
				return true, operator.Reconcile("Conditions updated")
			}
			return false, operator.Stop("Deployment Not Found")
		}

		return false, err
	}

	// Condition for Found should be set to true

	if status.Conditions.UpdateWithHash(platformApi.DeploymentFoundCondition, true, "ArangoDeployment found", "ArangoDeployment found", string(deployment.GetUID())) {
		return true, operator.Reconcile("Conditions updated")
	}

	return false, nil
}
