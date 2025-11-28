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
)

func applyImage(t *testing.T, template *core.PodTemplateSpec, ns ...*Image) func(in func(t *testing.T, pod *core.PodTemplateSpec)) {
	var i *Image

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

func Test_Image(t *testing.T) {
	t.Run("With Nil", func(t *testing.T) {
		applyImage(t, nil, nil)(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.ImagePullSecrets, 0)
		})
	})
	t.Run("With Empty", func(t *testing.T) {
		applyImage(t, &core.PodTemplateSpec{})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.ImagePullSecrets, 0)
		})
	})
	t.Run("With PS", func(t *testing.T) {
		applyImage(t, &core.PodTemplateSpec{}, &Image{
			ImagePullSecrets: []string{
				"secret",
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.ImagePullSecrets, 1)
			require.Equal(t, "secret", pod.Spec.ImagePullSecrets[0].Name)
		})
	})
	t.Run("With PS2", func(t *testing.T) {
		applyImage(t, &core.PodTemplateSpec{}, &Image{
			ImagePullSecrets: []string{
				"secret",
				"secret",
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.ImagePullSecrets, 1)
			require.Equal(t, "secret", pod.Spec.ImagePullSecrets[0].Name)
		})
	})
	t.Run("With Merge", func(t *testing.T) {
		applyImage(t, &core.PodTemplateSpec{}, &Image{
			ImagePullSecrets: []string{
				"secret",
			},
		}, &Image{
			ImagePullSecrets: []string{
				"secret2",
			},
		}, &Image{
			ImagePullSecrets: []string{
				"secret",
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Spec.ImagePullSecrets, 2)
			require.Equal(t, "secret", pod.Spec.ImagePullSecrets[0].Name)
			require.Equal(t, "secret2", pod.Spec.ImagePullSecrets[1].Name)
		})
	})
}
