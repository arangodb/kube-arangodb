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

package replication

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/arangodb/arangosync-client/client"
	"github.com/arangodb/arangosync-client/client/synccheck"
	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
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
	if dr.status.Phase == api.DeploymentReplicationPhaseFailed {
		dr.log.Debug("Deployment replication is in Failed state.")
		return nextInterval
	}

	// Is delete triggered?
	if dr.apiObject.GetDeletionTimestamp() != nil {
		// Deployment replication is triggered for deletion.
		if err := dr.runFinalizers(ctx, dr.apiObject); err != nil {
			dr.log.Err(err).Warn("Failed to run finalizers")
			hasError = true
		}
	} else {
		// Inspect configuration status
		destClient, err := dr.createSyncMasterClient(spec.Destination)
		if err != nil {
			dr.reportDeploymentReplicationErr(err, "Failed to create destination syncmaster client")
			hasError = true
		} else {
			destArangosyncVersion, err := destClient.Version(ctx)
			if err != nil {
				dr.reportDeploymentReplicationErr(err, "Failed to get destination arangosync version")
				hasError = true
			}

			// Fetch status of destination
			updateStatusNeeded := false
			configureSyncNeeded := false
			cancelSyncNeeded := false
			destEndpoint, err := destClient.Master().GetEndpoints(ctx)
			if err != nil {
				dr.log.Err(err).Warn("Failed to fetch endpoints from destination syncmaster")
				hasError = true
			}
			destStatus, err := destClient.Master().Status(ctx)
			if err != nil {
				dr.log.Err(err).Warn("Failed to fetch status from destination syncmaster")
				hasError = true
			} else {
				// Inspect destination status
				if destStatus.Status.IsActive() {
					isIncomingEndpoint, err := dr.isIncomingEndpoint(destStatus, spec.Source)
					if err != nil {
						dr.log.Err(err).Warn("Failed to check is-incoming-endpoint")
					} else {
						if isIncomingEndpoint {
							// Destination is correctly configured
							dr.status.Conditions.Update(api.ConditionTypeConfigured, true, "Active", "Destination syncmaster is configured correctly and active")
							dr.status.Destination = createEndpointStatus(destStatus, "")
							dr.status.IncomingSynchronization = dr.inspectIncomingSynchronizationStatus(ctx, destClient, driver.Version(destArangosyncVersion.Version), destStatus.Shards)
							updateStatusNeeded = true
						} else {
							// Sync is active, but from different source
							dr.log.Warn("Destination syncmaster is configured for different source")
							cancelSyncNeeded = true
							if dr.status.Conditions.Update(api.ConditionTypeConfigured, false, "Invalid", "Destination syncmaster is configured for different source") {
								updateStatusNeeded = true
							}
						}
					}
				} else {
					// Destination has correct source, but is inactive
					configureSyncNeeded = true
					if dr.status.Conditions.Update(api.ConditionTypeConfigured, false, "Inactive", "Destination syncmaster is configured correctly but in-active") {
						updateStatusNeeded = true
					}
				}
			}

			// Inspect source
			sourceClient, err := dr.createSyncMasterClient(spec.Source)
			if err != nil {
				dr.reportDeploymentReplicationErr(err, "Failed to create destination syncmaster client")
				hasError = true
				updateStatusNeeded = false
			} else {
				sourceStatus, err := sourceClient.Master().Status(ctx)
				if err != nil {
					dr.log.Err(err).Warn("Failed to fetch status from source syncmaster")
				}

				//if sourceStatus.Status.IsActive() {
				outgoingID, hasOutgoingEndpoint, err := dr.hasOutgoingEndpoint(sourceStatus, spec.Destination, destEndpoint)
				if err != nil {
					dr.log.Err(err).Warn("Failed to check has-outgoing-endpoint")
				} else if hasOutgoingEndpoint {
					// Destination is know in source
					// Fetch shard status
					dr.status.Source = createEndpointStatus(sourceStatus, outgoingID)
					updateStatusNeeded = true
				} else {
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
					dr.reportDeploymentReplicationErr(err, "Failed to create syncmaster endpoint")
					hasError = true
				} else {
					auth, err := dr.createArangoSyncTLSAuthentication(spec)
					if err != nil {
						dr.reportDeploymentReplicationErr(err, "Failed to configure synchronization authentication")
						hasError = true
					} else {
						req := client.SynchronizationRequest{
							Source:         source,
							Authentication: auth,
						}
						dr.log.Info("Configuring synchronization")
						if err := destClient.Master().Synchronize(ctx, req); err != nil {
							dr.reportDeploymentReplicationErr(err, "Failed to configure synchronization")
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
func (dr *DeploymentReplication) inspectIncomingSynchronizationStatus(ctx context.Context, syncClient client.API, arangosyncVersion driver.Version, localShards []client.ShardSyncInfo) api.SynchronizationStatus {
	dataCentersResp, err := syncClient.Master().GetDataCentersInfo(ctx)
	if err != nil {
		errMsg := "Failed to fetch data-centers info"
		dr.log.Err(err).Warn(errMsg)
		return api.SynchronizationStatus{
			Error: fmt.Sprintf("%s: %s", errMsg, err.Error()),
		}
	}

	ch := synccheck.NewSynchronizationChecker(syncClient, time.Minute)
	incomingSyncStatus, err := ch.CheckSync(ctx, &dataCentersResp, localShards)
	if err != nil {
		errMsg := "Failed to check synchronization status"
		dr.log.Err(err).Warn(errMsg)
		return api.SynchronizationStatus{
			Error: fmt.Sprintf("%s: %s", errMsg, err.Error()),
		}
	}
	return dr.createSynchronizationStatus(arangosyncVersion, incomingSyncStatus)
}

// createSynchronizationStatus returns aggregated info about DCSyncStatus
func (dr *DeploymentReplication) createSynchronizationStatus(arangosyncVersion driver.Version, dcSyncStatus *synccheck.DCSyncStatus) api.SynchronizationStatus {
	dbs := make(map[string]api.DatabaseSynchronizationStatus, len(dcSyncStatus.Databases))
	i := 0
	for dbName, dbSyncStatus := range dcSyncStatus.Databases {
		i++
		db := dbName
		if features.SensitiveInformationProtection().Enabled() {
			// internal IDs are not available in older versions
			if arangosyncVersion.CompareTo("2.12.0") >= 0 {
				db = dbSyncStatus.ID
			} else {
				db = fmt.Sprintf("<PROTECTED_INFO_%d>", i)
			}
		}
		dbs[db] = dr.createDatabaseSynchronizationStatus(dbSyncStatus)
	}
	return api.SynchronizationStatus{
		AllInSync: dcSyncStatus.AllInSync(),
		Databases: dbs,
		Error:     "",
	}
}

// createDatabaseSynchronizationStatus returns sync status for DB
func (dr *DeploymentReplication) createDatabaseSynchronizationStatus(dbSyncStatus synccheck.DatabaseSyncStatus) api.DatabaseSynchronizationStatus {
	// use limit for errors because the resulting status object should not be too big
	const maxReportedIncomingSyncErrors = 20

	var errs []api.DatabaseSynchronizationError
	var shardsTotal, shardsInSync int
	var errorsReportedToLog = 0
	for colName, colSyncStatus := range dbSyncStatus.Collections {
		if colSyncStatus.Error != "" && len(errs) < maxReportedIncomingSyncErrors {
			col := colName
			if features.SensitiveInformationProtection().Enabled() {
				col = colSyncStatus.ID
			}

			errs = append(errs, api.DatabaseSynchronizationError{
				Collection: col,
				Shard:      "",
				Message:    colSyncStatus.Error,
			})
		}

		shardsTotal += len(colSyncStatus.Shards)
		for shardIndex, shardSyncStatus := range colSyncStatus.Shards {
			if shardSyncStatus.InSync {
				shardsInSync++
			} else if errorsReportedToLog < maxReportedIncomingSyncErrors {
				dr.log.Str("db", dbSyncStatus.ID).
					Str("col", colSyncStatus.ID).
					Int("shard", shardIndex).
					Debug("incoming synchronization shard status is not in-sync: %s", shardSyncStatus.Message)
				errorsReportedToLog++
			}
		}
	}

	return api.DatabaseSynchronizationStatus{
		ShardsTotal:  shardsTotal,
		ShardsInSync: shardsInSync,
		Errors:       errs,
	}
}

// createEndpointStatus creates an api EndpointStatus from the given sync status.
func createEndpointStatus(status client.SyncInfo, outgoingID string) api.EndpointStatus {
	result := api.EndpointStatus{}
	if outgoingID == "" {
		return createEndpointStatusFromShards(status.Shards)
	}
	for _, o := range status.Outgoing {
		if o.ID != outgoingID {
			continue
		}
		return createEndpointStatusFromShards(o.Shards)
	}

	return result
}

// createEndpointStatusFromShards creates an api EndpointStatus from the given list of shard statuses.
func createEndpointStatusFromShards(shards []client.ShardSyncInfo) api.EndpointStatus {
	result := api.EndpointStatus{}

	getDatabase := func(name string) *api.DatabaseStatus {
		for i, d := range result.Databases {
			if d.Name == name {
				return &result.Databases[i]
			}
		}
		// Not found, add it
		result.Databases = append(result.Databases, api.DatabaseStatus{Name: name})
		return &result.Databases[len(result.Databases)-1]
	}

	getCollection := func(db *api.DatabaseStatus, name string) *api.CollectionStatus {
		for i, c := range db.Collections {
			if c.Name == name {
				return &db.Collections[i]
			}
		}
		// Not found, add it
		db.Collections = append(db.Collections, api.CollectionStatus{Name: name})
		return &db.Collections[len(db.Collections)-1]
	}

	// Sort shard by index
	sort.Slice(shards, func(i, j int) bool {
		return shards[i].ShardIndex < shards[j].ShardIndex
	})
	for _, s := range shards {
		db := getDatabase(s.Database)
		col := getCollection(db, s.Collection)

		// Add "missing" shards if needed
		for len(col.Shards) < s.ShardIndex {
			col.Shards = append(col.Shards, api.ShardStatus{Status: ""})
		}

		// Add current shard
		col.Shards = append(col.Shards, api.ShardStatus{Status: string(s.Status)})
	}

	// Sort result
	sort.Slice(result.Databases, func(i, j int) bool { return result.Databases[i].Name < result.Databases[j].Name })
	for i, db := range result.Databases {
		sort.Slice(db.Collections, func(i, j int) bool { return db.Collections[i].Name < db.Collections[j].Name })
		result.Databases[i] = db
	}
	return result
}
