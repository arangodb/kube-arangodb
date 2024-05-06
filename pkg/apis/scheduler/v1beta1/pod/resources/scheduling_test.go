//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func applyScheduling(t *testing.T, template *core.PodTemplateSpec, ns ...*Scheduling) func(in func(t *testing.T, pod *core.PodTemplateSpec)) {
	var i *Scheduling

	for _, n := range ns {
		require.NoError(t, n.Validate())

		i = i.With(n)

		require.NoError(t, i.Validate())
	}

	template = template.DeepCopy()

	if template == nil {
		template = &core.PodTemplateSpec{}
	}

	require.NoError(t, i.Apply(template))

	return func(in func(t *testing.T, spec *core.PodTemplateSpec)) {
		t.Run("Validate", func(t *testing.T) {
			in(t, template)
		})
	}
}

func Test_Scheduling_Affinity(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		applyScheduling(t, nil)(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Nil(t, pod.Spec.Affinity)
		})
	})
	t.Run("Empty Template", func(t *testing.T) {
		applyScheduling(t, &core.PodTemplateSpec{}, &Scheduling{
			Affinity: &core.Affinity{},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Nil(t, pod.Spec.Affinity)
		})
	})
	t.Run("Empty", func(t *testing.T) {
		applyScheduling(t, nil, &Scheduling{
			Affinity: &core.Affinity{},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Nil(t, pod.Spec.Affinity)
		})
	})
	t.Run("PodAffinity", func(t *testing.T) {
		t.Run("Empty", func(t *testing.T) {
			applyScheduling(t, nil, &Scheduling{
				Affinity: &core.Affinity{
					PodAffinity: &core.PodAffinity{},
				},
			})(func(t *testing.T, pod *core.PodTemplateSpec) {
				require.NotNil(t, pod.Spec.Affinity)
				require.NotNil(t, pod.Spec.Affinity.PodAffinity)
			})
		})

		t.Run("One", func(t *testing.T) {
			applyScheduling(t, nil, &Scheduling{
				Affinity: &core.Affinity{
					PodAffinity: &core.PodAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []core.PodAffinityTerm{
							{
								TopologyKey: "test",
							},
						},
						PreferredDuringSchedulingIgnoredDuringExecution: []core.WeightedPodAffinityTerm{
							{
								PodAffinityTerm: core.PodAffinityTerm{
									TopologyKey: "test",
								},
							},
						},
					},
				},
			})(func(t *testing.T, pod *core.PodTemplateSpec) {
				require.NotNil(t, pod.Spec.Affinity)
				require.NotNil(t, pod.Spec.Affinity.PodAffinity)

				require.Len(t, pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, 1)

				require.EqualValues(t, "test", pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].TopologyKey)

				require.Len(t, pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 1)

				require.EqualValues(t, "test", pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.TopologyKey)
			})
		})

		t.Run("Merge - with Empty", func(t *testing.T) {
			applyScheduling(t, nil, &Scheduling{
				Affinity: &core.Affinity{
					PodAffinity: &core.PodAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []core.PodAffinityTerm{
							{
								TopologyKey: "test",
							},
						},
						PreferredDuringSchedulingIgnoredDuringExecution: []core.WeightedPodAffinityTerm{
							{
								PodAffinityTerm: core.PodAffinityTerm{
									TopologyKey: "test",
								},
							},
						},
					},
				},
			}, &Scheduling{
				Affinity: &core.Affinity{
					PodAffinity: &core.PodAffinity{},
				},
			})(func(t *testing.T, pod *core.PodTemplateSpec) {
				require.NotNil(t, pod.Spec.Affinity)
				require.NotNil(t, pod.Spec.Affinity.PodAffinity)

				require.Len(t, pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, 1)

				require.EqualValues(t, "test", pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].TopologyKey)

				require.Len(t, pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 1)

				require.EqualValues(t, "test", pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.TopologyKey)
			})
		})

		t.Run("Merge", func(t *testing.T) {
			applyScheduling(t, nil, &Scheduling{
				Affinity: &core.Affinity{
					PodAffinity: &core.PodAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []core.PodAffinityTerm{
							{
								TopologyKey: "test",
							},
						},
						PreferredDuringSchedulingIgnoredDuringExecution: []core.WeightedPodAffinityTerm{
							{
								PodAffinityTerm: core.PodAffinityTerm{
									TopologyKey: "test",
								},
							},
						},
					},
				},
			}, &Scheduling{
				Affinity: &core.Affinity{
					PodAffinity: &core.PodAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []core.PodAffinityTerm{
							{
								TopologyKey: "test2",
							},
						},
						PreferredDuringSchedulingIgnoredDuringExecution: []core.WeightedPodAffinityTerm{
							{
								PodAffinityTerm: core.PodAffinityTerm{
									TopologyKey: "test2",
								},
							},
						},
					},
				},
			})(func(t *testing.T, pod *core.PodTemplateSpec) {
				require.NotNil(t, pod.Spec.Affinity)
				require.NotNil(t, pod.Spec.Affinity.PodAffinity)

				require.Len(t, pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, 2)

				require.EqualValues(t, "test", pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].TopologyKey)
				require.EqualValues(t, "test2", pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution[1].TopologyKey)

				require.Len(t, pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 2)

				require.EqualValues(t, "test", pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.TopologyKey)
				require.EqualValues(t, "test2", pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[1].PodAffinityTerm.TopologyKey)
			})
		})
	})

	t.Run("PodAntiAffinity", func(t *testing.T) {
		t.Run("Empty", func(t *testing.T) {
			applyScheduling(t, nil, &Scheduling{
				Affinity: &core.Affinity{
					PodAntiAffinity: &core.PodAntiAffinity{},
				},
			})(func(t *testing.T, pod *core.PodTemplateSpec) {
				require.NotNil(t, pod.Spec.Affinity)
				require.NotNil(t, pod.Spec.Affinity.PodAntiAffinity)
			})
		})

		t.Run("One", func(t *testing.T) {
			applyScheduling(t, nil, &Scheduling{
				Affinity: &core.Affinity{
					PodAntiAffinity: &core.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []core.PodAffinityTerm{
							{
								TopologyKey: "test",
							},
						},
						PreferredDuringSchedulingIgnoredDuringExecution: []core.WeightedPodAffinityTerm{
							{
								PodAffinityTerm: core.PodAffinityTerm{
									TopologyKey: "test",
								},
							},
						},
					},
				},
			})(func(t *testing.T, pod *core.PodTemplateSpec) {
				require.NotNil(t, pod.Spec.Affinity)
				require.NotNil(t, pod.Spec.Affinity.PodAntiAffinity)

				require.Len(t, pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, 1)

				require.EqualValues(t, "test", pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].TopologyKey)

				require.Len(t, pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 1)

				require.EqualValues(t, "test", pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.TopologyKey)
			})
		})

		t.Run("Merge - with Empty", func(t *testing.T) {
			applyScheduling(t, nil, &Scheduling{
				Affinity: &core.Affinity{
					PodAntiAffinity: &core.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []core.PodAffinityTerm{
							{
								TopologyKey: "test",
							},
						},
						PreferredDuringSchedulingIgnoredDuringExecution: []core.WeightedPodAffinityTerm{
							{
								PodAffinityTerm: core.PodAffinityTerm{
									TopologyKey: "test",
								},
							},
						},
					},
				},
			}, &Scheduling{
				Affinity: &core.Affinity{
					PodAntiAffinity: &core.PodAntiAffinity{},
				},
			})(func(t *testing.T, pod *core.PodTemplateSpec) {
				require.NotNil(t, pod.Spec.Affinity)
				require.NotNil(t, pod.Spec.Affinity.PodAntiAffinity)

				require.Len(t, pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, 1)

				require.EqualValues(t, "test", pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].TopologyKey)

				require.Len(t, pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 1)

				require.EqualValues(t, "test", pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.TopologyKey)
			})
		})

		t.Run("Merge", func(t *testing.T) {
			applyScheduling(t, nil, &Scheduling{
				Affinity: &core.Affinity{
					PodAntiAffinity: &core.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []core.PodAffinityTerm{
							{
								TopologyKey: "test",
							},
						},
						PreferredDuringSchedulingIgnoredDuringExecution: []core.WeightedPodAffinityTerm{
							{
								PodAffinityTerm: core.PodAffinityTerm{
									TopologyKey: "test",
								},
							},
						},
					},
				},
			}, &Scheduling{
				Affinity: &core.Affinity{
					PodAntiAffinity: &core.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []core.PodAffinityTerm{
							{
								TopologyKey: "test2",
							},
						},
						PreferredDuringSchedulingIgnoredDuringExecution: []core.WeightedPodAffinityTerm{
							{
								PodAffinityTerm: core.PodAffinityTerm{
									TopologyKey: "test2",
								},
							},
						},
					},
				},
			})(func(t *testing.T, pod *core.PodTemplateSpec) {
				require.NotNil(t, pod.Spec.Affinity)
				require.NotNil(t, pod.Spec.Affinity.PodAntiAffinity)

				require.Len(t, pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, 2)

				require.EqualValues(t, "test", pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].TopologyKey)
				require.EqualValues(t, "test2", pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[1].TopologyKey)

				require.Len(t, pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 2)

				require.EqualValues(t, "test", pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.TopologyKey)
				require.EqualValues(t, "test2", pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[1].PodAffinityTerm.TopologyKey)
			})
		})
	})

	t.Run("NodeAffinity", func(t *testing.T) {
		t.Run("Empty", func(t *testing.T) {
			applyScheduling(t, nil, &Scheduling{
				Affinity: &core.Affinity{
					NodeAffinity: &core.NodeAffinity{},
				},
			})(func(t *testing.T, pod *core.PodTemplateSpec) {
				require.NotNil(t, pod.Spec.Affinity)
				require.NotNil(t, pod.Spec.Affinity.NodeAffinity)
			})
		})

		t.Run("One", func(t *testing.T) {
			applyScheduling(t, nil, &Scheduling{
				Affinity: &core.Affinity{
					NodeAffinity: &core.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
							NodeSelectorTerms: []core.NodeSelectorTerm{
								{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key: "term1",
										},
									},
									MatchFields: []core.NodeSelectorRequirement{
										{
											Key: "term1",
										},
									},
								},
								{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key: "term2",
										},
									},
									MatchFields: []core.NodeSelectorRequirement{
										{
											Key: "term2",
										},
									},
								},
							},
						},
						PreferredDuringSchedulingIgnoredDuringExecution: []core.PreferredSchedulingTerm{
							{
								Preference: core.NodeSelectorTerm{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key: "term1",
										},
									},
									MatchFields: []core.NodeSelectorRequirement{
										{
											Key: "term1",
										},
									},
								},
							},
							{
								Preference: core.NodeSelectorTerm{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key: "term2",
										},
									},
									MatchFields: []core.NodeSelectorRequirement{
										{
											Key: "term2",
										},
									},
								},
							},
						},
					},
				},
			})(func(t *testing.T, pod *core.PodTemplateSpec) {
				require.NotNil(t, pod.Spec.Affinity)
				require.NotNil(t, pod.Spec.Affinity.NodeAffinity)

				require.Len(t, pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 2)

				require.Len(t, pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Preference.MatchFields, 1)
				require.EqualValues(t, "term1", pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Preference.MatchFields[0].Key)
				require.Len(t, pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Preference.MatchExpressions, 1)
				require.EqualValues(t, "term1", pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Preference.MatchExpressions[0].Key)

				require.Len(t, pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[1].Preference.MatchFields, 1)
				require.EqualValues(t, "term2", pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[1].Preference.MatchFields[0].Key)
				require.Len(t, pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[1].Preference.MatchExpressions, 1)
				require.EqualValues(t, "term2", pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[1].Preference.MatchExpressions[0].Key)

				require.NotNil(t, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
				require.Len(t, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, 2)

				require.Len(t, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchFields, 1)
				require.EqualValues(t, "term1", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchFields[0].Key)
				require.Len(t, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions, 1)
				require.EqualValues(t, "term1", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key)

				require.Len(t, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[1].MatchFields, 1)
				require.EqualValues(t, "term2", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[1].MatchFields[0].Key)
				require.Len(t, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[1].MatchExpressions, 1)
				require.EqualValues(t, "term2", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[1].MatchExpressions[0].Key)
			})
		})

		t.Run("Merge", func(t *testing.T) {
			applyScheduling(t, nil, &Scheduling{
				Affinity: &core.Affinity{
					NodeAffinity: &core.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
							NodeSelectorTerms: []core.NodeSelectorTerm{
								{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key: "term1",
										},
									},
									MatchFields: []core.NodeSelectorRequirement{
										{
											Key: "term1",
										},
									},
								},
								{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key: "term2",
										},
									},
									MatchFields: []core.NodeSelectorRequirement{
										{
											Key: "term2",
										},
									},
								},
							},
						},
						PreferredDuringSchedulingIgnoredDuringExecution: []core.PreferredSchedulingTerm{
							{
								Preference: core.NodeSelectorTerm{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key: "term1",
										},
									},
									MatchFields: []core.NodeSelectorRequirement{
										{
											Key: "term1",
										},
									},
								},
							},
							{
								Preference: core.NodeSelectorTerm{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key: "term2",
										},
									},
									MatchFields: []core.NodeSelectorRequirement{
										{
											Key: "term2",
										},
									},
								},
							},
						},
					},
				},
			}, &Scheduling{
				Affinity: &core.Affinity{
					NodeAffinity: &core.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
							NodeSelectorTerms: []core.NodeSelectorTerm{
								{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key: "term3",
										},
									},
									MatchFields: []core.NodeSelectorRequirement{
										{
											Key: "term3",
										},
									},
								},
								{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key: "term4",
										},
									},
									MatchFields: []core.NodeSelectorRequirement{
										{
											Key: "term4",
										},
									},
								},
							},
						},
						PreferredDuringSchedulingIgnoredDuringExecution: []core.PreferredSchedulingTerm{
							{
								Preference: core.NodeSelectorTerm{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key: "term3",
										},
									},
									MatchFields: []core.NodeSelectorRequirement{
										{
											Key: "term3",
										},
									},
								},
							},
							{
								Preference: core.NodeSelectorTerm{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key: "term4",
										},
									},
									MatchFields: []core.NodeSelectorRequirement{
										{
											Key: "term4",
										},
									},
								},
							},
						},
					},
				},
			})(func(t *testing.T, pod *core.PodTemplateSpec) {
				require.NotNil(t, pod.Spec.Affinity)
				require.NotNil(t, pod.Spec.Affinity.NodeAffinity)

				require.Len(t, pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 4)

				require.Len(t, pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Preference.MatchFields, 1)
				require.EqualValues(t, "term1", pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Preference.MatchFields[0].Key)
				require.Len(t, pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Preference.MatchExpressions, 1)
				require.EqualValues(t, "term1", pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Preference.MatchExpressions[0].Key)

				require.Len(t, pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[1].Preference.MatchFields, 1)
				require.EqualValues(t, "term2", pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[1].Preference.MatchFields[0].Key)
				require.Len(t, pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[1].Preference.MatchExpressions, 1)
				require.EqualValues(t, "term2", pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[1].Preference.MatchExpressions[0].Key)

				require.NotNil(t, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
				require.Len(t, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, 4)

				require.Len(t, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchFields, 2)
				require.EqualValues(t, "term1", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchFields[0].Key)
				require.EqualValues(t, "term3", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchFields[1].Key)
				require.Len(t, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions, 2)
				require.EqualValues(t, "term1", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key)
				require.EqualValues(t, "term3", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[1].Key)

				require.Len(t, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[1].MatchFields, 2)
				require.EqualValues(t, "term1", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[1].MatchFields[0].Key)
				require.EqualValues(t, "term4", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[1].MatchFields[1].Key)
				require.Len(t, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[1].MatchExpressions, 2)
				require.EqualValues(t, "term1", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[1].MatchExpressions[0].Key)
				require.EqualValues(t, "term4", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[1].MatchExpressions[1].Key)

				require.Len(t, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[2].MatchFields, 2)
				require.EqualValues(t, "term2", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[2].MatchFields[0].Key)
				require.EqualValues(t, "term3", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[2].MatchFields[1].Key)
				require.Len(t, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[2].MatchExpressions, 2)
				require.EqualValues(t, "term2", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[2].MatchExpressions[0].Key)
				require.EqualValues(t, "term3", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[2].MatchExpressions[1].Key)

				require.Len(t, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[3].MatchFields, 2)
				require.EqualValues(t, "term2", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[3].MatchFields[0].Key)
				require.EqualValues(t, "term4", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[3].MatchFields[1].Key)
				require.Len(t, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[3].MatchExpressions, 2)
				require.EqualValues(t, "term2", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[3].MatchExpressions[0].Key)
				require.EqualValues(t, "term4", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[3].MatchExpressions[1].Key)
			})
		})
	})
}

