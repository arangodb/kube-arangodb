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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs/sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

type Item interface {
	Inspect(ctx context.Context, deployment *api.ArangoDeployment, client kclient.Client, cachedStatus inspectorInterface.Inspector, acs *api.ArangoClusterSynchronization) (bool, error)
}
type item struct {
	last   time.Time
	uid    types.UID
	ok     bool
	client *itemClient
	cache  inspectorInterface.Inspector
}

func (i *item) Ready() bool {
	return i.ok
}
func (i *item) UID() types.UID {
	return i.uid
}
func (i *item) Cache() inspectorInterface.Inspector {
	return i.cache
}

type itemClient struct {
	factory kclient.Factory
}

func (i *item) Inspect(ctx context.Context, deployment *api.ArangoDeployment, client kclient.Client, cachedStatus inspectorInterface.Inspector, acs *api.ArangoClusterSynchronization) (bool, error) {
	status := acs.Status.DeepCopy()
	defer func() {
		i.ok = status.Conditions.IsTrue(sutil.ConnectionReadyCondition)
	}()
	if changed, cerr := i.applyInspectStatus(ctx, deployment, client, cachedStatus, acs, status, i.inspectReadyConditionPre, i.inspectExistingCondition, i.inspectKubeClient, i.inspectRemoteDeploymentStatus, i.inspectKubeCache, i.inspectReadyConditionPost); changed {
		acs.Status = *status
		nctx, cancel := globals.GetGlobals().Timeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()
		if _, err := client.Arango().DatabaseV1().ArangoClusterSynchronizations(acs.GetNamespace()).UpdateStatus(nctx, acs, meta.UpdateOptions{}); err != nil {
			return false, err
		}
		if cerr != nil {
			return true, cerr
		}
		return true, nil
	}
	return false, nil
}

type itemInspector func(ctx context.Context, deployment *api.ArangoDeployment, client kclient.Client, cachedStatus inspectorInterface.Inspector, acs *api.ArangoClusterSynchronization, status *api.ArangoClusterSynchronizationStatus) (bool, error)

func (i item) applyInspectStatus(ctx context.Context, deployment *api.ArangoDeployment, client kclient.Client, cachedStatus inspectorInterface.Inspector, acs *api.ArangoClusterSynchronization, status *api.ArangoClusterSynchronizationStatus, items ...itemInspector) (bool, error) {
	changed := false
	for _, item := range items {
		if c, err := item(ctx, deployment, client, cachedStatus, acs, status); err != nil {
			if errors.IsReconcile(err) {
				return changed, nil
			}
			return changed, err
		} else if c {
			changed = true
		}
	}
	return changed, nil
}
