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
	"github.com/rs/zerolog/log"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/fields"
	kwatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	"github.com/arangodb/k8s-operator/pkg/deployment"
	"github.com/arangodb/k8s-operator/pkg/generated/clientset/versioned"
	"github.com/arangodb/k8s-operator/pkg/metrics"
)

const (
	initRetryWaitTime = 30 * time.Second
)

var (
	deploymentsCreated  = metrics.MustRegisterCounter("controller", "deployments_created", "Number of deployments that have been created")
	deploymentsDeleted  = metrics.MustRegisterCounter("controller", "deployments_deleted", "Number of deployments that have been deleted")
	deploymentsFailed   = metrics.MustRegisterCounter("controller", "deployments_failed", "Number of deployments that have failed")
	deploymentsModified = metrics.MustRegisterCounter("controller", "deployments_modified", "Number of deployment modifications")
	deploymentsCurrent  = metrics.MustRegisterGauge("controller", "deployments", "Number of deployments currently being managed")
)

type Event struct {
	Type   kwatch.EventType
	Object *api.ArangoDeployment
}

type Controller struct {
	Config
	Dependencies

	deployments map[string]*deployment.Deployment
}

type Config struct {
	Namespace      string
	ServiceAccount string
	CreateCRD      bool
}

type Dependencies struct {
	Log           zerolog.Logger
	KubeCli       kubernetes.Interface
	KubeExtCli    apiextensionsclient.Interface
	DatabaseCRCli versioned.Interface
}

// NewController instantiates a new controller from given config & dependencies.
func NewController(config Config, deps Dependencies) (*Controller, error) {
	c := &Controller{
		Config:       config,
		Dependencies: deps,
		deployments:  make(map[string]*deployment.Deployment),
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
	log := c.Dependencies.Log

	log.Info().Msgf("Running controller in namespace '%s'", c.Config.Namespace)
	source := cache.NewListWatchFromClient(
		c.Dependencies.DatabaseCRCli.DatabaseV1alpha().RESTClient(),
		api.ArangoDeploymentResourcePlural,
		c.Config.Namespace,
		fields.Everything())

	_, informer := cache.NewIndexerInformer(source, &api.ArangoDeployment{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAddArangoDeployment,
		UpdateFunc: c.onUpdateArangoDeployment,
		DeleteFunc: c.onDeleteArangoDeployment,
	}, cache.Indexers{})

	ctx := context.TODO()
	// TODO: use workqueue to avoid blocking
	informer.Run(ctx.Done())
}

// onAddArangoDeployment deployment addition callback
func (c *Controller) onAddArangoDeployment(obj interface{}) {
	log := c.Dependencies.Log
	apiObject := obj.(*api.ArangoDeployment)
	log.Debug().
		Str("name", apiObject.GetObjectMeta().GetName()).
		Str("ns", apiObject.GetObjectMeta().GetNamespace()).
		Msg("ArangoDeployment added")
	c.syncArangoDeployment(apiObject)
}

// onUpdateArangoDeployment deployment update callback
func (c *Controller) onUpdateArangoDeployment(oldObj, newObj interface{}) {
	log := c.Dependencies.Log
	apiObject := newObj.(*api.ArangoDeployment)
	log.Debug().
		Str("name", apiObject.GetObjectMeta().GetName()).
		Str("ns", apiObject.GetObjectMeta().GetNamespace()).
		Msg("ArangoDeployment updated")
	c.syncArangoDeployment(apiObject)
}

// onDeleteArangoDeployment deployment delete callback
func (c *Controller) onDeleteArangoDeployment(obj interface{}) {
	apiObject, ok := obj.(*api.ArangoDeployment)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			panic(fmt.Sprintf("unknown object from ArangoDeployment delete event: %#v", obj))
		}
		apiObject, ok = tombstone.Obj.(*api.ArangoDeployment)
		if !ok {
			panic(fmt.Sprintf("Tombstone contained object that is not an ArangoDeployment: %#v", obj))
		}
	}
	log.Debug().
		Str("name", apiObject.GetObjectMeta().GetName()).
		Str("ns", apiObject.GetObjectMeta().GetNamespace()).
		Msg("ArangoDeployment deleted")
	ev := &Event{
		Type:   kwatch.Deleted,
		Object: apiObject,
	}

	//	pt.start()
	err := c.handleDeploymentEvent(ev)
	if err != nil {
		c.Dependencies.Log.Warn().Err(err).Msg("Failed to handle event")
	}
	//pt.stop()
}

