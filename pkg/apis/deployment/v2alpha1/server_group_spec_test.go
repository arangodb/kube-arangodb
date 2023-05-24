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

package v2alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func TestServerGroupSpecValidateCount(t *testing.T) {
	// Valid
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](1)}.WithGroup(ServerGroupSingle).Validate(ServerGroupSingle, true, DeploymentModeSingle, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](0)}.WithGroup(ServerGroupSingle).Validate(ServerGroupSingle, false, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](1)}.WithGroup(ServerGroupAgents).Validate(ServerGroupAgents, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](3)}.WithGroup(ServerGroupAgents).Validate(ServerGroupAgents, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](1)}.WithGroup(ServerGroupAgents).Validate(ServerGroupAgents, true, DeploymentModeActiveFailover, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](3)}.WithGroup(ServerGroupAgents).Validate(ServerGroupAgents, true, DeploymentModeActiveFailover, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](2)}.WithGroup(ServerGroupDBServers).Validate(ServerGroupDBServers, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](6)}.WithGroup(ServerGroupDBServers).Validate(ServerGroupDBServers, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](1)}.WithGroup(ServerGroupCoordinators).Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](2)}.WithGroup(ServerGroupCoordinators).Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](3)}.WithGroup(ServerGroupAgents).Validate(ServerGroupAgents, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](3)}.WithGroup(ServerGroupAgents).Validate(ServerGroupAgents, true, DeploymentModeActiveFailover, EnvironmentProduction))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](2)}.WithGroup(ServerGroupDBServers).Validate(ServerGroupDBServers, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](2)}.WithGroup(ServerGroupCoordinators).Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](2)}.WithGroup(ServerGroupSyncMasters).Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](2)}.WithGroup(ServerGroupSyncWorkers).Validate(ServerGroupSyncWorkers, true, DeploymentModeCluster, EnvironmentProduction))

	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](2), MinCount: util.NewType[int](2), MaxCount: util.NewType[int](5)}.WithGroup(ServerGroupCoordinators).Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](1), MaxCount: util.NewType[int](5)}.WithGroup(ServerGroupCoordinators).Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](6), MinCount: util.NewType[int](2)}.WithGroup(ServerGroupCoordinators).Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](5), MinCount: util.NewType[int](5), MaxCount: util.NewType[int](5)}.WithGroup(ServerGroupCoordinators).Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](2)}.WithGroup(ServerGroupCoordinators).Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))

	// Invalid
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1)}.WithGroup(ServerGroupSingle).Validate(ServerGroupSingle, false, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](2)}.WithGroup(ServerGroupSingle).Validate(ServerGroupSingle, true, DeploymentModeSingle, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1)}.WithGroup(ServerGroupSingle).Validate(ServerGroupSingle, true, DeploymentModeActiveFailover, EnvironmentProduction))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](0)}.WithGroup(ServerGroupAgents).Validate(ServerGroupAgents, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](0)}.WithGroup(ServerGroupAgents).Validate(ServerGroupAgents, true, DeploymentModeActiveFailover, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](0)}.WithGroup(ServerGroupDBServers).Validate(ServerGroupDBServers, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](0)}.WithGroup(ServerGroupCoordinators).Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](0)}.WithGroup(ServerGroupSyncMasters).Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](0)}.WithGroup(ServerGroupSyncWorkers).Validate(ServerGroupSyncWorkers, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](-1)}.WithGroup(ServerGroupAgents).Validate(ServerGroupAgents, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](-1)}.WithGroup(ServerGroupAgents).Validate(ServerGroupAgents, true, DeploymentModeActiveFailover, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](-1)}.WithGroup(ServerGroupDBServers).Validate(ServerGroupDBServers, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](-1)}.WithGroup(ServerGroupCoordinators).Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](-1)}.WithGroup(ServerGroupSyncMasters).Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](-1)}.WithGroup(ServerGroupSyncWorkers).Validate(ServerGroupSyncWorkers, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](2)}.WithGroup(ServerGroupAgents).Validate(ServerGroupAgents, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](2)}.WithGroup(ServerGroupAgents).Validate(ServerGroupAgents, true, DeploymentModeActiveFailover, EnvironmentProduction))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1)}.WithGroup(ServerGroupDBServers).Validate(ServerGroupDBServers, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1)}.WithGroup(ServerGroupCoordinators).Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1)}.WithGroup(ServerGroupSyncMasters).Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1)}.WithGroup(ServerGroupSyncWorkers).Validate(ServerGroupSyncWorkers, true, DeploymentModeCluster, EnvironmentProduction))

	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](2), MinCount: util.NewType[int](5), MaxCount: util.NewType[int](1)}.WithGroup(ServerGroupCoordinators).Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](6), MaxCount: util.NewType[int](5)}.WithGroup(ServerGroupCoordinators).Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1), MinCount: util.NewType[int](2)}.WithGroup(ServerGroupCoordinators).Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))

}

