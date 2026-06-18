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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs/sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func (i *item) inspectRemoteDeploymentStatus(ctx context.Context, deployment *api.ArangoDeployment, client kclient.Client, cachedStatus inspectorInterface.Inspector, acs *api.ArangoClusterSynchronization, status *api.ArangoClusterSynchronizationStatus) (bool, error) {
	rc, ok := i.client.factory.Client()
	if !ok {
		return false, errors.Reconcile()
	}
	if status.RemoteDeployment != nil {
		return false, nil
	}
	nctx, cancel := globals.GetGlobals().Timeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	// Already checked if not nil before
	racs, err := rc.Arango().DatabaseV1().ArangoClusterSynchronizations(acs.Spec.KubeConfig.Namespace).Get(nctx, acs.GetName(), meta.GetOptions{})
	if err != nil {
		if status.Conditions.Update(sutil.RemoteDeploymentReadyCondition, false, "ACS Is missing", "ACS is missing") {
			return true, errors.Reconcile()
		}
		return false, errors.Reconcile()
	}
	nctx, cancel = globals.GetGlobals().Timeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	rdepl, err := rc.Arango().DatabaseV1().ArangoDeployments(racs.GetNamespace()).Get(nctx, racs.Spec.DeploymentName, meta.GetOptions{})
	if err != nil {
		if status.Conditions.Update(sutil.RemoteDeploymentReadyCondition, false, "ArangoDeployment is missing", "ArangoDeployment is missing") {
			return true, errors.Reconcile()
		}
		return false, errors.Reconcile()
	}
	status.RemoteDeployment = &api.ArangoClusterSynchronizationDeploymentStatus{
		Name:      rdepl.GetName(),
		Namespace: rdepl.GetNamespace(),
		UID:       rdepl.GetUID(),
	}
	status.Conditions.Update(sutil.RemoteDeploymentReadyCondition, true, "ArangoDeployment is present", "")
	return true, nil
}
