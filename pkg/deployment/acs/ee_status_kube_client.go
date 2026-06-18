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

func (i *item) inspectKubeClient(ctx context.Context, deployment *api.ArangoDeployment, client kclient.Client, cachedStatus inspectorInterface.Inspector, acs *api.ArangoClusterSynchronization, status *api.ArangoClusterSynchronizationStatus) (bool, error) {
	if i.client == nil {
		i.client = &itemClient{
			factory: kclient.GetUnattachedFactory(),
		}
		i.client.factory.SetKubeConfigGetter(ACSConfigGetter(cachedStatus, acs.GetName()))
	}
	if err := i.client.factory.Refresh(); err != nil {
		if status.Conditions.Update(sutil.KubernetesConnectedCondition, false, "Kubernetes not connected", err.Error()) {
			return true, errors.Reconcile()
		}
		return false, errors.Reconcile()
	}
	return status.Conditions.Update(sutil.KubernetesConnectedCondition, true, "Kubernetes connected", ""), nil
}
