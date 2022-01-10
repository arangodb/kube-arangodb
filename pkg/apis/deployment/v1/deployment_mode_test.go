//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeploymentModeValidate(t *testing.T) {
	// Valid
	assert.Nil(t, DeploymentMode("Single").Validate())
	assert.Nil(t, DeploymentMode("ActiveFailover").Validate())
	assert.Nil(t, DeploymentMode("Cluster").Validate())

	// Not valid
	assert.Error(t, DeploymentMode("").Validate())
	assert.Error(t, DeploymentMode(" cluster").Validate())
	assert.Error(t, DeploymentMode("singles").Validate())
	assert.Error(t, DeploymentMode("single").Validate())
	assert.Error(t, DeploymentMode("activefailover").Validate())
	assert.Error(t, DeploymentMode("cluster").Validate())
}

func TestDeploymentModeHasX(t *testing.T) {
	assert.True(t, DeploymentModeSingle.HasSingleServers())
	assert.True(t, DeploymentModeActiveFailover.HasSingleServers())
	assert.False(t, DeploymentModeCluster.HasSingleServers())

	assert.False(t, DeploymentModeSingle.HasAgents())
	assert.True(t, DeploymentModeActiveFailover.HasAgents())
	assert.True(t, DeploymentModeCluster.HasAgents())

	assert.False(t, DeploymentModeSingle.HasDBServers())
	assert.False(t, DeploymentModeActiveFailover.HasDBServers())
	assert.True(t, DeploymentModeCluster.HasDBServers())

	assert.False(t, DeploymentModeSingle.HasCoordinators())
	assert.False(t, DeploymentModeActiveFailover.HasCoordinators())
	assert.True(t, DeploymentModeCluster.HasCoordinators())
}

func TestDeploymentModeSupportsSync(t *testing.T) {
	assert.False(t, DeploymentModeSingle.SupportsSync())
	assert.False(t, DeploymentModeActiveFailover.SupportsSync())
	assert.True(t, DeploymentModeCluster.SupportsSync())
}
