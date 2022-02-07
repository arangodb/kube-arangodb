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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconciler"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/rs/zerolog"
)

type StateInspector interface {
	RefreshState(ctx context.Context, members api.DeploymentStatusMemberElements)
	MemberState(id string) (State, bool)

	Log(logger zerolog.Logger)
}

func NewStateInspector(client reconciler.DeploymentMemberClient) StateInspector {
	return &stateInspector{
		client: client,
	}
}

type stateInspector struct {
	lock sync.Mutex

	members map[string]State

	client reconciler.DeploymentMemberClient
}

func (s *stateInspector) Log(logger zerolog.Logger) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for m, s := range s.members {
		if s.IsInvalid() {
			s.Log(logger.Info()).Str("member", m).Msgf("Member is in invalid state")
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
		results[id] = State{}

		c, err := s.client.GetServerClient(nctx, members[id].Group, members[id].Member.ID)
		if err != nil {
			results[id].Reachable = err
			return
		}

		if _, err := c.Version(nctx); err != nil {
			results[id].Reachable = err
			return
		}
	})

	current := map[string]State{}

	for id := range members {
		current[members[id].Member.ID] = results[id]
	}

	s.members = current
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

type State struct {
	Reachable error
}

func (s State) IsReachable() bool {
	return s.Reachable == nil
}

func (s State) Log(event *zerolog.Event) *zerolog.Event {
	if !s.IsReachable() {
		event = event.Bool("reachable", false).AnErr("reachableError", s.Reachable)
	} else {
		event = event.Bool("reachable", false)
	}
	return event
}

func (s State) IsInvalid() bool {
	return !s.IsReachable()
}
