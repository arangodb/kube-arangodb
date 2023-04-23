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

	"golang.org/x/exp/slices"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	// TODO make it configurable
	ttlRebuildOutSyncedShards                 time.Duration    = 600
	actionRebuildOutSyncedShardsLocalJobID    api.PlanLocalKey = "rebuildJobID"
	actionRebuildOutSyncedShardsLocalDatabase api.PlanLocalKey = "database"
	actionRebuildOutSyncedShardsLocalShard    api.PlanLocalKey = "shard"
	actionRebuildOutSyncedShardsBatchID       api.PlanLocalKey = "batchID"
)

// newRebuildOutSyncedShardsAction creates a new Action that implements the given
// planned RebuildOutSyncedShards action.
func newRebuildOutSyncedShardsAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionRecreateMember{}

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
	agencyState, ok := a.actionCtx.GetAgencyCache()
	if !ok {
		return false, errors.New("Agency cache not found")
	}

	notInSyncShards := agency.GetDBServerShardsNotInSync(agencyState, agency.Server(a.action.MemberID))
	if len(notInSyncShards) == 0 {
		a.log.Warn("Shards are in sync, action %s for member %s will BE SKIPPED!", a.action.Type, a.action.MemberID)
		return true, nil
	}

	for _, shardDetails := range notInSyncShards {
		a.action.AddParam(shardDetails.Shard, time.Now().Format(time.RFC3339))
		a.log.Info("Shard %s on member %s is not in sync, start monitoring it", shardDetails.Shard, a.action.MemberID)
	}

	return false, nil
}

// CheckProgress returns: ready, abort, error.
func (a *actionRebuildOutSyncedShards) CheckProgress(ctx context.Context) (bool, bool, error) {
	clientSync, err := a.actionCtx.GetMembersState().GetMemberClient(a.action.MemberID)
	if err != nil {
		return false, false, errors.Wrapf(err, "Unable to create client (SyncMode)")
	}

	clientAsync, err := a.actionCtx.GetServerAsyncClient(a.action.MemberID)
	if err != nil {
		return false, false, errors.Wrapf(err, "Unable to create client (AsyncMode)")
	}

	agencyState, ok := a.actionCtx.GetAgencyCache()
	if !ok {
		a.log.Error("Agency cache not found, can not check progress of action %s for member %s", a.action.Type, a.action.MemberID)
		return false, false, errors.New("Agency cache not found")
	}

	// check first if there is rebuild job running
	rebuildInProgress, err := a.checkRebuildShardProgress(ctx, clientAsync, clientSync)
	if rebuildInProgress {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}

	notInSyncShards := agency.GetDBServerShardsNotInSync(agencyState, agency.Server(a.action.MemberID))
	if len(notInSyncShards) == 0 {
		a.log.Warn("Shards are in sync, action %s for member %s is done", a.action.Type, a.action.MemberID)
		return true, false, nil
	}

	a.removeShardsWhichAreInSync(notInSyncShards)

	for _, shardDetails := range notInSyncShards {
		startTimeStr, startTimeExist := a.action.GetParam(shardDetails.Shard)
		if !startTimeExist {
			// there is new shard on the list, let's start monitoring it
			a.action.AddParam(shardDetails.Shard, time.Now().Format(time.RFC3339))
			a.log.Info("New shard %s on member %s is not in sync, start monitoring it", shardDetails.Shard, a.action.MemberID)
		} else {
			startTime, err := time.Parse(time.RFC3339, startTimeStr)
			if err != nil {
				a.log.Error("Unable to parse start time, ShardID %s, StartTime %s", shardDetails.Shard, startTimeStr)
				return false, false, errors.Newf("Unable to parse start time, ShardID %s, StartTime %s", shardDetails.Shard, startTimeStr)
			}
			if time.Since(startTime) > 15*time.Minute {
				// shard is not in sync for more than 15 minutes, let's rebuild it (only 1 shard at a time)
				return false, false, a.rebuildShard(ctx, shardDetails, clientSync, clientAsync)
			} else {
				// shard still not in sync
				a.log.Info("Shard %s on member %s is not in sync, we will keep monitoring it", shardDetails.Shard, a.action.MemberID)
			}
		}
	}

	return false, false, nil
}

