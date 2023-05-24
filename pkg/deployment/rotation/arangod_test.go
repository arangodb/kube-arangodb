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

package rotation

import (
	"testing"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/deployment/topology"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Test_ArangoD_SchedulerName(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Change SchedulerName from Empty",
			spec: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.SchedulerName = ""
			}),
			status: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.SchedulerName = "new"
			}),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SilentRotation,
			},
		},
		{
			name: "Change SchedulerName into Empty",
			spec: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.SchedulerName = "new"
			}),
			status: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.SchedulerName = ""
			}),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SilentRotation,
			},
		},
		{
			name: "SchedulerName equals",
			spec: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.SchedulerName = ""
			}),
			status: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.SchedulerName = ""
			}),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SkippedRotation,
			},
		},
	}

	runTestCases(t)(testCases...)
}

func Test_ArangoD_TerminationGracePeriodSeconds(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Add",
			spec: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.TerminationGracePeriodSeconds = nil
			}),
			status: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.TerminationGracePeriodSeconds = util.NewType[int64](30)
			}),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SilentRotation,
			},
		},
		{
			name: "Remove",
			spec: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.TerminationGracePeriodSeconds = util.NewType[int64](30)
			}),
			status: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.TerminationGracePeriodSeconds = nil
			}),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SilentRotation,
			},
		},
		{
			name: "Update",
			spec: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.TerminationGracePeriodSeconds = util.NewType[int64](33)
			}),
			status: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.TerminationGracePeriodSeconds = util.NewType[int64](30)
			}),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SilentRotation,
			},
		},
	}

	runTestCases(t)(testCases...)
}

func Test_ArangoD_Affinity(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Remove affinity",
			spec: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.Affinity = &core.Affinity{
					NodeAffinity: &core.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
							NodeSelectorTerms: []core.NodeSelectorTerm{
								{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key:      "beta.kubernetes.io/arch",
											Operator: core.NodeSelectorOpIn,
											Values: []string{
												"amd64",
											},
										},
									},
								},
							},
						},
					},
				}
			}),
			status: buildPodSpec(func(pod *core.PodTemplateSpec) {
			}),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SilentRotation,
			},
		},
		{
			name: "Add affinity",
			spec: buildPodSpec(func(pod *core.PodTemplateSpec) {
			}),
			status: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.Affinity = &core.Affinity{
					NodeAffinity: &core.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
							NodeSelectorTerms: []core.NodeSelectorTerm{
								{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key:      "beta.kubernetes.io/arch",
											Operator: core.NodeSelectorOpIn,
											Values: []string{
												"amd64",
											},
										},
									},
								},
							},
						},
					},
				}
			}),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SilentRotation,
			},
		},
		{
			name: "Change affinity",
			spec: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.Affinity = &core.Affinity{
					NodeAffinity: &core.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
							NodeSelectorTerms: []core.NodeSelectorTerm{
								{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key:      "beta.kubernetes.io/arch",
											Operator: core.NodeSelectorOpIn,
											Values: []string{
												"amd64",
											},
										},
									},
								},
							},
						},
					},
				}
			}),
			status: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.Affinity = &core.Affinity{
					NodeAffinity: &core.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
							NodeSelectorTerms: []core.NodeSelectorTerm{
								{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key:      "kubernetes.io/arch",
											Operator: core.NodeSelectorOpIn,
											Values: []string{
												"amd64",
											},
										},
									},
								},
							},
						},
					},
				}
			}),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SilentRotation,
			},
		},
		{
			name: "Change affinity archs",
			spec: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.Affinity = &core.Affinity{
					NodeAffinity: &core.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
							NodeSelectorTerms: []core.NodeSelectorTerm{
								{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key:      "beta.kubernetes.io/arch",
											Operator: core.NodeSelectorOpIn,
											Values: []string{
												"amd64",
											},
										},
									},
								},
							},
						},
					},
				}
			}),
			status: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.Affinity = &core.Affinity{
					NodeAffinity: &core.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
							NodeSelectorTerms: []core.NodeSelectorTerm{
								{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key:      "kubernetes.io/arch",
											Operator: core.NodeSelectorOpIn,
											Values: []string{
												"arm64",
											},
										},
									},
								},
							},
						},
					},
				}
			}),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SilentRotation,
			},
		},
		{
			name: "Change affinity archs - swap arch order",
			spec: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.Affinity = &core.Affinity{
					NodeAffinity: &core.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
							NodeSelectorTerms: []core.NodeSelectorTerm{
								{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key:      "beta.kubernetes.io/arch",
											Operator: core.NodeSelectorOpIn,
											Values: []string{
												"amd64", "arm64",
											},
										},
									},
								},
							},
						},
					},
				}
			}),
			status: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.Affinity = &core.Affinity{
					NodeAffinity: &core.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
							NodeSelectorTerms: []core.NodeSelectorTerm{
								{
									MatchExpressions: []core.NodeSelectorRequirement{
										{
											Key:      "kubernetes.io/arch",
											Operator: core.NodeSelectorOpIn,
											Values: []string{
												"arm64", "amd64",
											},
										},
									},
								},
							},
						},
					},
				}
			}),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SilentRotation,
			},
		},
	}

	runTestCases(t)(testCases...)
}

