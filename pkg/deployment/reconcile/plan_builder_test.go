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
	"errors"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/arangodb/arangosync-client/client"
	"github.com/arangodb/go-driver/agency"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	core "k8s.io/api/core/v1"
)

var _ PlanBuilderContext = &testContext{}

type testContext struct {
	Pods             []core.Pod
	ErrPods          error
	ArangoDeployment *api.ArangoDeployment
	PVC              *core.PersistentVolumeClaim
	PVCErr           error
	RecordedEvent    *k8sutil.Event
}

func (c *testContext) UpdatePvc(pvc *core.PersistentVolumeClaim) error {
	panic("implement me")
}

func (c *testContext) GetPv(pvName string) (*core.PersistentVolume, error) {
	panic("implement me")
}

func (c *testContext) GetAgencyData(ctx context.Context, i interface{}, keyParts ...string) error {
	return nil
}

func (c *testContext) GetAPIObject() k8sutil.APIObject {
	if c.ArangoDeployment == nil {
		return &api.ArangoDeployment{}
	}
	return c.ArangoDeployment
}

func (c *testContext) GetSpec() api.DeploymentSpec {
	return c.ArangoDeployment.Spec
}

func (c *testContext) UpdateStatus(status api.DeploymentStatus, lastVersion int32, force ...bool) error {
	c.ArangoDeployment.Status = status
	return nil
}

func (c *testContext) UpdateMember(member api.MemberStatus) error {
	panic("implement me")
}

func (c *testContext) GetDatabaseClient(ctx context.Context) (driver.Client, error) {
	panic("implement me")
}

func (c *testContext) GetServerClient(ctx context.Context, group api.ServerGroup, id string) (driver.Client, error) {
	panic("implement me")
}

func (c *testContext) GetAgencyClients(ctx context.Context, predicate func(id string) bool) ([]driver.Connection, error) {
	panic("implement me")
}

func (c *testContext) GetAgency(ctx context.Context) (agency.Agency, error) {
	panic("implement me")
}

func (c *testContext) GetSyncServerClient(ctx context.Context, group api.ServerGroup, id string) (client.API, error) {
	panic("implement me")
}

func (c *testContext) CreateMember(group api.ServerGroup, id string) (string, error) {
	panic("implement me")
}

func (c *testContext) DeletePod(podName string) error {
	panic("implement me")
}

func (c *testContext) DeletePvc(pvcName string) error {
	panic("implement me")
}

func (c *testContext) RemovePodFinalizers(podName string) error {
	panic("implement me")
}

func (c *testContext) GetOwnedPods() ([]core.Pod, error) {
	if c.ErrPods != nil {
		return nil, c.ErrPods
	}

	if c.Pods == nil {
		return make([]core.Pod, 0), c.ErrPods
	}
	return c.Pods, c.ErrPods
}

func (c *testContext) DeleteTLSKeyfile(group api.ServerGroup, member api.MemberStatus) error {
	panic("implement me")
}

func (c *testContext) DeleteSecret(secretName string) error {
	panic("implement me")
}

func (c *testContext) GetDeploymentHealth() (driver.ClusterHealth, error) {
	panic("implement me")
}

func (c *testContext) DisableScalingCluster() error {
	panic("implement me")
}

func (c *testContext) EnableScalingCluster() error {
	panic("implement me")
}

// GetTLSKeyfile returns the keyfile encoded TLS certificate+key for
// the given member.
func (c *testContext) GetTLSKeyfile(group api.ServerGroup, member api.MemberStatus) (string, error) {
	return "", maskAny(fmt.Errorf("Not implemented"))
}

// GetTLSCA returns the TLS CA certificate in the secret with given name.
// Returns: publicKey, privateKey, ownerByDeployment, error
func (c *testContext) GetTLSCA(secretName string) (string, string, bool, error) {
	return "", "", false, maskAny(fmt.Errorf("Not implemented"))
}

// CreateEvent creates a given event.
// On error, the error is logged.
func (c *testContext) CreateEvent(evt *k8sutil.Event) {
	c.RecordedEvent = evt
}

// GetPvc gets a PVC by the given name, in the samespace of the deployment.
func (c *testContext) GetPvc(pvcName string) (*core.PersistentVolumeClaim, error) {
	return c.PVC, c.PVCErr
}

// GetExpectedPodArguments creates command line arguments for a server in the given group with given ID.
func (c *testContext) GetExpectedPodArguments(apiObject metav1.Object, deplSpec api.DeploymentSpec, group api.ServerGroup,
	agents api.MemberStatusList, id string, version driver.Version) []string {
	return nil // not implemented
}

