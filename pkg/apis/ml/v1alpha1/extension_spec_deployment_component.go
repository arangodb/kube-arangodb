//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
)

type ArangoMLExtensionSpecDeploymentComponent struct {
	// Port defines on which port the container will be listening for connections
	Port *int32 `json:"port,omitempty"`

	// ContainerTemplate Keeps the information about Container configuration
	*sharedApi.ContainerTemplate `json:",inline"`
}

func (s *ArangoMLExtensionSpecDeploymentComponent) GetPort(def int32) int32 {
	if s == nil || s.Port == nil {
		return def
	}
	return *s.Port
}

func (s *ArangoMLExtensionSpecDeploymentComponent) GetContainerTemplate() *sharedApi.ContainerTemplate {
	if s == nil || s.ContainerTemplate == nil {
		return nil
	}

	return s.ContainerTemplate
}

func (s *ArangoMLExtensionSpecDeploymentComponent) Validate() error {
	if s == nil {
		return nil
	}

	var err []error

	err = append(err,
		s.GetContainerTemplate().Validate(),
	)

	return shared.WithErrors(err...)
}
