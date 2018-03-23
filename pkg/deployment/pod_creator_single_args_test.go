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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/stretchr/testify/assert"
)

// TestCreateArangodArgsSingle tests createArangodArgs for single server.
func TestCreateArangodArgsSingle(t *testing.T) {
	// Default deployment
	{
		apiObject := &api.ArangoDeployment{
			Spec: api.DeploymentSpec{
				XMode: api.NewMode(api.DeploymentModeSingle),
			},
		}
		apiObject.Spec.SetDefaults("test")
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupSingle, nil, "id1", false)
		assert.Equal(t,
			[]string{
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

	// Default+AutoUpgrade deployment
	{
		apiObject := &api.ArangoDeployment{
			Spec: api.DeploymentSpec{
				Mode: api.DeploymentModeSingle,
			},
		}
		apiObject.Spec.SetDefaults("test")
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupSingle, nil, "id1", true)
		assert.Equal(t,
			[]string{
				"--database.auto-upgrade=true",
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

	// Default+TLS disabled deployment
	{
		apiObject := &api.ArangoDeployment{
			Spec: api.DeploymentSpec{
				XMode: api.NewMode(api.DeploymentModeSingle),
				TLS: api.TLSSpec{
					XCASecretName: util.NewString("None"),
				},
			},
		}
		apiObject.Spec.SetDefaults("test")
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupSingle, nil, "id1", false)
		assert.Equal(t,
			[]string{
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

	// Default deployment with mmfiles
	{
		apiObject := &api.ArangoDeployment{
			Spec: api.DeploymentSpec{
				XMode:          api.NewMode(api.DeploymentModeSingle),
				XStorageEngine: api.NewStorageEngine(api.StorageEngineMMFiles),
			},
		}
		apiObject.Spec.SetDefaults("test")
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupSingle, nil, "id1", false)
		assert.Equal(t,
			[]string{
				"--database.directory=/data",
				"--foxx.queues=true",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=true",
				"--server.endpoint=ssl://[::]:8529",
				"--server.jwt-secret=$(ARANGOD_JWT_SECRET)",
				"--server.statistics=true",
				"--server.storage-engine=mmfiles",
				"--ssl.ecdh-curve=",
				"--ssl.keyfile=/secrets/tls/tls.keyfile",
			},
			cmdline,
		)
	}

	// No authentication
	{
		apiObject := &api.ArangoDeployment{
			Spec: api.DeploymentSpec{
				XMode: api.NewMode(api.DeploymentModeSingle),
			},
		}
		apiObject.Spec.Authentication.XJWTSecretName = util.NewString("None")
		apiObject.Spec.SetDefaults("test")
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupSingle, nil, "id1", false)
		assert.Equal(t,
			[]string{
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

	// Custom args
	{
		apiObject := &api.ArangoDeployment{
			Spec: api.DeploymentSpec{
				XMode: api.NewMode(api.DeploymentModeSingle),
			},
		}
		apiObject.Spec.Single.Args = []string{"--foo1", "--foo2"}
		apiObject.Spec.SetDefaults("test")
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupSingle, nil, "id1", false)
		assert.Equal(t,
			[]string{
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
				"--foo1",
				"--foo2",
			},
			cmdline,
		)
	}

	// Resilient single
	{
		apiObject := &api.ArangoDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "ns",
			},
			Spec: api.DeploymentSpec{
				XMode: api.NewMode(api.DeploymentModeResilientSingle),
			},
		}
		apiObject.Spec.SetDefaults("test")
		agents := api.MemberStatusList{
			api.MemberStatus{ID: "a1"},
			api.MemberStatus{ID: "a2"},
			api.MemberStatus{ID: "a3"},
		}
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupSingle, agents, "id1", false)
		assert.Equal(t,
			[]string{
				"--cluster.agency-endpoint=ssl://name-agent-a1.name-int.ns.svc:8529",
				"--cluster.agency-endpoint=ssl://name-agent-a2.name-int.ns.svc:8529",
				"--cluster.agency-endpoint=ssl://name-agent-a3.name-int.ns.svc:8529",
				"--cluster.my-address=ssl://name-single-id1.name-int.ns.svc:8529",
				"--cluster.my-role=SINGLE",
				"--database.directory=/data",
				"--foxx.queues=true",
				"--log.level=INFO",
				"--log.output=+",
				"--replication.automatic-failover=true",
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
}
