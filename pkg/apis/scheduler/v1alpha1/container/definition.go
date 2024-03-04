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
	core "k8s.io/api/core/v1"

	schedulerContainerResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/container/resources"
	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/interfaces"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/container"
)

type Containers map[string]Container

func (c Containers) Apply(template *core.PodTemplateSpec) error {
	if len(c) == 0 {
		return nil
	}

	for k, v := range c {
		if id := container.GetContainerIDByName(template.Spec.Containers, k); id >= 0 {
			if err := v.Apply(template, &template.Spec.Containers[id]); err != nil {
				return err
			}
		} else {
			id = len(template.Spec.Containers)

			template.Spec.Containers = append(template.Spec.Containers, core.Container{
				Name: k,
			})

			if err := v.Apply(template, &template.Spec.Containers[id]); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c Containers) With(other Containers) Containers {
	if len(c) == 0 && len(other) == 0 {
		return nil
	}

	if len(c) == 0 {
		return other.DeepCopy()
	}

	if len(other) == 0 {
		return c.DeepCopy()
	}

	ret := Containers{}

	for k, v := range c {
		if v1, ok := other[k]; !ok {
			ret[k] = v
		} else {
			ret[k] = util.TypeOrDefault(v.With(&v1))
		}
	}

	for k, v := range other {
		if _, ok := c[k]; !ok {
			ret[k] = v
		}
	}

	return ret
}

func (c Containers) Validate() error {
	for name, container := range c {
		if err := container.Validate(); err != nil {
			return errors.Wrapf(err, "Container %s failed", name)
		}
	}

	return nil
}

var _ interfaces.Container[Container] = &Container{}

type Container struct {
	// Security keeps the security settings for Container
	*schedulerContainerResourcesApi.Security `json:",inline"`

	// Environments keeps the environment variables for Container
	*schedulerContainerResourcesApi.Environments `json:",inline"`

	// Image define default image used for the Container
	*schedulerContainerResourcesApi.Image `json:",inline"`

	// Resources define resources assigned to the Container
	*schedulerContainerResourcesApi.Resources `json:",inline"`

	// VolumeMounts define volume mounts assigned to the Container
	*schedulerContainerResourcesApi.VolumeMounts `json:",inline"`

	// Probes define probes assigned to the Container
	*schedulerContainerResourcesApi.Probes `json:",inline"`

	// Networking define networking assigned to the Container
	*schedulerContainerResourcesApi.Networking `json:",inline"`

	// Lifecycle define lifecycle assigned to the Container
	*schedulerContainerResourcesApi.Lifecycle `json:",inline"`
}

func (c *Container) Apply(template *core.PodTemplateSpec, container *core.Container) error {
	if c == nil {
		return nil
	}

	return shared.WithErrors(
		c.Security.Apply(template, container),
		c.Environments.Apply(template, container),
		c.Image.Apply(template, container),
		c.Resources.Apply(template, container),
		c.VolumeMounts.Apply(template, container),
		c.Probes.Apply(template, container),
		c.Networking.Apply(template, container),
		c.Lifecycle.Apply(template, container),
	)
}

func (c *Container) GetImage() *schedulerContainerResourcesApi.Image {
	if c == nil || c.Image == nil {
		return nil
	}

	return c.Image
}

func (c *Container) GetResources() *schedulerContainerResourcesApi.Resources {
	if c == nil || c.Resources == nil {
		return nil
	}

	return c.Resources
}

func (c *Container) GetSecurity() *schedulerContainerResourcesApi.Security {
	if c == nil || c.Security == nil {
		return nil
	}

	return c.Security
}

func (c *Container) GetEnvironments() *schedulerContainerResourcesApi.Environments {
	if c == nil || c.Environments == nil {
		return nil
	}

	return c.Environments
}

func (c *Container) GetVolumeMounts() *schedulerContainerResourcesApi.VolumeMounts {
	if c == nil || c.VolumeMounts == nil {
		return nil
	}

	return c.VolumeMounts
}

func (c *Container) GetProbes() *schedulerContainerResourcesApi.Probes {
	if c == nil || c.Probes == nil {
		return nil
	}

	return c.Probes
}

func (c *Container) GetNetworking() *schedulerContainerResourcesApi.Networking {
	if c == nil || c.Networking == nil {
		return nil
	}

	return c.Networking
}

func (c *Container) GetLifecycle() *schedulerContainerResourcesApi.Lifecycle {
	if c == nil || c.Lifecycle == nil {
		return nil
	}

	return c.Lifecycle
}

func (c *Container) With(other *Container) *Container {
	if c == nil && other == nil {
		return nil
	}

	if c == nil {
		return other.DeepCopy()
	}

	if other == nil {
		return c.DeepCopy()
	}

	return &Container{
		Security:     c.Security.With(other.Security),
		Environments: c.Environments.With(other.Environments),
		Image:        c.Image.With(other.Image),
		Resources:    c.Resources.With(other.Resources),
		VolumeMounts: c.VolumeMounts.With(other.VolumeMounts),
		Lifecycle:    c.Lifecycle.With(other.Lifecycle),
		Networking:   c.Networking.With(other.Networking),
		Probes:       c.Probes.With(other.Probes),
	}
}

func (c *Container) Validate() error {
	if c == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceErrors("containerSecurity", c.Security.Validate()),
		shared.PrefixResourceErrors("containerEnvironments", c.Environments.Validate()),
		shared.PrefixResourceErrors("containerResources", c.Image.Validate()),
		shared.PrefixResourceErrors("containerImage", c.Resources.Validate()),
		shared.PrefixResourceErrors("volumeMounts", c.VolumeMounts.Validate()),
		shared.PrefixResourceErrors("lifecycle", c.Lifecycle.Validate()),
		shared.PrefixResourceErrors("networking", c.Networking.Validate()),
		shared.PrefixResourceErrors("probes", c.Probes.Validate()),
	)
}
