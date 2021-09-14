//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Tomasz Mielech
//

package reconcile

import (
	"context"
	"fmt"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	monitoringClient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"

	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"

	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	policy "k8s.io/api/policy/v1beta1"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"

	"github.com/arangodb/arangosync-client/client"
	"github.com/arangodb/go-driver/agency"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	core "k8s.io/api/core/v1"
)

const pvcName = "pvc_test"

var _ PlanBuilderContext = &testContext{}
var _ Context = &testContext{}

type testContext struct {
	Pods             []core.Pod
	ErrPods          error
	ArangoDeployment *api.ArangoDeployment
	PVC              *core.PersistentVolumeClaim
	PVCErr           error
	RecordedEvent    *k8sutil.Event
}

func (c *testContext) GetKubeCli() kubernetes.Interface {
	panic("implement me")
}

func (c *testContext) GetMonitoringV1Cli() monitoringClient.MonitoringV1Interface {
	panic("implement me")
}

func (c *testContext) GetArangoCli() versioned.Interface {
	panic("implement me")
}

func (c *testContext) RenderPodForMemberFromCurrent(ctx context.Context, cachedStatus inspectorInterface.Inspector, memberID string) (*core.Pod, error) {
	panic("implement me")
}

func (c *testContext) RenderPodTemplateForMemberFromCurrent(ctx context.Context, cachedStatus inspectorInterface.Inspector, memberID string) (*core.PodTemplateSpec, error) {
	return &core.PodTemplateSpec{}, nil
}

func (c *testContext) SelectImageForMember(spec api.DeploymentSpec, status api.DeploymentStatus, member api.MemberStatus) (api.ImageInfo, bool) {
	return c.SelectImage(spec, status)
}

func (c *testContext) RenderPodTemplateForMember(ctx context.Context, cachedStatus inspectorInterface.Inspector, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.PodTemplateSpec, error) {
	panic("implement me")
}

func (c *testContext) WithArangoMemberUpdate(ctx context.Context, namespace, name string, action resources.ArangoMemberUpdateFunc) error {
	panic("implement me")
}

func (c *testContext) WithArangoMemberStatusUpdate(ctx context.Context, namespace, name string, action resources.ArangoMemberStatusUpdateFunc) error {
	panic("implement me")
}

func (c *testContext) GetAgencyMaintenanceMode(ctx context.Context) (bool, error) {
	panic("implement me")
}

func (c *testContext) SetAgencyMaintenanceMode(ctx context.Context, enabled bool) error {
	panic("implement me")
}

func (c *testContext) WithStatusUpdate(ctx context.Context, action resources.DeploymentStatusUpdateFunc, force ...bool) error {
	panic("implement me")
}

func (c *testContext) GetPod(_ context.Context, podName string) (*core.Pod, error) {
	if c.ErrPods != nil {
		return nil, c.ErrPods
	}

	for _, p := range c.Pods {
		if p.Name == podName {
			return p.DeepCopy(), nil
		}
	}

	return nil, apiErrors.NewNotFound(schema.GroupResource{}, podName)
}

func (c *testContext) GetAuthentication() conn.Auth {
	return func() (authentication driver.Authentication, err error) {
		return nil, nil
	}
}

func (c *testContext) RenderPodForMember(_ context.Context, cachedStatus inspectorInterface.Inspector, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.Pod, error) {
	return &core.Pod{}, nil
}

func (c *testContext) GetName() string {
	panic("implement me")
}

func (c *testContext) GetBackup(_ context.Context, backup string) (*backupApi.ArangoBackup, error) {
	panic("implement me")
}

func (c *testContext) SecretsInterface() k8sutil.SecretInterface {
	panic("implement me")
}

func (c *testContext) SelectImage(spec api.DeploymentSpec, status api.DeploymentStatus) (api.ImageInfo, bool) {
	return api.ImageInfo{
		Image:           "",
		ImageID:         "",
		ArangoDBVersion: "",
		Enterprise:      false,
	}, true
}

