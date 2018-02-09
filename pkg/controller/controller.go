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

package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/fields"
	kwatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	"github.com/arangodb/k8s-operator/pkg/cluster"
	"github.com/arangodb/k8s-operator/pkg/generated/clientset/versioned"
	"github.com/arangodb/k8s-operator/pkg/metrics"
)

const (
	initRetryWaitTime = 30 * time.Second
)

var (
	clustersCreated  = metrics.MustRegisterCounter("controller", "clusters_created", "Number of clusters that have been created")
	clustersDeleted  = metrics.MustRegisterCounter("controller", "clusters_deleted", "Number of clusters that have been deleted")
	clustersFailed   = metrics.MustRegisterCounter("controller", "clusters_failed", "Number of clusters that have failed")
	clustersModified = metrics.MustRegisterCounter("controller", "clusters_modified", "Number of cluster modifications")
	clustersCurrent  = metrics.MustRegisterGauge("controller", "clusters", "Number of clusters currently being managed")
)

type Event struct {
	Type   kwatch.EventType
	Object *api.ArangoCluster
}

type Controller struct {
	Config
	Dependencies

	clusters map[string]*cluster.Cluster
}

type Config struct {
	Namespace      string
	ServiceAccount string
	CreateCRD      bool
}

type Dependencies struct {
	Log          zerolog.Logger
	KubeCli      kubernetes.Interface
	KubeExtCli   apiextensionsclient.Interface
	ClusterCRCli versioned.Interface
}

// NewController instantiates a new controller from given config & dependencies.
func NewController(config Config, deps Dependencies) (*Controller, error) {
	c := &Controller{
		Config:       config,
		Dependencies: deps,
		clusters:     make(map[string]*cluster.Cluster),
	}
	return c, nil
}

// Start the controller
func (c *Controller) Start() error {
	log := c.Dependencies.Log

	for {
		if err := c.initResourceIfNeeded(); err == nil {
			break
		} else {
			log.Error().Err(err).Msg("Resource initialization failed")
			log.Info().Msgf("Retrying in %s...", initRetryWaitTime)
			time.Sleep(initRetryWaitTime)
		}
	}

	//probe.SetReady()
	c.run()
	panic("unreachable")
}

// run the controller.
// This registers a listener and waits until the process stops.
func (c *Controller) run() {
	source := cache.NewListWatchFromClient(
		c.Dependencies.ClusterCRCli.ClusterV1alpha().RESTClient(),
		api.ArangoClusterResourcePlural,
		c.Config.Namespace,
		fields.Everything())

	_, informer := cache.NewIndexerInformer(source, &api.ArangoCluster{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAddArangoCluster,
		UpdateFunc: c.onUpdateArangoCluster,
		DeleteFunc: c.onDeleteArangoCluster,
	}, cache.Indexers{})

	ctx := context.TODO()
	// TODO: use workqueue to avoid blocking
	informer.Run(ctx.Done())
}

// onAddArangoCluster cluster addition callback
func (c *Controller) onAddArangoCluster(obj interface{}) {
	c.syncArangoCluster(obj.(*api.ArangoCluster))
}

// onUpdateArangoCluster cluster update callback
func (c *Controller) onUpdateArangoCluster(oldObj, newObj interface{}) {
	c.syncArangoCluster(newObj.(*api.ArangoCluster))
}

// onDeleteArangoCluster cluster delete callback
func (c *Controller) onDeleteArangoCluster(obj interface{}) {
	clus, ok := obj.(*api.ArangoCluster)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			panic(fmt.Sprintf("unknown object from EtcdCluster delete event: %#v", obj))
		}
		clus, ok = tombstone.Obj.(*api.ArangoCluster)
		if !ok {
			panic(fmt.Sprintf("Tombstone contained object that is not an ArangoCluster: %#v", obj))
		}
	}
	ev := &Event{
		Type:   kwatch.Deleted,
		Object: clus,
	}

	//	pt.start()
	err := c.handleClusterEvent(ev)
	if err != nil {
		c.Dependencies.Log.Warn().Err(err).Msg("Failed to handle event")
	}
	//pt.stop()
}

// syncArangoCluster synchronized the given cluster.
func (c *Controller) syncArangoCluster(apiCluster *api.ArangoCluster) {
	ev := &Event{
		Type:   kwatch.Added,
		Object: apiCluster,
	}
	// re-watch or restart could give ADD event.
	// If for an ADD event the cluster spec is invalid then it is not added to the local cache
	// so modifying that cluster will result in another ADD event
	if _, ok := c.clusters[apiCluster.Name]; ok {
		ev.Type = kwatch.Modified
	}

	//pt.start()
	err := c.handleClusterEvent(ev)
	if err != nil {
		c.Dependencies.Log.Warn().Err(err).Msg("Failed to handle event")
	}
	//pt.stop()
}

// handleClusterEvent processed the given event.
func (c *Controller) handleClusterEvent(event *Event) error {
	apiCluster := event.Object

	if apiCluster.Status.State.IsFailed() {
		clustersFailed.Inc()
		if event.Type == kwatch.Deleted {
			delete(c.clusters, apiCluster.Name)
			return nil
		}
		return maskAny(fmt.Errorf("ignore failed cluster (%s). Please delete its CR", apiCluster.Name))
	}

	// Fill in defaults
	apiCluster.Spec.SetDefaults()
	// Validate cluster spec
	if err := apiCluster.Spec.Validate(); err != nil {
		return maskAny(errors.Wrapf(err, "invalid cluster spec. please fix the following problem with the cluster spec: %v", err))
	}

	switch event.Type {
	case kwatch.Added:
		if _, ok := c.clusters[apiCluster.Name]; ok {
			return maskAny(fmt.Errorf("unsafe state. cluster (%s) was created before but we received event (%s)", apiCluster.Name, event.Type))
		}

		cfg, deps := c.makeClusterConfigAndDeps()
		nc, err := cluster.NewCluster(cfg, deps, apiCluster)
		if err != nil {
			return maskAny(fmt.Errorf("failed to create cluster: %s", err))
		}
		c.clusters[apiCluster.Name] = nc

		clustersCreated.Inc()
		clustersCurrent.Set(float64(len(c.clusters)))

	case kwatch.Modified:
		if _, ok := c.clusters[apiCluster.Name]; !ok {
			return maskAny(fmt.Errorf("unsafe state. cluster (%s) was never created but we received event (%s)", apiCluster.Name, event.Type))
		}
		c.clusters[apiCluster.Name].Update(apiCluster)
		clustersModified.Inc()

	case kwatch.Deleted:
		if _, ok := c.clusters[apiCluster.Name]; !ok {
			return maskAny(fmt.Errorf("unsafe state. cluster (%s) was never created but we received event (%s)", apiCluster.Name, event.Type))
		}
		c.clusters[apiCluster.Name].Delete()
		delete(c.clusters, apiCluster.Name)
		clustersDeleted.Inc()
		clustersCurrent.Set(float64(len(c.clusters)))
	}
	return nil
}

// makeClusterConfigAndDeps creates a Config & Dependencies object for a new cluster.
func (c *Controller) makeClusterConfigAndDeps() (cluster.Config, cluster.Dependencies) {
	cfg := cluster.Config{
		ServiceAccount: c.Config.ServiceAccount,
	}
	deps := cluster.Dependencies{
		Log:          c.Dependencies.Log,
		KubeCli:      c.Dependencies.KubeCli,
		ClusterCRCli: c.Dependencies.ClusterCRCli,
	}
	return cfg, deps
}
