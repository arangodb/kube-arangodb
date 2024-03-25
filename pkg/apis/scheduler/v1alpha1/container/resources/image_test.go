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

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func applyImage(t *testing.T, template *core.PodTemplateSpec, container *core.Container, ns ...*Image) func(in func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container)) {
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

func Test_Image(t *testing.T) {
	t.Run("With Nil", func(t *testing.T) {
		applyImage(t, nil, nil)(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Empty(t, container.Image)
			require.Empty(t, container.ImagePullPolicy)
			require.Len(t, pod.Spec.ImagePullSecrets, 0)
		})
	})
	t.Run("With Empty", func(t *testing.T) {
		applyImage(t, &core.PodTemplateSpec{}, &core.Container{})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Empty(t, container.Image)
			require.Empty(t, container.ImagePullPolicy)
			require.Len(t, pod.Spec.ImagePullSecrets, 0)
		})
	})
	t.Run("With Image", func(t *testing.T) {
		applyImage(t, &core.PodTemplateSpec{}, &core.Container{}, &Image{
			Image: util.NewType("image"),
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.EqualValues(t, "image", container.Image)
			require.Empty(t, container.ImagePullPolicy)
			require.Len(t, pod.Spec.ImagePullSecrets, 0)
		})
	})
	t.Run("With PS", func(t *testing.T) {
		applyImage(t, &core.PodTemplateSpec{}, &core.Container{}, &Image{
			Image:           util.NewType("image"),
			ImagePullPolicy: util.NewType(core.PullAlways),
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.EqualValues(t, "image", container.Image)
			require.EqualValues(t, core.PullAlways, container.ImagePullPolicy)
		})
	})
	t.Run("With PS2", func(t *testing.T) {
		applyImage(t, &core.PodTemplateSpec{}, &core.Container{}, &Image{
			Image:           util.NewType("image"),
			ImagePullPolicy: util.NewType(core.PullAlways),
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.EqualValues(t, "image", container.Image)
			require.EqualValues(t, core.PullAlways, container.ImagePullPolicy)
		})
	})
}
