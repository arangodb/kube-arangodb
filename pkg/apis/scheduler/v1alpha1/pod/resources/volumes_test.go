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

func applyVolumes(t *testing.T, template *core.PodTemplateSpec, ns ...*Volumes) func(in func(t *testing.T, pod *core.PodTemplateSpec)) {
	var i *Volumes

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

func Test_Volumes(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		applyVolumes(t, nil)(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.Volumes, 0)
		})
	})
	t.Run("Empty", func(t *testing.T) {
		applyVolumes(t, &core.PodTemplateSpec{})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.Volumes, 0)
		})
	})
	t.Run("Add volume", func(t *testing.T) {
		applyVolumes(t, &core.PodTemplateSpec{}, &Volumes{
			Volumes: []core.Volume{
				{
					Name: "test",
					VolumeSource: core.VolumeSource{
						EmptyDir: &core.EmptyDirVolumeSource{},
					},
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.Volumes, 1)

			require.EqualValues(t, "test", pod.Spec.Volumes[0].Name)
			require.NotNil(t, pod.Spec.Volumes[0].VolumeSource.EmptyDir)
			require.Nil(t, pod.Spec.Volumes[0].VolumeSource.Secret)
		})
	})
	t.Run("Append volume", func(t *testing.T) {
		applyVolumes(t, &core.PodTemplateSpec{}, &Volumes{
			Volumes: []core.Volume{
				{
					Name: "test",
					VolumeSource: core.VolumeSource{
						EmptyDir: &core.EmptyDirVolumeSource{},
					},
				},
			},
		}, &Volumes{
			Volumes: []core.Volume{
				{
					Name: "test2",
					VolumeSource: core.VolumeSource{
						Secret: &core.SecretVolumeSource{
							SecretName: "test",
						},
					},
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.Volumes, 2)

			require.EqualValues(t, "test", pod.Spec.Volumes[0].Name)
			require.NotNil(t, pod.Spec.Volumes[0].VolumeSource.EmptyDir)
			require.Nil(t, pod.Spec.Volumes[0].VolumeSource.Secret)

			require.EqualValues(t, "test2", pod.Spec.Volumes[1].Name)
			require.Nil(t, pod.Spec.Volumes[1].VolumeSource.EmptyDir)
			require.NotNil(t, pod.Spec.Volumes[1].VolumeSource.Secret)
			require.EqualValues(t, "test", pod.Spec.Volumes[1].VolumeSource.Secret.SecretName)
		})
	})
	t.Run("Update volume", func(t *testing.T) {
		applyVolumes(t, &core.PodTemplateSpec{}, &Volumes{
			Volumes: []core.Volume{
				{
					Name: "test",
					VolumeSource: core.VolumeSource{
						EmptyDir: &core.EmptyDirVolumeSource{},
					},
				},
			},
		}, &Volumes{
			Volumes: []core.Volume{
				{
					Name: "test",
					VolumeSource: core.VolumeSource{
						Secret: &core.SecretVolumeSource{
							SecretName: "test",
						},
					},
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.Volumes, 1)

			require.EqualValues(t, "test", pod.Spec.Volumes[0].Name)
			require.Nil(t, pod.Spec.Volumes[0].VolumeSource.EmptyDir)
			require.NotNil(t, pod.Spec.Volumes[0].VolumeSource.Secret)
			require.EqualValues(t, "test", pod.Spec.Volumes[0].VolumeSource.Secret.SecretName)
		})
	})
}
