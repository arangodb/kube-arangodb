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
	replication2 "github.com/arangodb/kube-arangodb/pkg/apis/replication"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	kwatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	api "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/replication"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

var (
	deploymentReplicationsCreated  = metrics.MustRegisterCounter("controller", "deployment_replications_created", "Number of deployment replications that have been created")
	deploymentReplicationsDeleted  = metrics.MustRegisterCounter("controller", "deployment_replications_deleted", "Number of deployment replications that have been deleted")
	deploymentReplicationsFailed   = metrics.MustRegisterCounter("controller", "deployment_replications_failed", "Number of deployment replications that have failed")
	deploymentReplicationsModified = metrics.MustRegisterCounter("controller", "deployment_replications_modified", "Number of deployment replication modifications")
	deploymentReplicationsCurrent  = metrics.MustRegisterGauge("controller", "deployment_replications", "Number of deployment replications currently being managed")
)

// run the deployment replications part of the operator.
// This registers a listener and waits until the process stops.
func (o *Operator) runDeploymentReplications(stop <-chan struct{}) {
	rw := k8sutil.NewResourceWatcher(
		o.Dependencies.Client.Arango().ReplicationV1().RESTClient(),
		replication2.ArangoDeploymentReplicationResourcePlural,
		o.Config.Namespace,
		&api.ArangoDeploymentReplication{},
		cache.ResourceEventHandlerFuncs{
			AddFunc:    o.onAddArangoDeploymentReplication,
			UpdateFunc: o.onUpdateArangoDeploymentReplication,
			DeleteFunc: o.onDeleteArangoDeploymentReplication,
		})

	o.Dependencies.DeploymentReplicationProbe.SetReady()
	rw.Run(stop)
}

// onAddArangoDeploymentReplication deployment replication addition callback
func (o *Operator) onAddArangoDeploymentReplication(obj interface{}) {
	o.Dependencies.LivenessProbe.Lock()
	defer o.Dependencies.LivenessProbe.Unlock()

	apiObject := obj.(*api.ArangoDeploymentReplication)
	o.log.
		Str("name", apiObject.GetObjectMeta().GetName()).
		Debug("ArangoDeploymentReplication added")
	o.syncArangoDeploymentReplication(apiObject)
}

// onUpdateArangoDeploymentReplication deployment replication update callback
func (o *Operator) onUpdateArangoDeploymentReplication(oldObj, newObj interface{}) {
	o.Dependencies.LivenessProbe.Lock()
	defer o.Dependencies.LivenessProbe.Unlock()

	apiObject := newObj.(*api.ArangoDeploymentReplication)
	o.log.
		Str("name", apiObject.GetObjectMeta().GetName()).
		Debug("ArangoDeploymentReplication updated")
	o.syncArangoDeploymentReplication(apiObject)
}

// onDeleteArangoDeploymentReplication deployment replication delete callback
func (o *Operator) onDeleteArangoDeploymentReplication(obj interface{}) {
	o.Dependencies.LivenessProbe.Lock()
	defer o.Dependencies.LivenessProbe.Unlock()

	log := o.log
	apiObject, ok := obj.(*api.ArangoDeploymentReplication)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			log.Interface("event-object", obj).Error("unknown object from ArangoDeploymentReplication delete event")
			return
		}
		apiObject, ok = tombstone.Obj.(*api.ArangoDeploymentReplication)
		if !ok {
			log.Interface("event-object", obj).Error("Tombstone contained object that is not an ArangoDeploymentReplication")
			return
		}
	}
	log.
		Str("name", apiObject.GetObjectMeta().GetName()).
		Debug("ArangoDeploymentReplication deleted")
	ev := &Event{
		Type:                  kwatch.Deleted,
		DeploymentReplication: apiObject,
	}

	//	pt.start()
	err := o.handleDeploymentReplicationEvent(ev)
	if err != nil {
		log.Err(err).Warn("Failed to handle event")
	}
	//pt.stop()
}

// syncArangoDeploymentReplication synchronized the given deployment replication.
func (o *Operator) syncArangoDeploymentReplication(apiObject *api.ArangoDeploymentReplication) {
	ev := &Event{
		Type:                  kwatch.Added,
		DeploymentReplication: apiObject,
	}
	// re-watch or restart could give ADD event.
	// If for an ADD event the cluster spec is invalid then it is not added to the local cache
	// so modifying that deployment will result in another ADD event
	if _, ok := o.deploymentReplications[apiObject.Name]; ok {
		ev.Type = kwatch.Modified
	}

	//pt.start()
	err := o.handleDeploymentReplicationEvent(ev)
	if err != nil {
		o.log.Err(err).Warn("Failed to handle event")
	}
	//pt.stop()
}

// handleDeploymentReplicationEvent processed the given event.
func (o *Operator) handleDeploymentReplicationEvent(event *Event) error {
	apiObject := event.DeploymentReplication

	if apiObject.Status.Phase.IsFailed() {
		deploymentReplicationsFailed.Inc()
		if event.Type == kwatch.Deleted {
			delete(o.deploymentReplications, apiObject.Name)
			return nil
		}
		return errors.WithStack(errors.Newf("ignore failed deployment replication (%s). Please delete its CR", apiObject.Name))
	}

	switch event.Type {
	case kwatch.Added:
		if _, ok := o.deploymentReplications[apiObject.Name]; ok {
			return errors.WithStack(errors.Newf("unsafe state. deployment replication (%s) was created before but we received event (%s)", apiObject.Name, event.Type))
		}

		// Fill in defaults
		apiObject.Spec.SetDefaults()
		// Validate deployment spec
		if err := apiObject.Spec.Validate(); err != nil {
			return errors.WithStack(errors.Wrapf(err, "invalid deployment replication spec. please fix the following problem with the deployment replication spec: %v", err))
		}

		cfg, deps := o.makeDeploymentReplicationConfigAndDeps()
		nc, err := replication.New(cfg, deps, apiObject)
		if err != nil {
			return errors.WithStack(errors.Newf("failed to create deployment: %s", err))
		}
		o.deploymentReplications[apiObject.Name] = nc

		deploymentReplicationsCreated.Inc()
		deploymentReplicationsCurrent.Set(float64(len(o.deploymentReplications)))

	case kwatch.Modified:
		repl, ok := o.deploymentReplications[apiObject.Name]
		if !ok {
			return errors.WithStack(errors.Newf("unsafe state. deployment replication (%s) was never created but we received event (%s)", apiObject.Name, event.Type))
		}
		repl.Update(apiObject)
		deploymentReplicationsModified.Inc()

	case kwatch.Deleted:
		repl, ok := o.deploymentReplications[apiObject.Name]
		if !ok {
			return errors.WithStack(errors.Newf("unsafe state. deployment replication (%s) was never created but we received event (%s)", apiObject.Name, event.Type))
		}
		repl.Delete()
		delete(o.deploymentReplications, apiObject.Name)
		deploymentReplicationsDeleted.Inc()
		deploymentReplicationsCurrent.Set(float64(len(o.deploymentReplications)))
	}
	return nil
}

// makeDeploymentReplicationConfigAndDeps creates a Config & Dependencies object for a new DeploymentReplication.
func (o *Operator) makeDeploymentReplicationConfigAndDeps() (replication.Config, replication.Dependencies) {
	cfg := replication.Config{
		Namespace: o.Config.Namespace,
	}
	deps := replication.Dependencies{
		Client:        o.Client,
		EventRecorder: o.Dependencies.EventRecorder,
	}

	return cfg, deps
}