func (c *testContext) UpdatePvc(_ context.Context, pvc *core.PersistentVolumeClaim) error {
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

func (c *testContext) UpdateStatus(_ context.Context, status api.DeploymentStatus, lastVersion int32, force ...bool) error {
	c.ArangoDeployment.Status = status
	return nil
}

func (c *testContext) UpdateMember(_ context.Context, member api.MemberStatus) error {
	panic("implement me")
}

func (c *testContext) GetDatabaseClient(ctx context.Context) (driver.Client, error) {
	return nil, errors.Newf("Client Not Found")
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

func (c *testContext) CreateMember(_ context.Context, group api.ServerGroup, id string) (string, error) {
	panic("implement me")
}

func (c *testContext) DeletePod(_ context.Context, podName string) error {
	panic("implement me")
}

func (c *testContext) DeletePvc(_ context.Context, pvcName string) error {
	panic("implement me")
}

func (c *testContext) RemovePodFinalizers(_ context.Context, podName string) error {
	panic("implement me")
}

func (c *testContext) GetOwnedPods(_ context.Context) ([]core.Pod, error) {
	if c.ErrPods != nil {
		return nil, c.ErrPods
	}

	if c.Pods == nil {
		return make([]core.Pod, 0), c.ErrPods
	}
	return c.Pods, c.ErrPods
}

func (c *testContext) DeleteTLSKeyfile(_ context.Context, group api.ServerGroup, member api.MemberStatus) error {
	panic("implement me")
}

func (c *testContext) DeleteSecret(secretName string) error {
	panic("implement me")
}

func (c *testContext) GetDeploymentHealth() (driver.ClusterHealth, error) {
	panic("implement me")
}

func (c *testContext) DisableScalingCluster(_ context.Context) error {
	panic("implement me")
}

func (c *testContext) EnableScalingCluster(_ context.Context) error {
	panic("implement me")
}

// GetTLSKeyfile returns the keyfile encoded TLS certificate+key for
// the given member.
func (c *testContext) GetTLSKeyfile(group api.ServerGroup, member api.MemberStatus) (string, error) {
	return "", errors.WithStack(errors.Newf("Not implemented"))
}

// GetTLSCA returns the TLS CA certificate in the secret with given name.
// Returns: publicKey, privateKey, ownerByDeployment, error
func (c *testContext) GetTLSCA(secretName string) (string, string, bool, error) {
	return "", "", false, errors.WithStack(errors.Newf("Not implemented"))
}

// CreateEvent creates a given event.
// On error, the error is logged.
func (c *testContext) CreateEvent(evt *k8sutil.Event) {
	c.RecordedEvent = evt
}

// GetPvc gets a PVC by the given name, in the samespace of the deployment.
func (c *testContext) GetPvc(_ context.Context, pvcName string) (*core.PersistentVolumeClaim, error) {
	return c.PVC, c.PVCErr
}

// GetExpectedPodArguments creates command line arguments for a server in the given group with given ID.
func (c *testContext) GetExpectedPodArguments(apiObject meta.Object, deplSpec api.DeploymentSpec, group api.ServerGroup,
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
			ID:             fmt.Sprintf("AGNT-%d", i),
			PodName:        fmt.Sprintf("agnt-depl-xxx-%d", i),
			PodSpecVersion: "random",
			Phase:          api.MemberPhaseCreated,
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := &testContext{}
	log := zerolog.Nop()
	spec := api.DeploymentSpec{
		Mode: api.NewMode(api.DeploymentModeSingle),
	}
	spec.SetDefaults("test")
	depl := &api.ArangoDeployment{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test_depl",
			Namespace: "test",
		},
		Spec: spec,
	}

	// Test with empty status
	var status api.DeploymentStatus

	status.Hashes.JWT.Propagated = true
	status.Hashes.TLS.Propagated = true
	status.Hashes.Encryption.Propagated = true

	newPlan, changed := createNormalPlan(ctx, log, depl, nil, spec, status, inspector.NewEmptyInspector(), c)
	assert.True(t, changed)
	assert.Len(t, newPlan, 0) // Single mode does not scale

	// Test with 1 single member
	status.Members.Single = api.MemberStatusList{
		api.MemberStatus{
			ID:      "id",
			PodName: "something",
		},
	}
	newPlan, changed = createNormalPlan(ctx, log, depl, nil, spec, status, inspector.NewEmptyInspector(), c)
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
	newPlan, changed = createNormalPlan(ctx, log, depl, nil, spec, status, inspector.NewEmptyInspector(), c)
	assert.True(t, changed)
	assert.Len(t, newPlan, 0) // Single mode does not scale
}

// TestCreatePlanActiveFailoverScale creates a `ActiveFailover` deployment to test the creating of scaling plan.
func TestCreatePlanActiveFailoverScale(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := &testContext{}
	log := zerolog.Nop()
	spec := api.DeploymentSpec{
		Mode: api.NewMode(api.DeploymentModeActiveFailover),
	}
	spec.SetDefaults("test")
	spec.Single.Count = util.NewInt(2)
	depl := &api.ArangoDeployment{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test_depl",
			Namespace: "test",
		},
		Spec: spec,
	}

	// Test with empty status
	var status api.DeploymentStatus
	addAgentsToStatus(t, &status, 3)

	newPlan, changed := createNormalPlan(ctx, log, depl, nil, spec, status, inspector.NewEmptyInspector(), c)
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
	newPlan, changed = createNormalPlan(ctx, log, depl, nil, spec, status, inspector.NewEmptyInspector(), c)
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
	newPlan, changed = createNormalPlan(ctx, log, depl, nil, spec, status, inspector.NewEmptyInspector(), c)
	assert.True(t, changed)
	require.Len(t, newPlan, 2) // Note: Downscaling is only down 1 at a time
	assert.Equal(t, api.ActionTypeShutdownMember, newPlan[0].Type)
	assert.Equal(t, api.ActionTypeRemoveMember, newPlan[1].Type)
	assert.Equal(t, api.ServerGroupSingle, newPlan[0].Group)
	assert.Equal(t, api.ServerGroupSingle, newPlan[1].Group)
}

