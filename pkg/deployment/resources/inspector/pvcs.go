//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Tomasz Mielech
//

package inspector

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util/globals"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (i *inspector) IteratePersistentVolumeClaims(action persistentvolumeclaim.Action, filters ...persistentvolumeclaim.Filter) error {
	for _, pvc := range i.PersistentVolumeClaims() {
		if err := i.iteratePersistentVolumeClaim(pvc, action, filters...); err != nil {
			return err
		}
	}
	return nil
}

func (i *inspector) iteratePersistentVolumeClaim(pvc *core.PersistentVolumeClaim, action persistentvolumeclaim.Action, filters ...persistentvolumeclaim.Filter) error {
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

func (i *inspector) PersistentVolumeClaimReadInterface() persistentvolumeclaim.ReadInterface {
	return &persistentVolumeClaimReadInterface{i: i}
}

type persistentVolumeClaimReadInterface struct {
	i *inspector
}

func (s persistentVolumeClaimReadInterface) Get(ctx context.Context, name string, opts meta.GetOptions) (*core.PersistentVolumeClaim, error) {
	if s, ok := s.i.PersistentVolumeClaim(name); !ok {
		return nil, apiErrors.NewNotFound(schema.GroupResource{
			Group:    core.GroupName,
			Resource: "persistentvolumeclaims",
		}, name)
	} else {
		return s, nil
	}
}

func pvcsToMap(ctx context.Context, inspector *inspector, k kubernetes.Interface, namespace string) func() error {
	return func() error {
		pvcs, err := getPersistentVolumeClaims(ctx, k, namespace, "")
		if err != nil {
			return err
		}

		pvcMap := map[string]*core.PersistentVolumeClaim{}

		for _, pvc := range pvcs {
			_, exists := pvcMap[pvc.GetName()]
			if exists {
				return errors.Newf("PersistentVolumeClaim %s already exists in map, error received", pvc.GetName())
			}

			pvcMap[pvc.GetName()] = pvcPointer(pvc)
		}

		inspector.pvcs = pvcMap

		return nil
	}
}

func pvcPointer(pvc core.PersistentVolumeClaim) *core.PersistentVolumeClaim {
	return &pvc
}

func getPersistentVolumeClaims(ctx context.Context, k kubernetes.Interface, namespace, cont string) ([]core.PersistentVolumeClaim, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	pvcs, err := k.CoreV1().PersistentVolumeClaims(namespace).List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, err
	}

	if pvcs.Continue != "" {
		nextPersistentVolumeClaimsLayer, err := getPersistentVolumeClaims(ctx, k, namespace, pvcs.Continue)
		if err != nil {
			return nil, err
		}

		return append(pvcs.Items, nextPersistentVolumeClaimsLayer...), nil
	}

	return pvcs.Items, nil
}

func FilterPersistentVolumeClaimsByLabels(labels map[string]string) persistentvolumeclaim.Filter {
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
