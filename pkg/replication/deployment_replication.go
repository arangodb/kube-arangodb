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

package replication

import (
	"context"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	"github.com/arangodb/arangosync-client/client"
	api "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
	"github.com/arangodb/kube-arangodb/pkg/util/trigger"
)

// Config holds configuration settings for a DeploymentReplication
type Config struct {
	Namespace string
}

// Dependencies holds dependent services for a DeploymentReplication
type Dependencies struct {
	Log           zerolog.Logger
	Client        kclient.Client
	EventRecorder record.EventRecorder
}

// deploymentReplicationEvent strongly typed type of event
type deploymentReplicationEventType string

const (
	eventArangoDeploymentReplicationUpdated deploymentReplicationEventType = "DeploymentReplicationUpdated"
)

// seploymentReplicationEvent holds an event passed from the controller to the deployment replication.
type deploymentReplicationEvent struct {
	Type                  deploymentReplicationEventType
	DeploymentReplication *api.ArangoDeploymentReplication
}

const (
	deploymentReplicationEventQueueSize = 100
	minInspectionInterval               = time.Second // Ensure we inspect the generated resources no less than with this interval
	maxInspectionInterval               = time.Minute // Ensure we inspect the generated resources no less than with this interval
)

// DeploymentReplication is the in process state of an ArangoDeploymentReplication.
type DeploymentReplication struct {
	apiObject *api.ArangoDeploymentReplication // API object
	status    api.DeploymentReplicationStatus  // Internal status of the CR
	config    Config
	deps      Dependencies

	eventCh chan *deploymentReplicationEvent
	stopCh  chan struct{}
	stopped int32

	inspectTrigger         trigger.Trigger
	recentInspectionErrors int
	clientCache            client.ClientCache
}

// New creates a new DeploymentReplication from the given API object.
func New(config Config, deps Dependencies, apiObject *api.ArangoDeploymentReplication) (*DeploymentReplication, error) {
	if err := apiObject.Spec.Validate(); err != nil {
		return nil, errors.WithStack(err)
	}
	dr := &DeploymentReplication{
		apiObject: apiObject,
		status:    *(apiObject.Status.DeepCopy()),
		config:    config,
		deps:      deps,
		eventCh:   make(chan *deploymentReplicationEvent, deploymentReplicationEventQueueSize),
		stopCh:    make(chan struct{}),
	}

	go dr.run()

	return dr, nil
}

// Update the deployment replication.
// This sends an update event in the event queue.
func (dr *DeploymentReplication) Update(apiObject *api.ArangoDeploymentReplication) {
	dr.send(&deploymentReplicationEvent{
		Type:                  eventArangoDeploymentReplicationUpdated,
		DeploymentReplication: apiObject,
	})
}

// Delete the deployment replication.
// Called when the local storage was deleted by the user.
func (dr *DeploymentReplication) Delete() {
	dr.deps.Log.Info().Msg("deployment replication is deleted by user")
	if atomic.CompareAndSwapInt32(&dr.stopped, 0, 1) {
		close(dr.stopCh)
	}
}

// send given event into the deployment replication event queue.
func (dr *DeploymentReplication) send(ev *deploymentReplicationEvent) {
	select {
	case dr.eventCh <- ev:
		l, ecap := len(dr.eventCh), cap(dr.eventCh)
		if l > int(float64(ecap)*0.8) {
			dr.deps.Log.Warn().
				Int("used", l).
				Int("capacity", ecap).
				Msg("event queue buffer is almost full")
		}
	case <-dr.stopCh:
	}
}

// run is the core the core worker.
// It processes the event queue and polls the state of generated
// resource on a regular basis.
func (dr *DeploymentReplication) run() {
	log := dr.deps.Log

	// Add finalizers
	if err := dr.addFinalizers(); err != nil {
		log.Warn().Err(err).Msg("Failed to add finalizers")
	}

	inspectionInterval := maxInspectionInterval
	dr.inspectTrigger.Trigger()
	for {
		select {
		case <-dr.stopCh:
			// We're being stopped.
			return

		case event := <-dr.eventCh:
			// Got event from event queue
			switch event.Type {
			case eventArangoDeploymentReplicationUpdated:
				if err := dr.handleArangoDeploymentReplicationUpdatedEvent(event); err != nil {
					dr.failOnError(err, "Failed to handle deployment replication update")
					return
				}
			default:
				panic("unknown event type" + event.Type)
			}

		case <-dr.inspectTrigger.Done():
			inspectionInterval = dr.inspectDeploymentReplication(inspectionInterval)

		case <-time.After(inspectionInterval):
			// Trigger inspection
			dr.inspectTrigger.Trigger()
			// Backoff with next interval
			inspectionInterval = time.Duration(float64(inspectionInterval) * 1.5)
			if inspectionInterval > maxInspectionInterval {
				inspectionInterval = maxInspectionInterval
			}
		}
	}
}

