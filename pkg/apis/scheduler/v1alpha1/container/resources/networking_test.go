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

func applyNetworking(t *testing.T, template *core.PodTemplateSpec, container *core.Container, ns ...*Networking) func(in func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container)) {
	var i *Networking

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

func Test_Networking(t *testing.T) {
	t.Run("With Nil", func(t *testing.T) {
		applyNetworking(t, nil, nil)(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Ports, 0)
		})
	})
	t.Run("With Empty", func(t *testing.T) {
		applyNetworking(t, &core.PodTemplateSpec{}, &core.Container{})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Ports, 0)
		})
	})
	t.Run("With Port", func(t *testing.T) {
		applyNetworking(t, &core.PodTemplateSpec{}, &core.Container{}, &Networking{
			Ports: []core.ContainerPort{
				{
					Name:          "TEST",
					ContainerPort: 1,
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Ports, 1)

			require.EqualValues(t, "TEST", container.Ports[0].Name)
			require.EqualValues(t, 1, container.Ports[0].ContainerPort)
		})
	})
	t.Run("With Port Append", func(t *testing.T) {
		applyNetworking(t, &core.PodTemplateSpec{}, &core.Container{}, &Networking{
			Ports: []core.ContainerPort{
				{
					Name:          "TEST",
					ContainerPort: 1,
				},
			},
		}, &Networking{
			Ports: []core.ContainerPort{
				{
					Name:          "TEST2",
					ContainerPort: 2,
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Ports, 2)

			require.EqualValues(t, "TEST", container.Ports[0].Name)
			require.EqualValues(t, 1, container.Ports[0].ContainerPort)

			require.EqualValues(t, "TEST2", container.Ports[1].Name)
			require.EqualValues(t, 2, container.Ports[1].ContainerPort)
		})
	})
	t.Run("With Port Update", func(t *testing.T) {
		applyNetworking(t, &core.PodTemplateSpec{}, &core.Container{}, &Networking{
			Ports: []core.ContainerPort{
				{
					Name:          "TEST",
					ContainerPort: 1,
					HostIP:        "IP",
				},
			},
		}, &Networking{
			Ports: []core.ContainerPort{
				{
					Name:          "TEST",
					ContainerPort: 2,
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Ports, 1)

			require.EqualValues(t, "TEST", container.Ports[0].Name)
			require.EqualValues(t, 2, container.Ports[0].ContainerPort)
			require.EqualValues(t, "", container.Ports[0].HostIP)
		})
	})
}