func Test_Scheduling_Tolerations(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		applyScheduling(t, nil)(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.Tolerations, 0)
		})
	})
	t.Run("Empty", func(t *testing.T) {
		applyScheduling(t, nil, &Scheduling{
			Tolerations: Tolerations{},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.Tolerations, 0)
		})
	})
	t.Run("Single", func(t *testing.T) {
		applyScheduling(t, nil, &Scheduling{
			Tolerations: Tolerations{
				{
					Operator: core.TolerationOpExists,
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.Tolerations, 1)

			require.EqualValues(t, core.TolerationOpExists, pod.Spec.Tolerations[0].Operator)
			require.EqualValues(t, "", pod.Spec.Tolerations[0].Key)
			require.EqualValues(t, "", pod.Spec.Tolerations[0].Value)
			require.EqualValues(t, "", pod.Spec.Tolerations[0].Effect)
			require.Nil(t, pod.Spec.Tolerations[0].TolerationSeconds)
		})
	})
	t.Run("Single - With Grace", func(t *testing.T) {
		applyScheduling(t, nil, &Scheduling{
			Tolerations: Tolerations{
				{
					Operator:          core.TolerationOpExists,
					TolerationSeconds: util.NewType[int64](5),
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.Tolerations, 1)

			require.EqualValues(t, core.TolerationOpExists, pod.Spec.Tolerations[0].Operator)
			require.EqualValues(t, "", pod.Spec.Tolerations[0].Key)
			require.EqualValues(t, "", pod.Spec.Tolerations[0].Value)
			require.EqualValues(t, "", pod.Spec.Tolerations[0].Effect)
			require.NotNil(t, pod.Spec.Tolerations[0].TolerationSeconds)
			require.EqualValues(t, 5, *pod.Spec.Tolerations[0].TolerationSeconds)
		})
	})
	t.Run("One - merge within spec", func(t *testing.T) {
		applyScheduling(t, nil, &Scheduling{
			Tolerations: Tolerations{
				{
					Operator: core.TolerationOpExists,
				},
				{
					Operator:          core.TolerationOpExists,
					TolerationSeconds: util.NewType[int64](5),
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.Tolerations, 1)

			require.EqualValues(t, core.TolerationOpExists, pod.Spec.Tolerations[0].Operator)
			require.EqualValues(t, "", pod.Spec.Tolerations[0].Key)
			require.EqualValues(t, "", pod.Spec.Tolerations[0].Value)
			require.EqualValues(t, "", pod.Spec.Tolerations[0].Effect)
			require.NotNil(t, pod.Spec.Tolerations[0].TolerationSeconds)
			require.EqualValues(t, 5, *pod.Spec.Tolerations[0].TolerationSeconds)
		})
	})
	t.Run("Multi - Update", func(t *testing.T) {
		applyScheduling(t, nil, &Scheduling{
			Tolerations: Tolerations{
				{
					Operator: core.TolerationOpExists,
				},
			},
		}, &Scheduling{
			Tolerations: Tolerations{
				{
					Operator:          core.TolerationOpExists,
					TolerationSeconds: util.NewType[int64](5),
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.Tolerations, 1)

			require.EqualValues(t, core.TolerationOpExists, pod.Spec.Tolerations[0].Operator)
			require.EqualValues(t, "", pod.Spec.Tolerations[0].Key)
			require.EqualValues(t, "", pod.Spec.Tolerations[0].Value)
			require.EqualValues(t, "", pod.Spec.Tolerations[0].Effect)
			require.NotNil(t, pod.Spec.Tolerations[0].TolerationSeconds)
			require.EqualValues(t, 5, *pod.Spec.Tolerations[0].TolerationSeconds)
		})
	})
	t.Run("Multi - Update", func(t *testing.T) {
		applyScheduling(t, nil, &Scheduling{
			Tolerations: Tolerations{
				{
					Operator:          core.TolerationOpExists,
					TolerationSeconds: util.NewType[int64](5),
				},
			},
		}, &Scheduling{
			Tolerations: Tolerations{
				{
					Operator: core.TolerationOpExists,
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.Tolerations, 1)

			require.EqualValues(t, core.TolerationOpExists, pod.Spec.Tolerations[0].Operator)
			require.EqualValues(t, "", pod.Spec.Tolerations[0].Key)
			require.EqualValues(t, "", pod.Spec.Tolerations[0].Value)
			require.EqualValues(t, "", pod.Spec.Tolerations[0].Effect)
			require.Nil(t, pod.Spec.Tolerations[0].TolerationSeconds)
		})
	})
}

func Test_Scheduling_NodeSelector(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		applyScheduling(t, nil)(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.NodeSelector, 0)
		})
	})
	t.Run("Empty", func(t *testing.T) {
		applyScheduling(t, nil, &Scheduling{
			NodeSelector: map[string]string{},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.NodeSelector, 0)
		})
	})
	t.Run("Selector", func(t *testing.T) {
		applyScheduling(t, nil, &Scheduling{
			NodeSelector: map[string]string{
				"1": "1",
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.NodeSelector, 1)

			require.Contains(t, pod.Spec.NodeSelector, "1")
			require.EqualValues(t, pod.Spec.NodeSelector["1"], "1")
		})
	})
	t.Run("Append", func(t *testing.T) {
		applyScheduling(t, nil, &Scheduling{
			NodeSelector: map[string]string{
				"1": "1",
			},
		}, &Scheduling{
			NodeSelector: map[string]string{
				"2": "1",
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.NodeSelector, 2)

			require.Contains(t, pod.Spec.NodeSelector, "1")
			require.EqualValues(t, pod.Spec.NodeSelector["1"], "1")

			require.Contains(t, pod.Spec.NodeSelector, "2")
			require.EqualValues(t, pod.Spec.NodeSelector["2"], "1")
		})
	})
	t.Run("Override", func(t *testing.T) {
		applyScheduling(t, nil, &Scheduling{
			NodeSelector: map[string]string{
				"1": "1",
			},
		}, &Scheduling{
			NodeSelector: map[string]string{
				"2": "1",
				"1": "2",
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.NodeSelector, 2)

			require.Contains(t, pod.Spec.NodeSelector, "1")
			require.EqualValues(t, pod.Spec.NodeSelector["1"], "2")

			require.Contains(t, pod.Spec.NodeSelector, "2")
			require.EqualValues(t, pod.Spec.NodeSelector["2"], "1")
		})
	})
}

func Test_Scheduling_SchedulerName(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		applyScheduling(t, nil)(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.EqualValues(t, "", pod.Spec.SchedulerName)
		})
	})
	t.Run("With Scheduler", func(t *testing.T) {
		applyScheduling(t, nil, &Scheduling{
			SchedulerName: util.NewType("example"),
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.EqualValues(t, "example", pod.Spec.SchedulerName)
		})
	})
	t.Run("With override", func(t *testing.T) {
		applyScheduling(t, nil, &Scheduling{
			SchedulerName: util.NewType("example"),
		}, &Scheduling{
			SchedulerName: util.NewType("example2"),
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.EqualValues(t, "example2", pod.Spec.SchedulerName)
		})
	})
	t.Run("With skip", func(t *testing.T) {
		applyScheduling(t, nil, &Scheduling{
			SchedulerName: util.NewType("example"),
		}, &Scheduling{})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.EqualValues(t, "example", pod.Spec.SchedulerName)
		})
	})
}
