//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

package v2alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func TestServerGroupSpecValidateCount(t *testing.T) {
	// Valid
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](1)}.New().WithGroup(ServerGroupSingle).New().New().Validate(ServerGroupSingle, true, DeploymentModeSingle, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](0)}.New().WithGroup(ServerGroupSingle).New().Validate(ServerGroupSingle, false, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](1)}.New().WithGroup(ServerGroupAgents).New().Validate(ServerGroupAgents, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](3)}.New().WithGroup(ServerGroupAgents).New().Validate(ServerGroupAgents, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](1)}.New().WithGroup(ServerGroupAgents).New().Validate(ServerGroupAgents, true, DeploymentModeActiveFailover, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](3)}.New().WithGroup(ServerGroupAgents).New().Validate(ServerGroupAgents, true, DeploymentModeActiveFailover, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](2)}.New().WithGroup(ServerGroupDBServers).New().Validate(ServerGroupDBServers, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](6)}.New().WithGroup(ServerGroupDBServers).New().Validate(ServerGroupDBServers, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](1)}.New().WithGroup(ServerGroupCoordinators).New().Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](2)}.New().WithGroup(ServerGroupCoordinators).New().Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](3)}.New().WithGroup(ServerGroupAgents).New().Validate(ServerGroupAgents, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](3)}.New().WithGroup(ServerGroupAgents).New().Validate(ServerGroupAgents, true, DeploymentModeActiveFailover, EnvironmentProduction))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](2)}.New().WithGroup(ServerGroupDBServers).New().Validate(ServerGroupDBServers, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](2)}.New().WithGroup(ServerGroupCoordinators).New().Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](2)}.New().WithGroup(ServerGroupSyncMasters).New().Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](2)}.New().WithGroup(ServerGroupSyncWorkers).New().Validate(ServerGroupSyncWorkers, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](2)}.New().WithGroup(ServerGroupGateways).New().Validate(ServerGroupGateways, true, DeploymentModeCluster, EnvironmentProduction))

	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](2), MinCount: util.NewType[int](2), MaxCount: util.NewType[int](5)}.New().WithGroup(ServerGroupCoordinators).New().Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](1), MaxCount: util.NewType[int](5)}.New().WithGroup(ServerGroupCoordinators).New().Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](6), MinCount: util.NewType[int](2)}.New().WithGroup(ServerGroupCoordinators).New().Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](5), MinCount: util.NewType[int](5), MaxCount: util.NewType[int](5)}.New().WithGroup(ServerGroupCoordinators).New().Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](2)}.New().WithGroup(ServerGroupCoordinators).New().Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))

	// Invalid
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1)}.New().WithGroup(ServerGroupSingle).New().Validate(ServerGroupSingle, false, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](2)}.New().WithGroup(ServerGroupSingle).New().Validate(ServerGroupSingle, true, DeploymentModeSingle, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1)}.New().WithGroup(ServerGroupSingle).New().Validate(ServerGroupSingle, true, DeploymentModeActiveFailover, EnvironmentProduction))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](0)}.New().WithGroup(ServerGroupAgents).New().Validate(ServerGroupAgents, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](0)}.New().WithGroup(ServerGroupAgents).New().Validate(ServerGroupAgents, true, DeploymentModeActiveFailover, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](0)}.New().WithGroup(ServerGroupDBServers).New().Validate(ServerGroupDBServers, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](0)}.New().WithGroup(ServerGroupCoordinators).New().Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](0)}.New().WithGroup(ServerGroupSyncMasters).New().Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](0)}.New().WithGroup(ServerGroupSyncWorkers).New().Validate(ServerGroupSyncWorkers, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](-1)}.New().WithGroup(ServerGroupAgents).New().Validate(ServerGroupAgents, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](-1)}.New().WithGroup(ServerGroupAgents).New().Validate(ServerGroupAgents, true, DeploymentModeActiveFailover, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](-1)}.New().WithGroup(ServerGroupDBServers).New().Validate(ServerGroupDBServers, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](-1)}.New().WithGroup(ServerGroupCoordinators).New().Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](-1)}.New().WithGroup(ServerGroupSyncMasters).New().Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](-1)}.New().WithGroup(ServerGroupSyncWorkers).New().Validate(ServerGroupSyncWorkers, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](2)}.New().WithGroup(ServerGroupAgents).New().Validate(ServerGroupAgents, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](2)}.New().WithGroup(ServerGroupAgents).New().Validate(ServerGroupAgents, true, DeploymentModeActiveFailover, EnvironmentProduction))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1)}.New().WithGroup(ServerGroupDBServers).New().Validate(ServerGroupDBServers, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1)}.New().WithGroup(ServerGroupCoordinators).New().Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1)}.New().WithGroup(ServerGroupSyncMasters).New().Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1)}.New().WithGroup(ServerGroupSyncWorkers).New().Validate(ServerGroupSyncWorkers, true, DeploymentModeCluster, EnvironmentProduction))

	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](2), MinCount: util.NewType[int](5), MaxCount: util.NewType[int](1)}.New().WithGroup(ServerGroupCoordinators).New().Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](6), MaxCount: util.NewType[int](5)}.New().WithGroup(ServerGroupCoordinators).New().Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1), MinCount: util.NewType[int](2)}.New().WithGroup(ServerGroupCoordinators).New().Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))

}

