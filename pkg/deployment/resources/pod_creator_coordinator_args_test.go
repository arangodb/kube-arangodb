//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

// TestCreateArangodArgsCoordinator tests createArangodArgs for coordinator.
func TestCreateArangodArgsCoordinator(t *testing.T) {
	jwtSecretFile := filepath.Join(shared.ClusterJWTSecretVolumeMountDir, utilConstants.SecretKeyToken)
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
			ApiObject:  apiObject,
			Deployment: apiObject.Spec,
			Status:     api.DeploymentStatus{Members: api.DeploymentStatusMembers{Agents: agents}},
			Group:      api.ServerGroupCoordinators,
			GroupSpec:  apiObject.Spec.Coordinators,
			Image: api.ImageInfo{
				ArangoDBVersion: "",
				Enterprise:      false,
			},
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
				"--cluster.my-address=ssl://name-coordinator-id1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.my-role=COORDINATOR",
				"--database.directory=/data",
				"--foxx.queues=true",
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
			ApiObject:  apiObject,
			Deployment: apiObject.Spec,
			Status:     api.DeploymentStatus{Members: api.DeploymentStatusMembers{Agents: agents}},
			Group:      api.ServerGroupCoordinators,
			GroupSpec:  apiObject.Spec.Coordinators,
			Image: api.ImageInfo{
				ArangoDBVersion: "",
				Enterprise:      false,
			},
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
				"--cluster.my-address=ssl://name-coordinator-id1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.my-role=COORDINATOR",
				"--database.auto-upgrade=true",
				"--database.directory=/data",
				"--foxx.queues=true",
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

	// Default+AutoUpgrade deployment for 3.6.0
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
			ApiObject:  apiObject,
			Deployment: apiObject.Spec,
			Status:     api.DeploymentStatus{Members: api.DeploymentStatusMembers{Agents: agents}},
			Group:      api.ServerGroupCoordinators,
			GroupSpec:  apiObject.Spec.Coordinators,
			Image: api.ImageInfo{
				ArangoDBVersion: "3.6.0",
				Enterprise:      false,
			},
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
				"--cluster.my-address=ssl://name-coordinator-id1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.my-role=COORDINATOR",
				"--cluster.upgrade=online",
				"--database.directory=/data",
				"--foxx.queues=true",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=true",
				"--server.endpoint=ssl://[::]:8529",
				"--server.jwt-secret-keyfile=/secrets/cluster/jwt/token",
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
			ApiObject:  apiObject,
			Deployment: apiObject.Spec,
			Status:     api.DeploymentStatus{Members: api.DeploymentStatusMembers{Agents: agents}},
			Group:      api.ServerGroupCoordinators,
			GroupSpec:  apiObject.Spec.Coordinators,
			Image: api.ImageInfo{
				ArangoDBVersion: "",
				Enterprise:      false,
			},
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
				"--cluster.my-address=tcp://name-coordinator-id1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.my-role=COORDINATOR",
				"--database.directory=/data",
				"--foxx.queues=true",
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
			ApiObject:  apiObject,
			Deployment: apiObject.Spec,
			Status:     api.DeploymentStatus{Members: api.DeploymentStatusMembers{Agents: agents}},
			Group:      api.ServerGroupCoordinators,
			GroupSpec:  apiObject.Spec.Coordinators,
			Image: api.ImageInfo{
				ArangoDBVersion: "",
				Enterprise:      false,
			},
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
				"--cluster.my-address=ssl://name-coordinator-id1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.my-role=COORDINATOR",
				"--database.directory=/data",
				"--foxx.queues=true",
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

	// Custom args, RocksDB
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
		apiObject.Spec.Coordinators.Args = []string{"--foo1", "--foo2"}
		apiObject.Spec.StorageEngine = api.NewStorageEngine(api.StorageEngineMMFiles)
		agents := api.MemberStatusList{
			api.MemberStatus{ID: "a1"},
			api.MemberStatus{ID: "a2"},
			api.MemberStatus{ID: "a3"},
		}
		input := pod.Input{
			ApiObject:  apiObject,
			Deployment: apiObject.Spec,
			Status:     api.DeploymentStatus{Members: api.DeploymentStatusMembers{Agents: agents}},
			Group:      api.ServerGroupCoordinators,
			GroupSpec:  apiObject.Spec.Coordinators,
			Image: api.ImageInfo{
				ArangoDBVersion: "",
				Enterprise:      false,
			},
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
				"--cluster.my-address=ssl://name-coordinator-id1.name-int." + tests.FakeNamespace + ".svc:8529",
				"--cluster.my-role=COORDINATOR",
				"--database.directory=/data",
				"--foxx.queues=true",
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
