//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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

package k8sutil

import (
	"github.com/rs/zerolog"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

// ResourceWatcher is a helper to watch for events in a specific type
// of resource. The handler functions are protected from panics.
type ResourceWatcher struct {
	informer cache.Controller
}

// NewResourceWatcher creates a helper that watches for changes in a resource of a specific type.
// If wraps the given handler functions, such that panics are caught and logged.
func NewResourceWatcher(log zerolog.Logger, getter cache.Getter, resource, namespace string,
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
					log.Error().Interface("error", err).Msg("Recovered from panic")
				}
			}()
			if h.AddFunc != nil {
				h.AddFunc(obj)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			defer func() {
				if err := recover(); err != nil {
					log.Error().Interface("error", err).Msg("Recovered from panic")
				}
			}()
			if h.UpdateFunc != nil {
				h.UpdateFunc(oldObj, newObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
			defer func() {
				if err := recover(); err != nil {
					log.Error().Interface("error", err).Msg("Recovered from panic")
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
