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
	"context"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	api "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
	"github.com/arangodb/kube-arangodb/pkg/util/trigger"
)

// Config holds configuration settings for a LocalStorage
type Config struct {
	Namespace      string
	PodName        string
	ServiceAccount string
}

// Dependencies holds dependent services for a LocalStorage
type Dependencies struct {
	Log           zerolog.Logger
	Client        kclient.Client
	EventRecorder record.EventRecorder
}

// localStorageEvent strongly typed type of event
type localStorageEventType string

const (
	eventArangoLocalStorageUpdated localStorageEventType = "ArangoLocalStorageUpdated"
	eventPVCAdded                  localStorageEventType = "pvcAdded"
	eventPVCUpdated                localStorageEventType = "pvcUpdated"
	eventPVUpdated                 localStorageEventType = "pvUpdated"
)

// localStorageEvent holds an event passed from the controller to the local storage.
type localStorageEvent struct {
	Type                  localStorageEventType
	LocalStorage          *api.ArangoLocalStorage
	PersistentVolume      *v1.PersistentVolume
	PersistentVolumeClaim *v1.PersistentVolumeClaim
}

const (
	localStorageEventQueueSize = 100
	minInspectionInterval      = time.Second // Ensure we inspect the generated resources no less than with this interval
	maxInspectionInterval      = time.Minute // Ensure we inspect the generated resources no less than with this interval
)

// LocalStorage is the in process state of an ArangoLocalStorage.
type LocalStorage struct {
	apiObject *api.ArangoLocalStorage // API object
	status    api.LocalStorageStatus  // Internal status of the CR
	config    Config
	deps      Dependencies

	eventCh chan *localStorageEvent
	stopCh  chan struct{}
	stopped int32

	image           string
	imagePullPolicy v1.PullPolicy
	inspectTrigger  trigger.Trigger
	pvCleaner       *pvCleaner
}

// New creates a new LocalStorage from the given API object.
func New(config Config, deps Dependencies, apiObject *api.ArangoLocalStorage) (*LocalStorage, error) {
	if err := apiObject.Spec.Validate(); err != nil {
		return nil, errors.WithStack(err)
	}
	ls := &LocalStorage{
		apiObject: apiObject,
		status:    *(apiObject.Status.DeepCopy()),
		config:    config,
		deps:      deps,
		eventCh:   make(chan *localStorageEvent, localStorageEventQueueSize),
		stopCh:    make(chan struct{}),
	}

	ls.pvCleaner = newPVCleaner(deps.Log, deps.Client.Kubernetes(), ls.GetClientByNodeName)

	go ls.run()
	go ls.listenForPvcEvents()
	go ls.listenForPvEvents()
	go ls.pvCleaner.Run(ls.stopCh)

	return ls, nil
}

// Update the local storage.
// This sends an update event in the event queue.
func (ls *LocalStorage) Update(apiObject *api.ArangoLocalStorage) {
	ls.send(&localStorageEvent{
		Type:         eventArangoLocalStorageUpdated,
		LocalStorage: apiObject,
	})
}

// Delete the local storage.
// Called when the local storage was deleted by the user.
func (ls *LocalStorage) Delete() {
	ls.deps.Log.Info().Msg("local storage is deleted by user")
	if atomic.CompareAndSwapInt32(&ls.stopped, 0, 1) {
		close(ls.stopCh)
	}
}

// send given event into the local storage event queue.
func (ls *LocalStorage) send(ev *localStorageEvent) {
	select {
	case ls.eventCh <- ev:
		l, ecap := len(ls.eventCh), cap(ls.eventCh)
		if l > int(float64(ecap)*0.8) {
			ls.deps.Log.Warn().
				Int("used", l).
				Int("capacity", ecap).
				Msg("event queue buffer is almost full")
		}
	case <-ls.stopCh:
	}
}

