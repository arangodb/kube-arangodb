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

package replication

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/arangosync/client"
	api "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	maxCancelFailures = 5 // After this amount of failed cancel-synchronization attempts, the operator switch to abort-sychronization.
)

// addFinalizers adds a stop-sync finalizer to the api object when needed.
func (dr *DeploymentReplication) addFinalizers() error {
	apiObject := dr.apiObject
	if apiObject.GetDeletionTimestamp() != nil {
		// Delete already triggered, cannot add.
		return nil
	}
	for _, f := range apiObject.GetFinalizers() {
		if f == constants.FinalizerDeplReplStopSync {
			// Finalizer already added
			return nil
		}
	}
	apiObject.SetFinalizers(append(apiObject.GetFinalizers(), constants.FinalizerDeplReplStopSync))
	if err := dr.updateCRSpec(apiObject.Spec); err != nil {
		return maskAny(err)
	}
	return nil
}

// runFinalizers goes through the list of ArangoDeploymentReplication finalizers to see if they can be removed.
func (dr *DeploymentReplication) runFinalizers(ctx context.Context, p *api.ArangoDeploymentReplication) error {
	log := dr.deps.Log.With().Str("replication-name", p.GetName()).Logger()
	var removalList []string
	for _, f := range p.ObjectMeta.GetFinalizers() {
		switch f {
		case constants.FinalizerDeplReplStopSync:
			log.Debug().Msg("Inspecting stop-sync finalizer")
			if err := dr.inspectFinalizerDeplReplStopSync(ctx, log, p); err == nil {
				removalList = append(removalList, f)
			} else {
				log.Debug().Err(err).Str("finalizer", f).Msg("Cannot remove finalizer yet")
			}
		}
	}
	// Remove finalizers (if needed)
	if len(removalList) > 0 {
		ignoreNotFound := false
		if err := removeDeploymentReplicationFinalizers(log, dr.deps.CRCli, p, removalList, ignoreNotFound); err != nil {
			log.Debug().Err(err).Msg("Failed to update deployment replication (to remove finalizers)")
			return maskAny(err)
		}
	}
	return nil
}

// inspectFinalizerDeplReplStopSync checks the finalizer condition for stop-sync.
// It returns nil if the finalizer can be removed.
func (dr *DeploymentReplication) inspectFinalizerDeplReplStopSync(ctx context.Context, log zerolog.Logger, p *api.ArangoDeploymentReplication) error {
	// Inspect phase
	if p.Status.Phase.IsFailed() {
		log.Debug().Msg("Deployment replication is already failed, safe to remove stop-sync finalizer")
		return nil
	}

	// Inspect deployment deletion state in source
	abort := dr.status.CancelFailures > maxCancelFailures
	depls := dr.deps.CRCli.DatabaseV1alpha().ArangoDeployments(p.GetNamespace())
	if name := p.Spec.Source.GetDeploymentName(); name != "" {
		depl, err := depls.Get(name, metav1.GetOptions{})
		if k8sutil.IsNotFound(err) {
			log.Debug().Msg("Source deployment is gone. Abort enabled")
			abort = true
		} else if err != nil {
			log.Warn().Err(err).Msg("Failed to get source deployment")
			return maskAny(err)
		} else if depl.GetDeletionTimestamp() != nil {
			log.Debug().Msg("Source deployment is being deleted. Abort enabled")
			abort = true
		}
	}

	// Inspect deployment deletion state in destination
	cleanupSource := false
	if name := p.Spec.Destination.GetDeploymentName(); name != "" {
		depl, err := depls.Get(name, metav1.GetOptions{})
		if k8sutil.IsNotFound(err) {
			log.Debug().Msg("Destination deployment is gone. Source cleanup enabled")
			cleanupSource = true
		} else if err != nil {
			log.Warn().Err(err).Msg("Failed to get destinaton deployment")
			return maskAny(err)
		} else if depl.GetDeletionTimestamp() != nil {
			log.Debug().Msg("Destination deployment is being deleted. Source cleanup enabled")
			cleanupSource = true
		}
	}

	// Cleanup source or stop sync
	if cleanupSource {
		// Destination is gone, cleanup source
		/*sourceClient, err := dr.createSyncMasterClient(p.Spec.Source)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to create source client")
			return maskAny(err)
		}*/
		//sourceClient.Master().C
		return maskAny(fmt.Errorf("TODO"))
	} else {
		// Destination still exists, stop/abort sync
		destClient, err := dr.createSyncMasterClient(p.Spec.Destination)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to create destination client")
			return maskAny(err)
		}
		req := client.CancelSynchronizationRequest{
			WaitTimeout:  time.Minute * 3,
			Force:        abort,
			ForceTimeout: time.Minute * 2,
		}
		log.Debug().Bool("abort", abort).Msg("Stopping synchronization...")
		_, err = destClient.Master().CancelSynchronization(ctx, req)
		if err != nil && !client.IsPreconditionFailed(err) {
			log.Warn().Err(err).Bool("abort", abort).Msg("Failed to stop synchronization")
			dr.status.CancelFailures++
			if err := dr.updateCRStatus(); err != nil {
				log.Warn().Err(err).Msg("Failed to update status to reflect cancel-failures increment")
			}
			return maskAny(err)
		}
		return nil
	}
}

// removeDeploymentReplicationFinalizers removes the given finalizers from the given DeploymentReplication.
func removeDeploymentReplicationFinalizers(log zerolog.Logger, crcli versioned.Interface, p *api.ArangoDeploymentReplication, finalizers []string, ignoreNotFound bool) error {
	repls := crcli.ReplicationV1alpha().ArangoDeploymentReplications(p.GetNamespace())
	getFunc := func() (metav1.Object, error) {
		result, err := repls.Get(p.GetName(), metav1.GetOptions{})
		if err != nil {
			return nil, maskAny(err)
		}
		return result, nil
	}
	updateFunc := func(updated metav1.Object) error {
		updatedRepl := updated.(*api.ArangoDeploymentReplication)
		result, err := repls.Update(updatedRepl)
		if err != nil {
			return maskAny(err)
		}
		*p = *result
		return nil
	}
	if err := k8sutil.RemoveFinalizers(log, finalizers, getFunc, updateFunc, ignoreNotFound); err != nil {
		return maskAny(err)
	}
	return nil
}
