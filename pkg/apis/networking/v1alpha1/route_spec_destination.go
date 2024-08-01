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

type ArangoRouteSpecDestination struct {
	// Service defines service upstream reference
	Service *ArangoRouteSpecDestinationService `json:"service,omitempty"`

	// Schema defines HTTP/S schema used for connection
	Schema *ArangoRouteSpecDestinationSchema `json:"schema,omitempty"`

	// TLS defines TLS Configuration
	TLS *ArangoRouteSpecDestinationTLS `json:"tls,omitempty"`
}

func (s *ArangoRouteSpecDestination) GetService() *ArangoRouteSpecDestinationService {
	if s == nil || s.Service == nil {
		return nil
	}

	return s.Service
}

func (s *ArangoRouteSpecDestination) GetSchema() *ArangoRouteSpecDestinationSchema {
	if s == nil || s.Schema == nil {
		return nil
	}

	return s.Schema
}

func (s *ArangoRouteSpecDestination) GetTLS() *ArangoRouteSpecDestinationTLS {
	if s == nil || s.TLS == nil {
		return nil
	}

	return s.TLS
}

func (a *ArangoRouteSpecDestination) Validate() error {
	if a == nil {
		a = &ArangoRouteSpecDestination{}
	}

	if err := shared.WithErrors(
		shared.ValidateOptionalInterfacePath("service", a.Service),
		shared.ValidateOptionalInterfacePath("schema", a.Schema),
		shared.ValidateOptionalInterfacePath("tls", a.TLS),
	); err != nil {
		return err
	}

	return nil
}
