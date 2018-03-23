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

package v1alpha

import (
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestServerGroupSpecValidateCount(t *testing.T) {
	// Valid
	assert.Nil(t, ServerGroupSpec{XCount: util.NewInt(1)}.Validate(ServerGroupSingle, true, DeploymentModeSingle, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{XCount: util.NewInt(0)}.Validate(ServerGroupSingle, false, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{XCount: util.NewInt(1)}.Validate(ServerGroupAgents, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{XCount: util.NewInt(3)}.Validate(ServerGroupAgents, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{XCount: util.NewInt(1)}.Validate(ServerGroupAgents, true, DeploymentModeResilientSingle, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{XCount: util.NewInt(3)}.Validate(ServerGroupAgents, true, DeploymentModeResilientSingle, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{XCount: util.NewInt(1)}.Validate(ServerGroupDBServers, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{XCount: util.NewInt(6)}.Validate(ServerGroupDBServers, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{XCount: util.NewInt(1)}.Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{XCount: util.NewInt(2)}.Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Nil(t, ServerGroupSpec{XCount: util.NewInt(3)}.Validate(ServerGroupAgents, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Nil(t, ServerGroupSpec{XCount: util.NewInt(3)}.Validate(ServerGroupAgents, true, DeploymentModeResilientSingle, EnvironmentProduction))
	assert.Nil(t, ServerGroupSpec{XCount: util.NewInt(2)}.Validate(ServerGroupDBServers, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Nil(t, ServerGroupSpec{XCount: util.NewInt(2)}.Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Nil(t, ServerGroupSpec{XCount: util.NewInt(2)}.Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Nil(t, ServerGroupSpec{XCount: util.NewInt(2)}.Validate(ServerGroupSyncWorkers, true, DeploymentModeCluster, EnvironmentProduction))

	// Invalid
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(1)}.Validate(ServerGroupSingle, false, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(2)}.Validate(ServerGroupSingle, true, DeploymentModeSingle, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(1)}.Validate(ServerGroupSingle, true, DeploymentModeResilientSingle, EnvironmentProduction))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(0)}.Validate(ServerGroupAgents, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(0)}.Validate(ServerGroupAgents, true, DeploymentModeResilientSingle, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(0)}.Validate(ServerGroupDBServers, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(0)}.Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(0)}.Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(0)}.Validate(ServerGroupSyncWorkers, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(-1)}.Validate(ServerGroupAgents, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(-1)}.Validate(ServerGroupAgents, true, DeploymentModeResilientSingle, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(-1)}.Validate(ServerGroupDBServers, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(-1)}.Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(-1)}.Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(-1)}.Validate(ServerGroupSyncWorkers, true, DeploymentModeCluster, EnvironmentDevelopment))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(2)}.Validate(ServerGroupAgents, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(2)}.Validate(ServerGroupAgents, true, DeploymentModeResilientSingle, EnvironmentProduction))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(1)}.Validate(ServerGroupDBServers, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(1)}.Validate(ServerGroupCoordinators, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(1)}.Validate(ServerGroupSyncMasters, true, DeploymentModeCluster, EnvironmentProduction))
	assert.Error(t, ServerGroupSpec{XCount: util.NewInt(1)}.Validate(ServerGroupSyncWorkers, true, DeploymentModeCluster, EnvironmentProduction))
}

func TestServerGroupSpecDefault(t *testing.T) {
	def := func(spec ServerGroupSpec, group ServerGroup, used bool, mode DeploymentMode) ServerGroupSpec {
		spec.SetDefaults(group, used, mode)
		return spec
	}

	assert.Equal(t, 1, def(ServerGroupSpec{}, ServerGroupSingle, true, DeploymentModeSingle).GetCount())
	assert.Equal(t, 2, def(ServerGroupSpec{}, ServerGroupSingle, true, DeploymentModeResilientSingle).GetCount())
	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupSingle, false, DeploymentModeCluster).GetCount())

	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupAgents, false, DeploymentModeSingle).GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupAgents, true, DeploymentModeResilientSingle).GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupAgents, true, DeploymentModeCluster).GetCount())

	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupDBServers, false, DeploymentModeSingle).GetCount())
	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupDBServers, false, DeploymentModeResilientSingle).GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupDBServers, true, DeploymentModeCluster).GetCount())

	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupCoordinators, false, DeploymentModeSingle).GetCount())
	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupCoordinators, false, DeploymentModeResilientSingle).GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupCoordinators, true, DeploymentModeCluster).GetCount())

	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupSyncMasters, false, DeploymentModeSingle).GetCount())
	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupSyncMasters, false, DeploymentModeResilientSingle).GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupSyncMasters, true, DeploymentModeCluster).GetCount())

	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupSyncWorkers, false, DeploymentModeSingle).GetCount())
	assert.Equal(t, 0, def(ServerGroupSpec{}, ServerGroupSyncWorkers, false, DeploymentModeResilientSingle).GetCount())
	assert.Equal(t, 3, def(ServerGroupSpec{}, ServerGroupSyncWorkers, true, DeploymentModeCluster).GetCount())

	for _, g := range AllServerGroups {
		assert.Equal(t, 0, len(def(ServerGroupSpec{}, g, true, DeploymentModeSingle).Args))
		assert.Equal(t, "", def(ServerGroupSpec{}, g, true, DeploymentModeSingle).GetStorageClassName())
	}
}
