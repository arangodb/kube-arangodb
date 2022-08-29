//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package storage

import (
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// listenForPvcEvents keep listening for changes in PVC's until the given channel is closed.
func (ls *LocalStorage) listenForPvcEvents() {
	getPvc := func(obj interface{}) (*core.PersistentVolumeClaim, bool) {
		pvc, ok := obj.(*core.PersistentVolumeClaim)
		if !ok {
			tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
			if !ok {
				return nil, false
			}
			pvc, ok = tombstone.Obj.(*core.PersistentVolumeClaim)
			return pvc, ok
		}
		return pvc, true
	}

	rw := k8sutil.NewResourceWatcher(
		ls.deps.Client.Kubernetes().CoreV1().RESTClient(),
		"persistentvolumeclaims",
		"", //ls.apiObject.GetNamespace(),
		&core.PersistentVolumeClaim{},
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if pvc, ok := getPvc(obj); ok {
					ls.send(&localStorageEvent{
						Type:                  eventPVCAdded,
						PersistentVolumeClaim: pvc,
					})
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				if pvc, ok := getPvc(newObj); ok {
					ls.send(&localStorageEvent{
						Type:                  eventPVCUpdated,
						PersistentVolumeClaim: pvc,
					})
				}
			},
			DeleteFunc: func(obj interface{}) {
				// Ignore
			},
		})

	rw.Run(ls.stopCh)
}
