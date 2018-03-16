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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

// TestCreateArangodArgsAgent tests createArangodArgs for agent.
func TestCreateArangodArgsAgent(t *testing.T) {
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
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupAgents, apiObject.Spec.Agents, agents, "a1")
		assert.Equal(t,
			[]string{
				"--agency.activate=true",
				"--agency.endpoint=tcp://name-agnt-a2.name-int.ns.svc:8529",
				"--agency.endpoint=tcp://name-agnt-a3.name-int.ns.svc:8529",
				"--agency.my-address=tcp://name-agnt-a1.name-int.ns.svc:8529",
				"--agency.size=3",
				"--agency.supervision=true",
				"--cluster.my-id=a1",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=true",
				"--server.endpoint=tcp://[::]:8529",
				"--server.jwt-secret=$(ARANGOD_JWT_SECRET)",
				"--server.statistics=false",
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
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupAgents, apiObject.Spec.Agents, agents, "a1")
		assert.Equal(t,
			[]string{
				"--agency.activate=true",
				"--agency.endpoint=ssl://name-agnt-a2.name-int.ns.svc:8529",
				"--agency.endpoint=ssl://name-agnt-a3.name-int.ns.svc:8529",
				"--agency.my-address=ssl://name-agnt-a1.name-int.ns.svc:8529",
				"--agency.size=3",
				"--agency.supervision=true",
				"--cluster.my-id=a1",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=true",
				"--server.endpoint=ssl://[::]:8529",
				"--server.jwt-secret=$(ARANGOD_JWT_SECRET)",
				"--server.statistics=false",
				"--server.storage-engine=rocksdb",
				"--ssl.ecdh-curve=",
				"--ssl.keyfile=/secrets/tls/tls.keyfile",
			},
			cmdline,
		)
	}

	// No authentication, mmfiles
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
		apiObject.Spec.StorageEngine = api.StorageEngineMMFiles
		agents := api.MemberStatusList{
			api.MemberStatus{ID: "a1"},
			api.MemberStatus{ID: "a2"},
			api.MemberStatus{ID: "a3"},
		}
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupAgents, apiObject.Spec.Agents, agents, "a1")
		assert.Equal(t,
			[]string{
				"--agency.activate=true",
				"--agency.endpoint=tcp://name-agnt-a2.name-int.ns.svc:8529",
				"--agency.endpoint=tcp://name-agnt-a3.name-int.ns.svc:8529",
				"--agency.my-address=tcp://name-agnt-a1.name-int.ns.svc:8529",
				"--agency.size=3",
				"--agency.supervision=true",
				"--cluster.my-id=a1",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=false",
				"--server.endpoint=tcp://[::]:8529",
				"--server.statistics=false",
				"--server.storage-engine=mmfiles",
			},
			cmdline,
		)
	}

	// Custom args
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
		apiObject.Spec.Agents.Args = []string{"--foo1", "--foo2"}
		agents := api.MemberStatusList{
			api.MemberStatus{ID: "a1"},
			api.MemberStatus{ID: "a2"},
			api.MemberStatus{ID: "a3"},
		}
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupAgents, apiObject.Spec.Agents, agents, "a1")
		assert.Equal(t,
			[]string{
				"--agency.activate=true",
				"--agency.endpoint=tcp://name-agnt-a2.name-int.ns.svc:8529",
				"--agency.endpoint=tcp://name-agnt-a3.name-int.ns.svc:8529",
				"--agency.my-address=tcp://name-agnt-a1.name-int.ns.svc:8529",
				"--agency.size=3",
				"--agency.supervision=true",
				"--cluster.my-id=a1",
				"--database.directory=/data",
				"--foxx.queues=false",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=true",
				"--server.endpoint=tcp://[::]:8529",
				"--server.jwt-secret=$(ARANGOD_JWT_SECRET)",
				"--server.statistics=false",
				"--server.storage-engine=rocksdb",
				"--foo1",
				"--foo2",
			},
			cmdline,
		)
	}
}
