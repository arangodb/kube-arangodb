//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package acs

import (
	"context"
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs/sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func (i *item) inspectExistingCondition(ctx context.Context, deployment *api.ArangoDeployment, client kclient.Client, cachedStatus inspectorInterface.Inspector, acs *api.ArangoClusterSynchronization, status *api.ArangoClusterSynchronizationStatus) (bool, error) {
	if d := status.Deployment; d == nil {
		// Fill deployment
		status.Deployment = &api.ArangoClusterSynchronizationDeploymentStatus{
			Name:      deployment.GetName(),
			Namespace: deployment.GetNamespace(),
			UID:       deployment.GetUID(),
		}
		status.Conditions.Update(sutil.DeploymentReadyCondition, true, "Deployment found", "")
		return true, nil
	} else {
		if d.Name != deployment.GetName() || d.Namespace != deployment.GetNamespace() {
			if status.Conditions.Update(sutil.DeploymentReadyCondition, false, "Deployment not found", "") {
				return true, nil
			}
			return false, nil
		}
		if d.UID != deployment.GetUID() {
			if status.Conditions.Update(sutil.DeploymentReadyCondition, false, "Deployment UUID changed", "") {
				return true, nil
			}
			return false, nil
		}
	}
	i.last = time.Now()
	return false, nil
}
