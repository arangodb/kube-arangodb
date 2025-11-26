//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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
	"k8s.io/apimachinery/pkg/api/resource"
)

func applyResources(t *testing.T, template *core.PodTemplateSpec, container *core.Container, ns ...*Resources) func(in func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container)) {
	var i *Resources

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

func Test_Resources(t *testing.T) {
	v1Mi := resource.MustParse("1Mi")
	v8Mi := resource.MustParse("8Mi")
	v0 := resource.MustParse("0")

	t.Run("With Nil", func(t *testing.T) {
		applyResources(t, nil, nil)(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Empty(t, container.Resources)
		})
	})
	t.Run("With Empty", func(t *testing.T) {
		applyResources(t, &core.PodTemplateSpec{}, &core.Container{})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Empty(t, container.Resources)
		})
	})
	t.Run("Add", func(t *testing.T) {
		applyResources(t, &core.PodTemplateSpec{}, &core.Container{}, &Resources{
			Resources: &core.ResourceRequirements{
				Limits: core.ResourceList{
					core.ResourceCPU: v1Mi,
				},
				Requests: core.ResourceList{
					core.ResourceCPU: v1Mi,
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Resources.Limits, 1)
			require.Contains(t, container.Resources.Limits, core.ResourceCPU)
			require.EqualValues(t, v1Mi, container.Resources.Limits[core.ResourceCPU])

			require.Len(t, container.Resources.Requests, 1)
			require.Contains(t, container.Resources.Requests, core.ResourceCPU)
			require.EqualValues(t, v1Mi, container.Resources.Requests[core.ResourceCPU])
		})
	})
	t.Run("Add New One", func(t *testing.T) {
		applyResources(t, &core.PodTemplateSpec{}, &core.Container{}, &Resources{
			Resources: &core.ResourceRequirements{
				Limits: core.ResourceList{
					core.ResourceCPU: v1Mi,
				},
				Requests: core.ResourceList{
					core.ResourceCPU: v1Mi,
				},
			},
		}, &Resources{
			Resources: &core.ResourceRequirements{
				Limits: core.ResourceList{
					core.ResourceMemory: v1Mi,
				},
				Requests: core.ResourceList{
					core.ResourceMemory: v1Mi,
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Resources.Limits, 2)
			require.Contains(t, container.Resources.Limits, core.ResourceCPU)
			require.EqualValues(t, v1Mi, container.Resources.Limits[core.ResourceCPU])
			require.Contains(t, container.Resources.Limits, core.ResourceMemory)
			require.EqualValues(t, v1Mi, container.Resources.Limits[core.ResourceMemory])

			require.Len(t, container.Resources.Requests, 2)
			require.Contains(t, container.Resources.Requests, core.ResourceCPU)
			require.EqualValues(t, v1Mi, container.Resources.Requests[core.ResourceCPU])
			require.Contains(t, container.Resources.Requests, core.ResourceMemory)
			require.EqualValues(t, v1Mi, container.Resources.Requests[core.ResourceMemory])
		})
	})
	t.Run("Update", func(t *testing.T) {
		applyResources(t, &core.PodTemplateSpec{}, &core.Container{}, &Resources{
			Resources: &core.ResourceRequirements{
				Limits: core.ResourceList{
					core.ResourceCPU: v1Mi,
				},
				Requests: core.ResourceList{
					core.ResourceCPU: v1Mi,
				},
			},
		}, &Resources{
			Resources: &core.ResourceRequirements{
				Limits: core.ResourceList{
					core.ResourceCPU: v8Mi,
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Resources.Limits, 1)
			require.Contains(t, container.Resources.Limits, core.ResourceCPU)
			require.EqualValues(t, v8Mi, container.Resources.Limits[core.ResourceCPU])

			require.Len(t, container.Resources.Requests, 1)
			require.Contains(t, container.Resources.Requests, core.ResourceCPU)
			require.EqualValues(t, v1Mi, container.Resources.Requests[core.ResourceCPU])
		})
	})
	t.Run("Remove", func(t *testing.T) {
		applyResources(t, &core.PodTemplateSpec{}, &core.Container{}, &Resources{
			Resources: &core.ResourceRequirements{
				Limits: core.ResourceList{
					core.ResourceCPU: v1Mi,
				},
				Requests: core.ResourceList{
					core.ResourceCPU: v1Mi,
				},
			},
		}, &Resources{
			Resources: &core.ResourceRequirements{
				Limits: core.ResourceList{
					core.ResourceCPU: v0,
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Resources.Limits, 0)

			require.Len(t, container.Resources.Requests, 1)
			require.Contains(t, container.Resources.Requests, core.ResourceCPU)
			require.EqualValues(t, v1Mi, container.Resources.Requests[core.ResourceCPU])
		})
	})
	t.Run("Upscale", func(t *testing.T) {
		applyResources(t, &core.PodTemplateSpec{}, &core.Container{}, &Resources{
			Resources: &core.ResourceRequirements{
				Limits: core.ResourceList{
					core.ResourceCPU: v1Mi,
				},
				Requests: core.ResourceList{
					core.ResourceCPU: v1Mi,
				},
			},
		}, &Resources{
			Resources: &core.ResourceRequirements{
				Requests: core.ResourceList{
					core.ResourceCPU: v8Mi,
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Resources.Limits, 1)
			require.Contains(t, container.Resources.Limits, core.ResourceCPU)
			require.EqualValues(t, v8Mi, container.Resources.Limits[core.ResourceCPU])

			require.Len(t, container.Resources.Requests, 1)
			require.Contains(t, container.Resources.Requests, core.ResourceCPU)
			require.EqualValues(t, v8Mi, container.Resources.Requests[core.ResourceCPU])
		})
	})
	t.Run("Downscale", func(t *testing.T) {
		applyResources(t, &core.PodTemplateSpec{}, &core.Container{
			Resources: core.ResourceRequirements{
				Requests: core.ResourceList{
					core.ResourceCPU: v8Mi,
				},
			},
		}, &Resources{
			Resources: &core.ResourceRequirements{
				Limits: core.ResourceList{
					core.ResourceCPU: v1Mi,
				},
				Requests: core.ResourceList{
					core.ResourceCPU: v1Mi,
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Resources.Limits, 1)
			require.Contains(t, container.Resources.Limits, core.ResourceCPU)
			require.EqualValues(t, v1Mi, container.Resources.Limits[core.ResourceCPU])

			require.Len(t, container.Resources.Requests, 1)
			require.Contains(t, container.Resources.Requests, core.ResourceCPU)
			require.EqualValues(t, v1Mi, container.Resources.Requests[core.ResourceCPU])
		})
	})
}
