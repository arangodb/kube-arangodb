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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

// enableFeatureWithDependencies enables f and all of its transitive dependencies for the duration of the
// test, restoring the previous state afterwards.
func enableFeatureWithDependencies(t *testing.T, f features.Feature) {
	seen := map[string]bool{}

	var enable func(x features.Feature)
	enable = func(x features.Feature) {
		if seen[x.Name()] {
			return
		}
		seen[x.Name()] = true

		p := x.EnabledPointer()
		old := *p
		*p = true
		t.Cleanup(func() { *p = old })

		for _, d := range x.Dependencies() {
			enable(d)
		}
	}

	enable(f)
}

var hardenBaseArgs = []string{
	"--server.harden",
	"--javascript.harden",
	"--javascript.startup-options-denylist=.*",
	"--javascript.environment-variables-allowlist=^HOSTNAME$",
	"--javascript.environment-variables-allowlist=^PATH$",
	"--log.api-enabled=jwt",
	"--backup.api-enabled=jwt",
}

func hardenArangodArgsFor(t *testing.T, mode api.DeploymentMode, group api.ServerGroup, dbServers int) []string {
	apiObject := &api.ArangoDeployment{
		ObjectMeta: meta.ObjectMeta{Name: "name", Namespace: tests.FakeNamespace},
		Spec:       api.DeploymentSpec{Mode: api.NewMode(mode)},
	}
	apiObject.Spec.SetDefaults("test")
	if mode == api.DeploymentModeCluster {
		apiObject.Spec.DBServers.Count = util.NewType(dbServers)
	}

	agents := api.MemberStatusList{{ID: "a1"}, {ID: "a2"}, {ID: "a3"}}

	var groupSpec api.ServerGroupSpec
	switch group {
	case api.ServerGroupCoordinators:
		groupSpec = apiObject.Spec.Coordinators
	case api.ServerGroupDBServers:
		groupSpec = apiObject.Spec.DBServers
	case api.ServerGroupSingle:
		groupSpec = apiObject.Spec.Single
	}

	input := pod.Input{
		ApiObject:  apiObject,
		Deployment: apiObject.Spec,
		Status:     api.DeploymentStatus{Members: api.DeploymentStatusMembers{Agents: agents}},
		Group:      group,
		GroupSpec:  groupSpec,
		Member:     api.MemberStatus{ID: "id1"},
	}

	f := kclient.NewFakeClientBuilder()
	f = createClient(f, apiObject, api.ServerGroupAgents, agents...)
	f = createClient(f, apiObject, group, input.Member)
	i := createInspector(t, f)

	cmdline, err := createArangodArgs(i, input)
	require.NoError(t, err)
	return cmdline
}

func TestCreateArangodArgs_Harden(t *testing.T) {
	t.Run("Disabled - no harden args", func(t *testing.T) {
		require.False(t, features.Harden().Enabled())

		cmdline := hardenArangodArgsFor(t, api.DeploymentModeSingle, api.ServerGroupSingle, 0)
		for _, a := range hardenBaseArgs {
			assert.NotContains(t, cmdline, a)
		}
	})

	t.Run("Enabled single - harden flags, no cluster flags", func(t *testing.T) {
		enableFeatureWithDependencies(t, features.Harden())
		require.True(t, features.Harden().Enabled())

		cmdline := hardenArangodArgsFor(t, api.DeploymentModeSingle, api.ServerGroupSingle, 0)
		for _, a := range hardenBaseArgs {
			assert.Contains(t, cmdline, a)
		}
		for _, a := range cmdline {
			assert.NotContains(t, a, "--cluster.default-replication-factor")
			assert.NotContains(t, a, "--cluster.min-replication-factor")
		}
	})

	t.Run("Enabled cluster 3 dbservers - default-rf=3, min-rf=2", func(t *testing.T) {
		enableFeatureWithDependencies(t, features.Harden())

		cmdline := hardenArangodArgsFor(t, api.DeploymentModeCluster, api.ServerGroupCoordinators, 3)
		for _, a := range hardenBaseArgs {
			assert.Contains(t, cmdline, a)
		}
		assert.Contains(t, cmdline, "--cluster.default-replication-factor=3")
		assert.Contains(t, cmdline, "--cluster.min-replication-factor=2")
		assert.Contains(t, cmdline, "--cluster.write-concern=2")
	})

	t.Run("Enabled cluster 5 dbservers - default-rf=3, min-rf=2, wc=2", func(t *testing.T) {
		enableFeatureWithDependencies(t, features.Harden())

		cmdline := hardenArangodArgsFor(t, api.DeploymentModeCluster, api.ServerGroupDBServers, 5)
		assert.Contains(t, cmdline, "--cluster.default-replication-factor=3")
		assert.Contains(t, cmdline, "--cluster.min-replication-factor=2")
		assert.Contains(t, cmdline, "--cluster.write-concern=2")
	})

	t.Run("Enabled cluster 2 dbservers - default-rf=2, min-rf=2, no wc", func(t *testing.T) {
		enableFeatureWithDependencies(t, features.Harden())

		cmdline := hardenArangodArgsFor(t, api.DeploymentModeCluster, api.ServerGroupCoordinators, 2)
		assert.Contains(t, cmdline, "--cluster.default-replication-factor=2")
		assert.Contains(t, cmdline, "--cluster.min-replication-factor=2")
		for _, a := range cmdline {
			assert.NotContains(t, a, "--cluster.write-concern")
		}
	})

	t.Run("Enabled cluster 1 dbserver - no replication flags", func(t *testing.T) {
		enableFeatureWithDependencies(t, features.Harden())

		cmdline := hardenArangodArgsFor(t, api.DeploymentModeCluster, api.ServerGroupCoordinators, 1)
		for _, a := range cmdline {
			assert.NotContains(t, a, "--cluster.default-replication-factor")
			assert.NotContains(t, a, "--cluster.min-replication-factor")
		}
	})
}
