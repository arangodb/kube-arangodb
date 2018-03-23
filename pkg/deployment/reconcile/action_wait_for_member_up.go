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

package reconcile

import (
	"context"
	"sync"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/rs/zerolog"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
)

// NewWaitForMemberUpAction creates a new Action that implements the given
// planned WaitForMemberUp action.
func NewWaitForMemberUpAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	return &actionWaitForMemberUp{
		log:       log,
		action:    action,
		actionCtx: actionCtx,
	}
}

const (
	maxAgentResponseTime = time.Second * 10
)

// actionWaitForMemberUp implements an WaitForMemberUp.
type actionWaitForMemberUp struct {
	log       zerolog.Logger
	action    api.Action
	actionCtx ActionContext
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionWaitForMemberUp) Start(ctx context.Context) (bool, error) {
	ready, err := a.CheckProgress(ctx)
	if err != nil {
		return false, maskAny(err)
	}
	return ready, nil
}

// CheckProgress checks the progress of the action.
// Returns true if the action is completely finished, false otherwise.
func (a *actionWaitForMemberUp) CheckProgress(ctx context.Context) (bool, error) {
	if a.action.Group.IsArangosync() {
		return a.checkProgressArangoSync(ctx)
	}
	switch a.actionCtx.GetMode() {
	case api.DeploymentModeSingle:
		return a.checkProgressSingle(ctx)
	default:
		if a.action.Group == api.ServerGroupAgents {
			return a.checkProgressAgent(ctx)
		}
		return a.checkProgressCluster(ctx)
	}
}

// checkProgressSingle checks the progress of the action in the case
// of a single server.
func (a *actionWaitForMemberUp) checkProgressSingle(ctx context.Context) (bool, error) {
	log := a.log
	c, err := a.actionCtx.GetDatabaseClient(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create database client")
		return false, maskAny(err)
	}
	if _, err := c.Version(ctx); err != nil {
		log.Debug().Err(err).Msg("Failed to get version")
		return false, maskAny(err)
	}
	return true, nil
}

type agentStatus struct {
	IsLeader       bool
	LeaderEndpoint string
	IsResponding   bool
}

// checkProgressAgent checks the progress of the action in the case
// of an agent.
func (a *actionWaitForMemberUp) checkProgressAgent(ctx context.Context) (bool, error) {
	log := a.log
	clients, err := a.actionCtx.GetAgencyClients(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create agency clients")
		return false, maskAny(err)
	}

	wg := sync.WaitGroup{}
	invalidKey := []string{"does-not-exists-149e97e8-4b81-5664-a8a8-9ba93881d64c"}
	statuses := make([]agentStatus, len(clients))
	for i, c := range clients {
		wg.Add(1)
		go func(i int, c arangod.Agency) {
			defer wg.Done()
			var trash interface{}
			lctx, cancel := context.WithTimeout(ctx, maxAgentResponseTime)
			defer cancel()
			if err := c.ReadKey(lctx, invalidKey, &trash); err == nil || arangod.IsKeyNotFound(err) {
				// We got a valid read from the leader
				statuses[i].IsLeader = true
				statuses[i].LeaderEndpoint = c.Endpoint()
				statuses[i].IsResponding = true
			} else {
				if location, ok := arangod.IsNotLeader(err); ok {
					// Valid response from a follower
					statuses[i].IsLeader = false
					statuses[i].LeaderEndpoint = location
					statuses[i].IsResponding = true
				} else {
					// Unexpected / invalid response
					log.Debug().Err(err).Str("endpoint", c.Endpoint()).Msg("Agent is not responding")
					statuses[i].IsResponding = false
				}
			}
		}(i, c)
	}
	wg.Wait()

	// Check the results
	noLeaders := 0
	for i, status := range statuses {
		if !status.IsResponding {
			log.Debug().Msg("Not all agents are responding")
			return false, nil
		}
		if status.IsLeader {
			noLeaders++
		}
		if i > 0 {
			// Compare leader endpoint with previous
			prev := statuses[i-1].LeaderEndpoint
			if !arangod.IsSameEndpoint(prev, status.LeaderEndpoint) {
				log.Debug().Msg("Not all agents report the same leader endpoint")
				return false, nil
			}
		}
	}
	if noLeaders != 1 {
		log.Debug().Int("leaders", noLeaders).Msg("Unexpected number of agency leaders")
		return false, nil
	}

	log.Debug().
		Int("leaders", noLeaders).
		Int("followers", len(statuses)-noLeaders).
		Msg("Agency is happy")

	return true, nil
}

// checkProgressCluster checks the progress of the action in the case
// of a cluster deployment (coordinator/dbserver).
func (a *actionWaitForMemberUp) checkProgressCluster(ctx context.Context) (bool, error) {
	log := a.log
	c, err := a.actionCtx.GetDatabaseClient(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create database client")
		return false, maskAny(err)
	}
	cluster, err := c.Cluster(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to access cluster")
		return false, maskAny(err)
	}
	h, err := cluster.Health(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get cluster health")
		return false, maskAny(err)
	}
	sh, found := h.Health[driver.ServerID(a.action.MemberID)]
	if !found {
		log.Debug().Msg("Member not yet found in cluster health")
		return false, nil
	}
	if sh.Status != driver.ServerStatusGood {
		log.Debug().Str("status", string(sh.Status)).Msg("Member set status not yet good")
		return false, nil
	}
	return true, nil
}

// checkProgressArangoSync checks the progress of the action in the case
// of a sync master / worker.
func (a *actionWaitForMemberUp) checkProgressArangoSync(ctx context.Context) (bool, error) {
	// TODO
	return true, nil
}
