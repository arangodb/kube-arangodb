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
	"sync"

	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/deployment/acs/sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

func NewACS(main types.UID, cache inspectorInterface.Inspector) sutil.ACS {
	return &acs{
		acsMain: acsMain{
			uid:   main,
			cache: cache,
		},
	}
}

type acsMain struct {
	uid   types.UID
	cache inspectorInterface.Inspector
}
type acs struct {
	lock sync.Mutex
	acsMain
	items map[types.UID]*item
}

func (a *acs) CurrentClusterCache() inspectorInterface.Inspector {
	return a.cache
}
func (a *acs) ClusterCache(uid types.UID) (inspectorInterface.Inspector, bool) {
	c, ok := a.Cluster(uid)
	if !ok || !c.Ready() {
		return nil, false
	}
	return c.Cache(), true
}
func (a *acs) ForEachHealthyCluster(f func(item sutil.ACSItem) error) error {
	if err := f(a); err != nil {
		return err
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	for _, c := range a.items {
		if !c.Ready() {
			continue
		}
		if err := f(c); err != nil {
			return err
		}
	}
	return nil
}
func (a *acs) Ready() bool {
	return true
}
func (a *acs) UID() types.UID {
	return a.acsMain.uid
}
func (a *acs) Cache() inspectorInterface.Inspector {
	return a.acsMain.cache
}
func (a *acs) Cluster(uid types.UID) (sutil.ACSItem, bool) {
	if uid == "" || uid == a.acsMain.uid {
		return a, true
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	c, ok := a.items[uid]
	return c, ok
}
func (a *acs) RemoteClusters() []types.UID {
	q := make([]types.UID, 0, len(a.items))
	for k := range a.items {
		q = append(q, k)
	}
	return q
}
