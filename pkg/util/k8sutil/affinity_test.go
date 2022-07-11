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

package k8sutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

func TestCreateAffinity(t *testing.T) {
	expectedNodeAffinity := &core.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
			NodeSelectorTerms: []core.NodeSelectorTerm{
				{
					MatchExpressions: []core.NodeSelectorRequirement{
						{
							Key:      shared.NodeArchAffinityLabel,
							Operator: "In",
							Values:   []string{"amd64"},
						},
					},
				},
			},
		},
	}
	// Required
	a := CreateAffinity("test", "role", true, "")
	assert.Nil(t, a.PodAffinity)
	require.NotNil(t, a.PodAntiAffinity)
	require.Len(t, a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, 1)
	assert.Len(t, a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 0)
	assert.Equal(t, "kubernetes.io/hostname", a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].TopologyKey)
	require.NotNil(t, a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector)
	assert.Equal(t, "test", a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels["arango_deployment"])
	assert.Equal(t, "arangodb", a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels["app"])
	assert.Equal(t, "role", a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels["role"])
	assert.Equal(t, expectedNodeAffinity, a.NodeAffinity)

	// Require & affinity with role dbserver
	a = CreateAffinity("test", "role", true, "dbserver")
	require.NotNil(t, a.PodAffinity)
	require.Len(t, a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, 1)
	assert.Len(t, a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 0)
	assert.Equal(t, "kubernetes.io/hostname", a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].TopologyKey)
	require.NotNil(t, a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector)
	assert.Equal(t, "test", a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels["arango_deployment"])
	assert.Equal(t, "arangodb", a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels["app"])
	assert.Equal(t, "dbserver", a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels["role"])
	assert.Equal(t, expectedNodeAffinity, a.NodeAffinity)

	require.NotNil(t, a.PodAntiAffinity)
	require.Len(t, a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, 1)
	assert.Len(t, a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 0)
	assert.Equal(t, "kubernetes.io/hostname", a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].TopologyKey)
	require.NotNil(t, a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector)
	assert.Equal(t, "test", a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels["arango_deployment"])
	assert.Equal(t, "arangodb", a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels["app"])
	assert.Equal(t, "role", a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels["role"])
	assert.Equal(t, expectedNodeAffinity, a.NodeAffinity)

	// Not Required
	a = CreateAffinity("test", "role", false, "")
	assert.Nil(t, a.PodAffinity)
	require.NotNil(t, a.PodAntiAffinity)
	assert.Len(t, a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, 0)
	require.Len(t, a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 1)
	assert.Equal(t, "kubernetes.io/hostname", a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.TopologyKey)
	require.NotNil(t, a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector)
	assert.Equal(t, "test", a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels["arango_deployment"])
	assert.Equal(t, "arangodb", a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels["app"])
	assert.Equal(t, "role", a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels["role"])
	assert.Equal(t, expectedNodeAffinity, a.NodeAffinity)

	// Not Required & affinity with role dbserver
	a = CreateAffinity("test", "role", false, "dbserver")
	require.NotNil(t, a.PodAffinity)
	require.Len(t, a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 1)
	assert.Len(t, a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, 0)
	assert.Equal(t, "kubernetes.io/hostname", a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.TopologyKey)
	require.NotNil(t, a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector)
	assert.Equal(t, "test", a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels["arango_deployment"])
	assert.Equal(t, "arangodb", a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels["app"])
	assert.Equal(t, "dbserver", a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels["role"])
	assert.Equal(t, expectedNodeAffinity, a.NodeAffinity)

	require.NotNil(t, a.PodAntiAffinity)
	require.Len(t, a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 1)
	assert.Len(t, a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, 0)
	assert.Equal(t, "kubernetes.io/hostname", a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.TopologyKey)
	require.NotNil(t, a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector)
	assert.Equal(t, "test", a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels["arango_deployment"])
	assert.Equal(t, "arangodb", a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels["app"])
	assert.Equal(t, "role", a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels["role"])
	assert.Equal(t, expectedNodeAffinity, a.NodeAffinity)
}
