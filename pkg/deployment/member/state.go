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

package member

import (
	"context"
	"math/rand"
	"sync"

	"github.com/rs/zerolog"

	"github.com/arangodb/arangosync-client/client"
	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconciler"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

type StateInspectorGetter interface {
	GetMembersState() StateInspector
}

type StateInspector interface {
	RefreshState(ctx context.Context, members api.DeploymentStatusMemberElements)

	// GetMemberClient returns member connection to an ArangoDB server.
	GetMemberClient(id string) (driver.Client, error)

	// GetMemberSyncClient returns member connection to an ArangoSync server.
	GetMemberSyncClient(id string) (client.API, error)

	MemberState(id string) (State, bool)

	// Health returns health of members and boolean value which describes if it was possible to fetch health.
	Health() (Health, bool)

	State() State

	Log(logger logging.Logger)
}

// NewStateInspector creates a new deployment inspector.
func NewStateInspector(deployment reconciler.DeploymentGetter) StateInspector {
	return &stateInspector{
		deployment: deployment,
	}
}

// stateInspector provides cache for a deployment.
type stateInspector struct {
	// lock protects internal fields of this structure.
	lock sync.RWMutex
	// members stores information about specific members of a deployment.
	members map[string]State
	// state stores information about a deployment.
	state State
	// health stores information about healthiness of a deployment.
	health Health
	// deployment provides a deployment resources.
	deployment reconciler.DeploymentGetter
}

// Health returns health of members and true or, it returns false when fetching cluster health
// is not possible (fail-over, single).
func (s *stateInspector) Health() (Health, bool) {
	if s.health.Error == nil && s.health.Members == nil {
		// The health is not ready in the cluster mode, or it will never be ready in fail-over or single mode.
		return Health{}, false
	}

	return s.health, true
}

func (s *stateInspector) State() State {
	return s.state
}

func (s *stateInspector) Log(log logging.Logger) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for m, s := range s.members {
		if !s.IsReachable() {
			log.WrapObj(s).Str("member", m).Info("Member is in invalid state")
		}
	}
}

func (s *stateInspector) RefreshState(ctx context.Context, members api.DeploymentStatusMemberElements) {
	s.lock.Lock()
	defer s.lock.Unlock()

	var cs State
	var h Health

	results := make([]State, len(members))
	mode := s.deployment.GetMode()
	servingGroup := mode.ServingGroup()
	var client driver.Client
	members.ForEach(func(id int) {
		ctxChild, cancel := globals.GetGlobalTimeouts().ArangoDCheck().WithTimeout(ctx)
		defer cancel()

		if members[id].Group.IsArangosync() {
			results[id] = s.fetchArangosyncMemberState(ctxChild, members[id])
		} else {
			results[id] = s.fetchServerMemberState(ctxChild, members[id], servingGroup)
			if results[id].IsServing() {
				client = results[id].client
			}
		}
	})

	if client == nil {
		cs.NotReachableErr = errors.New("ArangoDB is not reachable")
	} else if !mode.IsCluster() {
		// In non-cluster mode take first client which serves.
		cs.client = client
	} else {
		// Fetch health only in cluster mode.
		h.Error = globals.GetGlobalTimeouts().ArangoDCheck().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			if cluster, err := client.Cluster(ctxChild); err != nil {
				return err
			} else if health, err := cluster.Health(ctxChild); err != nil {
				return err
			} else {
				h.Members = health.Health
			}

			// Mark members if they are healthy.
			for i, m := range members {
				if !results[i].IsReachable() {
					continue
				}

				if members[i].Group.IsArangosync() {
					// ArangoSync is considered as healthy when it is possible to fetch version.
					results[i].IsClusterHealthy = true
					continue
				}

				if v, ok := h.Members[driver.ServerID(m.Member.ID)]; ok {
					results[i].IsClusterHealthy = v.Status == driver.ServerStatusGood
					if results[i].IsServing() && v.SyncStatus == driver.ServerSyncStatusServing {
						if cs.client == nil || rand.Intn(100) > 50 {
							// Set client from nil or take next client with 50% probability.
							cs.client = results[i].client
							cs.Version = results[i].Version
						}
					}
				}
			}

			return nil
		})
	}

	if cs.client == nil {
		cs.NotReachableErr = errors.New("ArangoDB is not reachable")
	}
	current := map[string]State{}

	for id := range members {
		current[members[id].Member.ID] = results[id]
	}

	s.members = current
	s.state = cs
	s.health = h
}

