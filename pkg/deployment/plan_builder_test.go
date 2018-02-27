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

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
)

// TestCreatePlanSingleScaling creates a `single` deployment to test the creating of scaling plan.
func TestCreatePlanSingleScaleUp(t *testing.T) {
	log := zerolog.Nop()
	spec := api.DeploymentSpec{
		Mode: api.DeploymentModeSingle,
	}
	spec.SetDefaults("test")

	// Test with empty status
	var status api.DeploymentStatus
	newPlan, changed := createPlan(log, nil, spec, status)
	assert.True(t, changed)
	assert.Len(t, newPlan, 0) // Single mode does not scale

	// Test with 1 single member
	status.Members.Single = api.MemberStatusList{
		api.MemberStatus{
			ID:      "id",
			PodName: "something",
		},
	}
	newPlan, changed = createPlan(log, nil, spec, status)
	assert.True(t, changed)
	assert.Len(t, newPlan, 0) // Single mode does not scale
}

// TestCreatePlanResilientSingleScaleUp creates a `resilientsingle` deployment to test the creating of scaling plan.
func TestCreatePlanResilientSingleScaleUp(t *testing.T) {
	log := zerolog.Nop()
	spec := api.DeploymentSpec{
		Mode: api.DeploymentModeResilientSingle,
	}
	spec.SetDefaults("test")
	spec.Single.Count = 2

	// Test with empty status
	var status api.DeploymentStatus
	newPlan, changed := createPlan(log, nil, spec, status)
	assert.True(t, changed)
	require.Len(t, newPlan, 2)
	assert.Equal(t, api.ActionTypeAddMember, newPlan[0].Type)
	assert.Equal(t, api.ActionTypeAddMember, newPlan[1].Type)

	// Test with 1 single member
	status.Members.Single = api.MemberStatusList{
		api.MemberStatus{
			ID:      "id",
			PodName: "something",
		},
	}
	newPlan, changed = createPlan(log, nil, spec, status)
	assert.True(t, changed)
	require.Len(t, newPlan, 1)
	assert.Equal(t, api.ActionTypeAddMember, newPlan[0].Type)
}
