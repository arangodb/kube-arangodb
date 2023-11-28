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

package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/storage/utils"
)

func Test_Nodes_OptionalScheduling(t *testing.T) {
	nodes := Nodes{
		{
			ObjectMeta: meta.ObjectMeta{
				Name: "NodeA",
			},
			Spec: core.NodeSpec{
				Taints: []core.Taint{
					{
						Key:    "test2",
						Value:  "test",
						Effect: core.TaintEffectPreferNoSchedule,
					},
				},
			},
		},
		{
			ObjectMeta: meta.ObjectMeta{
				Name: "NodeB",
			},
			Spec: core.NodeSpec{
				Taints: []core.Taint{
					{
						Key:    "test1",
						Value:  "test",
						Effect: core.TaintEffectNoSchedule,
					},
				},
			},
		},
		{
			ObjectMeta: meta.ObjectMeta{
				Name: "NodeC",
			},
			Spec: core.NodeSpec{
				Taints: []core.Taint{},
			},
		},
	}

	t.Run("Without tolerations", func(t *testing.T) {
		require.Len(t, nodes.FilterTaints(&core.Pod{
			Spec: core.PodSpec{
				Tolerations: []core.Toleration{},
			},
		}), 2)
	})

	t.Run("With toleration - match fully", func(t *testing.T) {
		require.Len(t, nodes.FilterTaints(&core.Pod{
			Spec: core.PodSpec{
				Tolerations: []core.Toleration{
					{
						Key:      "test1",
						Operator: core.TolerationOpEqual,
						Value:    "test",
						Effect:   core.TaintEffectNoSchedule,
					},
				},
			},
		}), 3)
	})

	t.Run("With toleration - invalid effect", func(t *testing.T) {
		require.Len(t, nodes.FilterTaints(&core.Pod{
			Spec: core.PodSpec{
				Tolerations: []core.Toleration{
					{
						Key:      "test1",
						Operator: core.TolerationOpEqual,
						Value:    "test",
						Effect:   core.TaintEffectNoExecute,
					},
				},
			},
		}), 2)
	})

	t.Run("With toleration - invalid value", func(t *testing.T) {
		require.Len(t, nodes.FilterTaints(&core.Pod{
			Spec: core.PodSpec{
				Tolerations: []core.Toleration{
					{
						Key:      "test1",
						Operator: core.TolerationOpEqual,
						Value:    "test1",
						Effect:   core.TaintEffectNoSchedule,
					},
				},
			},
		}), 2)
	})

	t.Run("With toleration - invalid key", func(t *testing.T) {
		require.Len(t, nodes.FilterTaints(&core.Pod{
			Spec: core.PodSpec{
				Tolerations: []core.Toleration{
					{
						Key:      "test",
						Operator: core.TolerationOpEqual,
						Value:    "test",
						Effect:   core.TaintEffectNoSchedule,
					},
				},
			},
		}), 2)
	})

	t.Run("With toleration - exists", func(t *testing.T) {
		require.Len(t, nodes.FilterTaints(&core.Pod{
			Spec: core.PodSpec{
				Tolerations: []core.Toleration{
					{
						Key:      "test1",
						Operator: core.TolerationOpExists,
						Value:    "test-445",
						Effect:   core.TaintEffectNoSchedule,
					},
				},
			},
		}), 3)
	})

	t.Run("Nodes order by optional", func(t *testing.T) {
		pod := &core.Pod{
			Spec: core.PodSpec{},
		}

		sorted := nodes.SortBySchedulablePodTaints(pod)

		actual := make([]utils.ScheduleOption, len(nodes))
		schedules := make([]utils.ScheduleOption, len(sorted))

		for id := range sorted {
			actual[id] = utils.IsNodeSchedulableForPod(nodes[id], pod)
			schedules[id] = utils.IsNodeSchedulableForPod(sorted[id], pod)
		}

		require.Equal(t, utils.ScheduleAllowed, schedules[0])
		require.Equal(t, utils.ScheduleOptional, schedules[1])
		require.Equal(t, utils.ScheduleBlocked, schedules[2])

		require.Equal(t, utils.ScheduleOptional, actual[0])
		require.Equal(t, utils.ScheduleBlocked, actual[1])
		require.Equal(t, utils.ScheduleAllowed, actual[2])
	})
}
