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
	"fmt"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// TestCreatePlanSingleScale creates a `single` deployment to test the creating of scaling plan.
func TestCreatePlanSingleScale(t *testing.T) {
	getTLSKeyfile := func(group api.ServerGroup, member api.MemberStatus) (string, error) {
		return "", maskAny(fmt.Errorf("Not implemented"))
	}
	getTLSCA := func(string) (string, string, bool, error) {
		return "", "", false, maskAny(fmt.Errorf("Not implemented"))
	}
	getPVC := func(pvcName string) (*v1.PersistentVolumeClaim, error) {
		return nil, maskAny(fmt.Errorf("Not implemented"))
	}
	createEvent := func(evt *k8sutil.Event) {}
	log := zerolog.Nop()
	spec := api.DeploymentSpec{
		Mode: api.NewMode(api.DeploymentModeSingle),
	}
	spec.SetDefaults("test")
	depl := &api.ArangoDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_depl",
			Namespace: "test",
		},
		Spec: spec,
	}

	// Test with empty status
	var status api.DeploymentStatus
	newPlan, changed := createPlan(log, depl, nil, spec, status, nil, getTLSKeyfile, getTLSCA, getPVC, createEvent)
	assert.True(t, changed)
	assert.Len(t, newPlan, 0) // Single mode does not scale

	// Test with 1 single member
	status.Members.Single = api.MemberStatusList{
		api.MemberStatus{
			ID:      "id",
			PodName: "something",
		},
	}
	newPlan, changed = createPlan(log, depl, nil, spec, status, nil, getTLSKeyfile, getTLSCA, getPVC, createEvent)
	assert.True(t, changed)
	assert.Len(t, newPlan, 0) // Single mode does not scale

	// Test with 2 single members (which should not happen) and try to scale down
	status.Members.Single = api.MemberStatusList{
		api.MemberStatus{
			ID:      "id1",
			PodName: "something1",
		},
		api.MemberStatus{
			ID:      "id1",
			PodName: "something1",
		},
	}
	newPlan, changed = createPlan(log, depl, nil, spec, status, nil, getTLSKeyfile, getTLSCA, getPVC, createEvent)
	assert.True(t, changed)
	assert.Len(t, newPlan, 0) // Single mode does not scale
}

// TestCreatePlanActiveFailoverScale creates a `ActiveFailover` deployment to test the creating of scaling plan.
func TestCreatePlanActiveFailoverScale(t *testing.T) {
	getTLSKeyfile := func(group api.ServerGroup, member api.MemberStatus) (string, error) {
		return "", maskAny(fmt.Errorf("Not implemented"))
	}
	getTLSCA := func(string) (string, string, bool, error) {
		return "", "", false, maskAny(fmt.Errorf("Not implemented"))
	}
	getPVC := func(pvcName string) (*v1.PersistentVolumeClaim, error) {
		return nil, maskAny(fmt.Errorf("Not implemented"))
	}
	createEvent := func(evt *k8sutil.Event) {}
	log := zerolog.Nop()
	spec := api.DeploymentSpec{
		Mode: api.NewMode(api.DeploymentModeActiveFailover),
	}
	spec.SetDefaults("test")
	spec.Single.Count = util.NewInt(2)
	depl := &api.ArangoDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_depl",
			Namespace: "test",
		},
		Spec: spec,
	}

	// Test with empty status
	var status api.DeploymentStatus
	newPlan, changed := createPlan(log, depl, nil, spec, status, nil, getTLSKeyfile, getTLSCA, getPVC, createEvent)
	assert.True(t, changed)
	require.Len(t, newPlan, 2)
	assert.Equal(t, api.ActionTypeAddMember, newPlan[0].Type)
	assert.Equal(t, api.ActionTypeAddMember, newPlan[1].Type)

	// Test with 1 single member
	status.Members.Single = api.MemberStatusList{
		api.MemberStatus{
			ID:      "id",
			PodName: "something",
		},
	}
	newPlan, changed = createPlan(log, depl, nil, spec, status, nil, getTLSKeyfile, getTLSCA, getPVC, createEvent)
	assert.True(t, changed)
	require.Len(t, newPlan, 1)
	assert.Equal(t, api.ActionTypeAddMember, newPlan[0].Type)
	assert.Equal(t, api.ServerGroupSingle, newPlan[0].Group)

	// Test scaling down from 4 members to 2
	status.Members.Single = api.MemberStatusList{
		api.MemberStatus{
			ID:      "id1",
			PodName: "something1",
		},
		api.MemberStatus{
			ID:      "id2",
			PodName: "something2",
		},
		api.MemberStatus{
			ID:      "id3",
			PodName: "something3",
		},
		api.MemberStatus{
			ID:      "id4",
			PodName: "something4",
		},
	}
	newPlan, changed = createPlan(log, depl, nil, spec, status, nil, getTLSKeyfile, getTLSCA, getPVC, createEvent)
	assert.True(t, changed)
	require.Len(t, newPlan, 2) // Note: Downscaling is only down 1 at a time
	assert.Equal(t, api.ActionTypeShutdownMember, newPlan[0].Type)
	assert.Equal(t, api.ActionTypeRemoveMember, newPlan[1].Type)
	assert.Equal(t, api.ServerGroupSingle, newPlan[0].Group)
	assert.Equal(t, api.ServerGroupSingle, newPlan[1].Group)
}

