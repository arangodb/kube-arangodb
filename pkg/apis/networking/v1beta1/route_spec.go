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

import shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"

type ArangoRouteSpec struct {
	// Deployment specifies the ArangoDeployment object name
	// +doc/required
	Deployment *string `json:"deployment,omitempty"`

	// Destination defines the route destination
	// +doc/required
	Destination *ArangoRouteSpecDestination `json:"destination,omitempty"`

	// Route defines the route spec
	Route *ArangoRouteSpecRoute `json:"route,omitempty"`

	// Options defines connection upgrade options
	Options *ArangoRouteSpecOptions `json:"options,omitempty"`
}

func (s *ArangoRouteSpec) GetDeployment() string {
	if s == nil || s.Deployment == nil {
		return ""
	}

	return *s.Deployment
}

func (s *ArangoRouteSpec) GetDestination() *ArangoRouteSpecDestination {
	if s == nil || s.Destination == nil {
		return nil
	}

	return s.Destination
}

func (s *ArangoRouteSpec) GetRoute() *ArangoRouteSpecRoute {
	if s == nil || s.Route == nil {
		return nil
	}
	return s.Route
}

func (s *ArangoRouteSpec) Validate() error {
	if s == nil {
		s = &ArangoRouteSpec{}
	}

	if err := shared.WithErrors(shared.PrefixResourceErrors("spec",
		shared.PrefixResourceErrors("deployment", shared.ValidateResourceNamePointer(s.Deployment)),
		shared.ValidateRequiredInterfacePath("destination", s.Destination),
		shared.ValidateOptionalInterfacePath("route", s.Route),
		shared.ValidateOptionalInterfacePath("options", s.Options),
	)); err != nil {
		return err
	}

	return nil
}
