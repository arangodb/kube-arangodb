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
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
)

type JobType string

const (
	MLJobTrainingType      JobType = "training"
	MLJobPredictionType    JobType = "prediction"
	MLJobFeaturizationType JobType = "featurization"
)

type ArangoMLJobsTemplates struct {
	// Prediction defines template for the prediction job
	Prediction *ArangoMLJobTemplates `json:"prediction,omitempty"`

	// Training defines template for the training job
	Training *ArangoMLJobTemplates `json:"training,omitempty"`

	// Featurization defines template for the featurization job
	Featurization *ArangoMLJobTemplates `json:"featurization,omitempty"`
}

func (a *ArangoMLJobsTemplates) GetJobTemplates(jobType JobType) *ArangoMLJobTemplates {
	if a == nil {
		return nil
	}

	switch jobType {
	case MLJobTrainingType:
		return a.Training
	case MLJobPredictionType:
		return a.Prediction
	case MLJobFeaturizationType:
		return a.Featurization
	default:
		return nil
	}
}

func (a *ArangoMLJobsTemplates) Validate() error {
	if a == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceErrors("prediction", a.Prediction.Validate()),
		shared.PrefixResourceErrors("training", a.Training.Validate()),
		shared.PrefixResourceErrors("featurization", a.Featurization.Validate()),
	)
}

type JobScheduleType string

const (
	MLJobScheduleCPU JobScheduleType = "cpu"
	MLJobScheduleGPU JobScheduleType = "gpu"
)

type ArangoMLJobTemplates struct {
	// CPU defines templates for CPU jobs
	CPU *ArangoMLExtensionTemplate `json:"cpu,omitempty"`

	// GPU defines templates for GPU jobs
	GPU *ArangoMLExtensionTemplate `json:"gpu,omitempty"`
}

func (a *ArangoMLJobTemplates) GetJobTemplateSpec(scheduleType JobScheduleType) *ArangoMLExtensionTemplate {
	if a == nil {
		return nil
	}

	switch scheduleType {
	case MLJobScheduleCPU:
		return a.CPU
	case MLJobScheduleGPU:
		return a.GPU
	default:
		return nil
	}
}

func (a *ArangoMLJobTemplates) Validate() error {
	if a == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceErrors("cpu", a.CPU.Validate()),
		shared.PrefixResourceErrors("gpu", a.GPU.Validate()),
	)
}

type ArangoMLExtensionTemplate struct {
	// PodTemplate keeps the information about Pod configuration
	*sharedApi.PodTemplate `json:",inline"`

	// ContainerTemplate Keeps the information about Container configuration
	*sharedApi.ContainerTemplate `json:",inline"`
}

func (a *ArangoMLExtensionTemplate) GetPodTemplate() *sharedApi.PodTemplate {
	if a == nil {
		return nil
	}

	return a.PodTemplate
}

func (a *ArangoMLExtensionTemplate) GetContainerTemplate() *sharedApi.ContainerTemplate {
	if a == nil {
		return nil
	}

	return a.ContainerTemplate
}

func (a *ArangoMLExtensionTemplate) Validate() error {
	if a == nil {
		return nil
	}

	return shared.WithErrors(
		a.GetPodTemplate().Validate(),
		a.GetContainerTemplate().Validate(),
	)
}
