//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package reconcile

import (
	"context"
	"net/http"
	"path"
	"time"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	actionRebuildOutSyncedShardsBatchTTL                       = 600 * time.Second
	actionRebuildOutSyncedShardsLocalJobID    api.PlanLocalKey = "rebuildJobID"
	actionRebuildOutSyncedShardsLocalDatabase api.PlanLocalKey = "database"
	actionRebuildOutSyncedShardsLocalShard    api.PlanLocalKey = "shard"
	actionRebuildOutSyncedShardsBatchID       api.PlanLocalKey = "batchID"
)

// newRebuildOutSyncedShardsAction creates a new Action that implements the given
// planned RebuildOutSyncedShards action.
func newRebuildOutSyncedShardsAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionRebuildOutSyncedShards{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionRebuildOutSyncedShards implements an RebuildOutSyncedShardsAction.
type actionRebuildOutSyncedShards struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionRebuildOutSyncedShards) Start(ctx context.Context) (bool, error) {
	if !features.RebuildOutSyncedShards().Enabled() {
		// RebuildOutSyncedShards feature is not enabled
		return true, nil
	}

	clientSync, err := a.actionCtx.GetMembersState().GetMemberClient(a.action.MemberID)
	if err != nil {
		return false, errors.Wrapf(err, "Unable to create client (SyncMode)")
	}

	clientAsync, err := a.actionCtx.GetServerAsyncClient(a.action.MemberID)
	if err != nil {
		return false, errors.Wrapf(err, "Unable to create client (AsyncMode)")
	}

	shardID, exist := a.action.GetParam("shardID")
	if !exist {
		a.log.Error("*shardID* key not found in action params")
		return true, nil
	}

	database, exist := a.action.GetParam("database")
	if !exist {
		a.log.Error("*database* key not found in action params")
		return true, nil
	}

	// trigger async rebuild job
	err = a.rebuildShard(ctx, clientSync, clientAsync, shardID, database)
	if err != nil {
		a.log.Err(err).Error("Rebuild Shard Tree action failed on start", shardID, database, a.action.MemberID)
		return true, err
	}

	a.log.Info("Triggering async job Shard Tree rebuild", shardID, database, a.action.MemberID)
	return false, nil
}

// CheckProgress returns: ready, abort, error.
func (a *actionRebuildOutSyncedShards) CheckProgress(ctx context.Context) (bool, bool, error) {
	if !features.RebuildOutSyncedShards().Enabled() {
		// RebuildOutSyncedShards feature is not enabled
		return true, false, nil
	}

	clientSync, err := a.actionCtx.GetMembersState().GetMemberClient(a.action.MemberID)
	if err != nil {
		return false, false, errors.Wrapf(err, "Unable to create client (SyncMode)")
	}

	clientAsync, err := a.actionCtx.GetServerAsyncClient(a.action.MemberID)
	if err != nil {
		return false, false, errors.Wrapf(err, "Unable to create client (AsyncMode)")
	}

	jobID, ok := a.actionCtx.Get(a.action, actionRebuildOutSyncedShardsLocalJobID)
	if !ok {
		return false, true, errors.Newf("Local Key is missing in action: %s", actionRebuildOutSyncedShardsLocalJobID)
	}

	batchID, ok := a.actionCtx.Get(a.action, actionRebuildOutSyncedShardsBatchID)
	if !ok {
		return false, true, errors.Newf("Local Key is missing in action: %s", actionRebuildOutSyncedShardsBatchID)
	}

	database, ok := a.actionCtx.Get(a.action, actionRebuildOutSyncedShardsLocalDatabase)
	if !ok {
		return false, true, errors.Newf("Local Key is missing in action: %s", actionRebuildOutSyncedShardsLocalDatabase)
	}

	shardID, ok := a.actionCtx.Get(a.action, actionRebuildOutSyncedShardsLocalShard)
	if !ok {
		return false, true, errors.Newf("Local Key is missing in action: %s", actionRebuildOutSyncedShardsLocalShard)
	}

	// check first if there is rebuild job running
	rebuildInProgress, err := a.checkRebuildShardProgress(ctx, clientAsync, clientSync, shardID, database, jobID, batchID)
	if err != nil {
		if rebuildInProgress {
			a.log.Err(err).Error("Rebuild job failed but we will retry", shardID, database, a.action.MemberID)
			return false, false, err
		} else {
			a.log.Err(err).Error("Rebuild job failed", shardID, database, a.action.MemberID)
			return false, true, err
		}

	}
	if rebuildInProgress {
		a.log.Debug("Rebuild job is still in progress", shardID, database, a.action.MemberID)
		return false, false, nil
	}

	// rebuild job is done
	a.log.Info("Rebuild Shard Tree is done", shardID, database, a.action.MemberID)
	return true, false, nil
}

