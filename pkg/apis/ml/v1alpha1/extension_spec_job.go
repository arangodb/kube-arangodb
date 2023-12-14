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

type ArangoMLJobsTemplates struct {
	// Prediction defines template for the prediction job
	Prediction map[string]*ArangoMLExtensionTemplateSpec `json:"prediction,omitempty"`

	// Training defines template for the training job
	Training map[string]*ArangoMLExtensionTemplateSpec `json:"training,omitempty"`
}

func (j *ArangoMLJobsTemplates) Validate() error {
	if j == nil {
		return nil
	}

	var errs []error
	for _, template := range j.Prediction {
		if err := template.Validate(); err != nil {
			errs = append(errs, shared.PrefixResourceErrors("prediction", err))
		}
	}

	for _, template := range j.Training {
		if err := template.Validate(); err != nil {
			errs = append(errs, shared.PrefixResourceErrors("training", err))
		}
	}

	return shared.WithErrors(errs...)
}

type ArangoMLExtensionTemplateSpec struct {
	// PodTemplate keeps the information about Pod configuration
	*sharedApi.PodTemplate `json:",inline"`

	// ContainerTemplate Keeps the information about Container configuration
	*sharedApi.ContainerTemplate `json:",inline"`
}

func (a *ArangoMLExtensionTemplateSpec) GetPodTemplate() *sharedApi.PodTemplate {
	if a == nil {
		return nil
	}

	return a.PodTemplate
}

func (a *ArangoMLExtensionTemplateSpec) GetContainerTemplate() *sharedApi.ContainerTemplate {
	if a == nil {
		return nil
	}

	return a.ContainerTemplate
}

func (a *ArangoMLExtensionTemplateSpec) Validate() error {
	if a == nil {
		return nil
	}

	return shared.WithErrors(
		a.GetPodTemplate().Validate(),
		a.GetContainerTemplate().Validate(),
	)
}
