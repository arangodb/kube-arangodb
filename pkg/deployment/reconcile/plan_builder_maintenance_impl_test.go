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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
)

// agencyContext builds a PlanBuilderContext whose agency Plan contains one shard per provided server
// list (shards[0] is the leader). Callers control the replica count per shard to model RF1 vs RF>1.
func agencyContext(shards ...state.Servers) *testContext {
	s := state.Shards{}
	for i, servers := range shards {
		s[fmt.Sprintf("s%d", i)] = servers
	}

	return &testContext{
		AgencyState: state.State{
			Plan: state.Plan{
				Collections: state.PlanCollections{
					"db": state.PlanDBCollections{
						"col": state.PlanCollection{
							Shards: s,
						},
					},
				},
			},
		},
	}
}

// leaderContext makes each given server the leader of its own failover-capable (RF2) shard.
func leaderContext(leaders ...string) *testContext {
	shards := make([]state.Servers, 0, len(leaders))
	for _, id := range leaders {
		shards = append(shards, state.Servers{state.Server(id), "follower"})
	}
	return agencyContext(shards...)
}

// rf1Context makes each given server the leader of its own single-replica (RF1) shard.
func rf1Context(leaders ...string) *testContext {
	shards := make([]state.Servers, 0, len(leaders))
	for _, id := range leaders {
		shards = append(shards, state.Servers{state.Server(id)})
	}
	return agencyContext(shards...)
}

// newDBServerMember builds a MemberStatus for a DBServer with the requested conditions.
func newDBServerMember(id string, maintenanceMode, ready bool) api.MemberStatus {
	m := api.MemberStatus{ID: id}
	if maintenanceMode {
		m.Conditions = append(m.Conditions, api.Condition{
			Type:   api.ConditionTypeMemberMaintenanceMode,
			Status: core.ConditionTrue,
		})
	}
	status := core.ConditionFalse
	if ready {
		status = core.ConditionTrue
	}
	m.Conditions = append(m.Conditions, api.Condition{
		Type:   api.ConditionTypeReady,
		Status: status,
	})
	return m
}

