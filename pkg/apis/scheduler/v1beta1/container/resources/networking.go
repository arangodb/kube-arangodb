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
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/interfaces"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

var _ interfaces.Container[Networking] = &Networking{}

type Networking struct {
	// Ports contains list of ports to expose from the container. Not specifying a port here
	// DOES NOT prevent that port from being exposed. Any port which is
	// listening on the default "0.0.0.0" address inside a container will be
	// accessible from the network.
	// +doc/type: []core.ContainerPort
	Ports []core.ContainerPort `json:"ports,omitempty"`
}

func (n *Networking) Apply(pod *core.PodTemplateSpec, template *core.Container) error {
	if n == nil {
		return nil
	}

	for _, port := range n.Ports {
		if port.Name == "" {
			continue
		}

		for _, container := range pod.Spec.Containers {
			for _, existingPort := range container.Ports {
				if port.Name == existingPort.Name {
					return errors.Errorf("Port with name `%s` already exposed in container `%s`", port.Name, container.Name)
				}
			}
		}
	}

	obj := n.DeepCopy()

	template.Ports = obj.Ports

	return nil
}

func (n *Networking) With(newResources *Networking) *Networking {
	if n == nil && newResources == nil {
		return nil
	}

	if n == nil {
		return newResources.DeepCopy()
	}

	if newResources == nil {
		return n.DeepCopy()
	}

	return &Networking{Ports: kresources.MergeContainerPorts(n.Ports, newResources.Ports...)}
}

func (n *Networking) Validate() error {
	return nil
}
