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

package k8sutil

import (
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"

	"github.com/arangodb/kube-arangodb/pkg/logging"
)

var (
	informerLogger = logging.Global().Get("kubernetes-informer")
)

// ResourceWatcher is a helper to watch for events in a specific type
// of resource. The handler functions are protected from panics.
type ResourceWatcher struct {
	informer cache.Controller
}

// NewResourceWatcher creates a helper that watches for changes in a resource of a specific type.
// If wraps the given handler functions, such that panics are caught and logged.
func NewResourceWatcher(getter cache.Getter, resource, namespace string,
	objType runtime.Object, h cache.ResourceEventHandlerFuncs) *ResourceWatcher {
	source := cache.NewListWatchFromClient(
		getter,
		resource,
		namespace,
		fields.Everything())

	_, informer := cache.NewIndexerInformer(source, objType, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			defer func() {
				if err := recover(); err != nil {
					informerLogger.Interface("error", err).Error("Recovered from panic")
				}
			}()
			if h.AddFunc != nil {
				h.AddFunc(obj)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			defer func() {
				if err := recover(); err != nil {
					informerLogger.Interface("error", err).Error("Recovered from panic")
				}
			}()
			if h.UpdateFunc != nil {
				h.UpdateFunc(oldObj, newObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
			defer func() {
				if err := recover(); err != nil {
					informerLogger.Interface("error", err).Error("Recovered from panic")
				}
			}()
			if h.DeleteFunc != nil {
				h.DeleteFunc(obj)
			}
		},
	}, cache.Indexers{})

	return &ResourceWatcher{
		informer: informer,
	}
}

// Run continues to watch for events on the selected type of resource
// until the given channel is closed.
func (rw *ResourceWatcher) Run(stopCh <-chan struct{}) {
	rw.informer.Run(stopCh)
}