// handleArangoDeploymentReplicationUpdatedEvent is called when the deployment replication is updated by the user.
func (dr *DeploymentReplication) handleArangoDeploymentReplicationUpdatedEvent(event *deploymentReplicationEvent) error {
	log := dr.deps.Log.With().Str("deployoment-replication", event.DeploymentReplication.GetName()).Logger()
	repls := dr.deps.Client.Arango().ReplicationV1().ArangoDeploymentReplications(dr.apiObject.GetNamespace())

	// Get the most recent version of the deployment replication from the API server
	current, err := repls.Get(context.Background(), dr.apiObject.GetName(), metav1.GetOptions{})
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get current version of deployment replication from API server")
		if k8sutil.IsNotFound(err) {
			return nil
		}
		return errors.WithStack(err)
	}

	newAPIObject := current.DeepCopy()
	newAPIObject.Spec.SetDefaults()
	newAPIObject.Status = dr.status
	resetFields := dr.apiObject.Spec.ResetImmutableFields(&newAPIObject.Spec)
	if len(resetFields) > 0 {
		log.Debug().Strs("fields", resetFields).Msg("Found modified immutable fields")
	}
	if err := newAPIObject.Spec.Validate(); err != nil {
		dr.createEvent(k8sutil.NewErrorEvent("Validation failed", err, dr.apiObject))
		// Try to reset object
		if err := dr.updateCRSpec(dr.apiObject.Spec); err != nil {
			log.Error().Err(err).Msg("Restore original spec failed")
			dr.createEvent(k8sutil.NewErrorEvent("Restore original failed", err, dr.apiObject))
		}
		return nil
	}
	if len(resetFields) > 0 {
		for _, fieldName := range resetFields {
			log.Debug().Str("field", fieldName).Msg("Reset immutable field")
			dr.createEvent(k8sutil.NewImmutableFieldEvent(fieldName, dr.apiObject))
		}
	}

	// Save updated spec
	if err := dr.updateCRSpec(newAPIObject.Spec); err != nil {
		return errors.WithStack(errors.Newf("failed to update ArangoDeploymentReplication spec: %v", err))
	}

	// Trigger inspect
	dr.inspectTrigger.Trigger()

	return nil
}

// createEvent creates a given event.
// On error, the error is logged.
func (dr *DeploymentReplication) createEvent(evt *k8sutil.Event) {
	dr.deps.EventRecorder.Event(evt.InvolvedObject, evt.Type, evt.Reason, evt.Message)
}

// Update the status of the API object from the internal status
func (dr *DeploymentReplication) updateCRStatus() error {
	if reflect.DeepEqual(dr.apiObject.Status, dr.status) {
		// Nothing has changed
		return nil
	}

	// Send update to API server
	log := dr.deps.Log
	repls := dr.deps.Client.Arango().ReplicationV1().ArangoDeploymentReplications(dr.apiObject.GetNamespace())
	update := dr.apiObject.DeepCopy()
	attempt := 0
	for {
		attempt++
		update.Status = dr.status
		newAPIObject, err := repls.Update(context.Background(), update, metav1.UpdateOptions{})
		if err == nil {
			// Update internal object
			dr.apiObject = newAPIObject
			return nil
		}
		if attempt < 10 && k8sutil.IsConflict(err) {
			// API object may have been changed already,
			// Reload api object and try again
			var current *api.ArangoDeploymentReplication
			current, err = repls.Get(context.Background(), update.GetName(), metav1.GetOptions{})
			if err == nil {
				update = current.DeepCopy()
				continue
			}
		}
		if err != nil {
			log.Debug().Err(err).Msg("failed to patch ArangoDeploymentReplication status")
			return errors.WithStack(errors.Newf("failed to patch ArangoDeploymentReplication status: %v", err))
		}
	}
}

// Update the spec part of the API object (d.apiObject)
// to the given object, while preserving the status.
// On success, d.apiObject is updated.
func (dr *DeploymentReplication) updateCRSpec(newSpec api.DeploymentReplicationSpec) error {
	log := dr.deps.Log
	repls := dr.deps.Client.Arango().ReplicationV1().ArangoDeploymentReplications(dr.apiObject.GetNamespace())

	// Send update to API server
	update := dr.apiObject.DeepCopy()
	attempt := 0
	for {
		attempt++
		update.Spec = newSpec
		update.Status = dr.status
		newAPIObject, err := repls.Update(context.Background(), update, metav1.UpdateOptions{})
		if err == nil {
			// Update internal object
			dr.apiObject = newAPIObject
			return nil
		}
		if attempt < 10 && k8sutil.IsConflict(err) {
			// API object may have been changed already,
			// Reload api object and try again
			var current *api.ArangoDeploymentReplication
			current, err = repls.Get(context.Background(), update.GetName(), metav1.GetOptions{})
			if err == nil {
				update = current.DeepCopy()
				continue
			}
		}
		if err != nil {
			log.Debug().Err(err).Msg("failed to patch ArangoDeploymentReplication spec")
			return errors.WithStack(errors.Newf("failed to patch ArangoDeploymentReplication spec: %v", err))
		}
	}
}

// failOnError reports the given error and sets the deployment replication status to failed.
func (dr *DeploymentReplication) failOnError(err error, msg string) {
	log := dr.deps.Log
	log.Error().Err(err).Msg(msg)
	dr.status.Reason = err.Error()
	dr.reportFailedStatus()
}

// reportFailedStatus sets the status of the deployment replication to Failed and keeps trying to forward
// that to the API server.
func (dr *DeploymentReplication) reportFailedStatus() {
	log := dr.deps.Log
	log.Info().Msg("local storage failed. Reporting failed reason...")
	repls := dr.deps.Client.Arango().ReplicationV1().ArangoDeploymentReplications(dr.apiObject.GetNamespace())

	op := func() error {
		dr.status.Phase = api.DeploymentReplicationPhaseFailed
		err := dr.updateCRStatus()
		if err == nil || k8sutil.IsNotFound(err) {
			// Status has been updated
			return nil
		}

		if !k8sutil.IsConflict(err) {
			log.Warn().Err(err).Msg("retry report status: fail to update")
			return errors.WithStack(err)
		}

		depl, err := repls.Get(context.Background(), dr.apiObject.Name, metav1.GetOptions{})
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
		dr.apiObject = depl
		return errors.WithStack(errors.Newf("retry needed"))
	}

	retry.Retry(op, time.Hour*24*365)
}
