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
	schedulerPodApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/pod"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	ArangoMLExtensionSpecDeploymentComponentPrediction = "prediction"
	ArangoMLExtensionSpecDeploymentComponentTraining   = "training"
	ArangoMLExtensionSpecDeploymentComponentProject    = "project"

	ArangoMLExtensionSpecDeploymentComponentPredictionDefaultPort = 8501
	ArangoMLExtensionSpecDeploymentComponentTrainingDefaultPort   = 8502
	ArangoMLExtensionSpecDeploymentComponentProjectDefaultPort    = 8503
)

func GetArangoMLExtensionSpecDeploymentComponentDefaultPort(component string) int32 {
	switch component {
	case ArangoMLExtensionSpecDeploymentComponentPrediction:
		return ArangoMLExtensionSpecDeploymentComponentPredictionDefaultPort
	case ArangoMLExtensionSpecDeploymentComponentTraining:
		return ArangoMLExtensionSpecDeploymentComponentTrainingDefaultPort
	case ArangoMLExtensionSpecDeploymentComponentProject:
		return ArangoMLExtensionSpecDeploymentComponentProjectDefaultPort
	}

	return 0
}

type ArangoMLExtensionSpecDeployment struct {
	// Replicas defines the number of replicas running specified components. No replicas created if no components are defined.
	// +doc/default: 1
	Replicas *int32 `json:"replicas,omitempty"`

	// Service defines how components will be exposed
	Service *ArangoMLExtensionSpecDeploymentService `json:"service,omitempty"`

	// Pod defines base template for pods
	*schedulerPodApi.Pod

	// Prediction defines how Prediction workload will be deployed
	Prediction *ArangoMLExtensionSpecDeploymentComponent `json:"prediction,omitempty"`
	// Training defines how Training workload will be deployed
	Training *ArangoMLExtensionSpecDeploymentComponent `json:"training,omitempty"`
	// Project defines how Project workload will be deployed
	Project *ArangoMLExtensionSpecDeploymentComponent `json:"project,omitempty"`
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

func (s *ArangoMLExtensionSpecDeployment) GetPrediction() *ArangoMLExtensionSpecDeploymentComponent {
	if s == nil {
		return nil
	}
	return s.Prediction
}

func (s *ArangoMLExtensionSpecDeployment) GetTraining() *ArangoMLExtensionSpecDeploymentComponent {
	if s == nil {
		return nil
	}
	return s.Training
}

func (s *ArangoMLExtensionSpecDeployment) GetProject() *ArangoMLExtensionSpecDeploymentComponent {
	if s == nil {
		return nil
	}
	return s.Project
}

func (s *ArangoMLExtensionSpecDeployment) GetComponents() map[string]*ArangoMLExtensionSpecDeploymentComponent {
	if s == nil {
		return nil
	}
	return map[string]*ArangoMLExtensionSpecDeploymentComponent{
		ArangoMLExtensionSpecDeploymentComponentPrediction: s.GetPrediction(),
		ArangoMLExtensionSpecDeploymentComponentTraining:   s.GetTraining(),
		ArangoMLExtensionSpecDeploymentComponentProject:    s.GetProject(),
	}
}

func (s *ArangoMLExtensionSpecDeployment) HasComponents() bool {
	if s == nil || len(s.GetComponents()) == 0 {
		return false
	}

	for _, c := range s.GetComponents() {
		if c != nil {
			return true
		}
	}
	return false
}

func (s *ArangoMLExtensionSpecDeployment) GetService() *ArangoMLExtensionSpecDeploymentService {
	if s == nil {
		return nil
	}
	return s.Service
}

func (s *ArangoMLExtensionSpecDeployment) Validate() error {
	if s == nil {
		return nil
	}

	errs := []error{
		shared.PrefixResourceErrors("service", shared.ValidateOptional(s.GetService(), func(s ArangoMLExtensionSpecDeploymentService) error { return s.Validate() })),
		s.GetPodTemplate().Validate(),
	}

	if s.GetReplicas() < 0 || s.GetReplicas() > 10 {
		errs = append(errs, shared.PrefixResourceErrors("replicas", errors.Errorf("out of range [0, 10]")))
	}

	var usedPorts util.List[int32]
	for prefix, component := range s.GetComponents() {
		err := component.Validate()
		if err != nil {
			errs = append(errs, shared.PrefixResourceErrors(prefix, err))
			continue
		}
		if err == nil {
			port := component.GetPort(GetArangoMLExtensionSpecDeploymentComponentDefaultPort(prefix))

			if port == 0 {
				errs = append(errs, shared.PrefixResourceErrors(prefix, errors.Errorf("port not defined")))
				continue
			}

			duplicateCount := usedPorts.Count(func(i int32) bool {
				return i == port
			})
			if duplicateCount > 0 {
				errs = append(errs, shared.PrefixResourceErrors(prefix, errors.Errorf("port %d already specified for other component", port)))
			} else {
				usedPorts.Append(port)
			}
		}
	}
	return shared.WithErrors(errs...)
}