// run is the core the core worker.
// It processes the event queue and polls the state of generated
// resource on a regular basis.
func (ls *LocalStorage) run() {
	//log := ls.deps.Log

	// Find out my image
	image, pullPolicy, err := ls.getMyImage()
	if err != nil {
		ls.failOnError(err, "Failed to get my own image")
		return
	}
	ls.image = image
	ls.imagePullPolicy = pullPolicy

	// Set state
	if ls.status.State == api.LocalStorageStateNone {
		ls.status.State = api.LocalStorageStateCreating
		if err := ls.updateCRStatus(); err != nil {
			ls.createEvent(k8sutil.NewErrorEvent("Failed to update LocalStorage state", err, ls.apiObject))
		}
	}

	// Create StorageClass
	if err := ls.ensureStorageClass(ls.apiObject); err != nil {
		ls.failOnError(err, "Failed to create storage class")
		return
	}

	// Create DaemonSet
	if err := ls.ensureDaemonSet(ls.apiObject); err != nil {
		ls.failOnError(err, "Failed to create daemon set")
		return
	}

	// Create Service to access provisioners
	if err := ls.ensureProvisionerService(ls.apiObject); err != nil {
		ls.failOnError(err, "Failed to create service")
		return
	}

	inspectionInterval := maxInspectionInterval
	recentInspectionErrors := 0
	var pvsNeededSince *time.Time
	for {
		select {
		case <-ls.stopCh:
			// We're being stopped.
			return

		case event := <-ls.eventCh:
			// Got event from event queue
			switch event.Type {
			case eventArangoLocalStorageUpdated:
				if err := ls.handleArangoLocalStorageUpdatedEvent(event); err != nil {
					ls.failOnError(err, "Failed to handle local storage update")
					return
				}
			case eventPVCAdded, eventPVCUpdated, eventPVUpdated:
				// Do an inspection of PVC's
				ls.inspectTrigger.Trigger()
			default:
				panic("unknown event type" + event.Type)
			}

		case <-ls.inspectTrigger.Done():
			hasError := false
			unboundPVCs, err := ls.inspectPVCs()
			if err != nil {
				hasError = true
				ls.createEvent(k8sutil.NewErrorEvent("PVC inspection failed", err, ls.apiObject))
			}
			pvsAvailable, err := ls.inspectPVs()
			if err != nil {
				hasError = true
				ls.createEvent(k8sutil.NewErrorEvent("PV inspection failed", err, ls.apiObject))
			}
			if len(unboundPVCs) == 0 {
				pvsNeededSince = nil
			} else if len(unboundPVCs) > 0 {
				createNow := false
				if pvsNeededSince != nil && time.Since(*pvsNeededSince) > time.Second*30 {
					// Create now
					createNow = true
				} else if pvsAvailable < len(unboundPVCs) {
					// Create now
					createNow = true
				} else {
					// Volumes are there, just may no be a match.
					// Wait for that
					if pvsNeededSince == nil {
						now := time.Now()
						pvsNeededSince = &now
					}
				}
				if createNow {
					ctx := context.Background()
					if err := ls.createPVs(ctx, ls.apiObject, unboundPVCs); err != nil {
						hasError = true
						ls.createEvent(k8sutil.NewErrorEvent("PV creation failed", err, ls.apiObject))
					}
				}
			}
			if hasError {
				if recentInspectionErrors == 0 {
					inspectionInterval = minInspectionInterval
					recentInspectionErrors++
				}
			} else {
				if ls.status.State == api.LocalStorageStateCreating || ls.status.State == api.LocalStorageStateNone {
					ls.status.State = api.LocalStorageStateRunning
					if err := ls.updateCRStatus(); err != nil {
						hasError = true
						ls.createEvent(k8sutil.NewErrorEvent("Failed to update LocalStorage state", err, ls.apiObject))
					}
				}
				recentInspectionErrors = 0
			}

		case <-time.After(inspectionInterval):
			// Trigger inspection
			ls.inspectTrigger.Trigger()
			// Backoff with next interval
			inspectionInterval = time.Duration(float64(inspectionInterval) * 1.5)
			if inspectionInterval > maxInspectionInterval {
				inspectionInterval = maxInspectionInterval
			}
		}
	}
}

// handleArangoLocalStorageUpdatedEvent is called when the local storage is updated by the user.
func (ls *LocalStorage) handleArangoLocalStorageUpdatedEvent(event *localStorageEvent) error {
	log := ls.deps.Log.With().Str("localStorage", event.LocalStorage.GetName()).Logger()

	// Get the most recent version of the local storage from the API server
	current, err := ls.deps.Client.Arango().StorageV1alpha().ArangoLocalStorages().Get(context.Background(), ls.apiObject.GetName(), metav1.GetOptions{})
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get current version of local storage from API server")
		if k8sutil.IsNotFound(err) {
			return nil
		}
		return errors.WithStack(err)
	}

	newAPIObject := current.DeepCopy()
	newAPIObject.Spec.SetDefaults(newAPIObject.GetName())
	newAPIObject.Status = ls.status
	resetFields := ls.apiObject.Spec.ResetImmutableFields(&newAPIObject.Spec)
	if len(resetFields) > 0 {
		log.Debug().Strs("fields", resetFields).Msg("Found modified immutable fields")
	}
	if err := newAPIObject.Spec.Validate(); err != nil {
		ls.createEvent(k8sutil.NewErrorEvent("Validation failed", err, ls.apiObject))
		// Try to reset object
		if err := ls.updateCRSpec(ls.apiObject.Spec); err != nil {
			log.Error().Err(err).Msg("Restore original spec failed")
			ls.createEvent(k8sutil.NewErrorEvent("Restore original failed", err, ls.apiObject))
		}
		return nil
	}
	if len(resetFields) > 0 {
		for _, fieldName := range resetFields {
			log.Debug().Str("field", fieldName).Msg("Reset immutable field")
			ls.createEvent(k8sutil.NewImmutableFieldEvent(fieldName, ls.apiObject))
		}
	}

	// Save updated spec
	if err := ls.updateCRSpec(newAPIObject.Spec); err != nil {
		return errors.WithStack(errors.Newf("failed to update ArangoLocalStorage spec: %v", err))
	}

	// Trigger inspect
	ls.inspectTrigger.Trigger()

	return nil
}