func Test_createHighMemberMaintenanceDisablePlan(t *testing.T) {
	ctx := context.Background()
	r := newTestReconciler()
	depl := &api.ArangoDeployment{}
	spec := api.DeploymentSpec{}

	// Ensure Version310 is enabled for all sub-tests (it is by default; be explicit).
	*features.Version310().EnabledPointer() = true
	t.Cleanup(func() { features.Version310().Reset() })

	addDBServer := func(status *api.DeploymentStatus, m api.MemberStatus) {
		if err := status.Members.Add(m, api.ServerGroupDBServers); err != nil {
			t.Fatalf("failed to add member: %v", err)
		}
	}

	t.Run("maintenance=true ready=false with active leaders emits DisableMemberMaintenance", func(t *testing.T) {
		var status api.DeploymentStatus
		addDBServer(&status, newDBServerMember("dbserver1", true, false))

		plan := r.createHighMemberMaintenanceDisablePlan(ctx, depl, spec, status, leaderContext("dbserver1"))

		require.Len(t, plan, 1)
		require.Equal(t, api.ActionTypeDisableMemberMaintenance, plan[0].Type)
		require.Equal(t, api.ServerGroupDBServers, plan[0].Group)
		require.Equal(t, "dbserver1", plan[0].MemberID)
	})

	t.Run("maintenance=true ready=false without active leaders does not emit action", func(t *testing.T) {
		var status api.DeploymentStatus
		addDBServer(&status, newDBServerMember("dbserver1", true, false))

		// dbserver1 is not Ready but hosts no leaders, so there is nothing to fail over.
		plan := r.createHighMemberMaintenanceDisablePlan(ctx, depl, spec, status, leaderContext("dbserver-other"))

		require.Empty(t, plan)
	})

	t.Run("maintenance=true ready=false leading only RF1 shard does not emit action", func(t *testing.T) {
		var status api.DeploymentStatus
		addDBServer(&status, newDBServerMember("dbserver1", true, false))

		// dbserver1 leads a single-replica shard - there is nowhere to fail over.
		plan := r.createHighMemberMaintenanceDisablePlan(ctx, depl, spec, status, rf1Context("dbserver1"))

		require.Empty(t, plan)
	})

	t.Run("maintenance=true ready=false leading both RF1 and failover-capable shards emits action", func(t *testing.T) {
		var status api.DeploymentStatus
		addDBServer(&status, newDBServerMember("dbserver1", true, false))

		// dbserver1 leads an RF1 shard AND a failover-capable one - the latter is enough to disable.
		agency := agencyContext(
			state.Servers{"dbserver1"},             // RF1
			state.Servers{"dbserver1", "follower"}, // RF2, can fail over
		)
		plan := r.createHighMemberMaintenanceDisablePlan(ctx, depl, spec, status, agency)

		require.Len(t, plan, 1)
		require.Equal(t, "dbserver1", plan[0].MemberID)
	})

	t.Run("maintenance=true ready=true does not emit action", func(t *testing.T) {
		var status api.DeploymentStatus
		addDBServer(&status, newDBServerMember("dbserver1", true, true))

		plan := r.createHighMemberMaintenanceDisablePlan(ctx, depl, spec, status, leaderContext("dbserver1"))

		require.Empty(t, plan)
	})

	t.Run("maintenance=false ready=false does not emit action", func(t *testing.T) {
		var status api.DeploymentStatus
		addDBServer(&status, newDBServerMember("dbserver1", false, false))

		plan := r.createHighMemberMaintenanceDisablePlan(ctx, depl, spec, status, leaderContext("dbserver1"))

		require.Empty(t, plan)
	})

	t.Run("only unready leader member with maintenance gets action; ready member untouched", func(t *testing.T) {
		var status api.DeploymentStatus
		addDBServer(&status, newDBServerMember("dbserver-down", true, false))
		addDBServer(&status, newDBServerMember("dbserver-ok", true, true))

		plan := r.createHighMemberMaintenanceDisablePlan(ctx, depl, spec, status, leaderContext("dbserver-down", "dbserver-ok"))

		require.Len(t, plan, 1)
		require.Equal(t, "dbserver-down", plan[0].MemberID)
	})

	t.Run("multiple unready leader members each get a disable action", func(t *testing.T) {
		var status api.DeploymentStatus
		addDBServer(&status, newDBServerMember("dbserver1", true, false))
		addDBServer(&status, newDBServerMember("dbserver2", true, false))

		plan := r.createHighMemberMaintenanceDisablePlan(ctx, depl, spec, status, leaderContext("dbserver1", "dbserver2"))

		require.Len(t, plan, 2)
		ids := []string{plan[0].MemberID, plan[1].MemberID}
		require.Contains(t, ids, "dbserver1")
		require.Contains(t, ids, "dbserver2")
	})

	t.Run("Version310 disabled returns empty plan", func(t *testing.T) {
		*features.Version310().EnabledPointer() = false
		defer func() { *features.Version310().EnabledPointer() = true }()

		var status api.DeploymentStatus
		addDBServer(&status, newDBServerMember("dbserver1", true, false))

		plan := r.createHighMemberMaintenanceDisablePlan(ctx, depl, spec, status, leaderContext("dbserver1"))

		require.Empty(t, plan)
	})

	t.Run("no members produces empty plan", func(t *testing.T) {
		var status api.DeploymentStatus

		plan := r.createHighMemberMaintenanceDisablePlan(ctx, depl, spec, status, leaderContext())

		require.Empty(t, plan)
	})

	t.Run("agency cache unavailable returns empty plan", func(t *testing.T) {
		var status api.DeploymentStatus
		addDBServer(&status, newDBServerMember("dbserver1", true, false))

		// Even a not-Ready leader with maintenance must not be touched when the agency cache is unavailable.
		plan := r.createHighMemberMaintenanceDisablePlan(ctx, depl, spec, status, &testContext{AgencyStateUnavailable: true})

		require.Empty(t, plan)
	})
}
