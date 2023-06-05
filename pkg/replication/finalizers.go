//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	"fmt"
	"strconv"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/arangosync-client/client"
	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

const (
	CancellationTimeout = time.Minute * 15
	AbortTimeout        = time.Minute * 2
)

// addFinalizer adds new finalizer if it does not exist.
func (dr *DeploymentReplication) addFinalizer(finalizer string) error {
	if dr.apiObject.GetDeletionTimestamp() != nil {
		// Delete already triggered, cannot add.
		return nil
	}
	apiObject := dr.apiObject

	if !finalizerExists(apiObject, finalizer) {
		apiObject.SetFinalizers(append(apiObject.GetFinalizers(), finalizer))
		if err := dr.updateCRSpec(apiObject.Spec); err != nil {
			return errors.WithMessage(err, "Failed to update CR Spec")
		}
	}

	return nil
}

// addFinalizers adds a required finalizers to the api object when needed.
func (dr *DeploymentReplication) addFinalizers() error {
	// Add stop sync replication finalizer automatically.
	return dr.addFinalizer(constants.FinalizerDeplReplStopSync)
}

// runFinalizers removes stop sync finalizer if it is possible.
func (dr *DeploymentReplication) runFinalizers(ctx context.Context, p *api.ArangoDeploymentReplication) (bool, error) {
	if !finalizerExists(p, constants.FinalizerDeplReplStopSync) {
		return false, nil
	}

	dr.log.Str("replication-name", p.GetName()).Debug("Inspecting stop-sync finalizer")
	if retrySoon, err := dr.inspectFinalizerDeplReplStopSync(ctx, p); err != nil {
		return true, errors.WithMessagef(err, "Cannot remove finalizer \"%s\" yet", constants.FinalizerDeplReplStopSync)
	} else if retrySoon {
		// No error, but not finished. Try to reconcile soon.
		dr.log.Debug("Synchronization is still cancelling")
		return true, nil
	}

	removalList := []string{constants.FinalizerDeplReplStopSync}
	if err := removeDeploymentReplicationFinalizers(dr.deps.Client.Arango(), p, removalList, false); err != nil {
		return true, errors.WithMessage(err, "Failed to update deployment replication (to remove finalizers)")
	}

	return false, nil
}

