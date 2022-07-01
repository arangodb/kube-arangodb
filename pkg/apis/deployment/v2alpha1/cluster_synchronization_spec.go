//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package v2alpha1

import (
	"github.com/pkg/errors"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

type ArangoClusterSynchronizationSpec struct {
	DeploymentName string                                      `json:"deploymentName,omitempty"`
	KubeConfig     *ArangoClusterSynchronizationKubeConfigSpec `json:"kubeconfig,omitempty"`
}

type ArangoClusterSynchronizationKubeConfigSpec struct {
	SecretName string `json:"secretName"`
	SecretKey  string `json:"secretKey"`
	Namespace  string `json:"namespace"`
}

func (a *ArangoClusterSynchronizationKubeConfigSpec) Validate() error {
	if a == nil {
		return errors.Errorf("KubeConfig Spec cannot be nil")
	}

	return shared.WithErrors(
		shared.PrefixResourceError("secretName", shared.ValidateResourceName(a.SecretName)),
		shared.PrefixResourceError("secretKey", shared.ValidateResourceName(a.SecretKey)),
		shared.PrefixResourceError("namespace", shared.ValidateResourceName(a.Namespace)),
	)
}
