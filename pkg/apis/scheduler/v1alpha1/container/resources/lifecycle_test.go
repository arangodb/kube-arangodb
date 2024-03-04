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
)

func applyLifecycle(t *testing.T, template *core.PodTemplateSpec, container *core.Container, ns ...*Lifecycle) func(in func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container)) {
	var i *Lifecycle

	for _, n := range ns {
		require.NoError(t, n.Validate())

		i = i.With(n)

		require.NoError(t, i.Validate())
	}

	template = template.DeepCopy()

	if template == nil {
		template = &core.PodTemplateSpec{}
	}

	container = container.DeepCopy()
	if container == nil {
		container = &core.Container{}
	}

	template.Spec.Containers = append(template.Spec.Containers, *container)

	container = &template.Spec.Containers[0]

	require.NoError(t, i.Apply(template, container))

	return func(in func(t *testing.T, spec *core.PodTemplateSpec, container *core.Container)) {
		t.Run("Validate", func(t *testing.T) {
			in(t, template, container)
		})
	}
}

func Test_Lifecycle(t *testing.T) {
	t.Run("With Nil", func(t *testing.T) {
		applyLifecycle(t, nil, nil)(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Nil(t, container.Lifecycle)
		})
	})
	t.Run("With Empty", func(t *testing.T) {
		applyLifecycle(t, &core.PodTemplateSpec{}, &core.Container{})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Nil(t, container.Lifecycle)
		})
	})
	t.Run("With lifecycles", func(t *testing.T) {
		applyLifecycle(t, &core.PodTemplateSpec{}, &core.Container{}, &Lifecycle{
			Lifecycle: &core.Lifecycle{
				PostStart: &core.LifecycleHandler{
					Exec: &core.ExecAction{
						Command: []string{"test"},
					},
				},
				PreStop: &core.LifecycleHandler{
					HTTPGet: &core.HTTPGetAction{
						Path: "/test",
					},
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.NotNil(t, container.Lifecycle)

			require.NotNil(t, container.Lifecycle.PostStart)
			require.NotNil(t, container.Lifecycle.PostStart.Exec)
			require.Nil(t, container.Lifecycle.PostStart.HTTPGet)
			require.Len(t, container.Lifecycle.PostStart.Exec.Command, 1)
			require.EqualValues(t, "test", container.Lifecycle.PostStart.Exec.Command[0])

			require.NotNil(t, container.Lifecycle.PreStop)
			require.Nil(t, container.Lifecycle.PreStop.Exec)
			require.NotNil(t, container.Lifecycle.PreStop.HTTPGet)
			require.EqualValues(t, "/test", container.Lifecycle.PreStop.HTTPGet.Path)
		})
	})
	t.Run("With merge", func(t *testing.T) {
		applyLifecycle(t, &core.PodTemplateSpec{}, &core.Container{},
			&Lifecycle{
				Lifecycle: &core.Lifecycle{
					PostStart: &core.LifecycleHandler{
						Exec: &core.ExecAction{
							Command: []string{"test"},
						},
					},
				},
			},
			&Lifecycle{
				Lifecycle: &core.Lifecycle{
					PreStop: &core.LifecycleHandler{
						HTTPGet: &core.HTTPGetAction{
							Path: "/test",
						},
					},
				},
			})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.NotNil(t, container.Lifecycle)

			require.NotNil(t, container.Lifecycle.PostStart)
			require.NotNil(t, container.Lifecycle.PostStart.Exec)
			require.Nil(t, container.Lifecycle.PostStart.HTTPGet)
			require.Len(t, container.Lifecycle.PostStart.Exec.Command, 1)
			require.EqualValues(t, "test", container.Lifecycle.PostStart.Exec.Command[0])

			require.NotNil(t, container.Lifecycle.PreStop)
			require.Nil(t, container.Lifecycle.PreStop.Exec)
			require.NotNil(t, container.Lifecycle.PreStop.HTTPGet)
			require.EqualValues(t, "/test", container.Lifecycle.PreStop.HTTPGet.Path)
		})
	})
	t.Run("With override", func(t *testing.T) {
		applyLifecycle(t, &core.PodTemplateSpec{}, &core.Container{}, &Lifecycle{
			Lifecycle: &core.Lifecycle{
				PostStart: &core.LifecycleHandler{
					Exec: &core.ExecAction{
						Command: []string{"test"},
					},
				},
				PreStop: &core.LifecycleHandler{
					HTTPGet: &core.HTTPGetAction{
						Path: "/test",
					},
				},
			},
		}, &Lifecycle{
			Lifecycle: &core.Lifecycle{
				PostStart: &core.LifecycleHandler{
					HTTPGet: &core.HTTPGetAction{
						Path: "/test",
					},
				},
				PreStop: &core.LifecycleHandler{
					Exec: &core.ExecAction{
						Command: []string{"test"},
					},
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.NotNil(t, container.Lifecycle)

			require.NotNil(t, container.Lifecycle.PreStop)
			require.NotNil(t, container.Lifecycle.PreStop.Exec)
			require.Nil(t, container.Lifecycle.PreStop.HTTPGet)
			require.Len(t, container.Lifecycle.PreStop.Exec.Command, 1)
			require.EqualValues(t, "test", container.Lifecycle.PreStop.Exec.Command[0])

			require.NotNil(t, container.Lifecycle.PostStart)
			require.Nil(t, container.Lifecycle.PostStart.Exec)
			require.NotNil(t, container.Lifecycle.PostStart.HTTPGet)
			require.EqualValues(t, "/test", container.Lifecycle.PostStart.HTTPGet.Path)
		})
	})
}
