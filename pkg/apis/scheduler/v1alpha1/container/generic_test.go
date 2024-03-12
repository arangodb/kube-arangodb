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
	"k8s.io/apimachinery/pkg/util/yaml"

	schedulerContainerResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/container/resources"
)

func applyGeneric(t *testing.T, template *core.PodTemplateSpec, ns ...*Generic) func(in func(t *testing.T, pod *core.PodTemplateSpec, spec *Generic)) {
	var i *Generic

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

	return func(in func(t *testing.T, pod *core.PodTemplateSpec, spec *Generic)) {
		t.Run("Validate", func(t *testing.T) {
			if i != nil {
				in(t, template, i)
			} else {
				in(t, template, &Generic{})
			}
		})
	}
}

func applyGenericYAML(t *testing.T, template *core.PodTemplateSpec, ns ...string) func(in func(t *testing.T, pod *core.PodTemplateSpec, spec *Generic)) {
	elements := make([]*Generic, len(ns))

	for id := range ns {
		var p Generic
		require.NoError(t, yaml.Unmarshal([]byte(ns[id]), &p))
		elements[id] = p.DeepCopy()
	}

	return applyGeneric(t, template, elements...)
}

func Test_Generic(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		applyGeneric(t, nil, nil)(func(t *testing.T, pod *core.PodTemplateSpec, spec *Generic) {
			require.Nil(t, spec.Environments)
			require.Nil(t, spec.VolumeMounts)
		})
	})
	t.Run("Empty template", func(t *testing.T) {
		applyGeneric(t, &core.PodTemplateSpec{})(func(t *testing.T, pod *core.PodTemplateSpec, spec *Generic) {
			require.Nil(t, spec.Environments)
			require.Nil(t, spec.VolumeMounts)
		})
	})
	t.Run("With fields", func(t *testing.T) {
		applyGeneric(t, &core.PodTemplateSpec{
			Spec: core.PodSpec{
				Containers: []core.Container{
					{},
				},
			},
		}, &Generic{
			Environments: &schedulerContainerResourcesApi.Environments{
				Env: []core.EnvVar{
					{
						Name:  "key1",
						Value: "value1",
					},
				},
			},
			VolumeMounts: &schedulerContainerResourcesApi.VolumeMounts{
				VolumeMounts: []core.VolumeMount{
					{
						Name:      "TEST",
						MountPath: "/data",
					},
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, spec *Generic) {
			require.NotNil(t, spec.VolumeMounts)
			require.Len(t, spec.VolumeMounts.VolumeMounts, 1)
			require.EqualValues(t, "TEST", spec.VolumeMounts.VolumeMounts[0].Name)

			// Spec
			require.NotNil(t, spec.Environments)
			require.Len(t, spec.Environments.Env, 1)
			require.EqualValues(t, "key1", spec.Environments.Env[0].Name)
			require.EqualValues(t, "value1", spec.Environments.Env[0].Value)

			// Check
			require.Len(t, pod.Spec.Containers, 1)
			require.Len(t, pod.Spec.Containers[0].Env, 1)
			require.EqualValues(t, "key1", pod.Spec.Containers[0].Env[0].Name)
			require.EqualValues(t, "value1", pod.Spec.Containers[0].Env[0].Value)
		})
	})
}

func Test_Generic_YAML(t *testing.T) {
	t.Run("With Override", func(t *testing.T) {
		applyGenericYAML(t, &core.PodTemplateSpec{
			Spec: core.PodSpec{
				Containers: []core.Container{
					{},
				},
			},
		}, `
---
env:
- name: key1
  value: value1
`, `
---
env:
- name: key1
  value: value2
`)(func(t *testing.T, pod *core.PodTemplateSpec, spec *Generic) {
			// Spec
			require.NotNil(t, spec.Environments)
			require.Len(t, spec.Environments.Env, 1)
			require.EqualValues(t, "key1", spec.Environments.Env[0].Name)
			require.EqualValues(t, "value2", spec.Environments.Env[0].Value)

			// Check
			require.Len(t, pod.Spec.Containers, 1)
			require.Len(t, pod.Spec.Containers[0].Env, 1)
			require.EqualValues(t, "key1", pod.Spec.Containers[0].Env[0].Name)
			require.EqualValues(t, "value2", pod.Spec.Containers[0].Env[0].Value)
		})
	})
	t.Run("With fields", func(t *testing.T) {
		applyGenericYAML(t, &core.PodTemplateSpec{
			Spec: core.PodSpec{
				Containers: []core.Container{
					{},
				},
			},
		}, `
---
env:
- name: key1
  value: value1
`)(func(t *testing.T, pod *core.PodTemplateSpec, spec *Generic) {
			// Spec
			require.NotNil(t, spec.Environments)
			require.Len(t, spec.Environments.Env, 1)
			require.EqualValues(t, "key1", spec.Environments.Env[0].Name)
			require.EqualValues(t, "value1", spec.Environments.Env[0].Value)

			// Check
			require.Len(t, pod.Spec.Containers, 1)
			require.Len(t, pod.Spec.Containers[0].Env, 1)
			require.EqualValues(t, "key1", pod.Spec.Containers[0].Env[0].Name)
			require.EqualValues(t, "value1", pod.Spec.Containers[0].Env[0].Value)
		})
	})
}
