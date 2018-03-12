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

package deployment

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
)

// TestCreateArangodArgsDBServer tests createArangodArgs for dbserver.
func TestCreateArangodArgsDBServer(t *testing.T) {
	// Default deployment
	{
		apiObject := &api.ArangoDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "ns",
			},
			Spec: api.DeploymentSpec{
				Mode: api.DeploymentModeCluster,
			},
		}
		apiObject.Spec.SetDefaults("test")
		agents := api.MemberStatusList{
			api.MemberStatus{ID: "a1"},
			api.MemberStatus{ID: "a2"},
			api.MemberStatus{ID: "a3"},
		}
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupDBServers, apiObject.Spec.DBServers, agents, "id1")
		assert.Equal(t,
			[]string{
				"--cluster.agency-endpoint=tcp://name-agent-a1.name-int.ns.svc:8529",
				"--cluster.agency-endpoint=tcp://name-agent-a2.name-int.ns.svc:8529",
				"--cluster.agency-endpoint=tcp://name-agent-a3.name-int.ns.svc:8529",
				"--cluster.my-address=tcp://name-dbserver-id1.name-int.ns.svc:8529",
				"--cluster.my-id=id1",
				"--cluster.my-role=PRIMARY",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=true",
				"--server.endpoint=tcp://[::]:8529",
				"--server.jwt-secret=$(ARANGOD_JWT_SECRET)",
				"--server.statistics=true",
				"--server.storage-engine=rocksdb",
			},
			cmdline,
		)
	}

	// Default+TLS deployment
	{
		apiObject := &api.ArangoDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "ns",
			},
			Spec: api.DeploymentSpec{
				Mode: api.DeploymentModeCluster,
				TLS: api.TLSSpec{
					CASecretName: "test-ca",
				},
			},
		}
		apiObject.Spec.SetDefaults("test")
		agents := api.MemberStatusList{
			api.MemberStatus{ID: "a1"},
			api.MemberStatus{ID: "a2"},
			api.MemberStatus{ID: "a3"},
		}
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupDBServers, apiObject.Spec.DBServers, agents, "id1")
		assert.Equal(t,
			[]string{
				"--cluster.agency-endpoint=ssl://name-agent-a1.name-int.ns.svc:8529",
				"--cluster.agency-endpoint=ssl://name-agent-a2.name-int.ns.svc:8529",
				"--cluster.agency-endpoint=ssl://name-agent-a3.name-int.ns.svc:8529",
				"--cluster.my-address=ssl://name-dbserver-id1.name-int.ns.svc:8529",
				"--cluster.my-id=id1",
				"--cluster.my-role=PRIMARY",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=true",
				"--server.endpoint=ssl://[::]:8529",
				"--server.jwt-secret=$(ARANGOD_JWT_SECRET)",
				"--server.statistics=true",
				"--server.storage-engine=rocksdb",
				"--ssl.ecdh-curve=",
				"--ssl.keyfile=/secrets/tls/tls.keyfile",
			},
			cmdline,
		)
	}

	// No authentication
	{
		apiObject := &api.ArangoDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "ns",
			},
			Spec: api.DeploymentSpec{
				Mode: api.DeploymentModeCluster,
			},
		}
		apiObject.Spec.SetDefaults("test")
		apiObject.Spec.Authentication.JWTSecretName = "None"
		agents := api.MemberStatusList{
			api.MemberStatus{ID: "a1"},
			api.MemberStatus{ID: "a2"},
			api.MemberStatus{ID: "a3"},
		}
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupDBServers, apiObject.Spec.DBServers, agents, "id1")
		assert.Equal(t,
			[]string{
				"--cluster.agency-endpoint=tcp://name-agent-a1.name-int.ns.svc:8529",
				"--cluster.agency-endpoint=tcp://name-agent-a2.name-int.ns.svc:8529",
				"--cluster.agency-endpoint=tcp://name-agent-a3.name-int.ns.svc:8529",
				"--cluster.my-address=tcp://name-dbserver-id1.name-int.ns.svc:8529",
				"--cluster.my-id=id1",
				"--cluster.my-role=PRIMARY",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=false",
				"--server.endpoint=tcp://[::]:8529",
				"--server.statistics=true",
				"--server.storage-engine=rocksdb",
			},
			cmdline,
		)
	}

	// Custom args, MMFiles
	{
		apiObject := &api.ArangoDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "ns",
			},
			Spec: api.DeploymentSpec{
				Mode: api.DeploymentModeCluster,
			},
		}
		apiObject.Spec.SetDefaults("test")
		apiObject.Spec.StorageEngine = api.StorageEngineMMFiles
		apiObject.Spec.DBServers.Args = []string{"--foo1", "--foo2"}
		agents := api.MemberStatusList{
			api.MemberStatus{ID: "a1"},
			api.MemberStatus{ID: "a2"},
			api.MemberStatus{ID: "a3"},
		}
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupDBServers, apiObject.Spec.DBServers, agents, "id1")
		assert.Equal(t,
			[]string{
				"--cluster.agency-endpoint=tcp://name-agent-a1.name-int.ns.svc:8529",
				"--cluster.agency-endpoint=tcp://name-agent-a2.name-int.ns.svc:8529",
				"--cluster.agency-endpoint=tcp://name-agent-a3.name-int.ns.svc:8529",
				"--cluster.my-address=tcp://name-dbserver-id1.name-int.ns.svc:8529",
				"--cluster.my-id=id1",
				"--cluster.my-role=PRIMARY",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=true",
				"--server.endpoint=tcp://[::]:8529",
				"--server.jwt-secret=$(ARANGOD_JWT_SECRET)",
				"--server.statistics=true",
				"--server.storage-engine=mmfiles",
				"--foo1",
				"--foo2",
			},
			cmdline,
		)
	}
}
