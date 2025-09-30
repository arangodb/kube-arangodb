//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package throttle

import (
	"sync"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
)

type Inspector interface {
	GetThrottles() Components
}

func NewAlwaysThrottleComponents() Components {
	return NewThrottleComponents(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
}

func NewThrottleComponents(acs, am, at, ar, ap, aps, apsvc, node, pvc, pod, pv, pdb, secret, cm, service, serviceAccount, sm, endpoints time.Duration) Components {
	return &throttleComponents{
		arangoClusterSynchronization: NewThrottle(acs),
		arangoMember:                 NewThrottle(am),
		arangoTask:                   NewThrottle(at),
		arangoRoute:                  NewThrottle(ar),
		arangoProfile:                NewThrottle(ap),
		arangoPlatformStorage:        NewThrottle(aps),
		arangoPlatformService:        NewThrottle(apsvc),
		node:                         NewThrottle(node),
		persistentVolume:             NewThrottle(pv),
		persistentVolumeClaim:        NewThrottle(pvc),
		pod:                          NewThrottle(pod),
		podDisruptionBudget:          NewThrottle(pdb),
		secret:                       NewThrottle(secret),
		configMap:                    NewThrottle(cm),
		service:                      NewThrottle(service),
		serviceAccount:               NewThrottle(serviceAccount),
		serviceMonitor:               NewThrottle(sm),
		endpoints:                    NewThrottle(endpoints),
	}
}

type Components interface {
	ArangoClusterSynchronization() Throttle
	ArangoMember() Throttle
	ArangoTask() Throttle
	ArangoRoute() Throttle
	ArangoProfile() Throttle
	ArangoPlatformStorage() Throttle
	ArangoPlatformService() Throttle
	Node() Throttle
	PersistentVolume() Throttle
	PersistentVolumeClaim() Throttle
	Pod() Throttle
	PodDisruptionBudget() Throttle
	Secret() Throttle
	ConfigMap() Throttle
	Service() Throttle
	ServiceAccount() Throttle
	ServiceMonitor() Throttle
	Endpoints() Throttle

	Get(c definitions.Component) Throttle
	Invalidate(components ...definitions.Component)

	Counts() definitions.ComponentCount
	Copy() Components
}

type throttleComponents struct {
	arangoClusterSynchronization Throttle
	arangoMember                 Throttle
	arangoTask                   Throttle
	arangoRoute                  Throttle
	arangoProfile                Throttle
	arangoPlatformStorage        Throttle
	arangoPlatformService        Throttle
	node                         Throttle
	persistentVolume             Throttle
	persistentVolumeClaim        Throttle
	pod                          Throttle
	podDisruptionBudget          Throttle
	secret                       Throttle
	configMap                    Throttle
	service                      Throttle
	serviceAccount               Throttle
	serviceMonitor               Throttle
	endpoints                    Throttle
}

func (t *throttleComponents) ConfigMap() Throttle {
	return t.configMap
}

func (t *throttleComponents) PersistentVolume() Throttle {
	return t.persistentVolume
}

func (t *throttleComponents) Endpoints() Throttle {
	return t.endpoints
}

func (t *throttleComponents) Counts() definitions.ComponentCount {
	z := definitions.ComponentCount{}

	for _, c := range definitions.AllComponents() {
		z[c] = t.Get(c).Count()
	}

	return z
}

func (t *throttleComponents) Invalidate(components ...definitions.Component) {
	for _, c := range components {
		t.Get(c).Invalidate()
	}
}

func (t *throttleComponents) Get(c definitions.Component) Throttle {
	if t == nil {
		return NewAlwaysThrottle()
	}
	switch c {
	case definitions.ArangoClusterSynchronization:
		return t.arangoClusterSynchronization
	case definitions.ArangoMember:
		return t.arangoMember
	case definitions.ArangoTask:
		return t.arangoTask
	case definitions.ArangoRoute:
		return t.arangoRoute
	case definitions.ArangoProfile:
		return t.arangoProfile
	case definitions.ArangoPlatformStorage:
		return t.arangoPlatformStorage
	case definitions.ArangoPlatformService:
		return t.arangoPlatformService
	case definitions.Node:
		return t.node
	case definitions.PersistentVolume:
		return t.persistentVolume
	case definitions.PersistentVolumeClaim:
		return t.persistentVolumeClaim
	case definitions.Pod:
		return t.pod
	case definitions.PodDisruptionBudget:
		return t.podDisruptionBudget
	case definitions.Secret:
		return t.secret
	case definitions.ConfigMap:
		return t.configMap
	case definitions.Service:
		return t.service
	case definitions.ServiceAccount:
		return t.serviceAccount
	case definitions.ServiceMonitor:
		return t.serviceMonitor
	case definitions.Endpoints:
		return t.endpoints
	default:
		return NewAlwaysThrottle()
	}
}

func (t *throttleComponents) Copy() Components {
	return &throttleComponents{
		arangoClusterSynchronization: t.arangoClusterSynchronization.Copy(),
		arangoMember:                 t.arangoMember.Copy(),
		arangoTask:                   t.arangoTask.Copy(),
		arangoRoute:                  t.arangoRoute.Copy(),
		arangoProfile:                t.arangoProfile.Copy(),
		arangoPlatformStorage:        t.arangoPlatformStorage.Copy(),
		arangoPlatformService:        t.arangoPlatformService.Copy(),
		node:                         t.node.Copy(),
		persistentVolume:             t.persistentVolume.Copy(),
		persistentVolumeClaim:        t.persistentVolumeClaim.Copy(),
		pod:                          t.pod.Copy(),
		podDisruptionBudget:          t.podDisruptionBudget.Copy(),
		secret:                       t.secret.Copy(),
		configMap:                    t.configMap.Copy(),
		service:                      t.service.Copy(),
		serviceAccount:               t.serviceAccount.Copy(),
		serviceMonitor:               t.serviceMonitor.Copy(),
		endpoints:                    t.endpoints.Copy(),
	}
}

func (t *throttleComponents) ArangoClusterSynchronization() Throttle {
	return t.arangoClusterSynchronization
}

func (t *throttleComponents) ArangoMember() Throttle {
	return t.arangoMember
}

func (t *throttleComponents) ArangoTask() Throttle {
	return t.arangoTask
}

func (t *throttleComponents) ArangoRoute() Throttle {
	return t.arangoRoute
}

func (t *throttleComponents) ArangoProfile() Throttle {
	return t.arangoProfile
}

func (t *throttleComponents) ArangoPlatformStorage() Throttle {
	return t.arangoPlatformStorage
}

func (t *throttleComponents) ArangoPlatformService() Throttle {
	return t.arangoPlatformService
}

func (t *throttleComponents) Node() Throttle {
	return t.node
}

func (t *throttleComponents) PersistentVolumeClaim() Throttle {
	return t.persistentVolumeClaim
}

func (t *throttleComponents) Pod() Throttle {
	return t.pod
}

func (t *throttleComponents) PodDisruptionBudget() Throttle {
	return t.podDisruptionBudget
}

func (t *throttleComponents) Secret() Throttle {
	return t.secret
}

func (t *throttleComponents) Service() Throttle {
	return t.service
}

func (t *throttleComponents) ServiceAccount() Throttle {
	return t.serviceAccount
}

func (t *throttleComponents) ServiceMonitor() Throttle {
	return t.serviceMonitor
}

type Throttle interface {
	Invalidate()
	Throttle() bool
	Delay()

	Copy() Throttle

	Count() int
}

func NewAlwaysThrottle() Throttle {
	return &alwaysThrottle{}
}

type alwaysThrottle struct {
	count int
}

func (a alwaysThrottle) Count() int {
	return a.count
}

func (a *alwaysThrottle) Copy() Throttle {
	return a
}

func (a alwaysThrottle) Invalidate() {

}

func (a alwaysThrottle) Throttle() bool {
	return true
}

func (a *alwaysThrottle) Delay() {
	a.count++
}

func NewThrottle(delay time.Duration) Throttle {
	if delay == 0 {
		return NewAlwaysThrottle()
	}
	return &throttle{
		delay: delay,
	}
}

type throttle struct {
	lock sync.Mutex

	delay time.Duration
	next  time.Time
	count int
}

func (t *throttle) Count() int {
	t.lock.Lock()
	defer t.lock.Unlock()

	return t.count
}

func (t *throttle) Copy() Throttle {
	return &throttle{
		delay: t.delay,
		next:  t.next,
		count: t.count,
	}
}

func (t *throttle) Delay() {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.next = time.Now().Add(t.delay)
	t.count++
}

func (t *throttle) Throttle() bool {
	t.lock.Lock()
	defer t.lock.Unlock()

	return t.next.IsZero() || t.next.Before(time.Now())
}

func (t *throttle) Invalidate() {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.next = time.UnixMilli(0)
}
