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

package deployment

import (
	core "k8s.io/api/core/v1"
	crdv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// listenForPodEvents keep listening for changes in pod until the given channel is closed.
func (d *Deployment) listenForPodEvents(stopCh <-chan struct{}) {
	getPod := func(obj interface{}) (*core.Pod, bool) {
		pod, ok := obj.(*core.Pod)
		if !ok {
			tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
			if !ok {
				return nil, false
			}
			pod, ok = tombstone.Obj.(*core.Pod)
			return pod, ok
		}
		return pod, true
	}

	rw := k8sutil.NewResourceWatcher(
		d.deps.Client.Kubernetes().CoreV1().RESTClient(),
		"pods",
		d.apiObject.GetNamespace(),
		&core.Pod{},
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				d.acs.CurrentClusterCache().GetThrottles().Pod().Invalidate()
				if p, ok := getPod(obj); ok && d.isOwnerOf(p) {
					d.triggerInspection()
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				d.acs.CurrentClusterCache().GetThrottles().Pod().Invalidate()
				if p, ok := getPod(newObj); ok && d.isOwnerOf(p) {
					d.triggerInspection()
				}
			},
			DeleteFunc: func(obj interface{}) {
				d.acs.CurrentClusterCache().GetThrottles().Pod().Invalidate()
				if p, ok := getPod(obj); ok && d.isOwnerOf(p) {
					d.triggerInspection()
				}
			},
		})

	rw.Run(stopCh)
}

// listenForPVCEvents keep listening for changes in PVC's until the given channel is closed.
func (d *Deployment) listenForPVCEvents(stopCh <-chan struct{}) {
	getPVC := func(obj interface{}) (*core.PersistentVolumeClaim, bool) {
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
		d.deps.Client.Kubernetes().CoreV1().RESTClient(),
		"persistentvolumeclaims",
		d.apiObject.GetNamespace(),
		&core.PersistentVolumeClaim{},
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				d.acs.CurrentClusterCache().GetThrottles().PersistentVolumeClaim().Invalidate()
				if p, ok := getPVC(obj); ok && d.isOwnerOf(p) {
					d.triggerInspection()
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				d.acs.CurrentClusterCache().GetThrottles().PersistentVolumeClaim().Invalidate()
				if p, ok := getPVC(newObj); ok && d.isOwnerOf(p) {
					d.triggerInspection()
				}
			},
			DeleteFunc: func(obj interface{}) {
				d.acs.CurrentClusterCache().GetThrottles().PersistentVolumeClaim().Invalidate()
				if p, ok := getPVC(obj); ok && d.isOwnerOf(p) {
					d.triggerInspection()
				}
			},
		})

	rw.Run(stopCh)
}

// listenForSecretEvents keep listening for changes in Secrets's until the given channel is closed.
func (d *Deployment) listenForSecretEvents(stopCh <-chan struct{}) {
	getSecret := func(obj interface{}) bool {
		_, ok := obj.(*core.Secret)
		if !ok {
			tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
			if !ok {
				return false
			}
			_, ok = tombstone.Obj.(*core.Secret)
			return ok
		}
		return true
	}

	rw := k8sutil.NewResourceWatcher(
		d.deps.Client.Kubernetes().CoreV1().RESTClient(),
		"secrets",
		d.apiObject.GetNamespace(),
		&core.Secret{},
		cache.ResourceEventHandlerFuncs{
			// Note: For secrets we look at all of them because they do not have to be owned by this deployment.
			AddFunc: func(obj interface{}) {
				d.acs.CurrentClusterCache().GetThrottles().Secret().Invalidate()
				if getSecret(obj) {
					d.triggerInspection()
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				d.acs.CurrentClusterCache().GetThrottles().Secret().Invalidate()
				if getSecret(newObj) {
					d.triggerInspection()
				}
			},
			DeleteFunc: func(obj interface{}) {
				d.acs.CurrentClusterCache().GetThrottles().Secret().Invalidate()
				if getSecret(obj) {
					d.triggerInspection()
				}
			},
		})

	rw.Run(stopCh)
}

// listenForServiceEvents keep listening for changes in Service's until the given channel is closed.
func (d *Deployment) listenForServiceEvents(stopCh <-chan struct{}) {
	getService := func(obj interface{}) (*core.Service, bool) {
		service, ok := obj.(*core.Service)
		if !ok {
			tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
			if !ok {
				return nil, false
			}
			service, ok = tombstone.Obj.(*core.Service)
			return service, ok
		}
		return service, true
	}

	rw := k8sutil.NewResourceWatcher(
		d.deps.Client.Kubernetes().CoreV1().RESTClient(),
		"services",
		d.apiObject.GetNamespace(),
		&core.Service{},
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				d.acs.CurrentClusterCache().GetThrottles().Service().Invalidate()
				if s, ok := getService(obj); ok && d.isOwnerOf(s) {
					d.triggerInspection()
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				d.acs.CurrentClusterCache().GetThrottles().Service().Invalidate()
				if s, ok := getService(newObj); ok && d.isOwnerOf(s) {
					d.triggerInspection()
				}
			},
			DeleteFunc: func(obj interface{}) {
				d.acs.CurrentClusterCache().GetThrottles().Service().Invalidate()
				if s, ok := getService(obj); ok && d.isOwnerOf(s) {
					d.triggerInspection()
				}
			},
		})

	rw.Run(stopCh)
}

// listenForCRDEvents keep listening for changes in CRDs until the given channel is closed.
func (d *Deployment) listenForCRDEvents(stopCh <-chan struct{}) {
	rw := k8sutil.NewResourceWatcher(
		d.deps.Client.KubernetesExtensions().ApiextensionsV1().RESTClient(),
		"customresourcedefinitions",
		"",
		&crdv1.CustomResourceDefinition{},
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				d.triggerCRDInspection()
			},
			DeleteFunc: func(obj interface{}) {
				d.triggerCRDInspection()
			},
		})

	rw.Run(stopCh)
}
