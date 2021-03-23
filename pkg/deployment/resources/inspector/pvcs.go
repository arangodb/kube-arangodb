//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package inspector

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (i *inspector) IteratePersistentVolumeClaims(action persistentvolumeclaim.PersistentVolumeClaimAction, filters ...persistentvolumeclaim.PersistentVolumeClaimFilter) error {
	for _, pvc := range i.PersistentVolumeClaims() {
		if err := i.iteratePersistentVolumeClaim(pvc, action, filters...); err != nil {
			return err
		}
	}
	return nil
}

func (i *inspector) iteratePersistentVolumeClaim(pvc *core.PersistentVolumeClaim, action persistentvolumeclaim.PersistentVolumeClaimAction, filters ...persistentvolumeclaim.PersistentVolumeClaimFilter) error {
	for _, filter := range filters {
		if !filter(pvc) {
			return nil
		}
	}

	return action(pvc)
}

func (i *inspector) PersistentVolumeClaims() []*core.PersistentVolumeClaim {
	i.lock.Lock()
	defer i.lock.Unlock()

	var r []*core.PersistentVolumeClaim
	for _, persistentVolumeClaim := range i.pvcs {
		r = append(r, persistentVolumeClaim)
	}

	return r
}

func (i *inspector) PersistentVolumeClaim(name string) (*core.PersistentVolumeClaim, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	pvc, ok := i.pvcs[name]
	if !ok {
		return nil, false
	}

	return pvc, true
}

func pvcsToMap(k kubernetes.Interface, namespace string) (map[string]*core.PersistentVolumeClaim, error) {
	pvcs, err := getPersistentVolumeClaims(k, namespace, "")
	if err != nil {
		return nil, err
	}

	pvcMap := map[string]*core.PersistentVolumeClaim{}

	for _, pvc := range pvcs {
		_, exists := pvcMap[pvc.GetName()]
		if exists {
			return nil, errors.Newf("PersistentVolumeClaim %s already exists in map, error received", pvc.GetName())
		}

		pvcMap[pvc.GetName()] = pvcPointer(pvc)
	}

	return pvcMap, nil
}

func pvcPointer(pvc core.PersistentVolumeClaim) *core.PersistentVolumeClaim {
	return &pvc
}

func getPersistentVolumeClaims(k kubernetes.Interface, namespace, cont string) ([]core.PersistentVolumeClaim, error) {
	pvcs, err := k.CoreV1().PersistentVolumeClaims(namespace).List(context.Background(), meta.ListOptions{
		Limit:    128,
		Continue: cont,
	})

	if err != nil {
		return nil, err
	}

	if pvcs.Continue != "" {
		nextPersistentVolumeClaimsLayer, err := getPersistentVolumeClaims(k, namespace, pvcs.Continue)
		if err != nil {
			return nil, err
		}

		return append(pvcs.Items, nextPersistentVolumeClaimsLayer...), nil
	}

	return pvcs.Items, nil
}

func FilterPersistentVolumeClaimsByLabels(labels map[string]string) persistentvolumeclaim.PersistentVolumeClaimFilter {
	return func(pvc *core.PersistentVolumeClaim) bool {
		for key, value := range labels {
			v, ok := pvc.Labels[key]
			if !ok {
				return false
			}

			if v != value {
				return false
			}
		}

		return true
	}
}
