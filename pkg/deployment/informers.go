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
func (d *Deployment) listenForPodEvents(stopCh <-chan struct{}) {
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
				d.triggerInspection()
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if p, ok := getPod(newObj); ok && d.isOwnerOf(p) {
				d.triggerInspection()
			}
		},
		DeleteFunc: func(obj interface{}) {
			if p, ok := getPod(obj); ok && d.isOwnerOf(p) {
				d.triggerInspection()
			}
		},
	}, cache.Indexers{})

	informer.Run(stopCh)
}

// listenForPVCEvents keep listening for changes in PVC's until the given channel is closed.
func (d *Deployment) listenForPVCEvents(stopCh <-chan struct{}) {
	source := cache.NewListWatchFromClient(
		d.deps.KubeCli.CoreV1().RESTClient(),
		"persistentvolumeclaims",
		d.apiObject.GetNamespace(),
		fields.Everything())

	getPVC := func(obj interface{}) (*v1.PersistentVolumeClaim, bool) {
		pvc, ok := obj.(*v1.PersistentVolumeClaim)
		if !ok {
			tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
			if !ok {
				return nil, false
			}
			pvc, ok = tombstone.Obj.(*v1.PersistentVolumeClaim)
			return pvc, ok
		}
		return pvc, true
	}

	_, informer := cache.NewIndexerInformer(source, &v1.PersistentVolumeClaim{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if p, ok := getPVC(obj); ok && d.isOwnerOf(p) {
				d.triggerInspection()
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if p, ok := getPVC(newObj); ok && d.isOwnerOf(p) {
				d.triggerInspection()
			}
		},
		DeleteFunc: func(obj interface{}) {
			if p, ok := getPVC(obj); ok && d.isOwnerOf(p) {
				d.triggerInspection()
			}
		},
	}, cache.Indexers{})

	informer.Run(stopCh)
}

// listenForServiceEvents keep listening for changes in Service's until the given channel is closed.
func (d *Deployment) listenForServiceEvents(stopCh <-chan struct{}) {
	source := cache.NewListWatchFromClient(
		d.deps.KubeCli.CoreV1().RESTClient(),
		"services",
		d.apiObject.GetNamespace(),
		fields.Everything())

	getService := func(obj interface{}) (*v1.Service, bool) {
		service, ok := obj.(*v1.Service)
		if !ok {
			tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
			if !ok {
				return nil, false
			}
			service, ok = tombstone.Obj.(*v1.Service)
			return service, ok
		}
		return service, true
	}

	_, informer := cache.NewIndexerInformer(source, &v1.PersistentVolumeClaim{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if s, ok := getService(obj); ok && d.isOwnerOf(s) {
				d.triggerInspection()
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if s, ok := getService(newObj); ok && d.isOwnerOf(s) {
				d.triggerInspection()
			}
		},
		DeleteFunc: func(obj interface{}) {
			if s, ok := getService(obj); ok && d.isOwnerOf(s) {
				d.triggerInspection()
			}
		},
	}, cache.Indexers{})

	informer.Run(stopCh)
}
