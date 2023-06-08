//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	"io"
	"testing"

	monitoringClient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/arangosync-client/client"
	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/agency"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs/sutil"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	agencyCache "github.com/arangodb/kube-arangodb/pkg/deployment/agency"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/deployment/member"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	pod2 "github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconciler"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	arangomemberv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangomember/v1"
	persistentvolumeclaimv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim/v1"
	podv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod/v1"
	poddisruptionbudgetv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/poddisruptionbudget/v1"
	secretv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret/v1"
	servicev1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service/v1"
	serviceaccountv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/serviceaccount/v1"
	servicemonitorv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/servicemonitor/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
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

	Inspector inspectorInterface.Inspector
	state     member.StateInspector
}

func (c *testContext) IsSyncEnabled() bool {
	return false
}

func (c *testContext) GetAgencyArangoDBCache() (state.DB, bool) {
	//TODO implement me
	panic("implement me")
}

func (c *testContext) WithMemberStatusUpdateErr(ctx context.Context, id string, group api.ServerGroup, action reconciler.DeploymentMemberStatusUpdateErrFunc) error {
	//TODO implement me
	panic("implement me")
}

func (c *testContext) WithMemberStatusUpdate(ctx context.Context, id string, group api.ServerGroup, action reconciler.DeploymentMemberStatusUpdateFunc) error {
	//TODO implement me
	panic("implement me")
}

func (c *testContext) CreateOperatorEngineOpsAlertEvent(message string, args ...interface{}) {
	//TODO implement me
	panic("implement me")
}

func (c *testContext) GetAgencyHealth() (agencyCache.Health, bool) {
	//TODO implement me
	panic("implement me")
}

func (c *testContext) ShardsInSyncMap() (state.ShardsSyncStatus, bool) {
	//TODO implement me
	panic("implement me")
}

func (c *testContext) RenderPodForMember(ctx context.Context, acs sutil.ACS, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.Pod, error) {
	//TODO implement me
	panic("implement me")
}

func (c *testContext) RenderPodTemplateForMember(ctx context.Context, acs sutil.ACS, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.PodTemplateSpec, error) {
	return &core.PodTemplateSpec{}, nil
}

func (c *testContext) ACS() sutil.ACS {
	return acs.NewACS("", c.Inspector)
}

func (c *testContext) GetDatabaseAsyncClient(ctx context.Context) (driver.Client, error) {
	//TODO implement me
	panic("implement me")
}

func (ac *testContext) GetServerAsyncClient(id string) (driver.Client, error) {
	//TODO implement me
	panic("implement me")
}

func (c *testContext) GetMembersState() member.StateInspector {
	return c.state
}

func (c *testContext) GetMode() api.DeploymentMode {
	//TODO implement me
	panic("implement me")
}

func (c *testContext) GetNamespace() string {
	//TODO implement me
	panic("implement me")
}

func (c *testContext) ApplyPatchOnPod(ctx context.Context, pod *core.Pod, p ...patch.Item) error {
	panic("implement me")
}

func (c *testContext) ApplyPatch(ctx context.Context, p ...patch.Item) error {
	panic("implement me")
}

func (c *testContext) GetStatusSnapshot() api.DeploymentStatus {
	s := c.GetStatus()
	return *s.DeepCopy()
}

func (c *testContext) GenerateMemberEndpoint(group api.ServerGroup, member api.MemberStatus) (string, error) {
	return pod2.GenerateMemberEndpoint(c.Inspector, c.ArangoDeployment, c.ArangoDeployment.Spec, group, member)
}

func (c *testContext) GetAgencyCache() (state.State, bool) {
	return state.State{}, true
}

func (c *testContext) SecretsModInterface() secretv1.ModInterface {
	panic("implement me")
}

func (c *testContext) PodsModInterface() podv1.ModInterface {
	panic("implement me")
}

func (c *testContext) ServiceAccountsModInterface() serviceaccountv1.ModInterface {
	panic("implement me")
}

func (c *testContext) ServicesModInterface() servicev1.ModInterface {
	panic("implement me")
}

func (c *testContext) PersistentVolumeClaimsModInterface() persistentvolumeclaimv1.ModInterface {
	panic("implement me")
}

