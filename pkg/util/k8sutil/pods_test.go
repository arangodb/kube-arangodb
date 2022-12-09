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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/handlers/utils"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

// TestIsPodReady tests IsPodReady.
func TestIsPodReady(t *testing.T) {
	assert.False(t, IsPodReady(&core.Pod{}))
	assert.False(t, IsPodReady(&core.Pod{
		Status: core.PodStatus{
			Conditions: []core.PodCondition{
				core.PodCondition{
					Type:   core.PodReady,
					Status: core.ConditionFalse,
				},
			},
		},
	}))
	assert.True(t, IsPodReady(&core.Pod{
		Status: core.PodStatus{
			Conditions: []core.PodCondition{
				core.PodCondition{
					Type:   core.PodReady,
					Status: core.ConditionTrue,
				},
			},
		},
	}))
}

// TestIsPodFailed tests IsPodFailed.
func TestIsPodFailed(t *testing.T) {
	type args struct {
		pod            *core.Pod
		coreContainers utils.StringList
	}
	tests := map[string]struct {
		args args
		want bool
	}{
		"empty pod": {
			args: args{
				pod: &core.Pod{},
			},
		},
		"pod is running": {
			args: args{
				pod: &core.Pod{
					Status: core.PodStatus{
						Phase: core.PodRunning,
					},
				},
			},
		},
		"pod is failed": {
			args: args{
				pod: &core.Pod{
					Status: core.PodStatus{
						Phase: core.PodFailed,
					},
				},
			},
			want: true,
		},
		"one core container failed": {
			args: args{
				pod: &core.Pod{
					Status: core.PodStatus{
						ContainerStatuses: []core.ContainerStatus{
							{
								Name: "core_container",
								State: core.ContainerState{
									Terminated: &core.ContainerStateTerminated{
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
				pod: &core.Pod{
					Status: core.PodStatus{
						ContainerStatuses: []core.ContainerStatus{
							{
								Name: "non_core_container",
								State: core.ContainerState{
									Terminated: &core.ContainerStateTerminated{
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
				pod: &core.Pod{
					Status: core.PodStatus{
						ContainerStatuses: []core.ContainerStatus{
							{
								Name: "core_container",
								State: core.ContainerState{
									Terminated: &core.ContainerStateTerminated{
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
				pod: &core.Pod{
					Status: core.PodStatus{
						ContainerStatuses: []core.ContainerStatus{
							{
								Name: "core_container1",
								State: core.ContainerState{
									Running: &core.ContainerStateRunning{},
								},
							},
							{
								Name: "core_container2",
								State: core.ContainerState{
									Terminated: &core.ContainerStateTerminated{
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
				pod: &core.Pod{
					Status: core.PodStatus{
						ContainerStatuses: []core.ContainerStatus{
							{
								Name: "core_container1",
								State: core.ContainerState{
									Terminated: &core.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
							{
								Name: "core_container2",
								State: core.ContainerState{
									Terminated: &core.ContainerStateTerminated{
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
		pod            *core.Pod
		coreContainers utils.StringList
	}
	tests := map[string]struct {
		args args
		want bool
	}{
		"empty pod": {
			args: args{
				pod: &core.Pod{},
			},
		},
		"pod is succeeded": {
			args: args{
				pod: &core.Pod{
					Status: core.PodStatus{
						Phase: core.PodSucceeded,
					},
				},
			},
			want: true,
		},
		"all core containers succeeded": {
			args: args{
				pod: &core.Pod{
					Status: core.PodStatus{
						ContainerStatuses: []core.ContainerStatus{
							{
								Name: "core_container1",
								State: core.ContainerState{
									Terminated: &core.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
							{
								Name: "core_container2",
								State: core.ContainerState{
									Terminated: &core.ContainerStateTerminated{
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
				pod: &core.Pod{
					Status: core.PodStatus{
						ContainerStatuses: []core.ContainerStatus{
							{
								Name: "core_container1",
							},
							{
								Name: "non-core_container",
								State: core.ContainerState{
									Terminated: &core.ContainerStateTerminated{
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
				pod: &core.Pod{
					Status: core.PodStatus{
						ContainerStatuses: []core.ContainerStatus{
							{
								Name: "core_container1",
								State: core.ContainerState{
									Terminated: &core.ContainerStateTerminated{
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

func Test_extractContainerNamesFromConditionMessage(t *testing.T) {
	t.Run("Valid name", func(t *testing.T) {
		c, ok := extractContainerNamesFromConditionMessage("containers with unready status: [sidecar2 sidecar3]")
		require.True(t, ok)
		require.Len(t, c, 2)
		require.Contains(t, c, "sidecar2")
		require.Contains(t, c, "sidecar3")
		require.NotContains(t, c, "sidecar")
	})
}

func Test_EnsureFinalizer(t *testing.T) {
	c := kclient.NewFakeClient()

	f := "test.arangodb.com/test"

	pod := &core.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test",
			Namespace: "test",

			Finalizers: nil,
		},
	}
	refresh := func(t *testing.T) {
		p, err := c.Kubernetes().CoreV1().Pods(pod.GetNamespace()).Get(context.Background(), pod.GetName(), meta.GetOptions{})
		require.NoError(t, err)
		pod = p
	}

	_, err := c.Kubernetes().CoreV1().Pods(pod.GetNamespace()).Create(context.Background(), pod, meta.CreateOptions{})
	require.NoError(t, err)

	t.Run("Ensure finalizers", func(t *testing.T) {
		require.NoError(t, EnsureFinalizerPresent(context.Background(), c.Kubernetes().CoreV1().Pods(pod.GetNamespace()), pod, constants.FinalizerPodGracefulShutdown))

		refresh(t)

		require.Len(t, pod.Finalizers, 1)
		require.Contains(t, pod.Finalizers, constants.FinalizerPodGracefulShutdown)
	})

	t.Run("Add finalizers", func(t *testing.T) {
		require.NoError(t, EnsureFinalizerPresent(context.Background(), c.Kubernetes().CoreV1().Pods(pod.GetNamespace()), pod, f))

		refresh(t)

		require.Len(t, pod.Finalizers, 2)
		require.Contains(t, pod.Finalizers, constants.FinalizerPodGracefulShutdown)
		require.Contains(t, pod.Finalizers, f)
	})

	t.Run("Re Add finalizers", func(t *testing.T) {
		require.NoError(t, EnsureFinalizerPresent(context.Background(), c.Kubernetes().CoreV1().Pods(pod.GetNamespace()), pod, f))

		refresh(t)

		require.Len(t, pod.Finalizers, 2)
		require.Contains(t, pod.Finalizers, constants.FinalizerPodGracefulShutdown)
		require.Contains(t, pod.Finalizers, f)
	})

	t.Run("Remove finalizers", func(t *testing.T) {
		require.NoError(t, EnsureFinalizerAbsent(context.Background(), c.Kubernetes().CoreV1().Pods(pod.GetNamespace()), pod, f))

		refresh(t)

		require.Len(t, pod.Finalizers, 1)
		require.Contains(t, pod.Finalizers, constants.FinalizerPodGracefulShutdown)
		require.NotContains(t, pod.Finalizers, f)
	})

	t.Run("Re - remove finalizers", func(t *testing.T) {
		require.NoError(t, EnsureFinalizerAbsent(context.Background(), c.Kubernetes().CoreV1().Pods(pod.GetNamespace()), pod, f))

		refresh(t)

		require.Len(t, pod.Finalizers, 1)
		require.Contains(t, pod.Finalizers, constants.FinalizerPodGracefulShutdown)
		require.NotContains(t, pod.Finalizers, f)
	})

	t.Run("Remove final finalizers", func(t *testing.T) {
		require.NoError(t, EnsureFinalizerAbsent(context.Background(), c.Kubernetes().CoreV1().Pods(pod.GetNamespace()), pod, constants.FinalizerPodGracefulShutdown))

		refresh(t)

		require.Len(t, pod.Finalizers, 0)
		require.NotContains(t, pod.Finalizers, constants.FinalizerPodGracefulShutdown)
		require.NotContains(t, pod.Finalizers, f)
	})
}
