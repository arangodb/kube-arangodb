//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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
	"strconv"
	goStrings "strings"

	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

const (
	actionTypeTopologyMemberAssignmentID        = "id"
	actionTypeTopologyMemberAssignmentZone      = "zone"
	actionTypeTopologyMemberAssignmentOperation = "operation"
	actionTypeTopologyMemberAssignmentLabel     = "label"
)

type actionTopologyMemberAssignment struct {
	actionImpl

	actionEmptyCheckProgress
}

func (t actionTopologyMemberAssignment) Start(ctx context.Context) (bool, error) {
	var add bool

	if v, ok := t.action.GetParam(actionTypeTopologyMemberAssignmentOperation); !ok {
		t.log.Str("key", actionTypeTopologyMemberAssignmentOperation).Warn("Key is missing")
		return true, nil
	} else {
		switch goStrings.ToLower(v) {
		case actionTypeTopologyMemberAssignmentOperationAdd:
			add = true
		case actionTypeTopologyMemberAssignmentOperationRemove:
			add = false
		default:
			t.log.Str("key", actionTypeTopologyMemberAssignmentOperation).Str("value", v).Warn("Unable to parse value")
		}
	}

	if add {
		return t.actionAdd(ctx)
	}

	return t.actionRemove(ctx)
}

func (t actionTopologyMemberAssignment) actionAdd(ctx context.Context) (bool, error) {
	var zone int
	var id types.UID
	var label string

	if v, ok := t.action.GetParam(actionTypeTopologyMemberAssignmentID); !ok {
		t.log.Str("key", actionTypeTopologyMemberAssignmentID).Warn("UID is missing")
		return true, nil
	} else {
		id = types.UID(v)
	}

	if v, ok := t.action.GetParam(actionTypeTopologyMemberAssignmentZone); !ok {
		t.log.Str("key", actionTypeTopologyMemberAssignmentZone).Warn("Zone is missing")
		return true, nil
	} else {
		if i, err := strconv.Atoi(v); err != nil {
			t.log.Str("key", actionTypeTopologyMemberAssignmentZone).Err(err).Warn("Unable to get int")
			return true, nil
		} else {
			zone = i
		}
	}

	if v, ok := t.action.GetParam(actionTypeTopologyMemberAssignmentLabel); ok {
		label = v
	}

	if err := t.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		m, g, ok := s.Members.ElementByID(t.action.MemberID)
		if !ok {
			t.log.Warn("Member is missing")
			return false
		}

		if !s.Topology.Enabled() {
			t.log.Warn("Cannot add topology if it is disabled")
			return false
		}

		if m.Topology != nil {
			t.log.Warn("Cannot add topology if one is assigned")
			return false
		}

		if zone < 0 || zone >= s.Topology.Size {
			t.log.Warn("Cannot add topology - it is out of range")
			return false
		}

		if s.Topology.ID != id {
			t.log.Warn("Cannot add topology - wrong id")
			return false
		}

		m.Topology = &api.TopologyMemberStatus{
			ID:    s.Topology.ID,
			Zone:  zone,
			Label: label,
		}

		if err := s.Members.Update(m, g); err != nil {
			t.log.Err(err).Warn("Cannot add topology")
			return false
		}

		s.Topology.Zones[zone].AddMember(g, m.ID)

		return true
	}); err != nil {
		t.log.Err(err).Error("Unable to propagate state of member")
		return true, nil
	}

	return true, nil
}

func (t actionTopologyMemberAssignment) actionRemove(ctx context.Context) (bool, error) {
	if err := t.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		m, g, ok := s.Members.ElementByID(t.action.MemberID)
		if !ok {
			// Check if member removal is required
			return s.Topology.RemoveMember(t.action.Group, t.action.MemberID)
		}

		if m.Topology == nil {
			// Check if member removal is required
			return s.Topology.RemoveMember(g, m.ID)
		}

		if s.Topology.IsTopologyOwned(m.Topology) {
			m.Topology = nil
			return true
		}

		s.Topology.RemoveMember(g, m.ID)

		m.Topology = nil

		if err := s.Members.Update(m, g); err != nil {
			t.log.Err(err).Warn("Cannot add topology")
			return false
		}

		return true
	}); err != nil {
		t.log.Err(err).Error("Unable to propagate state of member")
		return true, nil
	}

	return true, nil
}