func TestServerGroupSpecDefault(t *testing.T) {
	def := func(spec ServerGroupSpec, group ServerGroup, used bool, mode DeploymentMode) ServerGroupSpec {
		spec.SetDefaults(group, used, mode)
		return spec
	}

	assert.Equal(t, 1, def(ServerGroupSpec{}, ServerGroupSingle, true, DeploymentModeSingle).New().GetCount())
	assert.Equal(t, 2, def(ServerGroupSpec{}, ServerGroupSingle, true, DeploymentModeActiveFailover).New().GetCount())
	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupSingle, false, DeploymentModeCluster).New().GetCount())

	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupAgents, false, DeploymentModeSingle).New().GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupAgents, true, DeploymentModeActiveFailover).New().GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupAgents, true, DeploymentModeCluster).New().GetCount())

	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupDBServers, false, DeploymentModeSingle).New().GetCount())
	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupDBServers, false, DeploymentModeActiveFailover).New().GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupDBServers, true, DeploymentModeCluster).New().GetCount())

	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupCoordinators, false, DeploymentModeSingle).New().GetCount())
	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupCoordinators, false, DeploymentModeActiveFailover).New().GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupCoordinators, true, DeploymentModeCluster).New().GetCount())

	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupSyncMasters, false, DeploymentModeSingle).New().GetCount())
	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupSyncMasters, false, DeploymentModeActiveFailover).New().GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupSyncMasters, true, DeploymentModeCluster).New().GetCount())

	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupSyncWorkers, false, DeploymentModeSingle).New().GetCount())
	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupSyncWorkers, false, DeploymentModeActiveFailover).New().GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupSyncWorkers, true, DeploymentModeCluster).New().GetCount())

	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupGateways, false, DeploymentModeSingle).New().GetCount())
	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupGateways, false, DeploymentModeActiveFailover).New().GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupGateways, true, DeploymentModeCluster).New().GetCount())

	for _, g := range AllServerGroups {
		assert.Equal(t, 0, len(def(ServerGroupSpec{}, g, true, DeploymentModeSingle).Args))
		assert.Equal(t, "", def(ServerGroupSpec{}, g, true, DeploymentModeSingle).New().GetStorageClassName())
	}
}

func TestServerGroupSpecValidateArgs(t *testing.T) {
	// Valid
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](1), Args: []string{}}.New().WithDefaults(ServerGroupSingle, true, DeploymentModeSingle).New().Validate(ServerGroupSingle, true, DeploymentModeSingle, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](1), Args: []string{"--master.endpoint"}}.New().WithDefaults(ServerGroupSingle, true, DeploymentModeSingle).New().Validate(ServerGroupSingle, true, DeploymentModeSingle, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](1), Args: []string{}}.New().WithDefaults(ServerGroupSyncMasters, true, DeploymentModeCluster).New().Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](1), Args: []string{"--server.authentication=true"}}.New().WithDefaults(ServerGroupSyncMasters, true, DeploymentModeCluster).New().Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentDevelopment))
	// Invalid
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1), Args: []string{"--server.authentication=true"}}.New().WithDefaults(ServerGroupSingle, true, DeploymentModeSingle).New().Validate(ServerGroupSingle, true, DeploymentModeSingle, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1), Args: []string{"--server.authentication", "true"}}.New().WithDefaults(ServerGroupSingle, true, DeploymentModeSingle).New().Validate(ServerGroupSingle, true, DeploymentModeSingle, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1), Args: []string{"--master.endpoint=http://something"}}.New().WithDefaults(ServerGroupSyncMasters, true, DeploymentModeCluster).New().Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1), Args: []string{"--mq.type=strange"}}.New().WithDefaults(ServerGroupSyncMasters, true, DeploymentModeCluster).New().Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentDevelopment))
}

func TestServerGroupSpecMemoryReservation(t *testing.T) {
	// If group is nil
	assert.EqualValues(t, 1024, (*ServerGroupSpec)(nil).CalculateMemoryReservation(1024))

	// If reservation is nil
	assert.EqualValues(t, 1024, (&ServerGroupSpec{}).CalculateMemoryReservation(1024))

	// If reservation is 0
	assert.EqualValues(t, 1024, (&ServerGroupSpec{MemoryReservation: util.NewType[int64](0)}).CalculateMemoryReservation(1024))

	// If reservation is -1
	assert.EqualValues(t, 1024, (&ServerGroupSpec{MemoryReservation: util.NewType[int64](-1)}).CalculateMemoryReservation(1024))

	// If reservation is 10
	assert.EqualValues(t, 921, (&ServerGroupSpec{MemoryReservation: util.NewType[int64](10)}).CalculateMemoryReservation(1024))

	// If reservation is 25
	assert.EqualValues(t, 768, (&ServerGroupSpec{MemoryReservation: util.NewType[int64](25)}).CalculateMemoryReservation(1024))

	// If reservation is 50
	assert.EqualValues(t, 512, (&ServerGroupSpec{MemoryReservation: util.NewType[int64](50)}).CalculateMemoryReservation(1024))

	// If reservation is 51
	assert.EqualValues(t, 512, (&ServerGroupSpec{MemoryReservation: util.NewType[int64](51)}).CalculateMemoryReservation(1024))

	// If reservation is int max
	assert.EqualValues(t, 576460752303423488, (&ServerGroupSpec{MemoryReservation: util.NewType[int64](50)}).CalculateMemoryReservation(0xFFFFFFFFFFFFFFF))
	assert.EqualValues(t, 864691128455135232, (&ServerGroupSpec{MemoryReservation: util.NewType[int64](25)}).CalculateMemoryReservation(0xFFFFFFFFFFFFFFF))
	assert.EqualValues(t, 1037629354146162304, (&ServerGroupSpec{MemoryReservation: util.NewType[int64](10)}).CalculateMemoryReservation(0xFFFFFFFFFFFFFFF))
}
