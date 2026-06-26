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
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
)

// newDBServerMember builds a MemberStatus for a DBServer with the requested conditions.
func newDBServerMember(id string, maintenanceMode, ready bool) api.MemberStatus {
	m := api.MemberStatus{ID: id}
	if maintenanceMode {
		m.Conditions = append(m.Conditions, api.Condition{
			Type:   api.ConditionTypeMemberMaintenanceMode,
			Status: core.ConditionTrue,
		})
	}
	if ready {
		m.Conditions = append(m.Conditions, api.Condition{
			Type:   api.ConditionTypeReady,
			Status: core.ConditionTrue,
		})
	}
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

	t.Run("maintenance=true ready=false emits DisableMemberMaintenance", func(t *testing.T) {
		var status api.DeploymentStatus
		addDBServer(&status, newDBServerMember("dbserver1", true, false))

		plan := r.createHighMemberMaintenanceDisablePlan(ctx, depl, spec, status, nil)

		require.Len(t, plan, 1)
		require.Equal(t, api.ActionTypeDisableMemberMaintenance, plan[0].Type)
		require.Equal(t, api.ServerGroupDBServers, plan[0].Group)
		require.Equal(t, "dbserver1", plan[0].MemberID)
	})

	t.Run("maintenance=true ready=true does not emit action", func(t *testing.T) {
		var status api.DeploymentStatus
		addDBServer(&status, newDBServerMember("dbserver1", true, true))

		plan := r.createHighMemberMaintenanceDisablePlan(ctx, depl, spec, status, nil)

		require.Empty(t, plan)
	})

	t.Run("maintenance=false ready=false does not emit action", func(t *testing.T) {
		var status api.DeploymentStatus
		addDBServer(&status, newDBServerMember("dbserver1", false, false))

		plan := r.createHighMemberMaintenanceDisablePlan(ctx, depl, spec, status, nil)

		require.Empty(t, plan)
	})

	t.Run("only unready member with maintenance gets action; ready member untouched", func(t *testing.T) {
		var status api.DeploymentStatus
		addDBServer(&status, newDBServerMember("dbserver-down", true, false))
		addDBServer(&status, newDBServerMember("dbserver-ok", true, true))

		plan := r.createHighMemberMaintenanceDisablePlan(ctx, depl, spec, status, nil)

		require.Len(t, plan, 1)
		require.Equal(t, "dbserver-down", plan[0].MemberID)
	})

	t.Run("multiple unready members each get a disable action", func(t *testing.T) {
		var status api.DeploymentStatus
		addDBServer(&status, newDBServerMember("dbserver1", true, false))
		addDBServer(&status, newDBServerMember("dbserver2", true, false))

		plan := r.createHighMemberMaintenanceDisablePlan(ctx, depl, spec, status, nil)

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

		plan := r.createHighMemberMaintenanceDisablePlan(ctx, depl, spec, status, nil)

		require.Empty(t, plan)
	})

	t.Run("no members produces empty plan", func(t *testing.T) {
		var status api.DeploymentStatus

		plan := r.createHighMemberMaintenanceDisablePlan(ctx, depl, spec, status, nil)

		require.Empty(t, plan)
	})
}
