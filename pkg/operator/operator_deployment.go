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

package operator

import (
	"fmt"

	"github.com/pkg/errors"
	kwatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/deployment"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

var (
	deploymentsCreated  = metrics.MustRegisterCounter("controller", "deployments_created", "Number of deployments that have been created")
	deploymentsDeleted  = metrics.MustRegisterCounter("controller", "deployments_deleted", "Number of deployments that have been deleted")
	deploymentsFailed   = metrics.MustRegisterCounter("controller", "deployments_failed", "Number of deployments that have failed")
	deploymentsModified = metrics.MustRegisterCounter("controller", "deployments_modified", "Number of deployment modifications")
	deploymentsCurrent  = metrics.MustRegisterGauge("controller", "deployments", "Number of deployments currently being managed")
)

// run the deployments part of the operator.
// This registers a listener and waits until the process stops.
func (o *Operator) runDeployments(stop <-chan struct{}) {
	rw := k8sutil.NewResourceWatcher(
		o.log,
		o.Dependencies.CRCli.DatabaseV1alpha().RESTClient(),
		api.ArangoDeploymentResourcePlural,
		o.Config.Namespace,
		&api.ArangoDeployment{},
		cache.ResourceEventHandlerFuncs{
			AddFunc:    o.onAddArangoDeployment,
			UpdateFunc: o.onUpdateArangoDeployment,
			DeleteFunc: o.onDeleteArangoDeployment,
		})

	o.Dependencies.DeploymentProbe.SetReady()
	rw.Run(stop)
}

// onAddArangoDeployment deployment addition callback
func (o *Operator) onAddArangoDeployment(obj interface{}) {
	apiObject := obj.(*api.ArangoDeployment)
	o.log.Debug().
		Str("name", apiObject.GetObjectMeta().GetName()).
		Msg("ArangoDeployment added")
	o.syncArangoDeployment(apiObject)
}

// onUpdateArangoDeployment deployment update callback
func (o *Operator) onUpdateArangoDeployment(oldObj, newObj interface{}) {
	apiObject := newObj.(*api.ArangoDeployment)
	o.log.Debug().
		Str("name", apiObject.GetObjectMeta().GetName()).
		Msg("ArangoDeployment updated")
	o.syncArangoDeployment(apiObject)
}

// onDeleteArangoDeployment deployment delete callback
func (o *Operator) onDeleteArangoDeployment(obj interface{}) {
	log := o.log
	apiObject, ok := obj.(*api.ArangoDeployment)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			log.Error().Interface("event-object", obj).Msg("unknown object from ArangoDeployment delete event")
			return
		}
		apiObject, ok = tombstone.Obj.(*api.ArangoDeployment)
		if !ok {
			log.Error().Interface("event-object", obj).Msg("Tombstone contained object that is not an ArangoDeployment")
			return
		}
	}
	log.Debug().
		Str("name", apiObject.GetObjectMeta().GetName()).
		Msg("ArangoDeployment deleted")
	ev := &Event{
		Type:       kwatch.Deleted,
		Deployment: apiObject,
	}

	//	pt.start()
	err := o.handleDeploymentEvent(ev)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to handle event")
	}
	//pt.stop()
}

// syncArangoDeployment synchronized the given deployment.
func (o *Operator) syncArangoDeployment(apiObject *api.ArangoDeployment) {
	ev := &Event{
		Type:       kwatch.Added,
		Deployment: apiObject,
	}
	// re-watch or restart could give ADD event.
	// If for an ADD event the cluster spec is invalid then it is not added to the local cache
	// so modifying that deployment will result in another ADD event
	if _, ok := o.deployments[apiObject.Name]; ok {
		ev.Type = kwatch.Modified
	}

	//pt.start()
	err := o.handleDeploymentEvent(ev)
	if err != nil {
		o.log.Warn().Err(err).Msg("Failed to handle event")
	}
	//pt.stop()
}

// handleDeploymentEvent processed the given event.
func (o *Operator) handleDeploymentEvent(event *Event) error {
	apiObject := event.Deployment

	if apiObject.Status.Phase.IsFailed() {
		deploymentsFailed.Inc()
		if event.Type == kwatch.Deleted {
			delete(o.deployments, apiObject.Name)
			return nil
		}
		return maskAny(fmt.Errorf("ignore failed deployment (%s). Please delete its CR", apiObject.Name))
	}

	switch event.Type {
	case kwatch.Added:
		if _, ok := o.deployments[apiObject.Name]; ok {
			return maskAny(fmt.Errorf("unsafe state. deployment (%s) was created before but we received event (%s)", apiObject.Name, event.Type))
		}

		// Fill in defaults
		apiObject.Spec.SetDefaults(apiObject.GetName())
		// Validate deployment spec
		if err := apiObject.Spec.Validate(); err != nil {
			return maskAny(errors.Wrapf(err, "invalid deployment spec. please fix the following problem with the deployment spec: %v", err))
		}

		cfg, deps := o.makeDeploymentConfigAndDeps(apiObject)
		nc, err := deployment.New(cfg, deps, apiObject)
		if err != nil {
			return maskAny(fmt.Errorf("failed to create deployment: %s", err))
		}
		o.deployments[apiObject.Name] = nc

		deploymentsCreated.Inc()
		deploymentsCurrent.Set(float64(len(o.deployments)))

	case kwatch.Modified:
		depl, ok := o.deployments[apiObject.Name]
		if !ok {
			return maskAny(fmt.Errorf("unsafe state. deployment (%s) was never created but we received event (%s)", apiObject.Name, event.Type))
		}
		depl.Update(apiObject)
		deploymentsModified.Inc()

	case kwatch.Deleted:
		depl, ok := o.deployments[apiObject.Name]
		if !ok {
			return maskAny(fmt.Errorf("unsafe state. deployment (%s) was never created but we received event (%s)", apiObject.Name, event.Type))
		}
		depl.Delete()
		delete(o.deployments, apiObject.Name)
		deploymentsDeleted.Inc()
		deploymentsCurrent.Set(float64(len(o.deployments)))
	}
	return nil
}

// makeDeploymentConfigAndDeps creates a Config & Dependencies object for a new Deployment.
func (o *Operator) makeDeploymentConfigAndDeps(apiObject *api.ArangoDeployment) (deployment.Config, deployment.Dependencies) {
	cfg := deployment.Config{
		ServiceAccount: o.Config.ServiceAccount,
		AllowChaos:     o.Config.AllowChaos,
	}
	deps := deployment.Dependencies{
		Log: o.Dependencies.LogService.MustGetLogger("deployment").With().
			Str("deployment", apiObject.GetName()).
			Logger(),
		KubeCli:       o.Dependencies.KubeCli,
		DatabaseCRCli: o.Dependencies.CRCli,
	}
	return cfg, deps
}
