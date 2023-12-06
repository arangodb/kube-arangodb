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
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ArangoMLExtensionSpecDeployment struct {
	// Replicas defines the number of replicas running specified components. No replicas created if no components are defined.
	// +doc/default: 1
	Replicas *int32 `json:"replicas,omitempty"`

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
		"prediction": s.GetPrediction(),
		"training":   s.GetTraining(),
		"project":    s.GetProject(),
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

func (s *ArangoMLExtensionSpecDeployment) Validate() error {
	if s == nil {
		return nil
	}

	var errs []error

	if s.GetReplicas() < 0 || s.GetReplicas() > 10 {
		errs = append(errs, shared.PrefixResourceErrors("replicas", errors.Newf("out of range [0, 10]")))
	}

	var usedPorts util.List[int32]
	for prefix, component := range s.GetComponents() {
		err := component.Validate()
		errs = append(errs, shared.PrefixResourceErrors(prefix, err))
		if err == nil {
			if usedPorts.IndexOf(component.GetPort()) >= 0 {
				errs = append(errs, shared.PrefixResourceErrors(prefix, errors.Newf("port %d already specified for other component", component.GetPort())))
			} else {
				usedPorts.Append(component.GetPort())
			}
		}
	}
	return shared.WithErrors(errs...)
}
