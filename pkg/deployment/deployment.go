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
	c := &Deployment{
		apiObject: apiObject,
		status:    *(apiObject.Status.DeepCopy()),
		config:    config,
		deps:      deps,
		eventCh:   make(chan *deploymentEvent, deploymentEventQueueSize),
		stopCh:    make(chan struct{}),
	}
	return c, nil
}

// Update the deployment.
// This sends an update event in the deployment event queue.
func (c *Deployment) Update(apiObject *api.ArangoDeployment) {
	c.send(&deploymentEvent{
		Type:       eventModifyDeployment,
		Deployment: apiObject,
	})
}

// Delete the deployment.
// Called when the deployment was deleted by the user.
func (c *Deployment) Delete() {
	c.deps.Log.Info().Msg("deployment is deleted by user")
	close(c.stopCh)
}

// send given event into the deployment event queue.
func (c *Deployment) send(ev *deploymentEvent) {
	select {
	case c.eventCh <- ev:
		l, ecap := len(c.eventCh), cap(c.eventCh)
		if l > int(float64(ecap)*0.8) {
			c.deps.Log.Warn().
				Int("used", l).
				Int("capacity", ecap).
				Msg("event queue buffer is almost full")
		}
	case <-c.stopCh:
	}
}

// run is the core the core worker.
// It processes the event queue and polls the state of generated
// resource on a regular basis.
func (c *Deployment) run() {
	log := c.deps.Log

	c.status.State = api.DeploymentStateRunning
	if err := c.updateCRStatus(); err != nil {
		log.Warn().Err(err).Msg("update initial CR status failed")
	}
	log.Info().Msg("start running...")

	for {
		select {
		case <-c.stopCh:
			// We're being stopped.
			return

		case event := <-c.eventCh:
			// Got event from event queue
			switch event.Type {
			case eventModifyDeployment:
				if err := c.handleUpdateEvent(event); err != nil {
					log.Error().Err(err).Msg("handle update event failed")
					c.status.Reason = err.Error()
					c.reportFailedStatus()
					return
				}
			default:
				panic("unknown event type" + event.Type)
			}
		}
	}
}

// handleUpdateEvent processes the given event coming from the deployment event queue.
func (c *Deployment) handleUpdateEvent(event *deploymentEvent) error {
	// TODO
	return nil
}

// Update the status of the API object from the internal status
func (c *Deployment) updateCRStatus() error {
	if reflect.DeepEqual(c.apiObject.Status, c.status) {
		// Nothing has changed
		return nil
	}

	// Send update to API server
	update := *c.apiObject
	update.Status = c.status
	newAPIObject, err := c.deps.DatabaseCRCli.DatabaseV1alpha().ArangoDeployments(c.apiObject.Namespace).Update(&update)
	if err != nil {
		return maskAny(fmt.Errorf("failed to update CR status: %v", err))
	}

	// Update internal object
	c.apiObject = newAPIObject

	return nil
}

// reportFailedStatus sets the status of the deployment to Failed and keeps trying to forward
// that to the API server.
func (c *Deployment) reportFailedStatus() {
	log := c.deps.Log
	log.Info().Msg("deployment failed. Reporting failed reason...")

	op := func() error {
		c.status.State = api.DeploymentStateFailed
		err := c.updateCRStatus()
		if err == nil || k8sutil.IsNotFound(err) {
			// Status has been updated
			return nil
		}

		if !k8sutil.IsConflict(err) {
			log.Warn().Err(err).Msg("retry report status: fail to update")
			return maskAny(err)
		}

		depl, err := c.deps.DatabaseCRCli.DatabaseV1alpha().ArangoDeployments(c.apiObject.Namespace).Get(c.apiObject.Name, metav1.GetOptions{})
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
		c.apiObject = depl
		return maskAny(fmt.Errorf("retry needed"))
	}

	retry.Retry(op, time.Hour*24*365)
}