func TestServerGroupSpecDefault(t *testing.T) {
	def := func(spec ServerGroupSpec, group ServerGroup, used bool, mode DeploymentMode) ServerGroupSpec {
		spec.SetDefaults(group, used, mode)
		return spec
	}

	assert.Equal(t, 1, def(ServerGroupSpec{}, ServerGroupSingle, true, DeploymentModeSingle).GetCount())
	assert.Equal(t, 2, def(ServerGroupSpec{}, ServerGroupSingle, true, DeploymentModeActiveFailover).GetCount())
	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupSingle, false, DeploymentModeCluster).GetCount())

	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupAgents, false, DeploymentModeSingle).GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupAgents, true, DeploymentModeActiveFailover).GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupAgents, true, DeploymentModeCluster).GetCount())

	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupDBServers, false, DeploymentModeSingle).GetCount())
	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupDBServers, false, DeploymentModeActiveFailover).GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupDBServers, true, DeploymentModeCluster).GetCount())

	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupCoordinators, false, DeploymentModeSingle).GetCount())
	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupCoordinators, false, DeploymentModeActiveFailover).GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupCoordinators, true, DeploymentModeCluster).GetCount())

	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupSyncMasters, false, DeploymentModeSingle).GetCount())
	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupSyncMasters, false, DeploymentModeActiveFailover).GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupSyncMasters, true, DeploymentModeCluster).GetCount())

	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupSyncWorkers, false, DeploymentModeSingle).GetCount())
	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupSyncWorkers, false, DeploymentModeActiveFailover).GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupSyncWorkers, true, DeploymentModeCluster).GetCount())

	for _, g := range AllServerGroups {
		assert.Equal(t, 0, len(def(ServerGroupSpec{}, g, true, DeploymentModeSingle).Args))
		assert.Equal(t, "", def(ServerGroupSpec{}, g, true, DeploymentModeSingle).GetStorageClassName())
	}
}

func TestServerGroupSpecValidateArgs(t *testing.T) {
	// Valid
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](1), Args: []string{}}.WithDefaults(ServerGroupSingle, true, DeploymentModeSingle).Validate(ServerGroupSingle, true, DeploymentModeSingle, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](1), Args: []string{"--master.endpoint"}}.WithDefaults(ServerGroupSingle, true, DeploymentModeSingle).Validate(ServerGroupSingle, true, DeploymentModeSingle, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](1), Args: []string{}}.WithDefaults(ServerGroupSyncMasters, true, DeploymentModeCluster).Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{Count: util.NewType[int](1), Args: []string{"--server.authentication=true"}}.WithDefaults(ServerGroupSyncMasters, true, DeploymentModeCluster).Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentDevelopment))
	// Invalid
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1), Args: []string{"--server.authentication=true"}}.WithDefaults(ServerGroupSingle, true, DeploymentModeSingle).Validate(ServerGroupSingle, true, DeploymentModeSingle, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1), Args: []string{"--server.authentication", "true"}}.WithDefaults(ServerGroupSingle, true, DeploymentModeSingle).Validate(ServerGroupSingle, true, DeploymentModeSingle, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1), Args: []string{"--master.endpoint=http://something"}}.WithDefaults(ServerGroupSyncMasters, true, DeploymentModeCluster).Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{Count: util.NewType[int](1), Args: []string{"--mq.type=strange"}}.WithDefaults(ServerGroupSyncMasters, true, DeploymentModeCluster).Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentDevelopment))
}
