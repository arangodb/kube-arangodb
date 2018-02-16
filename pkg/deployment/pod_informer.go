//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package deployment

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

// listenForPodEvents keep listening for changes in pod until the given channel is closed.
func (d *Deployment) listenForPodEvents() {
	source := cache.NewListWatchFromClient(
		d.deps.KubeCli.CoreV1().RESTClient(),
		"pods",
		d.apiObject.GetNamespace(),
		fields.Everything())

	getPod := func(obj interface{}) (*v1.Pod, bool) {
		pod, ok := obj.(*v1.Pod)
		if !ok {
			tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
			if !ok {
				return nil, false
			}
			pod, ok = tombstone.Obj.(*v1.Pod)
			return pod, ok
		}
		return pod, true
	}

	_, informer := cache.NewIndexerInformer(source, &v1.Pod{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if p, ok := getPod(obj); ok && d.isOwnerOf(p) {
				d.send(&deploymentEvent{
					Type: eventPodAdded,
					Pod:  p,
				})
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if p, ok := getPod(newObj); ok && d.isOwnerOf(p) {
				d.send(&deploymentEvent{
					Type: eventPodUpdated,
					Pod:  p,
				})
			}
		},
		DeleteFunc: func(obj interface{}) {
			if p, ok := getPod(obj); ok && d.isOwnerOf(p) {
				d.send(&deploymentEvent{
					Type: eventPodDeleted,
					Pod:  p,
				})
			}
		},
	}, cache.Indexers{})

	informer.Run(d.stopCh)
}
