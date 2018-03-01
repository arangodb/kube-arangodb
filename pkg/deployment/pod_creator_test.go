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

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	"github.com/stretchr/testify/assert"
)

// TestCreateArangodArgs tests createArangodArgs.
func TestCreateArangodArgsSingle(t *testing.T) {
	// Default deployment
	{
		apiObject := &api.ArangoDeployment{
			Spec: api.DeploymentSpec{
				Mode: api.DeploymentModeSingle,
			},
		}
		apiObject.Spec.SetDefaults("test")
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupSingle, apiObject.Spec.Single, nil, "id1")
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
				"--server.storage-engine=mmfiles",
			},
			cmdline,
		)
	}

	// Default deployment with mmfiles
	{
		apiObject := &api.ArangoDeployment{
			Spec: api.DeploymentSpec{
				Mode:          api.DeploymentModeSingle,
				StorageEngine: api.StorageEngineMMFiles,
			},
		}
		apiObject.Spec.SetDefaults("test")
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupSingle, apiObject.Spec.Single, nil, "id1")
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
				"--server.storage-engine=mmfiles",
			},
			cmdline,
		)
	}

	// Default deployment with rocksdb
	{
		apiObject := &api.ArangoDeployment{
			Spec: api.DeploymentSpec{
				Mode:          api.DeploymentModeSingle,
				StorageEngine: api.StorageEngineRocksDB,
			},
		}
		apiObject.Spec.SetDefaults("test")
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupSingle, apiObject.Spec.Single, nil, "id1")
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

	// No authentication
	{
		apiObject := &api.ArangoDeployment{
			Spec: api.DeploymentSpec{
				Mode: api.DeploymentModeSingle,
			},
		}
		apiObject.Spec.Authentication.JWTSecretName = "None"
		apiObject.Spec.SetDefaults("test")
		cmdline := createArangodArgs(apiObject, apiObject.Spec, api.ServerGroupSingle, apiObject.Spec.Single, nil, "id1")
		assert.Equal(t,
			[]string{
				"--database.directory=/data",
				"--foxx.queues=true",
				"--log.level=INFO",
				"--log.output=+",
				"--server.authentication=false",
				"--server.endpoint=tcp://[::]:8529",
				"--server.statistics=true",
				"--server.storage-engine=mmfiles",
			},
			cmdline,
		)
	}
}
