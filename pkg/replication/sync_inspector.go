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

	"github.com/arangodb/arangosync-client/client"

	api "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// inspectDeploymentReplication inspects the entire deployment replication
// and configures the replication when needed.
// This function should be called when:
// - the deployment replication has changed
// - any of the underlying resources has changed
// - once in a while
// Returns the delay until this function should be called again.
func (dr *DeploymentReplication) inspectDeploymentReplication(lastInterval time.Duration) time.Duration {
	spec := dr.apiObject.Spec
	nextInterval := lastInterval
	hasError := false
	ctx := context.Background()

	// Add finalizers
	if err := dr.addFinalizers(); err != nil {
		dr.log.Err(err).Warn("Failed to add finalizers")
	}

	// Is the deployment in failed state, if so, give up.
	if dr.status.Phase.IsFailed() {
		dr.log.Debug("Deployment replication is in Failed state.")
		return nextInterval
	}

	// Is delete triggered?
	if timestamp := dr.apiObject.GetDeletionTimestamp(); timestamp != nil {
		// Resource is being deleted.
		retrySoon, err := dr.runFinalizers(ctx, dr.apiObject)
		if err != nil || retrySoon {
			if err != nil {
				dr.log.Err(err).Warn("Failed to run finalizers")
			}
			timeout := CancellationTimeout + AbortTimeout
			if isTimeExceeded(timestamp, timeout) {
				dr.failOnError(err, fmt.Sprintf("Failed to cancel synchronization in %s", timeout.String()))
			}
		}

		return cancellationInterval
	} else {
		// Inspect configuration status
		destClient, err := dr.createSyncMasterClient(spec.Destination)
		if err != nil {
			dr.log.Err(err).Warn("Failed to create destination syncmaster client")
		} else {
			// Fetch status of destination
			updateStatusNeeded := false
			configureSyncNeeded := false
			cancelSyncNeeded := false
			destEndpoint, err := destClient.Master().GetEndpoints(ctx)
			if err != nil {
				dr.log.Err(err).Warn("Failed to fetch endpoints from destination syncmaster")
			}
			destStatus, err := destClient.Master().Status(ctx, client.GetSyncStatusDetailsFull)
			if err != nil {
				dr.log.Err(err).Warn("Failed to fetch status from destination syncmaster")
			} else {
				// Inspect destination status
				if destStatus.Status.IsActive() {
					isIncomingEndpoint, err := dr.isIncomingEndpoint(destStatus, spec.Source)
					if err != nil {
						dr.log.Err(err).Warn("Failed to check is-incoming-endpoint")
					} else {
						if isIncomingEndpoint {
							// Destination is correctly configured
							dr.status.Conditions.Update(api.ConditionTypeConfigured, true, api.ConditionConfiguredReasonActive,
								"Destination syncmaster is configured correctly and active")
							dr.status.IncomingSynchronization = dr.inspectIncomingSynchronizationStatus(destStatus)
							updateStatusNeeded = true
						} else {
							// Sync is active, but from different source
							dr.log.Warn("Destination syncmaster is configured for different source")
							cancelSyncNeeded = true
							if dr.status.Conditions.Update(api.ConditionTypeConfigured, false, api.ConditionConfiguredReasonInvalid,
								"Destination syncmaster is configured for different source") {
								updateStatusNeeded = true
							}
						}
					}
				} else {
					// Destination has correct source, but is inactive
					configureSyncNeeded = true
					if dr.status.Conditions.Update(api.ConditionTypeConfigured, false, api.ConditionConfiguredReasonInactive,
						"Destination syncmaster is configured correctly but in-active") {
						updateStatusNeeded = true
					}
				}
			}

			// Inspect source
			sourceClient, err := dr.createSyncMasterClient(spec.Source)
			if err != nil {
				dr.log.Err(err).Warn("Failed to create source syncmaster client")
			} else {
				sourceStatus, err := sourceClient.Master().Status(ctx, client.GetSyncStatusDetailsShort)
				if err != nil {
					dr.log.Err(err).Warn("Failed to fetch status from source syncmaster")
				}

				//if sourceStatus.Status.IsActive() {
				_, hasOutgoingEndpoint, err := dr.hasOutgoingEndpoint(sourceStatus, spec.Destination, destEndpoint)
				if err != nil {
					dr.log.Err(err).Warn("Failed to check has-outgoing-endpoint")
				} else if !hasOutgoingEndpoint {
					// We cannot find the destination in the source status
					dr.log.Err(err).Info("Destination not yet known in source syncmasters")
				}
			}

			// Update status if needed
			if updateStatusNeeded {
				if err := dr.updateCRStatus(); err != nil {
					dr.log.Err(err).Warn("Failed to update status")
					hasError = true
				}
			}

			// Cancel sync if needed
			if cancelSyncNeeded {
				req := client.CancelSynchronizationRequest{}
				dr.log.Info("Canceling synchronization")
				if _, err := destClient.Master().CancelSynchronization(ctx, req); err != nil {
					dr.log.Err(err).Warn("Failed to cancel synchronization")
					hasError = true
				} else {
					dr.log.Info("Canceled synchronization")
					nextInterval = time.Second * 10
				}
			}

			// Configure sync if needed
			if configureSyncNeeded {
				source, err := dr.createArangoSyncEndpoint(spec.Source)
				if err != nil {
					dr.log.Err(err).Warn("Failed to create syncmaster endpoint")
					hasError = true
				} else {
					auth, err := dr.createArangoSyncTLSAuthentication(spec)
					if err != nil {
						msg := "Failed to configure synchronization authentication"
						dr.log.Err(err).Warn(msg)
						dr.reportInvalidConfigError(false, err, msg)
						hasError = true
					} else {
						req := client.SynchronizationRequest{
							Source:         source,
							Authentication: auth,
						}
						dr.log.Info("Configuring synchronization")
						if err := destClient.Master().Synchronize(ctx, req); err != nil {
							msg := "Failed to configure synchronization"
							dr.log.Err(err).Warn(msg)
							dr.reportInvalidConfigError(true, err, msg)
							hasError = true
						} else {
							dr.log.Info("Configured synchronization")
							nextInterval = time.Second * 10
						}
					}
				}
			}
		}
	}

	// Update next interval (on errors)
	if hasError {
		if dr.recentInspectionErrors == 0 {
			nextInterval = minInspectionInterval
			dr.recentInspectionErrors++
		}
	} else {
		dr.recentInspectionErrors = 0
	}
	if nextInterval > maxInspectionInterval {
		nextInterval = maxInspectionInterval
	}
	return nextInterval
}