// TestCreatePlanClusterScale creates a `cluster` deployment to test the creating of scaling plan.
func TestCreatePlanClusterScale(t *testing.T) {
	getTLSKeyfile := func(group api.ServerGroup, member api.MemberStatus) (string, error) {
		return "", maskAny(fmt.Errorf("Not implemented"))
	}
	getTLSCA := func(string) (string, string, bool, error) {
		return "", "", false, maskAny(fmt.Errorf("Not implemented"))
	}
	getPVC := func(pvcName string) (*v1.PersistentVolumeClaim, error) {
		return nil, maskAny(fmt.Errorf("Not implemented"))
	}
	createEvent := func(evt *k8sutil.Event) {}
	log := zerolog.Nop()
	spec := api.DeploymentSpec{
		Mode: api.NewMode(api.DeploymentModeCluster),
	}
	spec.SetDefaults("test")
	depl := &api.ArangoDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_depl",
			Namespace: "test",
		},
		Spec: spec,
	}

	// Test with empty status
	var status api.DeploymentStatus
	newPlan, changed := createPlan(log, depl, nil, spec, status, nil, getTLSKeyfile, getTLSCA, getPVC, createEvent)
	assert.True(t, changed)
	require.Len(t, newPlan, 6) // Adding 3 dbservers & 3 coordinators (note: agents do not scale now)
	assert.Equal(t, api.ActionTypeAddMember, newPlan[0].Type)
	assert.Equal(t, api.ActionTypeAddMember, newPlan[1].Type)
	assert.Equal(t, api.ActionTypeAddMember, newPlan[2].Type)
	assert.Equal(t, api.ActionTypeAddMember, newPlan[3].Type)
	assert.Equal(t, api.ActionTypeAddMember, newPlan[4].Type)
	assert.Equal(t, api.ActionTypeAddMember, newPlan[5].Type)
	assert.Equal(t, api.ServerGroupDBServers, newPlan[0].Group)
	assert.Equal(t, api.ServerGroupDBServers, newPlan[1].Group)
	assert.Equal(t, api.ServerGroupDBServers, newPlan[2].Group)
	assert.Equal(t, api.ServerGroupCoordinators, newPlan[3].Group)
	assert.Equal(t, api.ServerGroupCoordinators, newPlan[4].Group)
	assert.Equal(t, api.ServerGroupCoordinators, newPlan[5].Group)

	// Test with 2 dbservers & 1 coordinator
	status.Members.DBServers = api.MemberStatusList{
		api.MemberStatus{
			ID:      "db1",
			PodName: "something1",
		},
		api.MemberStatus{
			ID:      "db2",
			PodName: "something2",
		},
	}
	status.Members.Coordinators = api.MemberStatusList{
		api.MemberStatus{
			ID:      "cr1",
			PodName: "coordinator1",
		},
	}
	newPlan, changed = createPlan(log, depl, nil, spec, status, nil, getTLSKeyfile, getTLSCA, getPVC, createEvent)
	assert.True(t, changed)
	require.Len(t, newPlan, 3)
	assert.Equal(t, api.ActionTypeAddMember, newPlan[0].Type)
	assert.Equal(t, api.ActionTypeAddMember, newPlan[1].Type)
	assert.Equal(t, api.ActionTypeAddMember, newPlan[2].Type)
	assert.Equal(t, api.ServerGroupDBServers, newPlan[0].Group)
	assert.Equal(t, api.ServerGroupCoordinators, newPlan[1].Group)
	assert.Equal(t, api.ServerGroupCoordinators, newPlan[2].Group)

	// Now scale down
	status.Members.DBServers = api.MemberStatusList{
		api.MemberStatus{
			ID:      "db1",
			PodName: "something1",
		},
		api.MemberStatus{
			ID:      "db2",
			PodName: "something2",
		},
		api.MemberStatus{
			ID:      "db3",
			PodName: "something3",
		},
	}
	status.Members.Coordinators = api.MemberStatusList{
		api.MemberStatus{
			ID:      "cr1",
			PodName: "coordinator1",
		},
		api.MemberStatus{
			ID:      "cr2",
			PodName: "coordinator2",
		},
	}
	spec.DBServers.Count = util.NewInt(1)
	spec.Coordinators.Count = util.NewInt(1)
	newPlan, changed = createPlan(log, depl, nil, spec, status, nil, getTLSKeyfile, getTLSCA, getPVC, createEvent)
	assert.True(t, changed)
	require.Len(t, newPlan, 5) // Note: Downscaling is done 1 at a time
	assert.Equal(t, api.ActionTypeCleanOutMember, newPlan[0].Type)
	assert.Equal(t, api.ActionTypeShutdownMember, newPlan[1].Type)
	assert.Equal(t, api.ActionTypeRemoveMember, newPlan[2].Type)
	assert.Equal(t, api.ActionTypeShutdownMember, newPlan[3].Type)
	assert.Equal(t, api.ActionTypeRemoveMember, newPlan[4].Type)
	assert.Equal(t, api.ServerGroupDBServers, newPlan[0].Group)
	assert.Equal(t, api.ServerGroupDBServers, newPlan[1].Group)
	assert.Equal(t, api.ServerGroupDBServers, newPlan[2].Group)
	assert.Equal(t, api.ServerGroupCoordinators, newPlan[3].Group)
	assert.Equal(t, api.ServerGroupCoordinators, newPlan[4].Group)
}
