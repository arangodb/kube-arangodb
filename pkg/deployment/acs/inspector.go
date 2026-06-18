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
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func (a *acs) Inspect(ctx context.Context, deployment *api.ArangoDeployment, client kclient.Client, cachedStatus inspectorInterface.Inspector) error {
	if acss, err := cachedStatus.ArangoClusterSynchronization().V1(); err == nil {
		f := acss.Filter(arangoClusterSynchronizationFilter(deployment))
		switch s := len(f); s {
		case 0:
			return nil
		case 1:
			acs := f[0]
			obj := a.getItem(acs.GetUID())
			changed, err := obj.Inspect(ctx, deployment, client, cachedStatus, acs)
			if err != nil {
				return err
			}
			if changed {
				cachedStatus.GetThrottles().ArangoClusterSynchronization().Invalidate()
			}
		default:
			errors := make([]error, s)
			var wg sync.WaitGroup
			wg.Add(s)
			changed := false
			for id := 0; id < s; id++ {
				go func(i int) {
					defer wg.Done()
					acs := f[i]
					c, e := a.getItem(acs.GetUID()).Inspect(ctx, deployment, client, cachedStatus, acs)
					if e != nil {
						errors[i] = e
					}
					if c {
						changed = true
					}
				}(id)
			}
			wg.Wait()
			if changed {
				cachedStatus.GetThrottles().ArangoClusterSynchronization().Invalidate()
			}
			return shared.WithErrors(errors...)
		}
	}
	return nil
}
func (a *acs) getItem(uid types.UID) Item {
	a.lock.Lock()
	defer a.lock.Unlock()
	if v, ok := a.items[uid]; ok {
		return v
	}
	v := &item{
		last: time.Now(),
		uid:  uid,
	}
	if a.items == nil {
		a.items = map[types.UID]*item{}
	}
	a.items[uid] = v
	return v
}
func arangoClusterSynchronizationFilter(depl *api.ArangoDeployment) generic.Filter[*api.ArangoClusterSynchronization] {
	return func(acs *api.ArangoClusterSynchronization) bool {
		if d := acs.Status.Deployment; d == nil {
			return acs.Spec.DeploymentName == depl.GetName() && acs.GetNamespace() == depl.GetNamespace()
		} else {
			return d.Name == depl.GetName() && d.Namespace == depl.GetNamespace()
		}
	}
}
