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

package container

import (
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/yaml"

	schedulerContainerResourcesApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/container/resources"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func applyContainers(t *testing.T, template *core.PodTemplateSpec, ns ...Containers) func(in func(t *testing.T, pod *core.PodTemplateSpec, spec Containers)) {
	var i Containers

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

	return func(in func(t *testing.T, pod *core.PodTemplateSpec, spec Containers)) {
		t.Run("Validate", func(t *testing.T) {
			if i != nil {
				in(t, template, i)
			} else {
				in(t, template, Containers{})
			}
		})
	}
}

func applyContainersYAML(t *testing.T, template *core.PodTemplateSpec, ns ...string) func(in func(t *testing.T, pod *core.PodTemplateSpec, spec Containers)) {
	elements := make([]Containers, len(ns))

	for id := range ns {
		var p Containers
		require.NoError(t, yaml.Unmarshal([]byte(ns[id]), &p))
		elements[id] = p.DeepCopy()
	}

	return applyContainers(t, template, elements...)
}

func Test_Containers(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		applyContainers(t, nil)(func(t *testing.T, pod *core.PodTemplateSpec, spec Containers) {
			require.Len(t, spec, 0)
			require.Len(t, pod.Spec.Containers, 0)
		})
	})
	t.Run("Empty template", func(t *testing.T) {
		applyContainers(t, &core.PodTemplateSpec{})(func(t *testing.T, pod *core.PodTemplateSpec, spec Containers) {
			require.Len(t, spec, 0)
			require.Len(t, pod.Spec.Containers, 0)
		})
	})
	t.Run("Add container", func(t *testing.T) {
		applyContainers(t, &core.PodTemplateSpec{}, Containers{
			"test": {
				Image: &schedulerContainerResourcesApiv1alpha1.Image{
					Image: util.NewType("test"),
				},
				Resources: &schedulerContainerResourcesApiv1alpha1.Resources{
					Resources: &core.ResourceRequirements{
						Limits: map[core.ResourceName]resource.Quantity{
							core.ResourceCPU: resource.MustParse("1"),
						},
					},
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, spec Containers) {
			require.Len(t, spec, 1)
			require.Len(t, pod.Spec.Containers, 1)
			require.EqualValues(t, "test", pod.Spec.Containers[0].Name)
			require.EqualValues(t, "test", pod.Spec.Containers[0].Image)
			require.Len(t, pod.Spec.Containers[0].Resources.Limits, 1)
			require.Contains(t, pod.Spec.Containers[0].Resources.Limits, core.ResourceCPU)
			require.EqualValues(t, resource.MustParse("1"), pod.Spec.Containers[0].Resources.Limits[core.ResourceCPU])
		})
	})
	t.Run("Append container", func(t *testing.T) {
		applyContainers(t, &core.PodTemplateSpec{
			Spec: core.PodSpec{
				Containers: []core.Container{
					{
						Name: "example",
					},
				},
			},
		}, Containers{
			"test": {
				Image: &schedulerContainerResourcesApiv1alpha1.Image{
					Image: util.NewType("test"),
				},
				Resources: &schedulerContainerResourcesApiv1alpha1.Resources{
					Resources: &core.ResourceRequirements{
						Limits: map[core.ResourceName]resource.Quantity{
							core.ResourceCPU: resource.MustParse("1"),
						},
					},
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, spec Containers) {
			require.Len(t, spec, 1)
			require.Len(t, pod.Spec.Containers, 2)
			require.EqualValues(t, "test", pod.Spec.Containers[1].Name)
			require.EqualValues(t, "test", pod.Spec.Containers[1].Image)
			require.Len(t, pod.Spec.Containers[1].Resources.Limits, 1)
			require.Contains(t, pod.Spec.Containers[1].Resources.Limits, core.ResourceCPU)
			require.EqualValues(t, resource.MustParse("1"), pod.Spec.Containers[1].Resources.Limits[core.ResourceCPU])
		})
	})
}

func Test_Containers_YAML(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		applyContainersYAML(t, nil)(func(t *testing.T, pod *core.PodTemplateSpec, spec Containers) {
			require.Len(t, spec, 0)
			require.Len(t, pod.Spec.Containers, 0)
		})
	})
	t.Run("Empty template", func(t *testing.T) {
		applyContainersYAML(t, &core.PodTemplateSpec{})(func(t *testing.T, pod *core.PodTemplateSpec, spec Containers) {
			require.Len(t, spec, 0)
			require.Len(t, pod.Spec.Containers, 0)
		})
	})
	t.Run("Add container", func(t *testing.T) {
		applyContainersYAML(t, &core.PodTemplateSpec{}, `
---
test:
  image: test
  resources:
    limits:
      cpu: 1
`)(func(t *testing.T, pod *core.PodTemplateSpec, spec Containers) {
			require.Len(t, spec, 1)
			require.Len(t, pod.Spec.Containers, 1)
			require.EqualValues(t, "test", pod.Spec.Containers[0].Name)
			require.EqualValues(t, "test", pod.Spec.Containers[0].Image)
			require.Len(t, pod.Spec.Containers[0].Resources.Limits, 1)
			require.Contains(t, pod.Spec.Containers[0].Resources.Limits, core.ResourceCPU)
			require.EqualValues(t, resource.MustParse("1"), pod.Spec.Containers[0].Resources.Limits[core.ResourceCPU])
		})
	})
	t.Run("Append container", func(t *testing.T) {
		applyContainersYAML(t, &core.PodTemplateSpec{
			Spec: core.PodSpec{
				Containers: []core.Container{
					{
						Name: "example",
					},
				},
			},
		}, `
---
test:
  image: test
  resources:
    limits:
      cpu: 1
`)(func(t *testing.T, pod *core.PodTemplateSpec, spec Containers) {
			require.Len(t, spec, 1)
			require.Len(t, pod.Spec.Containers, 2)
			require.EqualValues(t, "test", pod.Spec.Containers[1].Name)
			require.EqualValues(t, "test", pod.Spec.Containers[1].Image)
			require.Len(t, pod.Spec.Containers[1].Resources.Limits, 1)
			require.Contains(t, pod.Spec.Containers[1].Resources.Limits, core.ResourceCPU)
			require.EqualValues(t, resource.MustParse("1"), pod.Spec.Containers[1].Resources.Limits[core.ResourceCPU])
		})
	})
}
