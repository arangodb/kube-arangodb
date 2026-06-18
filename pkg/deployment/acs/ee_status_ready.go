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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs/sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func (i *item) inspectReadyConditionPre(ctx context.Context, deployment *api.ArangoDeployment, client kclient.Client, cachedStatus inspectorInterface.Inspector, acs *api.ArangoClusterSynchronization, status *api.ArangoClusterSynchronizationStatus) (bool, error) {
	if acs.Status.Conditions.IsTrue(sutil.DeploymentReadyCondition) &&
		acs.Status.Conditions.IsTrue(sutil.KubernetesConnectedCondition) &&
		acs.Status.Conditions.IsTrue(sutil.RemoteDeploymentReadyCondition) &&
		acs.Status.Conditions.IsTrue(sutil.RemoteCacheReadyCondition) {
		return status.Conditions.Update(sutil.ConnectionReadyCondition, true, "Connection is ready", ""), nil
	} else {
		return status.Conditions.Update(sutil.ConnectionReadyCondition, false, "Connection is not ready", ""), nil
	}
}
func (i *item) inspectReadyConditionPost(ctx context.Context, deployment *api.ArangoDeployment, client kclient.Client, cachedStatus inspectorInterface.Inspector, acs *api.ArangoClusterSynchronization, status *api.ArangoClusterSynchronizationStatus) (bool, error) {
	if !status.Conditions.IsTrue(sutil.ConnectionReadyCondition) {
		return false, errors.Reconcile()
	}
	return false, nil
}
