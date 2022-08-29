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

package operator

import (
	kwatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	api "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/storage"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

var (
	localStoragesCreated  = metrics.MustRegisterCounter("controller", "local_storages_created", "Number of local storages that have been created")
	localStoragesDeleted  = metrics.MustRegisterCounter("controller", "local_storages_deleted", "Number of local storages that have been deleted")
	localStoragesFailed   = metrics.MustRegisterCounter("controller", "local_storages_failed", "Number of local storages that have failed")
	localStoragesModified = metrics.MustRegisterCounter("controller", "local_storages_modified", "Number of local storage modifications")
	localStoragesCurrent  = metrics.MustRegisterGauge("controller", "local_storages", "Number of local storages currently being managed")
)

// run the local storages part of the operator.
// This registers a listener and waits until the process stops.
func (o *Operator) runLocalStorages(stop <-chan struct{}) {
	rw := k8sutil.NewResourceWatcher(
		o.Dependencies.Client.Arango().StorageV1alpha().RESTClient(),
		api.ArangoLocalStorageResourcePlural,
		"", //o.Config.Namespace,
		&api.ArangoLocalStorage{},
		cache.ResourceEventHandlerFuncs{
			AddFunc:    o.onAddArangoLocalStorage,
			UpdateFunc: o.onUpdateArangoLocalStorage,
			DeleteFunc: o.onDeleteArangoLocalStorage,
		})

	o.Dependencies.StorageProbe.SetReady()
	rw.Run(stop)
}

// onAddArangoLocalStorage local storage addition callback
func (o *Operator) onAddArangoLocalStorage(obj interface{}) {
	o.Dependencies.LivenessProbe.Lock()
	defer o.Dependencies.LivenessProbe.Unlock()

	apiObject := obj.(*api.ArangoLocalStorage)
	o.log.
		Str("name", apiObject.GetObjectMeta().GetName()).
		Debug("ArangoLocalStorage added")
	o.syncArangoLocalStorage(apiObject)
}

// onUpdateArangoLocalStorage local storage update callback
func (o *Operator) onUpdateArangoLocalStorage(oldObj, newObj interface{}) {
	o.Dependencies.LivenessProbe.Lock()
	defer o.Dependencies.LivenessProbe.Unlock()

	apiObject := newObj.(*api.ArangoLocalStorage)
	o.log.
		Str("name", apiObject.GetObjectMeta().GetName()).
		Debug("ArangoLocalStorage updated")
	o.syncArangoLocalStorage(apiObject)
}

// onDeleteArangoLocalStorage local storage delete callback
func (o *Operator) onDeleteArangoLocalStorage(obj interface{}) {
	o.Dependencies.LivenessProbe.Lock()
	defer o.Dependencies.LivenessProbe.Unlock()

	log := o.log
	apiObject, ok := obj.(*api.ArangoLocalStorage)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			log.Interface("event-object", obj).Error("unknown object from ArangoLocalStorage delete event")
			return
		}
		apiObject, ok = tombstone.Obj.(*api.ArangoLocalStorage)
		if !ok {
			log.Interface("event-object", obj).Error("Tombstone contained object that is not an ArangoLocalStorage")
			return
		}
	}
	log.
		Str("name", apiObject.GetObjectMeta().GetName()).
		Debug("ArangoLocalStorage deleted")
	ev := &Event{
		Type:         kwatch.Deleted,
		LocalStorage: apiObject,
	}

	//	pt.start()
	err := o.handleLocalStorageEvent(ev)
	if err != nil {
		log.Err(err).Warn("Failed to handle event")
	}
	//pt.stop()
}

// syncArangoLocalStorage synchronizes the given local storage.
func (o *Operator) syncArangoLocalStorage(apiObject *api.ArangoLocalStorage) {
	ev := &Event{
		Type:         kwatch.Added,
		LocalStorage: apiObject,
	}
	// re-watch or restart could give ADD event.
	// If for an ADD event the cluster spec is invalid then it is not added to the local cache
	// so modifying that local storage will result in another ADD event
	if _, ok := o.localStorages[apiObject.Name]; ok {
		ev.Type = kwatch.Modified
	}

	//pt.start()
	err := o.handleLocalStorageEvent(ev)
	if err != nil {
		o.log.Err(err).Warn("Failed to handle event")
	}
	//pt.stop()
}

// handleLocalStorageEvent processed the given event.
func (o *Operator) handleLocalStorageEvent(event *Event) error {
	apiObject := event.LocalStorage

	if apiObject.Status.State.IsFailed() {
		localStoragesFailed.Inc()
		if event.Type == kwatch.Deleted {
			delete(o.localStorages, apiObject.Name)
			return nil
		}
		return errors.WithStack(errors.Newf("ignore failed local storage (%s). Please delete its CR", apiObject.Name))
	}

	// Fill in defaults
	apiObject.Spec.SetDefaults(apiObject.GetName())
	// Validate local storage spec
	if err := apiObject.Spec.Validate(); err != nil {
		return errors.WithStack(errors.Wrapf(err, "invalid local storage spec. please fix the following problem with the local storage spec: %v", err))
	}

	switch event.Type {
	case kwatch.Added:
		if _, ok := o.localStorages[apiObject.Name]; ok {
			return errors.WithStack(errors.Newf("unsafe state. local storage (%s) was created before but we received event (%s)", apiObject.Name, event.Type))
		}

		cfg, deps := o.makeLocalStorageConfigAndDeps()
		stg, err := storage.New(cfg, deps, apiObject)
		if err != nil {
			return errors.WithStack(errors.Newf("failed to create local storage: %s", err))
		}
		o.localStorages[apiObject.Name] = stg

		localStoragesCreated.Inc()
		localStoragesCurrent.Set(float64(len(o.localStorages)))

	case kwatch.Modified:
		stg, ok := o.localStorages[apiObject.Name]
		if !ok {
			return errors.WithStack(errors.Newf("unsafe state. local storage (%s) was never created but we received event (%s)", apiObject.Name, event.Type))
		}
		stg.Update(apiObject)
		localStoragesModified.Inc()

	case kwatch.Deleted:
		stg, ok := o.localStorages[apiObject.Name]
		if !ok {
			return errors.WithStack(errors.Newf("unsafe state. local storage (%s) was never created but we received event (%s)", apiObject.Name, event.Type))
		}
		stg.Delete()
		delete(o.localStorages, apiObject.Name)
		localStoragesDeleted.Inc()
		localStoragesCurrent.Set(float64(len(o.localStorages)))
	}
	return nil
}

// makeLocalStorageConfigAndDeps creates a Config & Dependencies object for a new LocalStorage.
func (o *Operator) makeLocalStorageConfigAndDeps() (storage.Config, storage.Dependencies) {
	cfg := storage.Config{
		Namespace:      o.Config.Namespace,
		PodName:        o.Config.PodName,
		ServiceAccount: o.Config.ServiceAccount,
	}
	deps := storage.Dependencies{
		Client:        o.Client,
		EventRecorder: o.Dependencies.EventRecorder,
	}
	return cfg, deps
}
