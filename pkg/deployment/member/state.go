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
	"sync"

	"github.com/rs/zerolog"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconciler"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

type StateInspectorGetter interface {
	GetMembersState() StateInspector
}

type StateInspector interface {
	RefreshState(ctx context.Context, members api.DeploymentStatusMemberElements)
	MemberState(id string) (State, bool)

	Health() Health

	State() State

	Log(logger logging.Logger)
}

func NewStateInspector(client reconciler.DeploymentClient) StateInspector {
	return &stateInspector{
		client: client,
	}
}

type stateInspector struct {
	lock sync.Mutex

	members map[string]State

	state State

	health Health

	client reconciler.DeploymentClient
}

func (s *stateInspector) Health() Health {
	return s.health
}

func (s *stateInspector) State() State {
	return s.state
}

func (s *stateInspector) Log(log logging.Logger) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for m, s := range s.members {
		if !s.IsReachable() {
			log.WrapObj(s).Str("member", m).Info("Member is in invalid state")
		}
	}
}

func (s *stateInspector) RefreshState(ctx context.Context, members api.DeploymentStatusMemberElements) {
	s.lock.Lock()
	defer s.lock.Unlock()

	results := make([]State, len(members))

	nctx, cancel := globals.GetGlobalTimeouts().ArangoDCheck().WithTimeout(ctx)
	defer cancel()

	members.ForEach(func(id int) {
		if members[id].Group.IsArangosync() {
			results[id] = s.fetchArangosyncMemberState(nctx, members[id])
		} else {
			results[id] = s.fetchServerMemberState(nctx, members[id])
		}
	})

	gctx, cancel := globals.GetGlobalTimeouts().ArangoDCheck().WithTimeout(ctx)
	defer cancel()

	var cs State
	var h Health

	c, err := s.client.GetDatabaseClient(ctx)
	if err != nil {
		cs.NotReachableErr = err
	} else {
		v, err := c.Version(gctx)
		if err != nil {
			cs.NotReachableErr = err
		} else {
			cs.Version = v
		}

		hctx, cancel := globals.GetGlobalTimeouts().ArangoDCheck().WithTimeout(ctx)
		defer cancel()
		if cluster, err := c.Cluster(hctx); err != nil {
			h.Error = err
		} else {
			if health, err := cluster.Health(hctx); err != nil {
				h.Error = err
			} else {
				h.Members = health.Health
			}
		}
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
	c, err := s.client.GetSyncServerClient(ctx, m.Group, m.Member.ID)
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
	}
	return state
}

func (s *stateInspector) fetchServerMemberState(ctx context.Context, m api.DeploymentStatusMemberElement) State {
	var state State
	c, err := s.client.GetServerClient(ctx, m.Group, m.Member.ID)
	if err != nil {
		state.NotReachableErr = err
		return state
	}

	if v, err := c.Version(ctx); err != nil {
		state.NotReachableErr = err
	} else {
		state.Version = v
	}
	return state
}

func (s *stateInspector) MemberState(id string) (State, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.members == nil {
		return State{}, false
	}

	v, ok := s.members[id]

	return v, ok
}

type Health struct {
	Members map[driver.ServerID]driver.ServerHealth

	Error error
}

type State struct {
	NotReachableErr error

	Version driver.VersionInfo
}

func (s State) IsReachable() bool {
	return s.NotReachableErr == nil
}

func (s State) WrapLogger(event *zerolog.Event) *zerolog.Event {
	return event.Bool("reachable", s.IsReachable()).AnErr("reachableError", s.NotReachableErr)
}
