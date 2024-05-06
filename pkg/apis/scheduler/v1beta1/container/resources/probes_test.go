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

func applyProbes(t *testing.T, template *core.PodTemplateSpec, container *core.Container, ns ...*Probes) func(in func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container)) {
	var i *Probes

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

func Test_Probes(t *testing.T) {
	t.Run("With Nil", func(t *testing.T) {
		applyProbes(t, nil, nil)(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Nil(t, container.ReadinessProbe)
			require.Nil(t, container.LivenessProbe)
			require.Nil(t, container.StartupProbe)
		})
	})
	t.Run("With Empty", func(t *testing.T) {
		applyProbes(t, &core.PodTemplateSpec{}, &core.Container{})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.Nil(t, container.ReadinessProbe)
			require.Nil(t, container.LivenessProbe)
			require.Nil(t, container.StartupProbe)
		})
	})
	t.Run("With Probes", func(t *testing.T) {
		applyProbes(t, &core.PodTemplateSpec{}, &core.Container{}, &Probes{
			ReadinessProbe: &core.Probe{
				ProbeHandler: core.ProbeHandler{
					Exec: &core.ExecAction{
						Command: []string{"test"},
					},
				},
				InitialDelaySeconds: 10,
			},
			LivenessProbe: &core.Probe{
				ProbeHandler: core.ProbeHandler{
					HTTPGet: &core.HTTPGetAction{
						Path: "/test",
					},
				},
				InitialDelaySeconds: 15,
			},
			StartupProbe: &core.Probe{
				ProbeHandler: core.ProbeHandler{
					GRPC: &core.GRPCAction{
						Port: 33,
					},
				},
				InitialDelaySeconds: 20,
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.NotNil(t, container.ReadinessProbe)
			require.EqualValues(t, 10, container.ReadinessProbe.InitialDelaySeconds)
			require.EqualValues(t, 0, container.ReadinessProbe.TimeoutSeconds)
			require.Nil(t, container.ReadinessProbe.HTTPGet)
			require.NotNil(t, container.ReadinessProbe.Exec)
			require.Nil(t, container.ReadinessProbe.GRPC)
			require.Len(t, container.ReadinessProbe.Exec.Command, 1)
			require.EqualValues(t, "test", container.ReadinessProbe.Exec.Command[0])

			require.NotNil(t, container.LivenessProbe)
			require.EqualValues(t, 15, container.LivenessProbe.InitialDelaySeconds)
			require.EqualValues(t, 0, container.LivenessProbe.TimeoutSeconds)
			require.NotNil(t, container.LivenessProbe.HTTPGet)
			require.Nil(t, container.LivenessProbe.Exec)
			require.Nil(t, container.LivenessProbe.GRPC)
			require.EqualValues(t, "/test", container.LivenessProbe.HTTPGet.Path)

			require.NotNil(t, container.StartupProbe)
			require.EqualValues(t, 20, container.StartupProbe.InitialDelaySeconds)
			require.EqualValues(t, 0, container.StartupProbe.TimeoutSeconds)
			require.Nil(t, container.StartupProbe.HTTPGet)
			require.Nil(t, container.StartupProbe.Exec)
			require.NotNil(t, container.StartupProbe.GRPC)
			require.EqualValues(t, 33, container.StartupProbe.GRPC.Port)
		})
	})
	t.Run("With Time Updates", func(t *testing.T) {
		applyProbes(t, &core.PodTemplateSpec{}, &core.Container{}, &Probes{
			ReadinessProbe: &core.Probe{
				ProbeHandler: core.ProbeHandler{
					Exec: &core.ExecAction{
						Command: []string{"test"},
					},
				},
				InitialDelaySeconds: 10,
			},
			LivenessProbe: &core.Probe{
				ProbeHandler: core.ProbeHandler{
					HTTPGet: &core.HTTPGetAction{
						Path: "/test",
					},
				},
				InitialDelaySeconds: 15,
			},
			StartupProbe: &core.Probe{
				ProbeHandler: core.ProbeHandler{
					GRPC: &core.GRPCAction{
						Port: 33,
					},
				},
				InitialDelaySeconds: 20,
			},
		}, &Probes{
			ReadinessProbe: &core.Probe{
				InitialDelaySeconds: 60,
			},
			LivenessProbe: &core.Probe{
				InitialDelaySeconds: 61,
			},
			StartupProbe: &core.Probe{
				TimeoutSeconds: 62,
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.NotNil(t, container.ReadinessProbe)
			require.EqualValues(t, 60, container.ReadinessProbe.InitialDelaySeconds)
			require.EqualValues(t, 0, container.ReadinessProbe.TimeoutSeconds)
			require.Nil(t, container.ReadinessProbe.HTTPGet)
			require.NotNil(t, container.ReadinessProbe.Exec)
			require.Nil(t, container.ReadinessProbe.GRPC)
			require.Len(t, container.ReadinessProbe.Exec.Command, 1)
			require.EqualValues(t, "test", container.ReadinessProbe.Exec.Command[0])

			require.NotNil(t, container.LivenessProbe)
			require.EqualValues(t, 61, container.LivenessProbe.InitialDelaySeconds)
			require.EqualValues(t, 0, container.LivenessProbe.TimeoutSeconds)
			require.NotNil(t, container.LivenessProbe.HTTPGet)
			require.Nil(t, container.LivenessProbe.Exec)
			require.Nil(t, container.LivenessProbe.GRPC)
			require.EqualValues(t, "/test", container.LivenessProbe.HTTPGet.Path)

			require.NotNil(t, container.StartupProbe)
			require.EqualValues(t, 20, container.StartupProbe.InitialDelaySeconds)
			require.EqualValues(t, 62, container.StartupProbe.TimeoutSeconds)
			require.Nil(t, container.StartupProbe.HTTPGet)
			require.Nil(t, container.StartupProbe.Exec)
			require.NotNil(t, container.StartupProbe.GRPC)
			require.EqualValues(t, 33, container.StartupProbe.GRPC.Port)
		})
	})
	t.Run("With Exec Updates", func(t *testing.T) {
		applyProbes(t, &core.PodTemplateSpec{}, &core.Container{}, &Probes{
			ReadinessProbe: &core.Probe{
				ProbeHandler: core.ProbeHandler{
					Exec: &core.ExecAction{
						Command: []string{"test"},
					},
				},
				InitialDelaySeconds: 10,
			},
			LivenessProbe: &core.Probe{
				ProbeHandler: core.ProbeHandler{
					HTTPGet: &core.HTTPGetAction{
						Path: "/test",
					},
				},
				InitialDelaySeconds: 15,
			},
			StartupProbe: &core.Probe{
				ProbeHandler: core.ProbeHandler{
					GRPC: &core.GRPCAction{
						Port: 33,
					},
				},
				InitialDelaySeconds: 20,
			},
		}, &Probes{
			ReadinessProbe: &core.Probe{
				ProbeHandler: core.ProbeHandler{
					GRPC: &core.GRPCAction{
						Port: 33,
					},
				},
			},
			LivenessProbe: &core.Probe{
				ProbeHandler: core.ProbeHandler{
					Exec: &core.ExecAction{
						Command: []string{"test"},
					},
				},
			},
			StartupProbe: &core.Probe{
				ProbeHandler: core.ProbeHandler{
					HTTPGet: &core.HTTPGetAction{
						Path: "/test",
					},
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container) {
			require.NotNil(t, container.ReadinessProbe)
			require.EqualValues(t, 10, container.ReadinessProbe.InitialDelaySeconds)
			require.EqualValues(t, 0, container.ReadinessProbe.TimeoutSeconds)
			require.Nil(t, container.ReadinessProbe.HTTPGet)
			require.Nil(t, container.ReadinessProbe.Exec)
			require.NotNil(t, container.ReadinessProbe.GRPC)
			require.EqualValues(t, 33, container.ReadinessProbe.GRPC.Port)

			require.NotNil(t, container.LivenessProbe)
			require.EqualValues(t, 15, container.LivenessProbe.InitialDelaySeconds)
			require.EqualValues(t, 0, container.LivenessProbe.TimeoutSeconds)
			require.Nil(t, container.LivenessProbe.HTTPGet)
			require.NotNil(t, container.LivenessProbe.Exec)
			require.Nil(t, container.LivenessProbe.GRPC)
			require.Len(t, container.LivenessProbe.Exec.Command, 1)
			require.EqualValues(t, "test", container.LivenessProbe.Exec.Command[0])

			require.NotNil(t, container.StartupProbe)
			require.EqualValues(t, 20, container.StartupProbe.InitialDelaySeconds)
			require.EqualValues(t, 0, container.StartupProbe.TimeoutSeconds)
			require.NotNil(t, container.StartupProbe.HTTPGet)
			require.Nil(t, container.StartupProbe.Exec)
			require.Nil(t, container.StartupProbe.GRPC)
			require.EqualValues(t, "/test", container.StartupProbe.HTTPGet.Path)
		})
	})
}