// inspectFinalizerDeplReplStopSync checks cancellation progress.
// When true is returned then function can be called after a few seconds to check progress.
// When it returns false and nil error then cancellation process is done.
func (dr *DeploymentReplication) inspectFinalizerDeplReplStopSync(ctx context.Context,
	p *api.ArangoDeploymentReplication) (bool, error) {

	abort := isTimeExceeded(p.GetDeletionTimestamp(), CancellationTimeout)
	// Inspect deployment deletion state in source.
	depls := dr.deps.Client.Arango().DatabaseV1().ArangoDeployments(p.GetNamespace())
	if name := p.Spec.Source.GetDeploymentName(); name != "" {
		depl, err := depls.Get(context.Background(), name, meta.GetOptions{})
		if kerrors.IsNotFound(err) {
			dr.log.Debug("Source deployment is gone. Abort enabled")
			abort = true
		} else if err != nil {
			dr.log.Err(err).Warn("Failed to get source deployment")
			return false, errors.WithStack(err)
		} else if depl.GetDeletionTimestamp() != nil {
			dr.log.Debug("Source deployment is being deleted. Abort enabled")
			abort = true
		}
	}

	// Inspect deployment deletion state in destination
	cleanupSource := false
	if name := p.Spec.Destination.GetDeploymentName(); name != "" {
		depl, err := depls.Get(context.Background(), name, meta.GetOptions{})
		if kerrors.IsNotFound(err) {
			dr.log.Debug("Destination deployment is gone. Source cleanup enabled")
			cleanupSource = true
		} else if err != nil {
			dr.log.Err(err).Warn("Failed to get destination deployment")
			return false, errors.WithStack(err)
		} else if depl.GetDeletionTimestamp() != nil {
			dr.log.Debug("Destination deployment is being deleted. Source cleanup enabled")
			cleanupSource = true
		}
	}

	// Cleanup source or stop sync
	if cleanupSource {
		// Destination is gone, cleanup source
		/*sourceClient, err := dr.createSyncMasterClient(p.Spec.Source)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to create source client")
			return errors.WithStack(err)
		}*/
		//sourceClient.Master().C
		return false, errors.WithStack(errors.Newf("TODO"))
	}

	// Destination still exists, stop/abort sync.
	// Create a client to the destination sync master.
	destClient, err := dr.createSyncMasterClient(p.Spec.Destination)
	if err != nil {
		return false, errors.WithMessage(err, "Failed to create destination synchronization master client")
	}

	// Get status from sync master.
	syncInfo, err := destClient.Master().Status(ctx, client.GetSyncStatusDetailsShort)
	if err != nil {
		return false, errors.WithMessage(err, "Failed to get status from target master")
	}

	// Check progress of a cancellation.
	if syncStatus, err := dr.getCancellationProgress(syncInfo); err != nil {
		return false, err
	} else if syncStatus == client.SyncStatusInactive {
		return false, nil
	} else if syncStatus == client.SyncStatusFailed {
		return false, errors.WithMessagef(err, "unexpected synchronization status \"%s\"", syncStatus)
	} else if syncStatus == client.SyncStatusCancelling {
		// Synchronization is cancelling, so request was already sent.
		if !abort {
			return true, nil
		}

		changed := dr.status.Conditions.Update(api.ConditionTypeAborted, abort, "Cancellation type",
			"Cancellation will wait for source data center to be canceled with a timeout")
		if !changed {
			return true, nil
		}
		// A Request must be sent once again because abort option has changed.
	}

	// Check whether data consistency must be ensured.
	if syncInfo.Status.IsActive() && util.TypeOrDefault[bool](p.Spec.Cancellation.EnsureInSync, true) {
		if inSync, inSyncShards, totalShards, err := dr.ensureInSync(ctx, destClient); err != nil {
			return false, err
		} else if !inSync {
			if time.Since(dr.lastLog) > time.Second*5 {
				dr.lastLog = time.Now()
				dr.log.Info("Consistency is being checked, %d of %d shards are in-sync", inSyncShards, totalShards)
			}

			// Retry soon.
			return true, nil
		}
	}

	// From here on this code should be launched only once unless abort option is changed
	// or replication is not in cancelling state.
	sourceServerMode := driver.ServerModeDefault
	if util.TypeOrDefault[bool](p.Spec.Cancellation.SourceReadOnly) {
		sourceServerMode = driver.ServerModeReadOnly
	}
	req := client.CancelSynchronizationRequest{
		Force:            abort,
		ForceTimeout:     AbortTimeout,
		SourceServerMode: sourceServerMode,
	}
	dr.log.Interface("request", req).Info("Stopping synchronization...")
	_, errCancel := destClient.Master().CancelSynchronization(ctx, req)
	if errCancel != nil {
		dr.status.Reason = fmt.Sprintf("Failed to stop synchronization: %s. Abort: %s", err.Error(), strconv.FormatBool(abort))
	} else {
		dr.status.Reason = "Stopping synchronization started"
	}

	// Update CR status.
	if err := dr.updateCRStatus(); err != nil {
		dr.log.Err(err).Warn("Failed to update replication status")
		// Don't return with this error because original error must be returned.
		// Not a big deal, because only reason was not saved.
		// It will be saved on next updateCRStatus call, because reason is kept in status memory
	}

	// If err is nil then nil will be returned.
	if errCancel != nil {
		if abort {
			return false, errors.WithMessage(errCancel, "Failed to abort synchronization")
		}

		return false, errors.WithMessage(errCancel, "Failed to stop synchronization")
	}

	return true, nil

}

