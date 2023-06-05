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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

// TestCreateArangodArgsDBServer tests createArangodArgs for dbserver.
func TestCreateArangodArgsDBServer(t *testing.T) {
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
			Group:       api.ServerGroupDBServers,
			GroupSpec:   apiObject.Spec.DBServers,
			Version:     "",
			Enterprise:  false,
			AutoUpgrade: false,
			Member:      api.MemberStatus{ID: "id1"},
		}

		f := kclient.NewFakeClientBuilder()
		f = createClient(f, apiObject, api.ServerGroupAgents, agents...)
		f = createClient(f, apiObject, input.Group, input.Member)
		i := createInspector(t, f)

		cmdline, err := createArangodArgs(i, input)
		require.NoError(t, err)
		assert.Equal(t,
			[]string{
				"--cluster.agency-endpoint=ssl://name-agent-a1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.agency-endpoint=ssl://name-agent-a2.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.agency-endpoint=ssl://name-agent-a3.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.my-address=ssl://name-dbserver-id1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.my-role=PRIMARY",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=true",
				"--server.endpoint=ssl://[::]:8529",
				"--server.jwt-secret-keyfile=" + jwtSecretFile,
				"--server.statistics=true",
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
			Group:       api.ServerGroupDBServers,
			GroupSpec:   apiObject.Spec.DBServers,
			Version:     "",
			Enterprise:  false,
			AutoUpgrade: true,
			Member:      api.MemberStatus{ID: "id1"},
		}

		f := kclient.NewFakeClientBuilder()
		f = createClient(f, apiObject, api.ServerGroupAgents, agents...)
		f = createClient(f, apiObject, input.Group, input.Member)
		i := createInspector(t, f)

		cmdline, err := createArangodArgsWithUpgrade(i, input)
		require.NoError(t, err)
		assert.Equal(t,
			[]string{
				"--cluster.agency-endpoint=ssl://name-agent-a1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.agency-endpoint=ssl://name-agent-a2.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.agency-endpoint=ssl://name-agent-a3.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.my-address=ssl://name-dbserver-id1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.my-role=PRIMARY",
				"--database.auto-upgrade=true",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=true",
				"--server.endpoint=ssl://[::]:8529",
				"--server.jwt-secret-keyfile=" + jwtSecretFile,
				"--server.statistics=true",
				"--server.storage-engine=rocksdb",
				"--ssl.ecdh-curve=",
				"--ssl.keyfile=/secrets/tls/tls.keyfile",
			},
			cmdline,
		)
	}

	// Default+ClusterDomain deployment
	{
		apiObject := &api.ArangoDeployment{
			ObjectMeta: meta.ObjectMeta{
				Name:      "name",
				Namespace: tests.FakeNamespace,
			},
			Spec: api.DeploymentSpec{
				Mode:          api.NewMode(api.DeploymentModeCluster),
				ClusterDomain: util.NewType[string]("cluster.local"),
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
			Group:       api.ServerGroupDBServers,
			GroupSpec:   apiObject.Spec.DBServers,
			Version:     "",
			Enterprise:  false,
			AutoUpgrade: true,
			Member:      api.MemberStatus{ID: "id1"},
		}

		f := kclient.NewFakeClientBuilder()
		f = createClient(f, apiObject, api.ServerGroupAgents, agents...)
		f = createClient(f, apiObject, input.Group, input.Member)
		i := createInspector(t, f)

		cmdline, err := createArangodArgsWithUpgrade(i, input)
		require.NoError(t, err)
		assert.Equal(t,
			[]string{
				"--cluster.agency-endpoint=ssl://name-agent-a1.name-int." + tests.FakeNamespace + ".svc.cluster.local:8529",
				"--cluster.agency-endpoint=ssl://name-agent-a2.name-int." + tests.FakeNamespace + ".svc.cluster.local:8529",
				"--cluster.agency-endpoint=ssl://name-agent-a3.name-int." + tests.FakeNamespace + ".svc.cluster.local:8529",
				"--cluster.my-address=ssl://name-dbserver-id1.name-int." + tests.FakeNamespace + ".svc.cluster.local:8529",
				"--cluster.my-role=PRIMARY",
				"--database.auto-upgrade=true",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=true",
				"--server.endpoint=ssl://[::]:8529",
				"--server.jwt-secret-keyfile=" + jwtSecretFile,
				"--server.statistics=true",
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
			Group:       api.ServerGroupDBServers,
			GroupSpec:   apiObject.Spec.DBServers,
			Version:     "",
			Enterprise:  false,
			AutoUpgrade: false,
			Member:      api.MemberStatus{ID: "id1"},
		}

		f := kclient.NewFakeClientBuilder()
		f = createClient(f, apiObject, api.ServerGroupAgents, agents...)
		f = createClient(f, apiObject, input.Group, input.Member)
		i := createInspector(t, f)

		cmdline, err := createArangodArgs(i, input)
		require.NoError(t, err)
		assert.Equal(t,
			[]string{
				"--cluster.agency-endpoint=tcp://name-agent-a1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.agency-endpoint=tcp://name-agent-a2.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.agency-endpoint=tcp://name-agent-a3.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.my-address=tcp://name-dbserver-id1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.my-role=PRIMARY",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=true",
				"--server.endpoint=tcp://[::]:8529",
				"--server.jwt-secret-keyfile=" + jwtSecretFile,
				"--server.statistics=true",
				"--server.storage-engine=rocksdb",
			},
			cmdline,
		)
	}

	// No authentication
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
		agents := api.MemberStatusList{
			api.MemberStatus{ID: "a1"},
			api.MemberStatus{ID: "a2"},
			api.MemberStatus{ID: "a3"},
		}
		input := pod.Input{
			ApiObject:   apiObject,
			Deployment:  apiObject.Spec,
			Status:      api.DeploymentStatus{Members: api.DeploymentStatusMembers{Agents: agents}},
			Group:       api.ServerGroupDBServers,
			GroupSpec:   apiObject.Spec.DBServers,
			Version:     "",
			Enterprise:  false,
			AutoUpgrade: false,
			Member:      api.MemberStatus{ID: "id1"},
		}

		f := kclient.NewFakeClientBuilder()
		f = createClient(f, apiObject, api.ServerGroupAgents, agents...)
		f = createClient(f, apiObject, input.Group, input.Member)
		i := createInspector(t, f)

		cmdline, err := createArangodArgs(i, input)
		require.NoError(t, err)
		assert.Equal(t,
			[]string{
				"--cluster.agency-endpoint=ssl://name-agent-a1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.agency-endpoint=ssl://name-agent-a2.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.agency-endpoint=ssl://name-agent-a3.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.my-address=ssl://name-dbserver-id1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.my-role=PRIMARY",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=false",
				"--server.endpoint=ssl://[::]:8529",
				"--server.statistics=true",
				"--server.storage-engine=rocksdb",
				"--ssl.ecdh-curve=",
				"--ssl.keyfile=/secrets/tls/tls.keyfile",
			},
			cmdline,
		)
	}

	// Custom args, MMFiles
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
		apiObject.Spec.StorageEngine = api.NewStorageEngine(api.StorageEngineMMFiles)
		apiObject.Spec.DBServers.Args = []string{"--foo1", "--foo2"}
		agents := api.MemberStatusList{
			api.MemberStatus{ID: "a1"},
			api.MemberStatus{ID: "a2"},
			api.MemberStatus{ID: "a3"},
		}
		input := pod.Input{
			ApiObject:   apiObject,
			Deployment:  apiObject.Spec,
			Status:      api.DeploymentStatus{Members: api.DeploymentStatusMembers{Agents: agents}},
			Group:       api.ServerGroupDBServers,
			GroupSpec:   apiObject.Spec.DBServers,
			Version:     "",
			Enterprise:  false,
			AutoUpgrade: false,
			Member:      api.MemberStatus{ID: "id1"},
		}

		f := kclient.NewFakeClientBuilder()
		f = createClient(f, apiObject, api.ServerGroupAgents, agents...)
		f = createClient(f, apiObject, input.Group, input.Member)
		i := createInspector(t, f)

		cmdline, err := createArangodArgs(i, input)
		require.NoError(t, err)
		assert.Equal(t,
			[]string{
				"--cluster.agency-endpoint=ssl://name-agent-a1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.agency-endpoint=ssl://name-agent-a2.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.agency-endpoint=ssl://name-agent-a3.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.my-address=ssl://name-dbserver-id1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.my-role=PRIMARY",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=true",
				"--server.endpoint=ssl://[::]:8529",
				"--server.jwt-secret-keyfile=" + jwtSecretFile,
				"--server.statistics=true",
				"--server.storage-engine=mmfiles",
				"--ssl.ecdh-curve=",
				"--ssl.keyfile=/secrets/tls/tls.keyfile",
				"--foo1",
				"--foo2",
			},
			cmdline,
		)
	}
}
