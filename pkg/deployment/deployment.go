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
	"fmt"
	"reflect"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	"github.com/arangodb/k8s-operator/pkg/generated/clientset/versioned"
	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"
	"github.com/arangodb/k8s-operator/pkg/util/retry"
)

// Config holds configuration settings for a Deployment
type Config struct {
	ServiceAccount string
}

// Dependencies holds dependent services for a Deployment
type Dependencies struct {
	Log           zerolog.Logger
	KubeCli       kubernetes.Interface
	DatabaseCRCli versioned.Interface
}

// deploymentEventType strongly typed type of event
type deploymentEventType string

const (
	eventModifyDeployment deploymentEventType = "Modify"
)

// deploymentEvent holds an event passed from the controller to the deployment.
type deploymentEvent struct {
	Type       deploymentEventType
	Deployment *api.ArangoDeployment
}

const (
	deploymentEventQueueSize = 100
)

// Deployment is the in process state of an ArangoDeployment.
type Deployment struct {
	apiObject *api.ArangoDeployment // API object
	status    api.DeploymentStatus  // Internal status of the CR
	config    Config
	deps      Dependencies

	eventCh chan *deploymentEvent
	stopCh  chan struct{}
}

// New creates a new Deployment from the given API object.
func New(config Config, deps Dependencies, apiObject *api.ArangoDeployment) (*Deployment, error) {
	if err := apiObject.Spec.Validate(); err != nil {
		return nil, maskAny(err)
	}
	d := &Deployment{
		apiObject: apiObject,
		status:    *(apiObject.Status.DeepCopy()),
		config:    config,
		deps:      deps,
		eventCh:   make(chan *deploymentEvent, deploymentEventQueueSize),
		stopCh:    make(chan struct{}),
	}

	go d.run()

	return d, nil
}

// Update the deployment.
// This sends an update event in the deployment event queue.
func (d *Deployment) Update(apiObject *api.ArangoDeployment) {
	d.send(&deploymentEvent{
		Type:       eventModifyDeployment,
		Deployment: apiObject,
	})
}

// Delete the deployment.
// Called when the deployment was deleted by the user.
func (d *Deployment) Delete() {
	d.deps.Log.Info().Msg("deployment is deleted by user")
	close(d.stopCh)
}

// send given event into the deployment event queue.
func (d *Deployment) send(ev *deploymentEvent) {
	select {
	case d.eventCh <- ev:
		l, ecap := len(d.eventCh), cap(d.eventCh)
		if l > int(float64(ecap)*0.8) {
			d.deps.Log.Warn().
				Int("used", l).
				Int("capacity", ecap).
				Msg("event queue buffer is almost full")
		}
	case <-d.stopCh:
	}
}

// run is the core the core worker.
// It processes the event queue and polls the state of generated
// resource on a regular basis.
func (d *Deployment) run() {
	log := d.deps.Log

	// Create services
	if err := d.createServices(d.apiObject); err != nil {
		d.failOnError(err, "Failed to create services")
		return
	}

	// Create members
	if err := d.createInitialMembers(d.apiObject); err != nil {
		d.failOnError(err, "Failed to create initial members")
		return
	}

	// Create PVCs
	if err := d.ensurePVCs(d.apiObject); err != nil {
		d.failOnError(err, "Failed to create persistent volume claims")
		return
	}

	d.status.State = api.DeploymentStateRunning
	if err := d.updateCRStatus(); err != nil {
		log.Warn().Err(err).Msg("update initial CR status failed")
	}
	log.Info().Msg("start running...")

	for {
		select {
		case <-d.stopCh:
			// We're being stopped.
			return

		case event := <-d.eventCh:
			// Got event from event queue
			switch event.Type {
			case eventModifyDeployment:
				if err := d.handleUpdateEvent(event); err != nil {
					d.failOnError(err, "Failed to handle deployment update")
					return
				}
			default:
				panic("unknown event type" + event.Type)
			}
		}
	}
}

// handleUpdateEvent processes the given event coming from the deployment event queue.
func (d *Deployment) handleUpdateEvent(event *deploymentEvent) error {
	// TODO
	return nil
}

// createServices creates all services needed to service the given deployment
func (d *Deployment) createServices(apiObject *api.ArangoDeployment) error {
	log := d.deps.Log
	kubecli := d.deps.KubeCli
	owner := apiObject.AsOwner()

	log.Debug().Msg("creating services...")

	if _, err := k8sutil.CreateHeadlessService(kubecli, apiObject, owner); err != nil {
		log.Debug().Err(err).Msg("Failed to create headless service")
		return maskAny(err)
	}
	single := apiObject.Spec.Mode.HasSingleServers()
	if svcName, err := k8sutil.CreateDatabaseClientService(kubecli, apiObject, single, owner); err != nil {
		log.Debug().Err(err).Msg("Failed to create database client service")
		return maskAny(err)
	} else {
		d.status.ServiceName = svcName
		if err := d.updateCRStatus(); err != nil {
			return maskAny(err)
		}
	}
	if apiObject.Spec.Sync.Enabled {
		if svcName, err := k8sutil.CreateSyncMasterClientService(kubecli, apiObject, owner); err != nil {
			log.Debug().Err(err).Msg("Failed to create syncmaster client service")
			return maskAny(err)
		} else {
			d.status.ServiceName = svcName
			if err := d.updateCRStatus(); err != nil {
				return maskAny(err)
			}
		}
	}
	return nil
}

// Update the status of the API object from the internal status
func (d *Deployment) updateCRStatus() error {
	if reflect.DeepEqual(d.apiObject.Status, d.status) {
		// Nothing has changed
		return nil
	}

	// Send update to API server
	update := d.apiObject.DeepCopy()
	update.Status = d.status
	newAPIObject, err := d.deps.DatabaseCRCli.DatabaseV1alpha().ArangoDeployments(d.apiObject.Namespace).Update(update)
	if err != nil {
		return maskAny(fmt.Errorf("failed to update ArangoDeployment status: %v", err))
	}

	// Update internal object
	d.apiObject = newAPIObject

	return nil
}

// failOnError reports the given error and sets the deployment status to failed.
func (d *Deployment) failOnError(err error, msg string) {
	log.Error().Err(err).Msg(msg)
	d.status.Reason = err.Error()
	d.reportFailedStatus()
}

// reportFailedStatus sets the status of the deployment to Failed and keeps trying to forward
// that to the API server.
func (d *Deployment) reportFailedStatus() {
	log := d.deps.Log
	log.Info().Msg("deployment failed. Reporting failed reason...")

	op := func() error {
		d.status.State = api.DeploymentStateFailed
		err := d.updateCRStatus()
		if err == nil || k8sutil.IsNotFound(err) {
			// Status has been updated
			return nil
		}

		if !k8sutil.IsConflict(err) {
			log.Warn().Err(err).Msg("retry report status: fail to update")
			return maskAny(err)
		}

		depl, err := d.deps.DatabaseCRCli.DatabaseV1alpha().ArangoDeployments(d.apiObject.Namespace).Get(d.apiObject.Name, metav1.GetOptions{})
		if err != nil {
			// Update (PUT) will return conflict even if object is deleted since we have UID set in object.
			// Because it will check UID first and return something like:
			// "Precondition failed: UID in precondition: 0xc42712c0f0, UID in object meta: ".
			if k8sutil.IsNotFound(err) {
				return nil
			}
			log.Warn().Err(err).Msg("retry report status: fail to get latest version")
			return maskAny(err)
		}
		d.apiObject = depl
		return maskAny(fmt.Errorf("retry needed"))
	}

	retry.Retry(op, time.Hour*24*365)
}