// syncArangoDeployment synchronized the given deployment.
func (c *Controller) syncArangoDeployment(apiObject *api.ArangoDeployment) {
	ev := &Event{
		Type:   kwatch.Added,
		Object: apiObject,
	}
	// re-watch or restart could give ADD event.
	// If for an ADD event the cluster spec is invalid then it is not added to the local cache
	// so modifying that deployment will result in another ADD event
	if _, ok := c.deployments[apiObject.Name]; ok {
		ev.Type = kwatch.Modified
	}

	//pt.start()
	err := c.handleDeploymentEvent(ev)
	if err != nil {
		c.Dependencies.Log.Warn().Err(err).Msg("Failed to handle event")
	}
	//pt.stop()
}

// handleDeploymentEvent processed the given event.
func (c *Controller) handleDeploymentEvent(event *Event) error {
	apiObject := event.Object

	if apiObject.Status.State.IsFailed() {
		deploymentsFailed.Inc()
		if event.Type == kwatch.Deleted {
			delete(c.deployments, apiObject.Name)
			return nil
		}
		return maskAny(fmt.Errorf("ignore failed deployment (%s). Please delete its CR", apiObject.Name))
	}

	// Fill in defaults
	apiObject.Spec.SetDefaults()
	// Validate deployment spec
	if err := apiObject.Spec.Validate(); err != nil {
		return maskAny(errors.Wrapf(err, "invalid deployment spec. please fix the following problem with the deployment spec: %v", err))
	}

	switch event.Type {
	case kwatch.Added:
		if _, ok := c.deployments[apiObject.Name]; ok {
			return maskAny(fmt.Errorf("unsafe state. deployment (%s) was created before but we received event (%s)", apiObject.Name, event.Type))
		}

		cfg, deps := c.makeDeploymentConfigAndDeps()
		nc, err := deployment.New(cfg, deps, apiObject)
		if err != nil {
			return maskAny(fmt.Errorf("failed to create deployment: %s", err))
		}
		c.deployments[apiObject.Name] = nc

		deploymentsCreated.Inc()
		deploymentsCurrent.Set(float64(len(c.deployments)))

	case kwatch.Modified:
		depl, ok := c.deployments[apiObject.Name]
		if !ok {
			return maskAny(fmt.Errorf("unsafe state. deployment (%s) was never created but we received event (%s)", apiObject.Name, event.Type))
		}
		depl.Update(apiObject)
		deploymentsModified.Inc()

	case kwatch.Deleted:
		depl, ok := c.deployments[apiObject.Name]
		if !ok {
			return maskAny(fmt.Errorf("unsafe state. deployment (%s) was never created but we received event (%s)", apiObject.Name, event.Type))
		}
		depl.Delete()
		delete(c.deployments, apiObject.Name)
		deploymentsDeleted.Inc()
		deploymentsCurrent.Set(float64(len(c.deployments)))
	}
	return nil
}

// makeDeploymentConfigAndDeps creates a Config & Dependencies object for a new cluster.
func (c *Controller) makeDeploymentConfigAndDeps() (deployment.Config, deployment.Dependencies) {
	cfg := deployment.Config{
		ServiceAccount: c.Config.ServiceAccount,
	}
	deps := deployment.Dependencies{
		Log:           c.Dependencies.Log,
		KubeCli:       c.Dependencies.KubeCli,
		DatabaseCRCli: c.Dependencies.DatabaseCRCli,
	}
	return cfg, deps
}