// createEvent creates a given event.
// On error, the error is logged.
func (ls *LocalStorage) createEvent(evt *k8sutil.Event) {
	ls.deps.EventRecorder.Event(evt.InvolvedObject, evt.Type, evt.Reason, evt.Message)
}

// Update the status of the API object from the internal status
func (ls *LocalStorage) updateCRStatus() error {
	if reflect.DeepEqual(ls.apiObject.Status, ls.status) {
		// Nothing has changed
		return nil
	}

	// Send update to API server
	update := ls.apiObject.DeepCopy()
	attempt := 0
	for {
		attempt++
		update.Status = ls.status
		newAPIObject, err := ls.deps.Client.Arango().StorageV1alpha().ArangoLocalStorages().Update(context.Background(), update, metav1.UpdateOptions{})
		if err == nil {
			// Update internal object
			ls.apiObject = newAPIObject
			return nil
		}
		if attempt < 10 && k8sutil.IsConflict(err) {
			// API object may have been changed already,
			// Reload api object and try again
			var current *api.ArangoLocalStorage
			current, err = ls.deps.Client.Arango().StorageV1alpha().ArangoLocalStorages().Get(context.Background(), update.GetName(), metav1.GetOptions{})
			if err == nil {
				update = current.DeepCopy()
				continue
			}
		}
		if err != nil {
			ls.deps.Log.Debug().Err(err).Msg("failed to patch ArangoLocalStorage status")
			return errors.WithStack(errors.Newf("failed to patch ArangoLocalStorage status: %v", err))
		}
	}
}

// Update the spec part of the API object (d.apiObject)
// to the given object, while preserving the status.
// On success, d.apiObject is updated.
func (ls *LocalStorage) updateCRSpec(newSpec api.LocalStorageSpec) error {
	// Send update to API server
	update := ls.apiObject.DeepCopy()
	attempt := 0
	for {
		attempt++
		update.Spec = newSpec
		update.Status = ls.status
		newAPIObject, err := ls.deps.Client.Arango().StorageV1alpha().ArangoLocalStorages().Update(context.Background(), update, metav1.UpdateOptions{})
		if err == nil {
			// Update internal object
			ls.apiObject = newAPIObject
			return nil
		}
		if attempt < 10 && k8sutil.IsConflict(err) {
			// API object may have been changed already,
			// Reload api object and try again
			var current *api.ArangoLocalStorage
			current, err = ls.deps.Client.Arango().StorageV1alpha().ArangoLocalStorages().Get(context.Background(), update.GetName(), metav1.GetOptions{})
			if err == nil {
				update = current.DeepCopy()
				continue
			}
		}
		if err != nil {
			ls.deps.Log.Debug().Err(err).Msg("failed to patch ArangoLocalStorage spec")
			return errors.WithStack(errors.Newf("failed to patch ArangoLocalStorage spec: %v", err))
		}
	}
}

// failOnError reports the given error and sets the local storage status to failed.
func (ls *LocalStorage) failOnError(err error, msg string) {
	log.Error().Err(err).Msg(msg)
	ls.status.Reason = err.Error()
	ls.reportFailedStatus()
}

// reportFailedStatus sets the status of the local storage to Failed and keeps trying to forward
// that to the API server.
func (ls *LocalStorage) reportFailedStatus() {
	log := ls.deps.Log
	log.Info().Msg("local storage failed. Reporting failed reason...")

	op := func() error {
		ls.status.State = api.LocalStorageStateFailed
		err := ls.updateCRStatus()
		if err == nil || k8sutil.IsNotFound(err) {
			// Status has been updated
			return nil
		}

		if !k8sutil.IsConflict(err) {
			log.Warn().Err(err).Msg("retry report status: fail to update")
			return errors.WithStack(err)
		}

		depl, err := ls.deps.Client.Arango().StorageV1alpha().ArangoLocalStorages().Get(context.Background(), ls.apiObject.Name, metav1.GetOptions{})
		if err != nil {
			// Update (PUT) will return conflict even if object is deleted since we have UID set in object.
			// Because it will check UID first and return something like:
			// "Precondition failed: UID in precondition: 0xc42712c0f0, UID in object meta: ".
			if k8sutil.IsNotFound(err) {
				return nil
			}
			log.Warn().Err(err).Msg("retry report status: fail to get latest version")
			return errors.WithStack(err)
		}
		ls.apiObject = depl
		return errors.WithStack(errors.Newf("retry needed"))
	}

	retry.Retry(op, time.Hour*24*365)
}

// isOwnerOf returns true if the given object belong to this local storage.
func (ls *LocalStorage) isOwnerOf(obj metav1.Object) bool {
	ownerRefs := obj.GetOwnerReferences()
	if len(ownerRefs) < 1 {
		return false
	}
	return ownerRefs[0].UID == ls.apiObject.UID
}
