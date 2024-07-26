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

import shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"

type ArangoRouteSpec struct {
	// Deployment specifies the ArangoDeployment object name
	Deployment string `json:"deployment,omitempty"`

	// Destination defines the route destination
	Destination *ArangoRouteSpecDestination `json:"destination,omitempty"`

	// Route defines the route spec
	Route *ArangoRouteSpecRoute `json:"route,omitempty"`
}

func (s *ArangoRouteSpec) Validate() error {
	if s == nil {
		s = &ArangoRouteSpec{}
	}

	if err := shared.WithErrors(shared.PrefixResourceErrors("spec",
		shared.PrefixResourceError("deployment", shared.ValidateResourceName(s.Deployment)),
		shared.ValidateRequiredInterfacePath("destination", s.Destination),
		shared.ValidateRequiredInterfacePath("route", s.Route),
	)); err != nil {
		return err
	}

	return nil
}
