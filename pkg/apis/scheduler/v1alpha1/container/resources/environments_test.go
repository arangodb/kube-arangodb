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
	"k8s.io/apimachinery/pkg/api/resource"
)

func applyEnvironments(t *testing.T, template *core.PodTemplateSpec, container *core.Container, ns ...*Environments) func(in func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container)) {
	var i *Environments

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

func Test_Environments(t *testing.T) {
	t.Run("With Nil", func(t *testing.T) {
		applyEnvironments(t, nil, nil)(func(t *testing.T, _ *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Env, 0)
			require.Len(t, container.EnvFrom, 0)
		})
	})
	t.Run("With Empty", func(t *testing.T) {
		applyEnvironments(t, &core.PodTemplateSpec{}, &core.Container{})(func(t *testing.T, _ *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Env, 0)
			require.Len(t, container.EnvFrom, 0)
		})
	})
	t.Run("With Env", func(t *testing.T) {
		applyEnvironments(t, &core.PodTemplateSpec{}, &core.Container{}, &Environments{
			Env: []core.EnvVar{
				{
					Name:  "var1",
					Value: "value1",
				},
			},
			EnvFrom: []core.EnvFromSource{
				{
					Prefix: "DATA_",
					SecretRef: &core.SecretEnvSource{
						LocalObjectReference: core.LocalObjectReference{
							Name: "secret1",
						},
					},
				},
			},
		})(func(t *testing.T, _ *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Env, 1)
			require.EqualValues(t, "var1", container.Env[0].Name)
			require.EqualValues(t, "value1", container.Env[0].Value)
			require.Nil(t, container.Env[0].ValueFrom)

			require.Len(t, container.EnvFrom, 1)
			require.EqualValues(t, "DATA_", container.EnvFrom[0].Prefix)
			require.NotNil(t, container.EnvFrom[0].SecretRef)
			require.Nil(t, container.EnvFrom[0].SecretRef.Optional)
			require.EqualValues(t, "secret1", container.EnvFrom[0].SecretRef.Name)
			require.Nil(t, container.EnvFrom[0].ConfigMapRef)
		})
	})
	t.Run("With Env Merge", func(t *testing.T) {
		applyEnvironments(t, &core.PodTemplateSpec{}, &core.Container{}, &Environments{
			Env: []core.EnvVar{
				{
					Name:  "var1",
					Value: "value1",
				},
			},
			EnvFrom: []core.EnvFromSource{
				{
					Prefix: "DATA_",
					SecretRef: &core.SecretEnvSource{
						LocalObjectReference: core.LocalObjectReference{
							Name: "secret1",
						},
					},
				},
			},
		}, &Environments{
			Env: []core.EnvVar{
				{
					Name:  "var2",
					Value: "value2",
				},
			},
			EnvFrom: []core.EnvFromSource{
				{
					Prefix: "DATA2_",
					ConfigMapRef: &core.ConfigMapEnvSource{
						LocalObjectReference: core.LocalObjectReference{
							Name: "cm1",
						},
					},
				},
			},
		})(func(t *testing.T, _ *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Env, 2)
			require.EqualValues(t, "var1", container.Env[0].Name)
			require.EqualValues(t, "value1", container.Env[0].Value)
			require.Nil(t, container.Env[0].ValueFrom)
			require.EqualValues(t, "var2", container.Env[1].Name)
			require.EqualValues(t, "value2", container.Env[1].Value)
			require.Nil(t, container.Env[1].ValueFrom)

			require.Len(t, container.EnvFrom, 2)
			require.EqualValues(t, "DATA_", container.EnvFrom[0].Prefix)
			require.NotNil(t, container.EnvFrom[0].SecretRef)
			require.Nil(t, container.EnvFrom[0].SecretRef.Optional)
			require.EqualValues(t, "secret1", container.EnvFrom[0].SecretRef.Name)
			require.Nil(t, container.EnvFrom[0].ConfigMapRef)

			require.EqualValues(t, "DATA2_", container.EnvFrom[1].Prefix)
			require.NotNil(t, container.EnvFrom[1].ConfigMapRef)
			require.Nil(t, container.EnvFrom[1].ConfigMapRef.Optional)
			require.EqualValues(t, "cm1", container.EnvFrom[1].ConfigMapRef.Name)
			require.Nil(t, container.EnvFrom[1].SecretRef)
		})
	})
	t.Run("With Env Replace", func(t *testing.T) {
		applyEnvironments(t, &core.PodTemplateSpec{}, &core.Container{}, &Environments{
			Env: []core.EnvVar{
				{
					Name:  "var1",
					Value: "value1",
				},
			},
		}, &Environments{
			Env: []core.EnvVar{
				{
					Name:  "var1",
					Value: "value2",
				},
			},
		})(func(t *testing.T, _ *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.EnvFrom, 0)
			require.Len(t, container.Env, 1)
			require.EqualValues(t, "var1", container.Env[0].Name)
			require.EqualValues(t, "value2", container.Env[0].Value)
			require.Nil(t, container.Env[0].ValueFrom)
		})
	})
	t.Run("With Env Replace Type", func(t *testing.T) {
		applyEnvironments(t, &core.PodTemplateSpec{}, &core.Container{}, &Environments{
			Env: []core.EnvVar{
				{
					Name:  "var1",
					Value: "value1",
				},
			},
		}, &Environments{
			Env: []core.EnvVar{
				{
					Name: "var1",
					ValueFrom: &core.EnvVarSource{
						ResourceFieldRef: &core.ResourceFieldSelector{
							ContainerName: "a",
							Resource:      "b",
							Divisor:       resource.Quantity{},
						},
					},
				},
			},
		})(func(t *testing.T, _ *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.EnvFrom, 0)
			require.Len(t, container.Env, 1)
			require.EqualValues(t, "var1", container.Env[0].Name)
			require.Empty(t, container.Env[0].Value)
			require.NotNil(t, container.Env[0].ValueFrom)
		})
	})
}
