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

package v1

type ServiceAccount struct {
	// Object keeps the reference to the ServiceAccount
	*Object `json:",inline"`

	// Namespaced keeps the reference to core.Role objects
	Namespaced *ServiceAccountRole `json:"namespaced,omitempty"`

	// Cluster keeps the reference to core.ClusterRole objects
	Cluster *ServiceAccountRole `json:"cluster,omitempty"`
}

type ServiceAccountRole struct {
	// Role keeps the reference to the Kubernetes Role
	Role *Object `json:"role,omitempty"`

	// Binding keeps the reference to the Kubernetes Binding
	Binding *Object `json:"binding,omitempty"`
}
