//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
)

type ArangoRouteStatus struct {
	// Conditions specific to the entire extension
	// +doc/type: api.Conditions
	Conditions api.ConditionList `json:"conditions,omitempty"`

	// Deployment keeps the ArangoDeployment reference
	Deployment *sharedApi.Object `json:"deployment,omitempty"`

	// Target keeps the target details
	Target *ArangoRouteStatusTarget `json:"target,omitempty"`
}