// isIncomingEndpoint returns true when given sync status's endpoint
// intersects with the given endpoint spec.
func (dr *DeploymentReplication) isIncomingEndpoint(status client.SyncInfo, epSpec api.EndpointSpec) (bool, error) {
	ep, err := dr.createArangoSyncEndpoint(epSpec)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return !status.Source.Intersection(ep).IsEmpty(), nil
}

// hasOutgoingEndpoint returns true when given sync status has an outgoing
// item that intersects with the given endpoint spec.
// Returns: outgoing-ID, outgoing-found, error
func (dr *DeploymentReplication) hasOutgoingEndpoint(status client.SyncInfo, epSpec api.EndpointSpec, reportedEndpoint client.Endpoint) (string, bool, error) {
	ep, err := dr.createArangoSyncEndpoint(epSpec)
	if err != nil {
		return "", false, errors.WithStack(err)
	}
	ep = ep.Merge(reportedEndpoint...)
	for _, o := range status.Outgoing {
		if !o.Endpoint.Intersection(ep).IsEmpty() {
			return o.ID, true, nil
		}
	}
	return "", false, nil
}

// inspectIncomingSynchronizationStatus returns the synchronization status for the incoming sync
func (dr *DeploymentReplication) inspectIncomingSynchronizationStatus(destStatus client.SyncInfo) api.SynchronizationStatus {
	const maxReportedIncomingSyncErrorsPerDatabase = 10

	var totalShardsFromStatus, shardsInSync int
	dbs := make(map[string]api.DatabaseSynchronizationStatus, 0)
	for _, s := range destStatus.Shards {
		db := dbs[s.Database]
		db.ShardsTotal++
		totalShardsFromStatus++
		if s.Status == client.SyncStatusRunning {
			db.ShardsInSync++
			shardsInSync++
		} else if s.Status == client.SyncStatusFailed && len(db.Errors) < maxReportedIncomingSyncErrorsPerDatabase {
			db.Errors = append(db.Errors, api.DatabaseSynchronizationError{
				Collection: s.Collection,
				Shard:      strconv.Itoa(s.ShardIndex),
				Message:    fmt.Sprintf("shard sync failed: %s", s.StatusMessage),
			})
		}
		dbs[s.Database] = db
	}

	var totalShards = destStatus.TotalShardsCount
	if totalShards == 0 {
		// can be zero for old versions of arangosync
		totalShards = totalShardsFromStatus
	}
	progress := float32(0.0)
	if totalShards > 0 {
		progress = float32(shardsInSync) / float32(totalShards)
	}
	return api.SynchronizationStatus{
		Progress:  progress,
		AllInSync: destStatus.Status == client.SyncStatusRunning,
		Databases: dbs,
		Error:     "",
	}
}
