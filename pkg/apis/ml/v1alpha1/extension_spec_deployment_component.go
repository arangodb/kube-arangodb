//
// DISCLAIMER
//
// Copyright 2023-2024 ArangoDB GmbH, Cologne, Germany
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

package v1alpha1

import (
	schedulerContainerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/container"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

type ArangoMLExtensionSpecDeploymentComponent struct {
	// GPU defined if GPU Jobs are enabled for component. In use only for ArangoMLExtensionSpecDeploymentComponentPrediction and ArangoMLExtensionSpecDeploymentComponentTraining
	// +doc/default: false
	GPU *bool `json:"gpu,omitempty"`

	// Port defines on which port the container will be listening for connections
	Port *int32 `json:"port,omitempty"`

	// Container Keeps the information about Container configuration
	*schedulerContainerApi.Container `json:",inline"`
}

func (s *ArangoMLExtensionSpecDeploymentComponent) GetGPU() bool {
	if s == nil || s.GPU == nil {
		return false
	}
	return *s.GPU
}

func (s *ArangoMLExtensionSpecDeploymentComponent) GetPort(def int32) int32 {
	if s == nil || s.Port == nil {
		return def
	}
	return *s.Port
}

func (s *ArangoMLExtensionSpecDeploymentComponent) GetContainer() *schedulerContainerApi.Container {
	if s == nil || s.Container == nil {
		return nil
	}

	return s.Container
}

func (s *ArangoMLExtensionSpecDeploymentComponent) Validate() error {
	if s == nil {
		return nil
	}

	var err []error

	err = append(err,
		s.GetContainer().Validate(),
	)

	return shared.WithErrors(err...)
}
