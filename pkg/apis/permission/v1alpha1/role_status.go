//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
)

type ArangoPermissionRoleStatus struct {
	// Conditions specific to the entire token
	// +doc/type: api.Conditions
	Conditions sharedApi.ConditionList `json:"conditions,omitempty"`

	// Deployment keeps the Deployment Reference
	Deployment *sharedApi.Object `json:"deployment,omitempty"`

	// Role keeps the Role Reference
	Role *sharedApi.Object `json:"role,omitempty"`

	// Policy keeps the Policy Reference
	Policy *sharedApi.Object `json:"policy,omitempty"`
}