// GetShardSyncStatus returns true if all shards are in sync
func (c *testContext) GetShardSyncStatus() bool {
	return true
}

// InvalidateSyncStatus resets the sync state to false and triggers an inspection
func (c *testContext) InvalidateSyncStatus() {}

// GetStatus returns the current status of the deployment
func (c *testContext) GetStatus() (api.DeploymentStatus, int32) {
	return c.ArangoDeployment.Status, 0
}

func addAgentsToStatus(t *testing.T, status *api.DeploymentStatus, count int) {
	for i := 0; i < count; i++ {
		require.NoError(t, status.Members.Add(api.MemberStatus{
			ID: fmt.Sprintf("AGNT-%d", i),
			PodName: fmt.Sprintf("agnt-depl-xxx-%d", i),
			Phase: api.MemberPhaseCreated,
			Conditions: []api.Condition{
				{
					Type:   api.ConditionTypeReady,
					Status: core.ConditionTrue,
				},
			},
		}, api.ServerGroupAgents))
	}
}

// TestCreatePlanSingleScale creates a `single` deployment to test the creating of scaling plan.
func TestCreatePlanSingleScale(t *testing.T) {
	c := &testContext{}
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
	newPlan, changed := createPlan(log, depl, nil, spec, status, nil, c)
	assert.True(t, changed)
	assert.Len(t, newPlan, 0) // Single mode does not scale

	// Test with 1 single member
	status.Members.Single = api.MemberStatusList{
		api.MemberStatus{
			ID:      "id",
			PodName: "something",
		},
	}
	newPlan, changed = createPlan(log, depl, nil, spec, status, nil, c)
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
	newPlan, changed = createPlan(log, depl, nil, spec, status, nil, c)
	assert.True(t, changed)
	assert.Len(t, newPlan, 0) // Single mode does not scale
}

// TestCreatePlanActiveFailoverScale creates a `ActiveFailover` deployment to test the creating of scaling plan.
func TestCreatePlanActiveFailoverScale(t *testing.T) {
	c := &testContext{}
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
	addAgentsToStatus(t, &status, 3)

	newPlan, changed := createPlan(log, depl, nil, spec, status, nil, c)
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
	newPlan, changed = createPlan(log, depl, nil, spec, status, nil, c)
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
	newPlan, changed = createPlan(log, depl, nil, spec, status, nil, c)
	assert.True(t, changed)
	require.Len(t, newPlan, 2) // Note: Downscaling is only down 1 at a time
	assert.Equal(t, api.ActionTypeShutdownMember, newPlan[0].Type)
	assert.Equal(t, api.ActionTypeRemoveMember, newPlan[1].Type)
	assert.Equal(t, api.ServerGroupSingle, newPlan[0].Group)
	assert.Equal(t, api.ServerGroupSingle, newPlan[1].Group)
}

// TestCreatePlanClusterScale creates a `cluster` deployment to test the creating of scaling plan.
func TestCreatePlanClusterScale(t *testing.T) {
	c := &testContext{}
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
	addAgentsToStatus(t, &status, 3)

	newPlan, changed := createPlan(log, depl, nil, spec, status, nil, c)
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
	newPlan, changed = createPlan(log, depl, nil, spec, status, nil, c)
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
	newPlan, changed = createPlan(log, depl, nil, spec, status, nil, c)
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

type LastLogRecord struct {
	msg string
}

func (l *LastLogRecord) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	l.msg = msg
}