// removeDeploymentReplicationFinalizers removes the given finalizers from the given DeploymentReplication.
func removeDeploymentReplicationFinalizers(crcli versioned.Interface, p *api.ArangoDeploymentReplication, finalizers []string, ignoreNotFound bool) error {
	repls := crcli.ReplicationV1().ArangoDeploymentReplications(p.GetNamespace())
	getFunc := func() (meta.Object, error) {
		result, err := repls.Get(context.Background(), p.GetName(), meta.GetOptions{})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return result, nil
	}
	updateFunc := func(updated meta.Object) error {
		updatedRepl := updated.(*api.ArangoDeploymentReplication)
		result, err := repls.Update(context.Background(), updatedRepl, meta.UpdateOptions{})
		if err != nil {
			return errors.WithStack(err)
		}
		*p = *result
		return nil
	}
	if _, err := k8sutil.RemoveFinalizers(finalizers, getFunc, updateFunc, ignoreNotFound); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// finalizerExists returns true if a given finalizer exists.
func finalizerExists(p *api.ArangoDeploymentReplication, finalizer string) bool {
	for _, f := range p.ObjectMeta.GetFinalizers() {
		if f == finalizer {
			return true
		}
	}

	return false
}

func (dr *DeploymentReplication) getCancellationProgress(syncInfo client.SyncInfo) (client.SyncStatus, error) {
	if syncInfo.IsInactive() {
		if len(syncInfo.Source) > 0 {
			return "", errors.New("Inactive target data center is still configured with the endpoint set to a source DC")
		}
		return client.SyncStatusInactive, nil
	}

	if syncInfo.Status == client.SyncStatusInactive {
		// There are some not finished shards but status is inactive, so it was na canceled.
		return "", errors.New("Target data center is inactive but some shards are not closed")
	}

	return syncInfo.Status, nil
}

// ensureInSync checks whether data is consistent on both data centers.
// During this check both data centers will be in read-only mode.
// Return nil when data is consistent or when consistency was already checked.
func (dr *DeploymentReplication) ensureInSync(ctx context.Context, c client.API) (bool, int, int, error) {
	if dr.status.Conditions.IsTrue(api.ConditionTypeEnsuredInSync) {
		return true, 0, 0, nil
	}

	cancelStatus, err := c.Master().GetSynchronizationBarrierStatus(ctx)
	if err != nil {
		return false, 0, 0, errors.WithMessage(err, "Can not get synchronization barrier status")
	}

	if !cancelStatus.SourceServerReadonly ||
		dr.status.Conditions.Update(api.ConditionTypeEnsuredInSync, false, "Consistent", "Data on both data centers is not the same") {
		// If `GetSynchronizationBarrierStatus` could return active barrier then it would not create the above condition.
		if err := c.Master().CreateSynchronizationBarrier(ctx); err != nil {
			if driver.IsPreconditionFailed(err) {
				dr.log.Info("Can not create synchronization barrier because synchronization is not running")
				return false, 0, 0, nil
			}

			return false, 0, 0, errors.WithMessage(err, "Can not create synchronization barrier")
		}

		if err := dr.updateCRStatus(); err != nil {
			return false, 0, 0, errors.WithMessage(err, "Failed to update ArangoDeploymentReplication status")
		}

		dr.log.Info("Synchronization barrier created, both data centers are in read-only mode")
	}

	totalShards := cancelStatus.InSyncShards + cancelStatus.NotInSyncShards
	if cancelStatus.InSyncShards > 0 && cancelStatus.NotInSyncShards == 0 {
		if dr.status.Conditions.Update(api.ConditionTypeEnsuredInSync, true, "Consistent", "Data on both data centers is the same") {
			if err := dr.updateCRStatus(); err != nil {
				return false, 0, 0, errors.WithMessage(err, "Failed to update ArangoDeploymentReplication status")
			}
		}

		return true, cancelStatus.InSyncShards, totalShards, nil
	}

	return false, cancelStatus.InSyncShards, totalShards, nil
}

// isTimeExceeded returns true when a time exceeds a given timeout.
func isTimeExceeded(t *meta.Time, timeout time.Duration) bool {
	return t != nil && time.Since(t.Time) > timeout
}
