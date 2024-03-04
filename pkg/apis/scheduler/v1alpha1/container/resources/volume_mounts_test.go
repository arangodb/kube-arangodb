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

func applyVolumeMounts(t *testing.T, template *core.PodTemplateSpec, container *core.Container, ns ...*VolumeMounts) func(in func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container)) {
	var i *VolumeMounts

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

func Test_VolumeMounts(t *testing.T) {
	t.Run("With Nil", func(t *testing.T) {
		applyVolumeMounts(t, nil, nil)(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.VolumeMounts, 0)
		})
	})
	t.Run("With Empty", func(t *testing.T) {
		applyVolumeMounts(t, &core.PodTemplateSpec{}, &core.Container{})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.VolumeMounts, 0)
		})
	})
	t.Run("Add mount", func(t *testing.T) {
		applyVolumeMounts(t, &core.PodTemplateSpec{}, &core.Container{}, &VolumeMounts{
			VolumeMounts: []core.VolumeMount{
				{
					Name:      "test",
					ReadOnly:  false,
					MountPath: "/var/test",
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.VolumeMounts, 1)

			require.EqualValues(t, "test", container.VolumeMounts[0].Name)
			require.EqualValues(t, "/var/test", container.VolumeMounts[0].MountPath)
			require.False(t, container.VolumeMounts[0].ReadOnly)
		})
	})
	t.Run("Append mount", func(t *testing.T) {
		applyVolumeMounts(t, &core.PodTemplateSpec{}, &core.Container{}, &VolumeMounts{
			VolumeMounts: []core.VolumeMount{
				{
					Name:      "test",
					ReadOnly:  false,
					MountPath: "/var/test",
				},
				{
					Name:      "test2",
					ReadOnly:  true,
					MountPath: "/var/test2",
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.VolumeMounts, 2)

			require.EqualValues(t, "test", container.VolumeMounts[0].Name)
			require.EqualValues(t, "/var/test", container.VolumeMounts[0].MountPath)
			require.False(t, container.VolumeMounts[0].ReadOnly)

			require.EqualValues(t, "test2", container.VolumeMounts[1].Name)
			require.EqualValues(t, "/var/test2", container.VolumeMounts[1].MountPath)
			require.True(t, container.VolumeMounts[1].ReadOnly)
		})
	})
	t.Run("Second mount", func(t *testing.T) {
		applyVolumeMounts(t, &core.PodTemplateSpec{}, &core.Container{}, &VolumeMounts{
			VolumeMounts: []core.VolumeMount{
				{
					Name:      "test",
					ReadOnly:  false,
					MountPath: "/var/test",
				},
				{
					Name:      "test",
					ReadOnly:  true,
					MountPath: "/var/test2",
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.VolumeMounts, 2)

			require.EqualValues(t, "test", container.VolumeMounts[0].Name)
			require.EqualValues(t, "/var/test", container.VolumeMounts[0].MountPath)
			require.False(t, container.VolumeMounts[0].ReadOnly)

			require.EqualValues(t, "test", container.VolumeMounts[1].Name)
			require.EqualValues(t, "/var/test2", container.VolumeMounts[1].MountPath)
			require.True(t, container.VolumeMounts[1].ReadOnly)
		})
	})
}
