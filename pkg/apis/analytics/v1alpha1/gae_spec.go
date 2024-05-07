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

package v1alpha1

import (
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type GraphAnalyticsEngineSpec struct {
	// DeploymentName define deployment name used in the object. Immutable
	DeploymentName *string `json:"deploymentName,omitempty"`
}

func (g *GraphAnalyticsEngineSpec) Validate() error {
	if g == nil {
		g = &GraphAnalyticsEngineSpec{}
	}

	return shared.WithErrors(shared.PrefixResourceErrors("spec",
		shared.PrefixResourceErrors("deploymentName", shared.ValidateRequired(g.DeploymentName, func(s string) error {
			if s == "" {
				return errors.Errorf("DeploymentName should be not empty")
			}

			return nil
		})),
	))
}
