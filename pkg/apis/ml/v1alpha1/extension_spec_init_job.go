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
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
)

type ArangoMLExtensionSpecInit struct {
	// PodTemplate keeps the information about Pod configuration
	*sharedApi.PodTemplate `json:",inline"`

	// ContainerTemplate Keeps the information about Container configuration
	*sharedApi.ContainerTemplate `json:",inline"`
}

func (a *ArangoMLExtensionSpecInit) GetPodTemplate() *sharedApi.PodTemplate {
	if a == nil {
		return nil
	}

	return a.PodTemplate
}

func (a *ArangoMLExtensionSpecInit) GetContainerTemplate() *sharedApi.ContainerTemplate {
	if a == nil {
		return nil
	}

	return a.ContainerTemplate
}

func (a *ArangoMLExtensionSpecInit) Validate() error {
	if a == nil {
		return nil
	}

	return shared.WithErrors(
		a.GetPodTemplate().Validate(),
		a.GetContainerTemplate().Validate(),
	)
}
