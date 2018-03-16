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

// TestCreateArangodArgsCoordinator tests createArangodArgs for coordinator.
func TestCreateArangodArgsCoordinator(t *testing.T) {
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
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupCoordinators, agents, "id1")
		assert.Equal(t,
			[]string{
				"--cluster.agency-endpoint=tcp://name-agnt-a1.name-int.ns.svc:8529",
				"--cluster.agency-endpoint=tcp://name-agnt-a2.name-int.ns.svc:8529",
				"--cluster.agency-endpoint=tcp://name-agnt-a3.name-int.ns.svc:8529",
				"--cluster.my-address=tcp://name-crdn-id1.name-int.ns.svc:8529",
				"--cluster.my-role=COORDINATOR",
				"--database.directory=/data",
				"--foxx.queues=true",
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
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupCoordinators, agents, "id1")
		assert.Equal(t,
			[]string{
				"--cluster.agency-endpoint=ssl://name-agnt-a1.name-int.ns.svc:8529",
				"--cluster.agency-endpoint=ssl://name-agnt-a2.name-int.ns.svc:8529",
				"--cluster.agency-endpoint=ssl://name-agnt-a3.name-int.ns.svc:8529",
				"--cluster.my-address=ssl://name-crdn-id1.name-int.ns.svc:8529",
				"--cluster.my-role=COORDINATOR",
				"--database.directory=/data",
				"--foxx.queues=true",
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
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupCoordinators, agents, "id1")
		assert.Equal(t,
			[]string{
				"--cluster.agency-endpoint=tcp://name-agnt-a1.name-int.ns.svc:8529",
				"--cluster.agency-endpoint=tcp://name-agnt-a2.name-int.ns.svc:8529",
				"--cluster.agency-endpoint=tcp://name-agnt-a3.name-int.ns.svc:8529",
				"--cluster.my-address=tcp://name-crdn-id1.name-int.ns.svc:8529",
				"--cluster.my-role=COORDINATOR",
				"--database.directory=/data",
				"--foxx.queues=true",
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

	// Custom args, RocksDB
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
		apiObject.Spec.Coordinators.Args = []string{"--foo1", "--foo2"}
		apiObject.Spec.StorageEngine = api.StorageEngineMMFiles
		agents := api.MemberStatusList{
			api.MemberStatus{ID: "a1"},
			api.MemberStatus{ID: "a2"},
			api.MemberStatus{ID: "a3"},
		}
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupCoordinators, agents, "id1")
		assert.Equal(t,
			[]string{
				"--cluster.agency-endpoint=tcp://name-agnt-a1.name-int.ns.svc:8529",
				"--cluster.agency-endpoint=tcp://name-agnt-a2.name-int.ns.svc:8529",
				"--cluster.agency-endpoint=tcp://name-agnt-a3.name-int.ns.svc:8529",
				"--cluster.my-address=tcp://name-crdn-id1.name-int.ns.svc:8529",
				"--cluster.my-role=COORDINATOR",
				"--database.directory=/data",
				"--foxx.queues=true",
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
