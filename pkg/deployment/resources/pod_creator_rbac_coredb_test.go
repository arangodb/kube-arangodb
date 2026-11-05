//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

// clusterArangodArgs builds the arangod command line for a member of the given group in a cluster.
// gatewayEnabled sets the deployment-status GatewaySidecarEnabled condition (a status-level condition,
// not a member-level one).
func clusterArangodArgs(t *testing.T, group api.ServerGroup, gatewayEnabled bool) []string {
	apiObject := &api.ArangoDeployment{
		ObjectMeta: meta.ObjectMeta{Name: "name", Namespace: tests.FakeNamespace},
		Spec:       api.DeploymentSpec{Mode: api.NewMode(api.DeploymentModeCluster)},
	}
	apiObject.Spec.SetDefaults("test")

	agents := api.MemberStatusList{{ID: "a1"}, {ID: "a2"}, {ID: "a3"}}
	member := api.MemberStatus{ID: "id1"}

	var groupSpec api.ServerGroupSpec
	switch group {
	case api.ServerGroupCoordinators:
		groupSpec = apiObject.Spec.Coordinators
	case api.ServerGroupDBServers:
		groupSpec = apiObject.Spec.DBServers
	}

	status := api.DeploymentStatus{Members: api.DeploymentStatusMembers{Agents: agents}}
	if gatewayEnabled {
		status.Conditions.Update(api.ConditionTypeGatewaySidecarEnabled, true, "test", "test")
	}

	input := pod.Input{
		ApiObject:  apiObject,
		Deployment: apiObject.Spec,
		Status:     status,
		Group:      group,
		GroupSpec:  groupSpec,
		Member:     member,
	}

	f := kclient.NewFakeClientBuilder()
	f = createClient(f, apiObject, api.ServerGroupAgents, agents...)
	f = createClient(f, apiObject, group, member)
	i := createInspector(t, f)

	cmdline, err := createArangodArgs(i, input)
	require.NoError(t, err)
	return cmdline
}

func TestCreateArangodArgs_RBACCoreDB(t *testing.T) {
	rbacArg := fmt.Sprintf("--server.external-rbac-service=http://127.0.0.1:%d", shared.InternalSidecarContainerPortHTTP)

	t.Run("Serving member with gateway sidecar - arg added", func(t *testing.T) {
		enableFeatureWithDependencies(t, features.RBACCoreDB())
		require.True(t, features.RBACCoreDB().Enabled())

		cmdline := clusterArangodArgs(t, api.ServerGroupCoordinators, true)
		assert.Contains(t, cmdline, rbacArg)
	})

	t.Run("Feature disabled - arg not added", func(t *testing.T) {
		require.False(t, features.RBACCoreDB().Enabled())

		cmdline := clusterArangodArgs(t, api.ServerGroupCoordinators, true)
		assert.NotContains(t, cmdline, rbacArg)
	})

	t.Run("Serving member without gateway sidecar - arg not added", func(t *testing.T) {
		enableFeatureWithDependencies(t, features.RBACCoreDB())

		cmdline := clusterArangodArgs(t, api.ServerGroupCoordinators, false)
		assert.NotContains(t, cmdline, rbacArg)
	})

	t.Run("Non-serving member (dbserver) - arg not added", func(t *testing.T) {
		enableFeatureWithDependencies(t, features.RBACCoreDB())

		cmdline := clusterArangodArgs(t, api.ServerGroupDBServers, true)
		assert.NotContains(t, cmdline, rbacArg)
	})
}