func (s *stateInspector) fetchArangosyncMemberState(ctx context.Context, m api.DeploymentStatusMemberElement) State {
	var state State
	c, err := s.deployment.GetSyncServerClient(ctx, m.Group, m.Member.ID)
	if err != nil {
		state.NotReachableErr = err
		return state
	}

	if v, err := c.Version(ctx); err != nil {
		state.NotReachableErr = err
	} else {
		// convert arangosync VersionInfo to go-driver VersionInfo for simplicity:
		state.Version = driver.VersionInfo{
			Server:  m.Group.AsRole(),
			Version: driver.Version(v.Version),
			License: GetImageLicense(m.Member.Image),
			Details: map[string]interface{}{
				"arangosync-build": v.Build,
			},
		}
		state.syncClient = c
	}
	return state
}

func (s *stateInspector) fetchServerMemberState(ctx context.Context, m api.DeploymentStatusMemberElement,
	servingGroup api.ServerGroup) State {
	// by default, it is not serving. It will be changed if it serves.
	var state State
	c, err := s.deployment.GetServerClient(ctx, m.Group, m.Member.ID)
	if err != nil {
		state.NotReachableErr = err
		return state
	}

	if v, err := c.Version(ctx); err != nil {
		state.NotReachableErr = err
		return state
	} else {
		state.Version = v
		state.client = c
	}

	if m.Group == servingGroup {
		// Server belongs to group of servers which should serve requests.
		globals.GetGlobalTimeouts().ArangoDCheck().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			state.serving, _ = arangod.IsServerAvailable(ctxChild, state.client)
			return nil
		})
	}

	return state
}

// GetMemberClient returns member client to a server.
func (s *stateInspector) GetMemberClient(id string) (driver.Client, error) {
	if state, ok := s.MemberState(id); ok {
		if state.NotReachableErr != nil {
			// ArangoDB client can be set, but it might be old value.
			return nil, state.NotReachableErr
		}

		if state.client != nil {
			return state.client, nil
		}
	}

	return nil, errors.Newf("failed to get ArangoDB member client: %s", id)
}

// GetMemberSyncClient returns member client to a server.
func (s *stateInspector) GetMemberSyncClient(id string) (client.API, error) {
	if state, ok := s.MemberState(id); ok {
		if state.NotReachableErr != nil {
			// ArangoSync client can be set, but it might be old value.
			return nil, state.NotReachableErr
		}

		if state.syncClient != nil {
			return state.syncClient, nil
		}
	}

	return nil, errors.Newf("failed to get ArangoSync member client: %s", id)
}

func (s *stateInspector) MemberState(id string) (State, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.members == nil {
		return State{}, false
	}

	v, ok := s.members[id]

	return v, ok
}

// Health describes a cluster health. In the fail-over or single mode fields members and error will be nil.
// In the cluster mode only one field should be set: error or members.
type Health struct {
	// Members is a map of members of the cluster.
	Members map[driver.ServerID]driver.ServerHealth
	// Errors is set when it is not possible to fetch a cluster info.
	Error error
}

// State describes a state of a member.
type State struct {
	// NotReachableErr set to non-nil if a member is not reachable.
	NotReachableErr error
	// IsClusterHealthy describes if member is healthy in a cluster. It is relevant only in cluster mode.
	IsClusterHealthy bool
	// Version of this specific member.
	Version driver.VersionInfo
	// client to this specific ArangoDB member.
	client driver.Client
	// client to this specific ArangoSync member.
	syncClient client.API
	// serving describes if a member can serve requests.
	serving bool
}

// GetDatabaseClient returns client to the database.
func (s State) GetDatabaseClient() (driver.Client, error) {
	if s.client != nil {
		return s.client, nil
	}

	if s.NotReachableErr != nil {
		return nil, s.NotReachableErr
	}

	return nil, errors.Newf("ArangoDB is not reachable")
}

func (s State) IsReachable() bool {
	return s.NotReachableErr == nil
}

// IsServing returns true when server can serve requests.
func (s State) IsServing() bool {
	return s.serving
}

func (s State) WrapLogger(event *zerolog.Event) *zerolog.Event {
	return event.Bool("reachable", s.IsReachable()).AnErr("reachableError", s.NotReachableErr)
}
