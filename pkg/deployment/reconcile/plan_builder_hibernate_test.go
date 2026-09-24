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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// singleModeHibernateDeployment builds a minimal single-mode deployment with one Single member in the
// given phase. Single mode is used so the hibernate plan builder does not touch (cluster-only)
// maintenance mode, keeping the test independent of the agency cache.
func singleModeHibernateDeployment(hibernate bool, phase api.MemberPhase) (*api.ArangoDeployment, api.DeploymentSpec, api.DeploymentStatus) {
	spec := api.DeploymentSpec{
		Mode:      api.NewMode(api.DeploymentModeSingle),
		Hibernate: util.NewType(hibernate),
	}
	spec.SetDefaults("hibernate-test")

	depl := &api.ArangoDeployment{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test_depl",
			Namespace: "test",
		},
		Spec: spec,
	}

	var status api.DeploymentStatus
	status.Members.Single = api.MemberStatusList{
		api.MemberStatus{
			ID:    "id",
			Phase: phase,
			Pod: &api.MemberPodStatus{
				Name: "something",
			},
		},
	}

	return depl, spec, status
}

// TestCreateHibernatePlan_Enable verifies that requesting hibernation on a running single-mode member
// emits a HibernateMember action for it.
func TestCreateHibernatePlan_Enable(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := newTestReconciler()
	c := newTC(t)

	depl, spec, status := singleModeHibernateDeployment(true, api.MemberPhaseCreated)

	plan := r.createHibernatePlan(ctx, depl, spec, status, c)

	require.Len(t, plan, 1)
	assert.Equal(t, api.ActionTypeHibernateMember, plan[0].Type)
	assert.Equal(t, api.ServerGroupSingle, plan[0].Group)
	assert.Equal(t, "id", plan[0].MemberID)
}

// TestCreateHibernatePlan_FullyHibernated verifies that once all members are hibernated the plan is an
// Idle action, so no other actions run while the deployment is hibernated.
func TestCreateHibernatePlan_FullyHibernated(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := newTestReconciler()
	c := newTC(t)

	depl, spec, status := singleModeHibernateDeployment(true, api.MemberPhaseHibernated)

	plan := r.createHibernatePlan(ctx, depl, spec, status, c)

	require.Len(t, plan, 1)
	assert.Equal(t, api.ActionTypeIdle, plan[0].Type)
}

// TestCreateHibernatePlan_Dehibernate verifies that clearing hibernation wakes a hibernated member up
// via a DehibernateMember action.
func TestCreateHibernatePlan_Dehibernate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := newTestReconciler()
	c := newTC(t)

	depl, spec, status := singleModeHibernateDeployment(false, api.MemberPhaseHibernated)

	plan := r.createHibernatePlan(ctx, depl, spec, status, c)

	require.Len(t, plan, 1)
	assert.Equal(t, api.ActionTypeDehibernateMember, plan[0].Type)
	assert.Equal(t, api.ServerGroupSingle, plan[0].Group)
	assert.Equal(t, "id", plan[0].MemberID)
}

// TestCreateHibernatePlan_DehibernateNoOp verifies that a running deployment without hibernated members
// produces no hibernation actions.
func TestCreateHibernatePlan_DehibernateNoOp(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := newTestReconciler()
	c := newTC(t)

	depl, spec, status := singleModeHibernateDeployment(false, api.MemberPhaseCreated)

	plan := r.createHibernatePlan(ctx, depl, spec, status, c)

	require.Empty(t, plan)
}
