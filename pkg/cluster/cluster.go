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

package cluster

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

type Config struct {
	ServiceAccount string
}

type Dependencies struct {
	Log          zerolog.Logger
	KubeCli      kubernetes.Interface
	ClusterCRCli versioned.Interface
}

// clusterEventType strongly typed type of event
type clusterEventType string

const (
	eventModifyCluster clusterEventType = "Modify"
)

// clusterType holds an event passed from the controller to the cluster.
type clusterEvent struct {
	Type    clusterEventType
	Cluster *api.ArangoCluster
}

const (
	clusterEventQueueSize = 100
)

// Cluster is the in process state of an ArangoDB cluster.
type Cluster struct {
	cluster *api.ArangoCluster // API object
	status  api.ClusterStatus  // Internal status of the CR
	config  Config
	deps    Dependencies

	eventCh chan *clusterEvent
	stopCh  chan struct{}
}

// NewCluster creates a new Cluster from the given API object.
func NewCluster(config Config, deps Dependencies, cluster *api.ArangoCluster) (*Cluster, error) {
	if err := cluster.Spec.Validate(); err != nil {
		return nil, maskAny(err)
	}
	c := &Cluster{
		cluster: cluster,
		status:  *(cluster.Status.DeepCopy()),
		config:  config,
		deps:    deps,
		eventCh: make(chan *clusterEvent, clusterEventQueueSize),
		stopCh:  make(chan struct{}),
	}
	return c, nil
}

// Update the cluster.
// This sends an update event in the cluster event queue.
func (c *Cluster) Update(cluster *api.ArangoCluster) {
	c.send(&clusterEvent{
		Type:    eventModifyCluster,
		Cluster: cluster,
	})
}

// Delete the cluster.
// Called when the cluster was deleted by the user.
func (c *Cluster) Delete() {
	c.deps.Log.Info().Msg("cluster is deleted by user")
	close(c.stopCh)
}

// send given event into the cluster event queue.
func (c *Cluster) send(ev *clusterEvent) {
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
func (c *Cluster) run() {
	log := c.deps.Log

	c.status.State = api.ClusterStateRunning
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
			case eventModifyCluster:
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

// handleUpdateEvent processes the given event coming from the cluster event queue.
func (c *Cluster) handleUpdateEvent(event *clusterEvent) error {
	// TODO
	return nil
}

// Update the status of the API object from the internal status
func (c *Cluster) updateCRStatus() error {
	if reflect.DeepEqual(c.cluster.Status, c.status) {
		// Nothing has changed
		return nil
	}

	// Send update to API server
	newCluster := c.cluster
	newCluster.Status = c.status
	newCluster, err := c.deps.ClusterCRCli.ClusterV1alpha().ArangoClusters(c.cluster.Namespace).Update(c.cluster)
	if err != nil {
		return maskAny(fmt.Errorf("failed to update CR status: %v", err))
	}

	// Update internal object
	c.cluster = newCluster

	return nil
}

// reportFailedStatus sets the status of the cluster to Failed and keeps trying to forward
// that to the API server.
func (c *Cluster) reportFailedStatus() {
	log := c.deps.Log
	log.Info().Msg("cluster failed. Reporting failed reason...")

	op := func() error {
		c.status.State = api.ClusterStateFailed
		err := c.updateCRStatus()
		if err == nil || k8sutil.IsNotFound(err) {
			// Status has been updated
			return nil
		}

		if !k8sutil.IsConflict(err) {
			log.Warn().Err(err).Msg("retry report status: fail to update")
			return maskAny(err)
		}

		cl, err := c.deps.ClusterCRCli.ClusterV1alpha().ArangoClusters(c.cluster.Namespace).Get(c.cluster.Name, metav1.GetOptions{})
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
		c.cluster = cl
		return maskAny(fmt.Errorf("retry needed"))
	}

	retry.Retry(op, time.Hour*24*365)
}