func TestCreatePlan(t *testing.T) {
	// Arrange
	threeCoordinators := api.MemberStatusList{
		{
			ID: "1",
		},
		{
			ID: "2",
		},
		{
			ID: "3",
		},
	}
	threeDBServers := api.MemberStatusList{
		{
			ID: "1",
		},
		{
			ID: "2",
		},
		{
			ID: "3",
		},
	}

	deploymentTemplate := &api.ArangoDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_depl",
			Namespace: "test",
		},
		Spec: api.DeploymentSpec{
			Mode: api.NewMode(api.DeploymentModeCluster),
			TLS: api.TLSSpec{
				CASecretName: util.NewString(api.CASecretNameDisabled),
			},
		},
		Status: api.DeploymentStatus{
			Members: api.DeploymentStatusMembers{
				DBServers:    threeDBServers,
				Coordinators: threeCoordinators,
			},
		},
	}
	addAgentsToStatus(t, &deploymentTemplate.Status, 3)
	deploymentTemplate.Spec.SetDefaults("createPlanTest")

	testCases := []struct {
		Name          string
		context       *testContext
		Helper        func(*api.ArangoDeployment)
		ExpectedError error
		ExpectedPlan  api.Plan
		ExpectedLog   string
		ExpectedEvent *k8sutil.Event
	}{
		{
			Name: "Can not get pods",
			context: &testContext{
				ErrPods: errors.New("fake error"),
			},
			ExpectedError: errors.New("fake error"),
			ExpectedLog:   "Failed to get owned pods",
		},
		{
			Name: "Can not create plan for single deployment",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.Mode = api.NewMode(api.DeploymentModeSingle)
			},
			ExpectedPlan: []api.Action{},
		},
		{
			Name: "Can not create plan for not created member",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Status.Members.DBServers[0].Phase = api.MemberPhaseNone
			},
			ExpectedPlan: []api.Action{},
		},
		{
			Name: "Can not create plan without PVC name",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Status.Members.DBServers[0].Phase = api.MemberPhaseCreated
				//ad.Status.Members.DBServers[0].PersistentVolumeClaimName = ""
			},
			ExpectedPlan: []api.Action{},
		},
		{
			Name: "Getting PVC from kubernetes failed",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
				PVCErr:           errors.New("fake error"),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Status.Members.DBServers[0].Phase = api.MemberPhaseCreated
				ad.Status.Members.DBServers[0].PersistentVolumeClaimName = "pvc_test"
			},
			ExpectedLog: "Failed to get PVC",
		},
		{
			Name: "Change Storage for DBServers",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
				PVC: &core.PersistentVolumeClaim{
					Spec: core.PersistentVolumeClaimSpec{
						StorageClassName: util.NewString("oldStorage"),
					},
				},
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.DBServers = api.ServerGroupSpec{
					Count: util.NewInt(3),
					VolumeClaimTemplate: &core.PersistentVolumeClaim{
						Spec: core.PersistentVolumeClaimSpec{
							StorageClassName: util.NewString("newStorage"),
						},
					},
				}
				ad.Status.Members.DBServers[0].Phase = api.MemberPhaseCreated
				ad.Status.Members.DBServers[0].PersistentVolumeClaimName = "pvc_test"
			},
			ExpectedPlan: []api.Action{
				api.NewAction(api.ActionTypeDisableClusterScaling, api.ServerGroupDBServers, ""),
				api.NewAction(api.ActionTypeAddMember, api.ServerGroupDBServers, ""),
				api.NewAction(api.ActionTypeWaitForMemberUp, api.ServerGroupDBServers, ""),
				api.NewAction(api.ActionTypeCleanOutMember, api.ServerGroupDBServers, ""),
				api.NewAction(api.ActionTypeShutdownMember, api.ServerGroupDBServers, ""),
				api.NewAction(api.ActionTypeRemoveMember, api.ServerGroupDBServers, ""),
				api.NewAction(api.ActionTypeEnableClusterScaling, api.ServerGroupDBServers, ""),
			},
			ExpectedLog: "Storage class has changed - pod needs replacement",
		},
		{
			Name: "Change Storage for Agents with deprecated storage class name",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
				PVC: &core.PersistentVolumeClaim{
					Spec: core.PersistentVolumeClaimSpec{
						StorageClassName: util.NewString("oldStorage"),
					},
				},
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.Agents = api.ServerGroupSpec{
					Count:            util.NewInt(2),
					StorageClassName: util.NewString("newStorage"),
				}
				ad.Status.Members.Agents[0].Phase = api.MemberPhaseCreated
				ad.Status.Members.Agents[0].PersistentVolumeClaimName = "pvc_test"
			},
			ExpectedPlan: []api.Action{
				api.NewAction(api.ActionTypeShutdownMember, api.ServerGroupAgents, ""),
				api.NewAction(api.ActionTypeRemoveMember, api.ServerGroupAgents, ""),
				api.NewAction(api.ActionTypeAddMember, api.ServerGroupAgents, ""),
				api.NewAction(api.ActionTypeWaitForMemberUp, api.ServerGroupAgents, ""),
			},
			ExpectedLog: "Storage class has changed - pod needs replacement",
		},
		{
			Name: "Storage for Coordinators is not possible",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
				PVC: &core.PersistentVolumeClaim{
					Spec: core.PersistentVolumeClaimSpec{
						StorageClassName: util.NewString("oldStorage"),
					},
				},
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.Coordinators = api.ServerGroupSpec{
					Count: util.NewInt(3),
					VolumeClaimTemplate: &core.PersistentVolumeClaim{
						Spec: core.PersistentVolumeClaimSpec{
							StorageClassName: util.NewString("newStorage"),
						},
					},
				}
				ad.Status.Members.Coordinators[0].Phase = api.MemberPhaseCreated
				ad.Status.Members.Coordinators[0].PersistentVolumeClaimName = "pvc_test"
			},
			ExpectedPlan: []api.Action{},
			ExpectedLog:  "Storage class has changed - pod needs replacement",
			ExpectedEvent: &k8sutil.Event{
				Type:    core.EventTypeNormal,
				Reason:  "Coordinator Member StorageClass Cannot Change",
				Message: "Member 1 with role coordinator should use a different StorageClass, but is cannot because: Not supported",
			},
		},
		{
			Name: "Create rotation plan",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
				PVC: &core.PersistentVolumeClaim{
					Spec: core.PersistentVolumeClaimSpec{
						StorageClassName: util.NewString("oldStorage"),
					},
					Status: core.PersistentVolumeClaimStatus{
						Conditions: []core.PersistentVolumeClaimCondition{
							{
								Type:   core.PersistentVolumeClaimFileSystemResizePending,
								Status: core.ConditionTrue,
							},
						},
					},
				},
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.Agents = api.ServerGroupSpec{
					Count: util.NewInt(2),
					VolumeClaimTemplate: &core.PersistentVolumeClaim{
						Spec: core.PersistentVolumeClaimSpec{
							StorageClassName: util.NewString("oldStorage"),
						},
					},
				}
				ad.Status.Members.Agents[0].Phase = api.MemberPhaseCreated
				ad.Status.Members.Agents[0].PersistentVolumeClaimName = "pvc_test"
			},
			ExpectedPlan: []api.Action{
				api.NewAction(api.ActionTypeRotateMember, api.ServerGroupAgents, ""),
				api.NewAction(api.ActionTypeWaitForMemberUp, api.ServerGroupAgents, ""),
			},
			ExpectedLog: "Creating rotation plan",
		},
		{
			Name: "Member in failed state",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.Agents = api.ServerGroupSpec{
					Count: util.NewInt(2),
				}
				ad.Status.Members.Agents[0].Phase = api.MemberPhaseFailed
			},
			ExpectedPlan: []api.Action{
				api.NewAction(api.ActionTypeRemoveMember, api.ServerGroupAgents, ""),
				api.NewAction(api.ActionTypeAddMember, api.ServerGroupAgents, ""),
			},
			ExpectedLog: "Creating member replacement plan because member has failed",
		},
		{
			Name: "Scale down DBservers",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.DBServers = api.ServerGroupSpec{
					Count: util.NewInt(2),
				}
				ad.Status.Members.DBServers[0].Phase = api.MemberPhaseCreated
				ad.Status.Members.DBServers[0].Conditions = api.ConditionList{
					{
						Type:   api.ConditionTypeCleanedOut,
						Status: core.ConditionTrue,
					},
				}
			},
			ExpectedPlan: []api.Action{
				api.NewAction(api.ActionTypeRemoveMember, api.ServerGroupDBServers, ""),
				api.NewAction(api.ActionTypeAddMember, api.ServerGroupDBServers, ""),
			},
			ExpectedLog: "Creating dbserver replacement plan because server is cleanout in created phase",
		},
	}

	for _, testCase := range testCases {
		//nolint:scopelint
		t.Run(testCase.Name, func(t *testing.T) {
			// Arrange
			h := &LastLogRecord{}
			logger := zerolog.New(ioutil.Discard).Hook(h)
			r := NewReconciler(logger, testCase.context)

			// Act
			if testCase.Helper != nil {
				testCase.Helper(testCase.context.ArangoDeployment)
			}
			err := r.CreatePlan()

			// Assert
			if testCase.ExpectedEvent != nil {
				require.NotNil(t, testCase.context.RecordedEvent)
				require.Equal(t, testCase.ExpectedEvent.Type, testCase.context.RecordedEvent.Type)
				require.Equal(t, testCase.ExpectedEvent.Message, testCase.context.RecordedEvent.Message)
				require.Equal(t, testCase.ExpectedEvent.Reason, testCase.context.RecordedEvent.Reason)
			}
			if len(testCase.ExpectedLog) > 0 {
				require.Equal(t, testCase.ExpectedLog, h.msg)
			}
			if testCase.ExpectedError != nil {
				assert.EqualError(t, err, testCase.ExpectedError.Error())
				return
			}

			require.NoError(t, err)
			status, _ := testCase.context.GetStatus()
			require.Len(t, status.Plan, len(testCase.ExpectedPlan))
			for i, v := range testCase.ExpectedPlan {
				assert.Equal(t, v.Type, status.Plan[i].Type)
				assert.Equal(t, v.Group, status.Plan[i].Group)
			}
		})
	}
}
