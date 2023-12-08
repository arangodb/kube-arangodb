//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
)

func Test_Scheduling_Validate(t *testing.T) {
	initSelector := map[string]string{
		"init": "true",
	}
	initAffinity := &core.Affinity{
		PodAffinity: &core.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []core.PodAffinityTerm{
				{
					TopologyKey: "init-key-pod",
				},
			},
		},
	}

	s := &Scheduling{}
	var noSched *Scheduling
	sched := Scheduling{
		NodeSelector: initSelector,
		Affinity:     initAffinity.DeepCopy(),
	}

	t.Run("basic", func(t *testing.T) {
		require.NoError(t, s.Validate())

		require.Nil(t, s.GetNodeSelector())
		require.Nil(t, s.GetAffinity())
		require.Nil(t, s.GetTolerations())
	})

	t.Run("nodeSelector", func(t *testing.T) {
		def := map[string]string{
			"test": "true",
		}
		require.Equal(t, def, s.WithNodeSelector(def))
		require.Nil(t, s.GetNodeSelector())

		require.Nil(t, noSched.GetNodeSelector())

		require.Equal(t, def, noSched.WithNodeSelector(def))
		require.Nil(t, noSched.GetNodeSelector())

		expected := map[string]string{
			"test": "true",
			"init": "true",
		}
		require.Equal(t, expected, sched.WithNodeSelector(def))
		require.Equal(t, initSelector, sched.GetNodeSelector())
	})

	t.Run("affinity", func(t *testing.T) {
		def := &core.Affinity{
			NodeAffinity: &core.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
					NodeSelectorTerms: []core.NodeSelectorTerm{
						{
							MatchExpressions: []core.NodeSelectorRequirement{
								{
									Key:      "default-key-node",
									Operator: core.NodeSelectorOpExists,
								},
							},
						},
					},
				},
			},
		}
		require.Equal(t, def, s.WithAffinity(def))
		require.Nil(t, s.GetAffinity())

		require.Nil(t, noSched.GetAffinity())

		require.Equal(t, def, noSched.WithAffinity(def))
		require.Nil(t, noSched.GetAffinity())

		expected := &core.Affinity{
			NodeAffinity: &core.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
					NodeSelectorTerms: []core.NodeSelectorTerm{
						{
							MatchExpressions: []core.NodeSelectorRequirement{
								{
									Key:      "default-key-node",
									Operator: core.NodeSelectorOpExists,
								},
							},
						},
					},
				},
			},
			PodAffinity: &core.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []core.PodAffinityTerm{
					{
						TopologyKey: "init-key-pod",
					},
				},
			},
		}
		actual := sched.WithAffinity(def)
		require.Equal(t, expected, actual)
		require.Equal(t, initAffinity, sched.GetAffinity())
	})
}
