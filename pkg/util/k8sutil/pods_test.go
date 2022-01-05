//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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
	v1 "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/handlers/utils"
)

// TestIsPodReady tests IsPodReady.
func TestIsPodReady(t *testing.T) {
	assert.False(t, IsPodReady(&v1.Pod{}))
	assert.False(t, IsPodReady(&v1.Pod{
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				v1.PodCondition{
					Type:   v1.PodReady,
					Status: v1.ConditionFalse,
				},
			},
		},
	}))
	assert.True(t, IsPodReady(&v1.Pod{
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				v1.PodCondition{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}))
}

// TestIsPodFailed tests IsPodFailed.
func TestIsPodFailed(t *testing.T) {
	type args struct {
		pod            *v1.Pod
		coreContainers utils.StringList
	}
	tests := map[string]struct {
		args args
		want bool
	}{
		"empty pod": {
			args: args{
				pod: &v1.Pod{},
			},
		},
		"pod is running": {
			args: args{
				pod: &v1.Pod{
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				},
			},
		},
		"pod is failed": {
			args: args{
				pod: &v1.Pod{
					Status: v1.PodStatus{
						Phase: v1.PodFailed,
					},
				},
			},
			want: true,
		},
		"one core container failed": {
			args: args{
				pod: &v1.Pod{
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							{
								Name: "core_container",
								State: v1.ContainerState{
									Terminated: &v1.ContainerStateTerminated{
										ExitCode: 1,
									},
								},
							},
						},
					},
				},
				coreContainers: utils.StringList{"something", "core_container"},
			},
			want: true,
		},
		"one non-core container failed": {
			args: args{
				pod: &v1.Pod{
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							{
								Name: "non_core_container",
								State: v1.ContainerState{
									Terminated: &v1.ContainerStateTerminated{
										ExitCode: 1,
									},
								},
							},
						},
					},
				},
				coreContainers: utils.StringList{"something", "core_container"},
			},
		},
		"one core container succeeded": {
			args: args{
				pod: &v1.Pod{
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							{
								Name: "core_container",
								State: v1.ContainerState{
									Terminated: &v1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
						},
					},
				},
				coreContainers: utils.StringList{"something", "core_container"},
			},
		},
		"first core container succeeded and second is still running": {
			args: args{
				pod: &v1.Pod{
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							{
								Name: "core_container1",
								State: v1.ContainerState{
									Running: &v1.ContainerStateRunning{},
								},
							},
							{
								Name: "core_container2",
								State: v1.ContainerState{
									Terminated: &v1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
						},
					},
				},
				coreContainers: utils.StringList{"core_container1", "core_container2"},
			},
			want: true,
		},
		"all containers succeeded": {
			args: args{
				pod: &v1.Pod{
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							{
								Name: "core_container1",
								State: v1.ContainerState{
									Terminated: &v1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
							{
								Name: "core_container2",
								State: v1.ContainerState{
									Terminated: &v1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
						},
					},
				},
				coreContainers: utils.StringList{"core_container1", "core_container2"},
			},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			got := IsPodFailed(test.args.pod, test.args.coreContainers)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestIsPodSucceeded(t *testing.T) {
	type args struct {
		pod            *v1.Pod
		coreContainers utils.StringList
	}
	tests := map[string]struct {
		args args
		want bool
	}{
		"empty pod": {
			args: args{
				pod: &v1.Pod{},
			},
		},
		"pod is succeeded": {
			args: args{
				pod: &v1.Pod{
					Status: v1.PodStatus{
						Phase: v1.PodSucceeded,
					},
				},
			},
			want: true,
		},
		"all core containers succeeded": {
			args: args{
				pod: &v1.Pod{
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							{
								Name: "core_container1",
								State: v1.ContainerState{
									Terminated: &v1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
							{
								Name: "core_container2",
								State: v1.ContainerState{
									Terminated: &v1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
							{
								Name: "non-core_container",
							},
						},
					},
				},
				coreContainers: utils.StringList{"core_container1", "core_container2"},
			},
			want: true,
		},
		"non-core container succeeded": {
			args: args{
				pod: &v1.Pod{
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							{
								Name: "core_container1",
							},
							{
								Name: "non-core_container",
								State: v1.ContainerState{
									Terminated: &v1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
						},
					},
				},
				coreContainers: utils.StringList{"core_container1"},
			},
		},
		"the only one core container succeeded": {
			args: args{
				pod: &v1.Pod{
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							{
								Name: "core_container1",
								State: v1.ContainerState{
									Terminated: &v1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
							{
								Name: "non-core_container",
							},
						},
					},
				},
				coreContainers: utils.StringList{"core_container1"},
			},
			want: true,
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			got := IsPodSucceeded(test.args.pod, test.args.coreContainers)
			assert.Equal(t, test.want, got)
		})
	}
}