func Test_ArangoD_Labels(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Add label",

			spec: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Labels = map[string]string{}
			}),

			status: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Labels = map[string]string{
					"A": "B",
				}
			}),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SkippedRotation,
			},
		},
		{
			name: "Remove label",

			spec: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Labels = map[string]string{
					"A": "B",
				}
			}),

			status: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Labels = map[string]string{}
			}),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SkippedRotation,
			},
		},
		{
			name: "Change label",

			spec: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Labels = map[string]string{
					"A": "A",
				}
			}),

			status: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Labels = map[string]string{
					"A": "B",
				}
			}),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SkippedRotation,
			},
		},
	}

	runTestCases(t)(testCases...)
}

func Test_ArangoD_Envs_Zone(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Add Zone env",

			spec: buildPodSpec(addContainer("server", func(c *core.Container) {
				c.Env = []core.EnvVar{}
			})),

			status: buildPodSpec(addContainer("server", func(c *core.Container) {
				c.Env = []core.EnvVar{
					{
						Name:  topology.ArangoDBZone,
						Value: "A",
					},
				}
			})),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SilentRotation,
			},
		},
		{
			name: "Remove Zone env",

			spec: buildPodSpec(addContainer("server", func(c *core.Container) {
				c.Env = []core.EnvVar{
					{
						Name:  topology.ArangoDBZone,
						Value: "A",
					},
				}
			})),

			status: buildPodSpec(addContainer("server", func(c *core.Container) {
				c.Env = []core.EnvVar{}
			})),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SilentRotation,
			},
		},
		{
			name: "Update Zone env",

			spec: buildPodSpec(addContainer("server", func(c *core.Container) {
				c.Env = []core.EnvVar{
					{
						Name:  topology.ArangoDBZone,
						Value: "A",
					},
				}
			})),

			status: buildPodSpec(addContainer("server", func(c *core.Container) {
				c.Env = []core.EnvVar{
					{
						Name:  topology.ArangoDBZone,
						Value: "B",
					},
				}
			})),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SilentRotation,
			},
		},
		{
			name: "Update other env",

			spec: buildPodSpec(addContainer("server", func(c *core.Container) {
				c.Env = []core.EnvVar{
					{
						Name:  "Q",
						Value: "A",
					},
					{
						Name:  topology.ArangoDBZone,
						Value: "A",
					},
				}
			})),

			status: buildPodSpec(addContainer("server", func(c *core.Container) {
				c.Env = []core.EnvVar{
					{
						Name:  topology.ArangoDBZone,
						Value: "A",
					},
				}
			})),

			TestCaseOverride: TestCaseOverride{
				expectedMode: GracefulRotation,
			},
		},
	}

	runTestCases(t)(testCases...)
}
