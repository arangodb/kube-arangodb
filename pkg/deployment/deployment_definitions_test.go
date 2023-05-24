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

package deployment

import (
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

var (
	defaultAgentTerminationTimeout       = int64(api.ServerGroupAgents.DefaultTerminationGracePeriod().Seconds())
	defaultDBServerTerminationTimeout    = int64(api.ServerGroupDBServers.DefaultTerminationGracePeriod().Seconds())
	defaultCoordinatorTerminationTimeout = int64(api.ServerGroupCoordinators.DefaultTerminationGracePeriod().Seconds())
	defaultSingleTerminationTimeout      = int64(api.ServerGroupSingle.DefaultTerminationGracePeriod().Seconds())
	defaultSyncMasterTerminationTimeout  = int64(api.ServerGroupSyncMasters.DefaultTerminationGracePeriod().Seconds())
	defaultSyncWorkerTerminationTimeout  = int64(api.ServerGroupSyncWorkers.DefaultTerminationGracePeriod().Seconds())

	securityContext api.ServerGroupSpecSecurityContext

	nodeSelectorTest = map[string]string{
		"test": "test",
	}

	arch = api.ArangoDeploymentArchitectureDefault

	firstAgentStatus = api.MemberStatus{
		ID:           "agent1",
		Phase:        api.MemberPhaseNone,
		Architecture: &arch,
	}

	firstCoordinatorStatus = api.MemberStatus{
		ID:           "coordinator1",
		Phase:        api.MemberPhaseNone,
		Architecture: &arch,
	}

	singleStatus = api.MemberStatus{
		ID:           "single1",
		Phase:        api.MemberPhaseNone,
		Architecture: &arch,
	}

	firstSyncMaster = api.MemberStatus{
		ID:           "syncMaster1",
		Phase:        api.MemberPhaseNone,
		Architecture: &arch,
	}

	firstSyncWorker = api.MemberStatus{
		ID:           "syncWorker1",
		Phase:        api.MemberPhaseNone,
		Architecture: &arch,
	}

	firstDBServerStatus = api.MemberStatus{
		ID:           "DBserver1",
		Phase:        api.MemberPhaseNone,
		Architecture: &arch,
	}

	noAuthentication = api.AuthenticationSpec{
		JWTSecretName: util.NewType[string](api.JWTSecretNameDisabled),
	}

	noTLS = api.TLSSpec{
		CASecretName: util.NewType[string](api.CASecretNameDisabled),
	}

	authenticationSpec = api.AuthenticationSpec{
		JWTSecretName: util.NewType[string](testJWTSecretName),
	}
	tlsSpec = api.TLSSpec{
		CASecretName: util.NewType[string](testCASecretName),
	}

	rocksDBSpec = api.RocksDBSpec{
		Encryption: api.RocksDBEncryptionSpec{
			KeySecretName: util.NewType[string](testRocksDBEncryptionKey),
		},
	}

	metricsSpec = api.MetricsSpec{
		Enabled: util.NewType[bool](true),
		Image:   util.NewType[string](testImage),
		Authentication: api.MetricsAuthenticationSpec{
			JWTTokenSecretName: util.NewType[string](testExporterToken),
		},
	}

	resourcesUnfiltered = core.ResourceRequirements{
		Limits: core.ResourceList{
			core.ResourceCPU:    resource.MustParse("500m"),
			core.ResourceMemory: resource.MustParse("2Gi"),
		},
		Requests: core.ResourceList{
			core.ResourceCPU:    resource.MustParse("100m"),
			core.ResourceMemory: resource.MustParse("1Gi"),
		},
	}

	emptyResources = core.ResourceRequirements{
		Limits:   make(core.ResourceList),
		Requests: make(core.ResourceList),
	}

	sidecarName1 = "sidecar1"
	sidecarName2 = "sidecar2"
)
