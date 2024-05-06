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

func applyContainer(t *testing.T, template *core.PodTemplateSpec, container *core.Container, ns ...*Container) func(in func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container, spec *Container)) {
	var i *Container

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

	return func(in func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container, spec *Container)) {
		t.Run("Validate", func(t *testing.T) {
			if i != nil {
				in(t, template, container, i)
			} else {
				in(t, template, container, &Container{})
			}
		})
	}
}

func applyContainerYAML(t *testing.T, template *core.PodTemplateSpec, container *core.Container, ns ...string) func(in func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container, spec *Container)) {
	elements := make([]*Container, len(ns))

	for id := range ns {
		var p Container
		require.NoError(t, yaml.Unmarshal([]byte(ns[id]), &p))
		elements[id] = p.DeepCopy()
	}

	return applyContainer(t, template, container, elements...)
}

func Test_Container(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		applyContainer(t, nil, nil)(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container, spec *Container) {
			require.Nil(t, spec.Resources)
			require.Nil(t, spec.Image)
			require.Nil(t, spec.Security)
			require.Nil(t, spec.Environments)
			require.Nil(t, spec.VolumeMounts)
			require.Nil(t, spec.Core)

			require.Len(t, container.Env, 0)
		})
	})
	t.Run("Empty template", func(t *testing.T) {
		applyContainer(t, &core.PodTemplateSpec{}, &core.Container{})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container, spec *Container) {
			require.Nil(t, spec.Resources)
			require.Nil(t, spec.Image)
			require.Nil(t, spec.Security)
			require.Nil(t, spec.Environments)
			require.Nil(t, spec.VolumeMounts)
			require.Nil(t, spec.Core)

			require.Len(t, container.Env, 0)
		})
	})
	t.Run("With fields", func(t *testing.T) {
		applyContainer(t, &core.PodTemplateSpec{}, &core.Container{}, &Container{
			Core: &schedulerContainerResourcesApiv1alpha1.Core{
				Args: []string{"A"},
			},
			Security: &schedulerContainerResourcesApiv1alpha1.Security{
				SecurityContext: &core.SecurityContext{
					RunAsUser: util.NewType[int64](50),
				},
			},
			Environments: &schedulerContainerResourcesApiv1alpha1.Environments{
				Env: []core.EnvVar{
					{
						Name:  "key1",
						Value: "value1",
					},
				},
			},
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
			VolumeMounts: &schedulerContainerResourcesApiv1alpha1.VolumeMounts{
				VolumeMounts: []core.VolumeMount{
					{
						Name:      "TEST",
						MountPath: "/data",
					},
				},
			},
			Probes: &schedulerContainerResourcesApiv1alpha1.Probes{
				LivenessProbe: &core.Probe{
					InitialDelaySeconds: 1,
				},
				ReadinessProbe: &core.Probe{
					InitialDelaySeconds: 2,
				},
				StartupProbe: &core.Probe{
					InitialDelaySeconds: 3,
				},
			},
			Lifecycle: &schedulerContainerResourcesApiv1alpha1.Lifecycle{
				Lifecycle: &core.Lifecycle{
					PostStart: &core.LifecycleHandler{
						HTTPGet: &core.HTTPGetAction{
							Path: "test1",
						},
					},
					PreStop: &core.LifecycleHandler{
						HTTPGet: &core.HTTPGetAction{
							Path: "test2",
						},
					},
				},
			},
			Networking: &schedulerContainerResourcesApiv1alpha1.Networking{
				Ports: []core.ContainerPort{
					{
						Name: "TEST",
					},
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container, spec *Container) {
			// Spec
			require.NotNil(t, spec.Core)
			require.NotNil(t, spec.Core.Args)
			require.Contains(t, spec.Core.Args, "A")
			require.Empty(t, spec.Core.Command)
			require.Empty(t, spec.Core.WorkingDir)

			require.NotNil(t, spec.Resources)
			require.NotNil(t, spec.Resources.Resources)
			require.Contains(t, spec.Resources.Resources.Limits, core.ResourceCPU)
			require.EqualValues(t, resource.MustParse("1"), spec.Resources.Resources.Limits[core.ResourceCPU])

			require.NotNil(t, spec.Image)
			require.NotNil(t, spec.Image.Image)
			require.EqualValues(t, "test", *spec.Image.Image)

			require.NotNil(t, spec.Security)
			require.NotNil(t, spec.Security.SecurityContext)
			require.NotNil(t, spec.Security.SecurityContext.RunAsUser)
			require.EqualValues(t, 50, *spec.Security.SecurityContext.RunAsUser)

			require.NotNil(t, spec.Environments)
			require.Len(t, spec.Environments.Env, 1)
			require.EqualValues(t, "key1", spec.Environments.Env[0].Name)
			require.EqualValues(t, "value1", spec.Environments.Env[0].Value)

			require.NotNil(t, spec.VolumeMounts)
			require.Len(t, spec.VolumeMounts.VolumeMounts, 1)
			require.EqualValues(t, "TEST", spec.VolumeMounts.VolumeMounts[0].Name)

			require.NotNil(t, spec.Probes)
			require.NotNil(t, spec.Probes.LivenessProbe)
			require.EqualValues(t, 1, spec.Probes.LivenessProbe.InitialDelaySeconds)
			require.NotNil(t, spec.Probes.ReadinessProbe)
			require.EqualValues(t, 2, spec.Probes.ReadinessProbe.InitialDelaySeconds)
			require.NotNil(t, spec.Probes.StartupProbe)
			require.EqualValues(t, 3, spec.Probes.StartupProbe.InitialDelaySeconds)

			require.NotNil(t, spec.Lifecycle)
			require.NotNil(t, spec.Lifecycle.Lifecycle)
			require.NotNil(t, spec.Lifecycle.Lifecycle.PostStart)
			require.NotNil(t, spec.Lifecycle.Lifecycle.PostStart.HTTPGet)
			require.EqualValues(t, "test1", spec.Lifecycle.Lifecycle.PostStart.HTTPGet.Path)
			require.NotNil(t, spec.Lifecycle.Lifecycle.PreStop)
			require.NotNil(t, spec.Lifecycle.Lifecycle.PreStop.HTTPGet)
			require.EqualValues(t, "test2", spec.Lifecycle.Lifecycle.PreStop.HTTPGet.Path)

			require.NotNil(t, spec.Networking)
			require.Len(t, spec.Networking.Ports, 1)
			require.EqualValues(t, "TEST", spec.Networking.Ports[0].Name)
		})
	})
}

func Test_Container_YAML(t *testing.T) {
	t.Run("With Override", func(t *testing.T) {
		applyContainerYAML(t, &core.PodTemplateSpec{}, &core.Container{}, `
---
securityContext:
  runAsUser: 50

args:
- A

env:
- name: key1
  value: value1

image: test

resources:
  limits:
    cpu: 1

volumeMounts:
  - name: TEST

livenessProbe:
  initialDelaySeconds: 1

readinessProbe:
  initialDelaySeconds: 2

startupProbe:
  initialDelaySeconds: 3

lifecycle:
  postStart:
    httpGet:
      path: test1
  preStop:
    httpGet:
      path: test2

ports:
  - name: TEST
`, `
---

securityContext:
  runAsUser: 10
`)(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container, spec *Container) {
			// Spec
			require.NotNil(t, spec.Core)
			require.NotNil(t, spec.Core.Args)
			require.Contains(t, spec.Core.Args, "A")
			require.Empty(t, spec.Core.Command)
			require.Empty(t, spec.Core.WorkingDir)

			require.NotNil(t, spec.Resources)
			require.NotNil(t, spec.Resources.Resources)
			require.Contains(t, spec.Resources.Resources.Limits, core.ResourceCPU)
			require.EqualValues(t, resource.MustParse("1"), spec.Resources.Resources.Limits[core.ResourceCPU])

			require.NotNil(t, spec.Image)
			require.NotNil(t, spec.Image.Image)
			require.EqualValues(t, "test", *spec.Image.Image)

			require.NotNil(t, spec.Security)
			require.NotNil(t, spec.Security.SecurityContext)
			require.NotNil(t, spec.Security.SecurityContext.RunAsUser)
			require.EqualValues(t, 10, *spec.Security.SecurityContext.RunAsUser)

			require.NotNil(t, spec.Environments)
			require.Len(t, spec.Environments.Env, 1)
			require.EqualValues(t, "key1", spec.Environments.Env[0].Name)
			require.EqualValues(t, "value1", spec.Environments.Env[0].Value)

			require.NotNil(t, spec.VolumeMounts)
			require.Len(t, spec.VolumeMounts.VolumeMounts, 1)
			require.EqualValues(t, "TEST", spec.VolumeMounts.VolumeMounts[0].Name)

			require.NotNil(t, spec.Probes)
			require.NotNil(t, spec.Probes.LivenessProbe)
			require.EqualValues(t, 1, spec.Probes.LivenessProbe.InitialDelaySeconds)
			require.NotNil(t, spec.Probes.ReadinessProbe)
			require.EqualValues(t, 2, spec.Probes.ReadinessProbe.InitialDelaySeconds)
			require.NotNil(t, spec.Probes.StartupProbe)
			require.EqualValues(t, 3, spec.Probes.StartupProbe.InitialDelaySeconds)

			require.NotNil(t, spec.Lifecycle)
			require.NotNil(t, spec.Lifecycle.Lifecycle)
			require.NotNil(t, spec.Lifecycle.Lifecycle.PostStart)
			require.NotNil(t, spec.Lifecycle.Lifecycle.PostStart.HTTPGet)
			require.EqualValues(t, "test1", spec.Lifecycle.Lifecycle.PostStart.HTTPGet.Path)
			require.NotNil(t, spec.Lifecycle.Lifecycle.PreStop)
			require.NotNil(t, spec.Lifecycle.Lifecycle.PreStop.HTTPGet)
			require.EqualValues(t, "test2", spec.Lifecycle.Lifecycle.PreStop.HTTPGet.Path)

			require.NotNil(t, spec.Networking)
			require.Len(t, spec.Networking.Ports, 1)
			require.EqualValues(t, "TEST", spec.Networking.Ports[0].Name)
		})
	})
	t.Run("With fields", func(t *testing.T) {
		applyContainerYAML(t, &core.PodTemplateSpec{}, &core.Container{}, `
---
securityContext:
  runAsUser: 50

env:
- name: key1
  value: value1

image: test

resources:
  limits:
    cpu: 1
`)(func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container, spec *Container) {
			// Spec
			require.NotNil(t, spec.Resources)
			require.NotNil(t, spec.Resources.Resources)
			require.Contains(t, spec.Resources.Resources.Limits, core.ResourceCPU)
			require.EqualValues(t, resource.MustParse("1"), spec.Resources.Resources.Limits[core.ResourceCPU])

			require.NotNil(t, spec.Image)
			require.NotNil(t, spec.Image.Image)
			require.EqualValues(t, "test", *spec.Image.Image)

			require.NotNil(t, spec.Security)
			require.NotNil(t, spec.Security.SecurityContext)
			require.NotNil(t, spec.Security.SecurityContext.RunAsUser)
			require.EqualValues(t, 50, *spec.Security.SecurityContext.RunAsUser)

			require.NotNil(t, spec.Environments)
			require.Len(t, spec.Environments.Env, 1)
			require.EqualValues(t, "key1", spec.Environments.Env[0].Name)
			require.EqualValues(t, "value1", spec.Environments.Env[0].Value)
		})
	})
}