func (a *actionRebuildOutSyncedShards) rebuildShard(ctx context.Context, clientSync, clientAsync driver.Client, shardID, database string) error {
	batchID, err := a.createBatch(ctx, clientSync)
	if err != nil {
		return errors.Wrapf(err, "Unable to create batch")
	}

	req, err := a.createShardRebuildRequest(clientAsync, shardID, database, batchID)
	if err != nil {
		return err
	}
	_, err = clientAsync.Connection().Do(ctx, req)
	if id, ok := conn.IsAsyncJobInProgress(err); ok {
		a.actionCtx.Add(actionRebuildOutSyncedShardsLocalJobID, id, true)
		a.actionCtx.Add(actionRebuildOutSyncedShardsLocalDatabase, database, true)
		a.actionCtx.Add(actionRebuildOutSyncedShardsLocalShard, shardID, true)
		a.actionCtx.Add(actionRebuildOutSyncedShardsBatchID, batchID, true)
		// Async request has been sent
		return nil
	} else {
		return errors.Wrapf(err, "Unknown rebuild request error")
	}
}

// checkRebuildShardProgress returns: inProgress, error.
func (a *actionRebuildOutSyncedShards) checkRebuildShardProgress(ctx context.Context, clientAsync, clientSync driver.Client, shardID, database, jobID, batchID string) (bool, error) {
	req, err := a.createShardRebuildRequest(clientAsync, shardID, database, batchID)
	if err != nil {
		return false, err
	}

	resp, err := clientAsync.Connection().Do(conn.WithAsyncID(ctx, jobID), req)
	if err != nil {
		if _, ok := conn.IsAsyncJobInProgress(err); ok {
			return true, nil
		}

		// Add wait grace period
		if ok := conn.IsAsyncErrorNotFound(err); ok {
			if s := a.action.StartTime; s != nil && !s.Time.IsZero() {
				if time.Since(s.Time) < 10*time.Second {
					// Retry
					return true, nil
				}
			}
		}

		return false, errors.Wrapf(err, "check rebuild progress error")
	}

	// cleanup batch
	_ = a.deleteBatch(ctx, clientSync, batchID)

	if resp.StatusCode() == http.StatusNoContent {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "rebuild progress failed with status code %d", resp.StatusCode())
	}
}

//************************** API Calls ************************************

// createShardRebuildRequest creates request for rebuilding shard. Returns request, error.
func (a *actionRebuildOutSyncedShards) createShardRebuildRequest(clientAsync driver.Client, shardID, database, batchID string) (driver.Request, error) {
	req, err := clientAsync.Connection().NewRequest("POST", path.Join("_db", database, "_api/replication/revisions/tree"))
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create rebuild shard request, shard %s, database: %s", shardID, database)
	}
	req = req.SetQuery("batchId", batchID)
	req = req.SetQuery("collection", shardID)
	return req, nil
}

// createBatch creates batch on the server. Returns batchID, error.
func (a *actionRebuildOutSyncedShards) createBatch(ctx context.Context, clientSync driver.Client) (string, error) {
	req, err := clientSync.Connection().NewRequest("POST", path.Join("_api/replication/batch"))
	if err != nil {
		return "", errors.Wrapf(err, "Unable to create request for batch creation")
	}
	params := struct {
		TTL float64 `json:"ttl"`
	}{TTL: actionRebuildOutSyncedShardsBatchTTL.Seconds()}
	req, err = req.SetBody(params)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to add body to the batch creation request")
	}

	resp, err := clientSync.Connection().Do(ctx, req)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to create batch, request failed")
	}
	if err := resp.CheckStatus(200); err != nil {
		return "", errors.Wrapf(err, "Unable to create batch, wrong status code %d", resp.StatusCode())
	}
	var batch struct {
		ID string `json:"id"`
	}

	if err := resp.ParseBody("", &batch); err != nil {
		return "", errors.Wrapf(err, "Unable to parse batch creation response")
	}
	return batch.ID, nil
}

// deleteBatch removes batch from the server
func (a *actionRebuildOutSyncedShards) deleteBatch(ctx context.Context, clientSync driver.Client, batchID string) error {
	req, err := clientSync.Connection().NewRequest("DELETE", path.Join("_api/replication/batch", batchID))
	if err != nil {
		return errors.Wrapf(err, "Unable to create request for batch removal")
	}

	resp, err := clientSync.Connection().Do(ctx, req)
	if err != nil {
		return errors.Wrapf(err, "Unable to remove batch, request failed")
	}
	if err := resp.CheckStatus(204); err != nil {
		return errors.Wrapf(err, "Unable to remove batch, wrong status code %d", resp.StatusCode())
	}
	return nil
}