// TestCreatePlanClusterScale creates a `cluster` deployment to test the creating of scaling plan.
func TestCreatePlanClusterScale(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := &testContext{}
	log := zerolog.Nop()
	spec := api.DeploymentSpec{
		Mode: api.NewMode(api.DeploymentModeCluster),
	}
	spec.SetDefaults("test")
	depl := &api.ArangoDeployment{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test_depl",
			Namespace: "test",
		},
		Spec: spec,
	}

	// Test with empty status
	var status api.DeploymentStatus
	addAgentsToStatus(t, &status, 3)

	newPlan, changed := createNormalPlan(ctx, log, depl, nil, spec, status, inspector.NewEmptyInspector(), c)
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
	newPlan, changed = createNormalPlan(ctx, log, depl, nil, spec, status, inspector.NewEmptyInspector(), c)
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
	newPlan, changed = createNormalPlan(ctx, log, depl, nil, spec, status, inspector.NewEmptyInspector(), c)
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

type testCase struct {
	Name             string
	context          *testContext
	Helper           func(*api.ArangoDeployment)
	ExpectedError    error
	ExpectedPlan     api.Plan
	ExpectedHighPlan api.Plan
	ExpectedLog      string
	ExpectedEvent    *k8sutil.Event

	Pods            map[string]*core.Pod
	Secrets         map[string]*core.Secret
	Services        map[string]*core.Service
	PVCS            map[string]*core.PersistentVolumeClaim
	ServiceAccounts map[string]*core.ServiceAccount
	PDBS            map[string]*policy.PodDisruptionBudget
	ServiceMonitors map[string]*monitoring.ServiceMonitor
	ArangoMembers   map[string]*api.ArangoMember

	Extender func(t *testing.T, r *Reconciler, c *testCase)
}

func (t testCase) Inspector() inspectorInterface.Inspector {
	return inspector.NewInspectorFromData(t.Pods, t.Secrets, t.PVCS, t.Services, t.ServiceAccounts, t.PDBS, t.ServiceMonitors, t.ArangoMembers)
}