func (a *actionRebuildOutSyncedShards) rebuildShard(ctx context.Context, shardDetails agency.CollectionShardDetail, clientSync, clientAsync driver.Client) error {
	batchID, err := a.createBatch(ctx, clientSync)
	if err != nil {
		return errors.Wrapf(err, "Unable to create batch")
	}

	req, err := a.createShardRebuildRequest(clientAsync, shardDetails.Shard, shardDetails.Database, batchID)
	if err != nil {
		return err
	}
	_, err = clientAsync.Connection().Do(ctx, req)
	if id, ok := conn.IsAsyncJobInProgress(err); ok {
		a.actionCtx.Add(actionRebuildOutSyncedShardsLocalJobID, id, true)
		a.actionCtx.Add(actionRebuildOutSyncedShardsLocalDatabase, shardDetails.Database, true)
		a.actionCtx.Add(actionRebuildOutSyncedShardsLocalShard, shardDetails.Shard, true)
		a.actionCtx.Add(actionRebuildOutSyncedShardsBatchID, batchID, true)
		// Async request has been sent
		return nil
	} else {
		return errors.Wrapf(err, "Unknown rebuild request error")
	}
}

// checkRebuildShardProgress returns: inProgress, error.
func (a *actionRebuildOutSyncedShards) checkRebuildShardProgress(ctx context.Context, clientAsync, clientSync driver.Client) (bool, error) {
	job, ok := a.actionCtx.Get(a.action, actionRebuildOutSyncedShardsLocalJobID)
	if !ok {
		// there is no job in progress
		return false, nil
	}

	batchID, ok := a.actionCtx.Get(a.action, actionRebuildOutSyncedShardsBatchID)
	if !ok {
		return false, errors.Newf("Local Key is missing in action: %s", actionRebuildOutSyncedShardsBatchID)
	}

	database, ok := a.actionCtx.Get(a.action, actionRebuildOutSyncedShardsLocalDatabase)
	if !ok {
		return false, errors.Newf("Local Key is missing in action: %s", actionRebuildOutSyncedShardsLocalDatabase)
	}

	shardID, ok := a.actionCtx.Get(a.action, actionRebuildOutSyncedShardsLocalShard)
	if !ok {
		return false, errors.Newf("Local Key is missing in action: %s", actionRebuildOutSyncedShardsLocalShard)
	}

	req, err := a.createShardRebuildRequest(clientAsync, shardID, database, batchID)
	if err != nil {
		return false, err
	}

	resp, err := clientAsync.Connection().Do(conn.WithAsyncID(ctx, job), req)
	if err != nil {
		if id, ok := conn.IsAsyncJobInProgress(err); ok {
			a.log.Info("Rebuild shard %s is still in progress, jobID %s", shardID, id)
			return true, nil
		} else {
			a.log.Err(err).Error("check rebuild progress error for shard %s", shardID)
			return true, errors.Wrapf(err, "check rebuild progress error")
		}
	}
	if resp.StatusCode() == http.StatusNoContent {
		a.log.Info("Rebuild shard %s is finished", shardID)

		// remove local keys
		a.actionCtx.Add(actionRebuildOutSyncedShardsLocalJobID, "", true)
		a.actionCtx.Add(actionRebuildOutSyncedShardsLocalDatabase, "", true)
		a.actionCtx.Add(actionRebuildOutSyncedShardsLocalShard, "", true)
		a.actionCtx.Add(actionRebuildOutSyncedShardsBatchID, "", true)

		_ = a.deleteBatch(ctx, clientSync, batchID)
		return false, nil
	} else {
		return false, errors.Wrapf(err, "rebuild progress failed with status code %d", resp.StatusCode())
	}
}

// removeShardsWhichAreInSync removes shards which are in sync from monitoring them
func (a *actionRebuildOutSyncedShards) removeShardsWhichAreInSync(notInSyncShards agency.CollectionShardDetails) {
	var outSyncedShardsIDs []string

	for _, shard := range notInSyncShards {
		outSyncedShardsIDs = append(outSyncedShardsIDs, shard.Shard)
	}

	for shardID := range a.action.Params {
		if !slices.Contains(outSyncedShardsIDs, shardID) {
			delete(a.action.Params, shardID)
		}
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
	}{TTL: ttlRebuildOutSyncedShards.Seconds()}
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
	req, err := clientSync.Connection().NewRequest("POST", path.Join("_api/replication/batch", batchID))
	if err != nil {
		return errors.Wrapf(err, "Unable to create request for batch removal")
	}

	_, err = clientSync.Connection().Do(ctx, req)
	return err
}