func (c *testContext) PodDisruptionBudgetsModInterface() poddisruptionbudgetv1.ModInterface {
	panic("implement me")
}

func (c *testContext) ServiceMonitorsModInterface() servicemonitorv1.ModInterface {
	panic("implement me")
}

func (c *testContext) ArangoMembersModInterface() arangomemberv1.ModInterface {
	panic("implement me")
}

func (c *testContext) WithStatusUpdateErr(ctx context.Context, action reconciler.DeploymentStatusUpdateErrFunc) error {
	_, err := action(&c.ArangoDeployment.Status)
	return err
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

func (c *testContext) SelectImageForMember(spec api.DeploymentSpec, status api.DeploymentStatus, member api.MemberStatus) (api.ImageInfo, bool) {
	return c.SelectImage(spec, status)
}

func (c *testContext) GetAgencyMaintenanceMode(ctx context.Context) (bool, error) {
	panic("implement me")
}

func (c *testContext) SetAgencyMaintenanceMode(ctx context.Context, enabled bool) error {
	panic("implement me")
}

func (c *testContext) WithStatusUpdate(ctx context.Context, action reconciler.DeploymentStatusUpdateFunc) error {
	action(&c.ArangoDeployment.Status)
	return nil
}

func (c *testContext) GetAuthentication() conn.Auth {
	return func() (authentication driver.Authentication, err error) {
		return nil, nil
	}
}

func (c *testContext) GetName() string {
	return "name"
}

func (c *testContext) GetBackup(_ context.Context, backup string) (*backupApi.ArangoBackup, error) {
	panic("implement me")
}

func (c *testContext) SecretsInterface() secretv1.Interface {
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

func (c *testContext) UpdateStatus(_ context.Context, status api.DeploymentStatus) error {
	c.ArangoDeployment.Status = status
	return nil
}

func (c *testContext) UpdateMember(_ context.Context, member api.MemberStatus) error {
	panic("implement me")
}

func (c *testContext) GetAgency(_ context.Context, _ ...string) (agency.Agency, error) {
	panic("implement me")
}

func (c *testContext) CreateMember(_ context.Context, group api.ServerGroup, id string, mods ...CreateMemberMod) (string, error) {
	panic("implement me")
}

func (c *testContext) DeletePod(_ context.Context, _ string, _ meta.DeleteOptions) error {
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

// GetStatus returns the current status of the deployment
func (c *testContext) GetStatus() api.DeploymentStatus {
	return c.ArangoDeployment.Status
}

func addAgentsToStatus(t *testing.T, status *api.DeploymentStatus, count int) {
	for i := 0; i < count; i++ {
		require.NoError(t, status.Members.Add(api.MemberStatus{
			ID: fmt.Sprintf("AGNT-%d", i),
			Pod: &api.MemberPodStatus{
				Name:        fmt.Sprintf("agnt-depl-xxx-%d", i),
				SpecVersion: "random",
			},
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

func newTC(t *testing.T) *testContext {
	return &testContext{
		Inspector: tests.NewEmptyInspector(t),
	}
}

// TestCreatePlanSingleScale creates a `single` deployment to test the creating of scaling plan.
func TestCreatePlanSingleScale(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := newTestReconciler()

	c := newTC(t)
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

	newPlan, _, changed := r.createNormalPlan(ctx, depl, nil, spec, status, c)
	assert.True(t, changed)
	assert.Len(t, newPlan, 1)

	// Test with 1 single member
	status.Members.Single = api.MemberStatusList{
		api.MemberStatus{
			ID: "id",
			Pod: &api.MemberPodStatus{
				Name: "something",
			},
		},
	}
	newPlan, _, changed = r.createNormalPlan(ctx, depl, nil, spec, status, c)
	assert.True(t, changed)
	assert.Len(t, newPlan, 0) // Single mode does not scale

	spec.Single.Count = util.NewType[int](2)
	newPlan, _, changed = r.createNormalPlan(ctx, depl, nil, spec, status, c)
	assert.True(t, changed)
	assert.Len(t, newPlan, 0) // Single mode does not scale

	spec.Single.Count = util.NewType[int](1)
	// Test with 2 single members (which should not happen) and try to scale down
	status.Members.Single = api.MemberStatusList{
		api.MemberStatus{
			ID: "id1",
			Pod: &api.MemberPodStatus{
				Name: "something1",
			},
		},
		api.MemberStatus{
			ID: "id1",
			Pod: &api.MemberPodStatus{
				Name: "something1",
			},
		},
	}
	newPlan, _, changed = r.createNormalPlan(ctx, depl, nil, spec, status, c)
	assert.True(t, changed)
	assert.Len(t, newPlan, 0) // Single mode does not scale down
}

// TestCreatePlanActiveFailoverScale creates a `ActiveFailover` deployment to test the creating of scaling plan.
func TestCreatePlanActiveFailoverScale(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := newTC(t)
	r := newTestReconciler()
	spec := api.DeploymentSpec{
		Mode: api.NewMode(api.DeploymentModeActiveFailover),
	}
	spec.SetDefaults("test")
	spec.Single.Count = util.NewType[int](2)
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

	newPlan, _, changed := r.createNormalPlan(ctx, depl, nil, spec, status, c)
	assert.True(t, changed)
	require.Len(t, newPlan, 2)
	assert.Equal(t, api.ActionTypeAddMember, newPlan[0].Type)
	assert.Equal(t, api.ActionTypeAddMember, newPlan[1].Type)

	// Test with 1 single member
	status.Members.Single = api.MemberStatusList{
		api.MemberStatus{
			ID: "id",
			Pod: &api.MemberPodStatus{
				Name: "something",
			},
		},
	}
	newPlan, _, changed = r.createNormalPlan(ctx, depl, nil, spec, status, c)
	assert.True(t, changed)
	require.Len(t, newPlan, 1)
	assert.Equal(t, api.ActionTypeAddMember, newPlan[0].Type)
	assert.Equal(t, api.ServerGroupSingle, newPlan[0].Group)

	// Test scaling down from 4 members to 2
	status.Members.Single = api.MemberStatusList{
		api.MemberStatus{
			ID: "id1",
			Pod: &api.MemberPodStatus{
				Name: "something1",
			},
		},
		api.MemberStatus{
			ID: "id2",
			Pod: &api.MemberPodStatus{
				Name: "something2",
			},
		},
		api.MemberStatus{
			ID: "id3",
			Pod: &api.MemberPodStatus{
				Name: "something3",
			},
		},
		api.MemberStatus{
			ID: "id4",
			Pod: &api.MemberPodStatus{
				Name: "something4",
			},
		},
	}
	newPlan, _, changed = r.createNormalPlan(ctx, depl, nil, spec, status, c)
	assert.True(t, changed)
	require.Len(t, newPlan, 3) // Note: Downscaling is only down 1 at a time
	assert.Equal(t, api.ActionTypeKillMemberPod, newPlan[0].Type)
	assert.Equal(t, api.ActionTypeShutdownMember, newPlan[1].Type)
	assert.Equal(t, api.ActionTypeRemoveMember, newPlan[2].Type)
	assert.Equal(t, api.ServerGroupSingle, newPlan[0].Group)
	assert.Equal(t, api.ServerGroupSingle, newPlan[1].Group)
	assert.Equal(t, api.ServerGroupSingle, newPlan[2].Group)
}

// TestCreatePlanClusterScale creates a `cluster` deployment to test the creating of scaling plan.
func TestCreatePlanClusterScale(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := newTC(t)
	r := newTestReconciler()
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

	newPlan, _, changed := r.createNormalPlan(ctx, depl, nil, spec, status, c)
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
			ID: "db1",
			Pod: &api.MemberPodStatus{
				Name: "something1",
			},
		},
		api.MemberStatus{
			ID: "db2",
			Pod: &api.MemberPodStatus{
				Name: "something2",
			},
		},
	}
	status.Members.Coordinators = api.MemberStatusList{
		api.MemberStatus{
			ID: "cr1",
			Pod: &api.MemberPodStatus{
				Name: "coordinator1",
			},
		},
	}
	newPlan, _, changed = r.createNormalPlan(ctx, depl, nil, spec, status, c)
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
			ID: "db1",
			Pod: &api.MemberPodStatus{
				Name: "something1",
			},
		},
		api.MemberStatus{
			ID: "db2",
			Pod: &api.MemberPodStatus{
				Name: "something2",
			},
		},
		api.MemberStatus{
			ID: "db3",
			Pod: &api.MemberPodStatus{
				Name: "something3",
			},
		},
	}
	status.Members.Coordinators = api.MemberStatusList{
		api.MemberStatus{
			ID: "cr1",
			Pod: &api.MemberPodStatus{
				Name: "coordinator1",
			},
		},
		api.MemberStatus{
			ID: "cr2",
			Pod: &api.MemberPodStatus{
				Name: "coordinator2",
			},
		},
	}
	spec.DBServers.Count = util.NewType[int](1)
	spec.Coordinators.Count = util.NewType[int](1)
	newPlan, _, changed = r.createNormalPlan(ctx, depl, nil, spec, status, c)
	assert.True(t, changed)
	require.Len(t, newPlan, 7) // Note: Downscaling is done 1 at a time
	assert.Equal(t, api.ActionTypeCleanOutMember, newPlan[0].Type)
	assert.Equal(t, api.ActionTypeKillMemberPod, newPlan[1].Type)
	assert.Equal(t, api.ActionTypeShutdownMember, newPlan[2].Type)
	assert.Equal(t, api.ActionTypeRemoveMember, newPlan[3].Type)
	assert.Equal(t, api.ActionTypeKillMemberPod, newPlan[4].Type)
	assert.Equal(t, api.ActionTypeShutdownMember, newPlan[5].Type)
	assert.Equal(t, api.ActionTypeRemoveMember, newPlan[6].Type)
	assert.Equal(t, api.ServerGroupDBServers, newPlan[0].Group)
	assert.Equal(t, api.ServerGroupDBServers, newPlan[1].Group)
	assert.Equal(t, api.ServerGroupDBServers, newPlan[2].Group)
	assert.Equal(t, api.ServerGroupDBServers, newPlan[3].Group)
	assert.Equal(t, api.ServerGroupCoordinators, newPlan[4].Group)
	assert.Equal(t, api.ServerGroupCoordinators, newPlan[5].Group)
	assert.Equal(t, api.ServerGroupCoordinators, newPlan[6].Group)
}

type LastLogRecord struct {
	t   *testing.T
	msg string
}

func (l *LastLogRecord) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	l.t.Log(msg)
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

	kclient.FakeDataInput
	Extender func(t *testing.T, r *Reconciler, c *testCase)
}

func (t testCase) Inspector(test *testing.T) inspectorInterface.Inspector {
	t.FakeDataInput.Namespace = tests.FakeNamespace
	return tests.NewInspector(test, t.FakeDataInput.Client())
}

func TestCreatePlan(t *testing.T) {
	// Arrange
	threeCoordinators := api.MemberStatusList{
		{
			ID: "1",
			Pod: &api.MemberPodStatus{
				Name: "coordinator1",
			},
		},
		{
			ID: "2",
			Pod: &api.MemberPodStatus{
				Name: "coordinator2",
			},
		},
		{
			ID: "3",
			Pod: &api.MemberPodStatus{
				Name: "coordinator3",
			},
		},
	}
	threeDBServers := api.MemberStatusList{
		{
			ID: "1",
			Pod: &api.MemberPodStatus{
				Name: "dbserver1",
			},
		},
		{
			ID: "2",
			Pod: &api.MemberPodStatus{
				Name: "dbserver2",
			},
		},
		{
			ID: "3",
			Pod: &api.MemberPodStatus{
				Name: "dbserver3",
			},
		},
	}

	deploymentTemplate := &api.ArangoDeployment{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test_depl",
			Namespace: tests.FakeNamespace,
		},
		Spec: api.DeploymentSpec{
			Mode: api.NewMode(api.DeploymentModeCluster),
			TLS: api.TLSSpec{
				CASecretName: util.NewType[string](api.CASecretNameDisabled),
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
				ad.Status.Members.Single = append(ad.Status.Members.Single, api.MemberStatus{})
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
			},
			ExpectedPlan: []api.Action{},
		},
		{
			Name: "Change Storage for DBServers",
			FakeDataInput: kclient.FakeDataInput{
				PVCS: map[string]*core.PersistentVolumeClaim{
					pvcName: {
						Spec: core.PersistentVolumeClaimSpec{
							StorageClassName: util.NewType[string]("oldStorage"),
						},
					},
				},
				Pods: map[string]*core.Pod{
					"dbserver1": {},
				},
			},
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.DBServers = api.ServerGroupSpec{
					Count: util.NewType[int](3),
					VolumeClaimTemplate: &core.PersistentVolumeClaim{
						Spec: core.PersistentVolumeClaimSpec{
							StorageClassName: util.NewType[string]("newStorage"),
						},
					},
				}
				ad.Status.Members.DBServers[0].Phase = api.MemberPhaseCreated
				ad.Status.Members.DBServers[0].PersistentVolumeClaim = &api.MemberPersistentVolumeClaimStatus{
					Name: pvcName,
				}
				ad.Status.Members.DBServers[1].Phase = api.MemberPhasePending
				ad.Status.Members.DBServers[1].PersistentVolumeClaim = &api.MemberPersistentVolumeClaimStatus{
					Name: pvcName,
				}
				ad.Status.Members.DBServers[2].Phase = api.MemberPhasePending
				ad.Status.Members.DBServers[2].PersistentVolumeClaim = &api.MemberPersistentVolumeClaimStatus{
					Name: pvcName,
				}
				ad.Status.Members.Coordinators[0].Phase = api.MemberPhasePending
				ad.Status.Members.Coordinators[1].Phase = api.MemberPhasePending
				ad.Status.Members.Coordinators[2].Phase = api.MemberPhasePending
			},
			ExpectedEvent: &k8sutil.Event{
				Type:   core.EventTypeNormal,
				Reason: "Plan Action added",
				Message: "A plan item of type SetMemberConditionV2 for member dbserver with role 1 has been added " +
					"with reason: Member replacement required",
			},
			ExpectedHighPlan: []api.Action{
				actions.NewAction(api.ActionTypeSetMemberConditionV2, api.ServerGroupDBServers, shared.WithPredefinedMember(""), "Member replacement required"),
			},
			ExpectedLog: "Member replacement required",
		},
		{
			Name: "Wait for changing Storage for DBServers",
			FakeDataInput: kclient.FakeDataInput{
				PVCS: map[string]*core.PersistentVolumeClaim{
					pvcName: {
						Spec: core.PersistentVolumeClaimSpec{
							StorageClassName: util.NewType[string]("oldStorage"),
						},
					},
				},
			},
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.DBServers = api.ServerGroupSpec{
					Count: util.NewType[int](3),
					VolumeClaimTemplate: &core.PersistentVolumeClaim{
						Spec: core.PersistentVolumeClaimSpec{
							StorageClassName: util.NewType[string]("newStorage"),
						},
					},
				}
				ad.Status.Members.DBServers[0].Phase = api.MemberPhaseCreated
				ad.Status.Members.DBServers[0].PersistentVolumeClaim = &api.MemberPersistentVolumeClaimStatus{
					Name: pvcName,
				}
				cond := api.Condition{
					Type:   api.MemberReplacementRequired,
					Status: conditionTrue,
				}
				ad.Status.Members.DBServers[0].Conditions = append(ad.Status.Members.DBServers[0].Conditions, cond)
			},
		},
		{
			Name: "Change Storage for Agents with deprecated storage class name",
			FakeDataInput: kclient.FakeDataInput{
				PVCS: map[string]*core.PersistentVolumeClaim{
					pvcName: {
						Spec: core.PersistentVolumeClaimSpec{
							StorageClassName: util.NewType[string](""),
						},
					},
				},
			},
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.Agents = api.ServerGroupSpec{
					Count:            util.NewType[int](2),
					StorageClassName: util.NewType[string]("newStorage"),
				}
				ad.Status.Members.Agents[0].Phase = api.MemberPhaseCreated
				ad.Status.Members.Agents[0].PersistentVolumeClaim = &api.MemberPersistentVolumeClaimStatus{
					Name: pvcName,
				}
				ad.Status.Members.Agents[1].Phase = api.MemberPhasePending
				ad.Status.Members.Agents[1].PersistentVolumeClaim = &api.MemberPersistentVolumeClaimStatus{
					Name: pvcName,
				}
				ad.Status.Members.Agents[2].Phase = api.MemberPhasePending
				ad.Status.Members.Agents[2].PersistentVolumeClaim = &api.MemberPersistentVolumeClaimStatus{
					Name: pvcName,
				}
				ad.Status.Members.Coordinators[0].Phase = api.MemberPhasePending
				ad.Status.Members.Coordinators[1].Phase = api.MemberPhasePending
				ad.Status.Members.Coordinators[2].Phase = api.MemberPhasePending
				ad.Status.Members.DBServers[0].Phase = api.MemberPhasePending
				ad.Status.Members.DBServers[1].Phase = api.MemberPhasePending
				ad.Status.Members.DBServers[2].Phase = api.MemberPhasePending
			},
		},
		{
			Name: "Storage for Coordinators is not possible",
			FakeDataInput: kclient.FakeDataInput{
				PVCS: map[string]*core.PersistentVolumeClaim{
					pvcName: {
						Spec: core.PersistentVolumeClaimSpec{
							StorageClassName: util.NewType[string]("oldStorage"),
						},
					},
				},
			},
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.Coordinators = api.ServerGroupSpec{
					Count: util.NewType[int](3),
					VolumeClaimTemplate: &core.PersistentVolumeClaim{
						Spec: core.PersistentVolumeClaimSpec{
							StorageClassName: util.NewType[string]("newStorage"),
						},
					},
				}
				ad.Status.Members.Coordinators[0].Phase = api.MemberPhaseCreated
				ad.Status.Members.Coordinators[0].PersistentVolumeClaim = &api.MemberPersistentVolumeClaimStatus{
					Name: pvcName,
				}
				ad.Status.Members.Coordinators[1].Phase = api.MemberPhasePending
				ad.Status.Members.Coordinators[2].Phase = api.MemberPhasePending
				ad.Status.Members.DBServers[0].Phase = api.MemberPhasePending
				ad.Status.Members.DBServers[1].Phase = api.MemberPhasePending
				ad.Status.Members.DBServers[2].Phase = api.MemberPhasePending
				ad.Status.Members.Agents[0].Phase = api.MemberPhasePending
				ad.Status.Members.Agents[1].Phase = api.MemberPhasePending
				ad.Status.Members.Agents[2].Phase = api.MemberPhasePending
			},
			ExpectedPlan: []api.Action{},
			ExpectedEvent: &k8sutil.Event{
				Type:    core.EventTypeNormal,
				Reason:  "Coordinator Member StorageClass Cannot Change",
				Message: "Member 1 with role coordinator should use a different StorageClass, but is cannot because: Not supported",
			},
		},
		{
			Name: "Create rotation plan",
			FakeDataInput: kclient.FakeDataInput{
				PVCS: map[string]*core.PersistentVolumeClaim{
					"pvc_test": {
						Spec: core.PersistentVolumeClaimSpec{
							StorageClassName: util.NewType[string]("oldStorage"),
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
			},
			Extender: func(t *testing.T, r *Reconciler, c *testCase) {
				// Add ArangoMember
				imageInfo, _ := c.context.SelectImage(c.context.ArangoDeployment.Spec, c.context.ArangoDeployment.Status)
				template, err := newPlanBuilderContext(r.context).RenderPodTemplateForMember(context.Background(), c.context.ACS(),
					c.context.ArangoDeployment.Spec, c.context.ArangoDeployment.Status,
					c.context.ArangoDeployment.Status.Members.Agents[0].ID, imageInfo)
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
					c.context.ArangoDeployment.Status.Members.Agents[0].Pod.GetName(): {
						ObjectMeta: meta.ObjectMeta{
							Name: c.context.ArangoDeployment.Status.Members.Agents[0].Pod.GetName(),
						},
					},
				}

				for _, e := range c.context.ArangoDeployment.Status.Members.AsList() {
					e.Member.Phase = api.MemberPhaseCreated
					require.NoError(t, c.context.ArangoDeployment.Status.Members.Update(e.Member, e.Group))
				}
			},
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.Agents = api.ServerGroupSpec{
					Count: util.NewType[int](2),
					VolumeClaimTemplate: &core.PersistentVolumeClaim{
						Spec: core.PersistentVolumeClaimSpec{
							StorageClassName: util.NewType[string]("oldStorage"),
						},
					},
				}
				ad.Status.Members.Agents[0].Phase = api.MemberPhaseCreated
				ad.Status.Members.Agents[0].PersistentVolumeClaim = &api.MemberPersistentVolumeClaimStatus{
					Name: "pvc_test",
				}
			},
			ExpectedHighPlan: []api.Action{
				actions.NewAction(api.ActionTypeSetMemberConditionV2, api.ServerGroupAgents, deploymentTemplate.Status.Members.Agents[0], "PVC Resize pending"),
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
					Count: util.NewType[int](2),
				}
				ad.Status.Members.Agents[0].Phase = api.MemberPhaseFailed
				ad.Status.Members.Agents[0].ID = "id"
				for i := range ad.Status.Members.Coordinators {
					ad.Status.Members.Coordinators[i].Phase = api.MemberPhaseCreated
				}
				for i := range ad.Status.Members.DBServers {
					ad.Status.Members.DBServers[i].Phase = api.MemberPhaseCreated
				}
			},
			ExpectedHighPlan: []api.Action{
				actions.NewAction(api.ActionTypeRecreateMember, api.ServerGroupAgents, shared.WithPredefinedMember("id")),
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
					Count: util.NewType[int](2),
				}
				ad.Status.Members.Coordinators[0].Phase = api.MemberPhaseFailed
				ad.Status.Members.Coordinators[0].ID = "id"
			},
			ExpectedPlan: []api.Action{
				actions.NewAction(api.ActionTypeRemoveMember, api.ServerGroupCoordinators, shared.WithPredefinedMember("id")),
				actions.NewAction(api.ActionTypeAddMember, api.ServerGroupCoordinators, shared.WithPredefinedMember("")),
				actions.NewAction(api.ActionTypeWaitForMemberUp, api.ServerGroupCoordinators,
					shared.WithPredefinedMember(api.MemberIDPreviousAction)),
			},
			ExpectedLog: "Creating member replacement plan because member has failed",
		},
		{
			Name: "DBServer in failed state - recreate",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.DBServers = api.ServerGroupSpec{
					Count: util.NewType[int](3),
				}
				ad.Status.Members.DBServers[0].Phase = api.MemberPhaseFailed
				ad.Status.Members.DBServers[0].ID = "id"
			},
			ExpectedPlan: []api.Action{
				actions.NewAction(api.ActionTypeRemoveMember, api.ServerGroupDBServers, shared.WithPredefinedMember("id")),
				actions.NewAction(api.ActionTypeAddMember, api.ServerGroupDBServers, shared.WithPredefinedMember("")),
				actions.NewAction(api.ActionTypeWaitForMemberUp, api.ServerGroupDBServers,
					shared.WithPredefinedMember(api.MemberIDPreviousAction)),
			},
			ExpectedLog: "Creating member replacement plan because member has failed",
		},
		{
			Name: "DBServer in failed state - remove",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.DBServers = api.ServerGroupSpec{
					Count: util.NewType[int](2),
				}
				ad.Status.Members.DBServers[2].Phase = api.MemberPhaseFailed
				ad.Status.Members.DBServers[2].ID = "id3"
			},
			ExpectedPlan: []api.Action{
				actions.NewAction(api.ActionTypeCleanOutMember, api.ServerGroupDBServers, shared.WithPredefinedMember("id3")),
				actions.NewAction(api.ActionTypeKillMemberPod, api.ServerGroupDBServers, shared.WithPredefinedMember("id3")),
				actions.NewAction(api.ActionTypeShutdownMember, api.ServerGroupDBServers, shared.WithPredefinedMember("id3")),
				actions.NewAction(api.ActionTypeRemoveMember, api.ServerGroupDBServers, shared.WithPredefinedMember("id3")),
			},
			ExpectedLog: "Creating scale-down plan",
		},
		{
			Name: "CleanOut DBserver",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.DBServers = api.ServerGroupSpec{
					Count: util.NewType[int](3),
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
				actions.NewAction(api.ActionTypeCleanOutMember, api.ServerGroupDBServers, shared.WithPredefinedMember("id")),
				actions.NewAction(api.ActionTypeKillMemberPod, api.ServerGroupDBServers, shared.WithPredefinedMember("")),
				actions.NewAction(api.ActionTypeShutdownMember, api.ServerGroupDBServers, shared.WithPredefinedMember("")),
				actions.NewAction(api.ActionTypeRemoveMember, api.ServerGroupDBServers, shared.WithPredefinedMember("")),
			},
			ExpectedLog: "Creating dbserver replacement plan because server is cleanout in created phase",
		},
		{
			Name: "CleanOut DBserver - scale down",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.DBServers = api.ServerGroupSpec{
					Count: util.NewType[int](2),
				}
				ad.Status.Members.DBServers[2].ID = "id3"
				ad.Status.Members.DBServers[2].Phase = api.MemberPhaseCreated
				ad.Status.Members.DBServers[2].Conditions = api.ConditionList{
					{
						Type:   api.ConditionTypeCleanedOut,
						Status: core.ConditionTrue,
					},
				}
			},
			ExpectedPlan: []api.Action{
				actions.NewAction(api.ActionTypeCleanOutMember, api.ServerGroupDBServers, shared.WithPredefinedMember("id3")),
				actions.NewAction(api.ActionTypeKillMemberPod, api.ServerGroupDBServers, shared.WithPredefinedMember("id3")),
				actions.NewAction(api.ActionTypeShutdownMember, api.ServerGroupDBServers, shared.WithPredefinedMember("id3")),
				actions.NewAction(api.ActionTypeRemoveMember, api.ServerGroupDBServers, shared.WithPredefinedMember("id3")),
			},
			ExpectedLog: "Creating scale-down plan",
		},
		{
			Name: "Scale down DBservers",
			context: &testContext{
				ArangoDeployment: deploymentTemplate.DeepCopy(),
			},
			Helper: func(ad *api.ArangoDeployment) {
				ad.Spec.DBServers = api.ServerGroupSpec{
					Count: util.NewType[int](2),
				}
				ad.Status.Members.DBServers[0].Phase = api.MemberPhaseCreated
			},
			ExpectedPlan: []api.Action{
				actions.NewAction(api.ActionTypeCleanOutMember, api.ServerGroupDBServers, shared.WithPredefinedMember("id")),
				actions.NewAction(api.ActionTypeKillMemberPod, api.ServerGroupDBServers, shared.WithPredefinedMember("")),
				actions.NewAction(api.ActionTypeShutdownMember, api.ServerGroupDBServers, shared.WithPredefinedMember("")),
				actions.NewAction(api.ActionTypeRemoveMember, api.ServerGroupDBServers, shared.WithPredefinedMember("")),
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

			i := testCase.Inspector(t)

			testCase.context.Inspector = i
			testCase.context.state = &FakeStateInspector{
				state: member.State{
					NotReachableErr: errors.New("Client Not Found"),
				},
			}

			h := &LastLogRecord{t: t}
			logger := logging.NewFactory(zerolog.New(io.Discard).Hook(h)).RegisterAndGetLogger("test", logging.Debug)
			r := &Reconciler{
				log:        logger,
				planLogger: logger,
				context:    testCase.context,
			}

			if testCase.Extender != nil {
				testCase.Extender(t, r, &testCase)
			}

			// Act
			if testCase.Helper != nil {
				testCase.Helper(testCase.context.ArangoDeployment)
			}

			err, _ := r.CreatePlan(ctx)

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
			status := testCase.context.GetStatus()

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

type FakeStateInspector struct {
	state member.State
}

func (FakeStateInspector) RefreshState(_ context.Context, _ api.DeploymentStatusMemberElements) {
	//TODO implement me
	panic("implement me")
}

func (FakeStateInspector) GetMemberClient(_ string) (driver.Client, error) {
	//TODO implement me
	panic("implement me")
}

func (FakeStateInspector) GetMemberSyncClient(_ string) (client.API, error) {
	//TODO implement me
	panic("implement me")
}

func (FakeStateInspector) MemberState(_ string) (member.State, bool) {
	//TODO implement me
	panic("implement me")
}

func (FakeStateInspector) Health() (member.Health, bool) {
	//TODO implement me
	panic("implement me")
}

func (f FakeStateInspector) State() member.State {
	return f.state
}

func (FakeStateInspector) Log(_ logging.Logger) {
	//TODO implement me
	panic("implement me")
}
