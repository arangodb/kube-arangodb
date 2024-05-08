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

package v1beta1

import (
	schedulerContainerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container"
	schedulerPodApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/pod"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	ArangoMLExtensionSpecDeploymentApi = "api"

	ArangoMLExtensionSpecDeploymentComponentDefaultPort = 8502
)

type ArangoMLExtensionSpecDeployment struct {
	// Replicas defines the number of replicas running specified components. No replicas created if no components are defined.
	// +doc/default: 1
	Replicas *int32 `json:"replicas,omitempty"`

	// Service defines how components will be exposed
	Service *ArangoMLExtensionSpecDeploymentService `json:"service,omitempty"`

	// TLS defined TLS Settings for extension
	TLS *sharedApi.TLS `json:"tls,omitempty"`

	// Pod defines base template for pods
	*schedulerPodApi.Pod

	// Container Keeps the information about Container configuration
	*schedulerContainerApi.Container `json:",inline"`

	// GPU defined if GPU Jobs are enabled.
	// +doc/default: false
	GPU *bool `json:"gpu,omitempty"`

	// Port defines on which port the container will be listening for connections
	Port *int32 `json:"port,omitempty"`
}

func (s *ArangoMLExtensionSpecDeployment) GetReplicas() int32 {
	if s == nil || s.Replicas == nil {
		return 1
	}
	return *s.Replicas
}

func (s *ArangoMLExtensionSpecDeployment) GetPodTemplate() *schedulerPodApi.Pod {
	if s == nil || s.Pod == nil {
		return nil
	}

	return s.Pod
}

func (s *ArangoMLExtensionSpecDeployment) GetGPU() bool {
	if s == nil || s.GPU == nil {
		return false
	}
	return *s.GPU
}

func (s *ArangoMLExtensionSpecDeployment) GetPort(def int32) int32 {
	if s == nil || s.Port == nil {
		return def
	}
	return *s.Port
}

func (s *ArangoMLExtensionSpecDeployment) GetContainer() *schedulerContainerApi.Container {
	if s == nil || s.Container == nil {
		return nil
	}

	return s.Container
}

func (s *ArangoMLExtensionSpecDeployment) GetService() *ArangoMLExtensionSpecDeploymentService {
	if s == nil {
		return nil
	}
	return s.Service
}

func (s *ArangoMLExtensionSpecDeployment) GetTLS() *sharedApi.TLS {
	if s == nil {
		return nil
	}
	return s.TLS
}

func (s *ArangoMLExtensionSpecDeployment) Validate() error {
	if s == nil {
		return nil
	}

	errs := []error{
		shared.PrefixResourceErrors("service", shared.ValidateOptional(s.GetService(), func(s ArangoMLExtensionSpecDeploymentService) error { return s.Validate() })),
		s.GetPodTemplate().Validate(),
		s.GetContainer().Validate(),
	}

	if s.GetReplicas() < 0 || s.GetReplicas() > 10 {
		errs = append(errs, shared.PrefixResourceErrors("replicas", errors.Errorf("out of range [0, 10]")))
	}
	return shared.WithErrors(errs...)
}