func TestCreatePlan(t *testing.T) {
	// Arrange
	threeCoordinators := api.MemberStatusList{
		{
			ID:      "1",
			PodName: "coordinator1",
		},
		{
			ID:      "2",
			PodName: "coordinator2",
		},
		{
			ID:      "3",
			PodName: "coordinator3",
		},
	}
	threeDBServers := api.MemberStatusList{
		{
			ID:      "1",
			PodName: "dbserver1",
		},
		{
			ID:      "2",
			PodName: "dbserver2",
		},
		{
			ID:      "3",
			PodName: "dbserver3",
		},
	}

	deploymentTemplate := &api.ArangoDeployment{
		ObjectMeta: meta.ObjectMeta{
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

	testCases := []testCase{
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
			Name: "Change Storage for DBServers",
			PVCS: map[string]*core.PersistentVolumeClaim{
				pvcName: {
					Spec: core.PersistentVolumeClaimSpec{
						StorageClassName: util.NewString("oldStorage"),
					},
				},
			},
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
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
				ad.Status.Members.DBServers[0].PersistentVolumeClaimName = pvcName
			},
			ExpectedPlan: []api.Action{
				api.NewAction(api.ActionTypeMarkToRemoveMember, api.ServerGroupDBServers, ""),
			},
			ExpectedLog: "Storage class has changed - pod needs replacement",
		},
		{
			Name: "Change Storage for Agents with deprecated storage class name",
			PVCS: map[string]*core.PersistentVolumeClaim{
				pvcName: {
					Spec: core.PersistentVolumeClaimSpec{
						StorageClassName: util.NewString("oldStorage"),
					},
				},
			},
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.Agents = api.ServerGroupSpec{
					Count:            util.NewInt(2),
					StorageClassName: util.NewString("newStorage"),
				}
				ad.Status.Members.Agents[0].Phase = api.MemberPhaseCreated
				ad.Status.Members.Agents[0].PersistentVolumeClaimName = pvcName
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
			PVCS: map[string]*core.PersistentVolumeClaim{
				pvcName: {
					Spec: core.PersistentVolumeClaimSpec{
						StorageClassName: util.NewString("oldStorage"),
					},
				},
			},
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
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
				ad.Status.Members.Coordinators[0].PersistentVolumeClaimName = pvcName
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
			PVCS: map[string]*core.PersistentVolumeClaim{
				"pvc_test": {
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
			Extender: func(t *testing.T, r *Reconciler, c *testCase) {
				// Add ArangoMember
				builderCtx := newPlanBuilderContext(r.context)

				template, err := builderCtx.RenderPodTemplateForMemberFromCurrent(context.Background(), c.Inspector(), c.context.ArangoDeployment.Status.Members.Agents[0].ID)
				require.NoError(t, err)

				checksum, err := resources.ChecksumArangoPod(c.context.ArangoDeployment.Spec.Agents, resources.CreatePodFromTemplate(template))
				require.NoError(t, err)

				templateSpec, err := api.GetArangoMemberPodTemplate(template, checksum)
				require.NoError(t, err)

				name := c.context.ArangoDeployment.Status.Members.Agents[0].ArangoMemberName(c.context.ArangoDeployment.Name, api.ServerGroupAgents)

				c.ArangoMembers = map[string]*api.ArangoMember{
					name: {
						ObjectMeta: meta.ObjectMeta{
							Name: name,
						},
						Spec: api.ArangoMemberSpec{
							Template: templateSpec,
						},
						Status: api.ArangoMemberStatus{
							Template: templateSpec,
						},
					},
				}

				c.Pods = map[string]*core.Pod{
					c.context.ArangoDeployment.Status.Members.Agents[0].PodName: {
						ObjectMeta: meta.ObjectMeta{
							Name: c.context.ArangoDeployment.Status.Members.Agents[0].PodName,
						},
					},
				}

				require.NoError(t, c.context.ArangoDeployment.Status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
					for _, m := range list {
						m.Phase = api.MemberPhaseCreated
						require.NoError(t, c.context.ArangoDeployment.Status.Members.Update(m, group))
					}

					return nil
				}))
			},
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
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
			ExpectedHighPlan: []api.Action{
				api.NewAction(api.ActionTypeSetMemberCondition, api.ServerGroupAgents, deploymentTemplate.Status.Members.Agents[0].ID, "PVC Resize pending"),
			},
			ExpectedLog: "PVC Resize pending",
		},
		{
			Name: "Agent in failed state",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.Agents = api.ServerGroupSpec{
					Count: util.NewInt(2),
				}
				ad.Status.Members.Agents[0].Phase = api.MemberPhaseFailed
				ad.Status.Members.Agents[0].ID = "id"
			},
			ExpectedPlan: []api.Action{
				api.NewAction(api.ActionTypeRecreateMember, api.ServerGroupAgents, "id"),
			},
			ExpectedLog: "Restoring old member. For agency members recreation of PVC is not supported - to prevent DataLoss",
		},
		{
			Name: "Coordinator in failed state",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.Coordinators = api.ServerGroupSpec{
					Count: util.NewInt(2),
				}
				ad.Status.Members.Coordinators[0].Phase = api.MemberPhaseFailed
				ad.Status.Members.Coordinators[0].ID = "id"
			},
			ExpectedPlan: []api.Action{
				api.NewAction(api.ActionTypeRemoveMember, api.ServerGroupCoordinators, "id"),
				api.NewAction(api.ActionTypeAddMember, api.ServerGroupCoordinators, ""),
				api.NewAction(api.ActionTypeWaitForMemberUp, api.ServerGroupCoordinators, api.MemberIDPreviousAction),
			},
			ExpectedLog: "Creating member replacement plan because member has failed",
		},
		{
			Name: "DBServer in failed state",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.DBServers = api.ServerGroupSpec{
					Count: util.NewInt(2),
				}
				ad.Status.Members.DBServers[0].Phase = api.MemberPhaseFailed
				ad.Status.Members.DBServers[0].ID = "id"
			},
			ExpectedPlan: []api.Action{
				api.NewAction(api.ActionTypeRemoveMember, api.ServerGroupDBServers, "id"),
				api.NewAction(api.ActionTypeAddMember, api.ServerGroupDBServers, ""),
				api.NewAction(api.ActionTypeWaitForMemberUp, api.ServerGroupDBServers, api.MemberIDPreviousAction),
			},
			ExpectedLog: "Creating member replacement plan because member has failed",
		},
		{
			Name: "CleanOut DBserver",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.DBServers = api.ServerGroupSpec{
					Count: util.NewInt(3),
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
				api.NewAction(api.ActionTypeCleanOutMember, api.ServerGroupDBServers, "id"),
				api.NewAction(api.ActionTypeShutdownMember, api.ServerGroupDBServers, ""),
				api.NewAction(api.ActionTypeRemoveMember, api.ServerGroupDBServers, ""),
			},
			ExpectedLog: "Creating dbserver replacement plan because server is cleanout in created phase",
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
			},
			ExpectedPlan: []api.Action{
				api.NewAction(api.ActionTypeCleanOutMember, api.ServerGroupDBServers, "id"),
				api.NewAction(api.ActionTypeShutdownMember, api.ServerGroupDBServers, ""),
				api.NewAction(api.ActionTypeRemoveMember, api.ServerGroupDBServers, ""),
			},
			ExpectedLog: "Creating scale-down plan",
		},
	}

	for _, testCase := range testCases {
		//nolint:scopelint
		t.Run(testCase.Name, func(t *testing.T) {
			// Arrange
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := &LastLogRecord{}
			logger := zerolog.New(ioutil.Discard).Hook(h)
			r := NewReconciler(logger, testCase.context)

			if testCase.Extender != nil {
				testCase.Extender(t, r, &testCase)
			}

			// Act
			if testCase.Helper != nil {
				testCase.Helper(testCase.context.ArangoDeployment)
			}

			err, _ := r.CreatePlan(ctx, testCase.Inspector())

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

			if len(testCase.ExpectedHighPlan) > 0 {
				require.Len(t, status.HighPriorityPlan, len(testCase.ExpectedHighPlan))
				for i, v := range testCase.ExpectedHighPlan {
					assert.Equal(t, v.Type, status.HighPriorityPlan[i].Type)
					assert.Equal(t, v.Group, status.HighPriorityPlan[i].Group)
					if v.Reason != "*" {
						assert.Equal(t, v.Reason, status.HighPriorityPlan[i].Reason)
					}
				}
			}

			require.Len(t, status.Plan, len(testCase.ExpectedPlan))
			for i, v := range testCase.ExpectedPlan {
				assert.Equal(t, v.Type, status.Plan[i].Type)
				assert.Equal(t, v.Group, status.Plan[i].Group)
				if v.Reason != "*" {
					assert.Equal(t, v.Reason, status.Plan[i].Reason)
				}
			}
		})
	}
}
