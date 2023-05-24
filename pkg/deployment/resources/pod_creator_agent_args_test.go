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

package resources

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func createClient(f kclient.FakeClientBuilder, apiObject *api.ArangoDeployment, group api.ServerGroup, statuses ...api.MemberStatus) kclient.FakeClientBuilder {
	for _, a := range statuses {
		memberName := a.ArangoMemberName(apiObject.GetName(), group)

		svc := core.Service{
			ObjectMeta: meta.ObjectMeta{
				Name:      memberName,
				Namespace: tests.FakeNamespace,
			},
			Spec: core.ServiceSpec{
				ClusterIP: "127.0.0.1",
			},
		}

		f = f.Add(&svc)
	}

	return f
}

func createInspector(t *testing.T, f kclient.FakeClientBuilder) inspector.Inspector {
	c := f.Client()
	i := tests.NewInspector(t, c)
	require.NoError(t, i.Refresh(context.Background()))
	return i
}

// TestCreateArangodArgsAgent tests createArangodArgs for agent.
func TestCreateArangodArgsAgent(t *testing.T) {
	jwtSecretFile := filepath.Join(shared.ClusterJWTSecretVolumeMountDir, constants.SecretKeyToken)
	// Default deployment
	{
		apiObject := &api.ArangoDeployment{
			ObjectMeta: meta.ObjectMeta{
				Name:      "name",
				Namespace: tests.FakeNamespace,
			},
			Spec: api.DeploymentSpec{
				Mode: api.NewMode(api.DeploymentModeCluster),
			},
		}
		apiObject.Spec.SetDefaults("test")
		agents := api.MemberStatusList{
			api.MemberStatus{ID: "a1"},
			api.MemberStatus{ID: "a2"},
			api.MemberStatus{ID: "a3"},
		}
		input := pod.Input{
			ApiObject:   apiObject,
			Deployment:  apiObject.Spec,
			Status:      api.DeploymentStatus{Members: api.DeploymentStatusMembers{Agents: agents}},
			Group:       api.ServerGroupAgents,
			GroupSpec:   apiObject.Spec.Agents,
			Version:     "",
			Enterprise:  false,
			AutoUpgrade: false,
			Member:      api.MemberStatus{ID: "a1"},
		}

		f := kclient.NewFakeClientBuilder()
		f = createClient(f, apiObject, api.ServerGroupAgents, agents...)
		i := createInspector(t, f)

		cmdline, err := createArangodArgs(i, input)
		require.NoError(t, err)
		assert.Equal(t,
			[]string{
				"--agency.activate=true",
				"--agency.disaster-recovery-id=a1",
				"--agency.endpoint=ssl://name-agent-a2.name-int." + tests.FakeNamespace + ".svc:8529",
				"--agency.endpoint=ssl://name-agent-a3.name-int." + tests.FakeNamespace + ".svc:8529",
				"--agency.my-address=ssl://name-agent-a1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--agency.size=3",
				"--agency.supervision=true",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=true",
				"--server.endpoint=ssl://[::]:8529",
				"--server.jwt-secret-keyfile=" + jwtSecretFile,
				"--server.statistics=false",
				"--server.storage-engine=rocksdb",
				"--ssl.ecdh-curve=",
				"--ssl.keyfile=/secrets/tls/tls.keyfile",
			},
			cmdline,
		)
	}

	// Default+AutoUpgrade deployment
	{
		apiObject := &api.ArangoDeployment{
			ObjectMeta: meta.ObjectMeta{
				Name:      "name",
				Namespace: tests.FakeNamespace,
			},
			Spec: api.DeploymentSpec{
				Mode: api.NewMode(api.DeploymentModeCluster),
			},
		}
		apiObject.Spec.SetDefaults("test")
		agents := api.MemberStatusList{
			api.MemberStatus{ID: "a1"},
			api.MemberStatus{ID: "a2"},
			api.MemberStatus{ID: "a3"},
		}

		input := pod.Input{
			ApiObject:   apiObject,
			Deployment:  apiObject.Spec,
			Status:      api.DeploymentStatus{Members: api.DeploymentStatusMembers{Agents: agents}},
			Group:       api.ServerGroupAgents,
			GroupSpec:   apiObject.Spec.Agents,
			Version:     "",
			Enterprise:  false,
			AutoUpgrade: true,
			Member:      api.MemberStatus{ID: "a1"},
		}

		f := kclient.NewFakeClientBuilder()
		f = createClient(f, apiObject, api.ServerGroupAgents, agents...)
		i := createInspector(t, f)

		cmdline, err := createArangodArgsWithUpgrade(i, input)
		require.NoError(t, err)
		assert.Equal(t,
			[]string{
				"--agency.activate=true",
				"--agency.disaster-recovery-id=a1",
				"--agency.endpoint=ssl://name-agent-a2.name-int." + tests.FakeNamespace + ".svc:8529",
				"--agency.endpoint=ssl://name-agent-a3.name-int." + tests.FakeNamespace + ".svc:8529",
				"--agency.my-address=ssl://name-agent-a1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--agency.size=3",
				"--agency.supervision=true",
				"--database.auto-upgrade=true",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=true",
				"--server.endpoint=ssl://[::]:8529",
				"--server.jwt-secret-keyfile=" + jwtSecretFile,
				"--server.statistics=false",
				"--server.storage-engine=rocksdb",
				"--ssl.ecdh-curve=",
				"--ssl.keyfile=/secrets/tls/tls.keyfile",
			},
			cmdline,
		)
	}

	// Default+TLS disabled deployment
	{
		apiObject := &api.ArangoDeployment{
			ObjectMeta: meta.ObjectMeta{
				Name:      "name",
				Namespace: tests.FakeNamespace,
			},
			Spec: api.DeploymentSpec{
				Mode: api.NewMode(api.DeploymentModeCluster),
				TLS: api.TLSSpec{
					CASecretName: util.NewType[string]("None"),
				},
			},
		}
		apiObject.Spec.SetDefaults("test")
		agents := api.MemberStatusList{
			api.MemberStatus{ID: "a1"},
			api.MemberStatus{ID: "a2"},
			api.MemberStatus{ID: "a3"},
		}

		input := pod.Input{
			ApiObject:   apiObject,
			Deployment:  apiObject.Spec,
			Status:      api.DeploymentStatus{Members: api.DeploymentStatusMembers{Agents: agents}},
			Group:       api.ServerGroupAgents,
			GroupSpec:   apiObject.Spec.Agents,
			Version:     "",
			Enterprise:  false,
			AutoUpgrade: false,
			Member:      api.MemberStatus{ID: "a1"},
		}

		f := kclient.NewFakeClientBuilder()
		f = createClient(f, apiObject, api.ServerGroupAgents, agents...)
		i := createInspector(t, f)

		cmdline, err := createArangodArgs(i, input)
		require.NoError(t, err)
		assert.Equal(t,
			[]string{
				"--agency.activate=true",
				"--agency.disaster-recovery-id=a1",
				"--agency.endpoint=tcp://name-agent-a2.name-int." + tests.FakeNamespace + ".svc:8529",
				"--agency.endpoint=tcp://name-agent-a3.name-int." + tests.FakeNamespace + ".svc:8529",
				"--agency.my-address=tcp://name-agent-a1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--agency.size=3",
				"--agency.supervision=true",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=true",
				"--server.endpoint=tcp://[::]:8529",
				"--server.jwt-secret-keyfile=" + jwtSecretFile,
				"--server.statistics=false",
				"--server.storage-engine=rocksdb",
			},
			cmdline,
		)
	}

	// No authentication, mmfiles
	{
		apiObject := &api.ArangoDeployment{
			ObjectMeta: meta.ObjectMeta{
				Name:      "name",
				Namespace: tests.FakeNamespace,
			},
			Spec: api.DeploymentSpec{
				Mode: api.NewMode(api.DeploymentModeCluster),
			},
		}
		apiObject.Spec.SetDefaults("test")
		apiObject.Spec.Authentication.JWTSecretName = util.NewType[string]("None")
		apiObject.Spec.StorageEngine = api.NewStorageEngine(api.StorageEngineMMFiles)
		agents := api.MemberStatusList{
			api.MemberStatus{ID: "a1"},
			api.MemberStatus{ID: "a2"},
			api.MemberStatus{ID: "a3"},
		}
		input := pod.Input{
			ApiObject:   apiObject,
			Deployment:  apiObject.Spec,
			Status:      api.DeploymentStatus{Members: api.DeploymentStatusMembers{Agents: agents}},
			Group:       api.ServerGroupAgents,
			GroupSpec:   apiObject.Spec.Agents,
			Version:     "",
			Enterprise:  false,
			AutoUpgrade: false,
			Member:      api.MemberStatus{ID: "a1"},
		}

		f := kclient.NewFakeClientBuilder()
		f = createClient(f, apiObject, api.ServerGroupAgents, agents...)
		i := createInspector(t, f)

		cmdline, err := createArangodArgs(i, input)
		require.NoError(t, err)
		assert.Equal(t,
			[]string{
				"--agency.activate=true",
				"--agency.disaster-recovery-id=a1",
				"--agency.endpoint=ssl://name-agent-a2.name-int." + tests.FakeNamespace + ".svc:8529",
				"--agency.endpoint=ssl://name-agent-a3.name-int." + tests.FakeNamespace + ".svc:8529",
				"--agency.my-address=ssl://name-agent-a1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--agency.size=3",
				"--agency.supervision=true",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=false",
				"--server.endpoint=ssl://[::]:8529",
				"--server.statistics=false",
				"--server.storage-engine=mmfiles",
				"--ssl.ecdh-curve=",
				"--ssl.keyfile=/secrets/tls/tls.keyfile",
			},
			cmdline,
		)
	}

	// Custom args
	{
		apiObject := &api.ArangoDeployment{
			ObjectMeta: meta.ObjectMeta{
				Name:      "name",
				Namespace: tests.FakeNamespace,
			},
			Spec: api.DeploymentSpec{
				Mode: api.NewMode(api.DeploymentModeCluster),
			},
		}
		apiObject.Spec.SetDefaults("test")
		apiObject.Spec.Agents.Args = []string{"--foo1", "--foo2"}
		agents := api.MemberStatusList{
			api.MemberStatus{ID: "a1"},
			api.MemberStatus{ID: "a2"},
			api.MemberStatus{ID: "a3"},
		}
		input := pod.Input{
			ApiObject:   apiObject,
			Deployment:  apiObject.Spec,
			Status:      api.DeploymentStatus{Members: api.DeploymentStatusMembers{Agents: agents}},
			Group:       api.ServerGroupAgents,
			GroupSpec:   apiObject.Spec.Agents,
			Version:     "",
			Enterprise:  false,
			AutoUpgrade: false,
			Member:      api.MemberStatus{ID: "a1"},
		}

		f := kclient.NewFakeClientBuilder()
		f = createClient(f, apiObject, api.ServerGroupAgents, agents...)
		i := createInspector(t, f)

		cmdline, err := createArangodArgs(i, input)
		require.NoError(t, err)
		assert.Equal(t,
			[]string{
				"--agency.activate=true",
				"--agency.disaster-recovery-id=a1",
				"--agency.endpoint=ssl://name-agent-a2.name-int." + tests.FakeNamespace + ".svc:8529",
				"--agency.endpoint=ssl://name-agent-a3.name-int." + tests.FakeNamespace + ".svc:8529",
				"--agency.my-address=ssl://name-agent-a1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--agency.size=3",
				"--agency.supervision=true",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=true",
				"--server.endpoint=ssl://[::]:8529",
				"--server.jwt-secret-keyfile=" + jwtSecretFile,
				"--server.statistics=false",
				"--server.storage-engine=rocksdb",
				"--ssl.ecdh-curve=",
				"--ssl.keyfile=/secrets/tls/tls.keyfile",
				"--foo1",
				"--foo2",
			},
			cmdline,
		)
	}
}
