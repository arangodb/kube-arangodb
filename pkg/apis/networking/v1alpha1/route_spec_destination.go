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

	// Endpoints defines service upstream reference - which is used to find endpoints
	Endpoints *ArangoRouteSpecDestinationEndpoints `json:"endpoints,omitempty"`

	// Schema defines HTTP/S schema used for connection
	// +doc/enum: http|HTTP Connection
	// +doc/enum: https|HTTPS Connection (HTTP with TLS)
	Schema *ArangoRouteSpecDestinationSchema `json:"schema,omitempty"`

	// Protocol defines http protocol used for the route
	// +doc/enum: http1|HTTP 1.1 Protocol
	// +doc/enum: http2|HTTP 2 Protocol
	Protocol *ArangoRouteDestinationProtocol `json:"protocol,omitempty"`

	// TLS defines TLS Configuration
	TLS *ArangoRouteSpecDestinationTLS `json:"tls,omitempty"`

	// Path defines service path used for overrides
	Path *string `json:"path,omitempty"`

	// Authentication defines auth methods
	Authentication *ArangoRouteSpecDestinationAuthentication `json:"authentication,omitempty"`
}

func (a *ArangoRouteSpecDestination) GetService() *ArangoRouteSpecDestinationService {
	if a == nil || a.Service == nil {
		return nil
	}

	return a.Service
}

func (a *ArangoRouteSpecDestination) GetEndpoints() *ArangoRouteSpecDestinationEndpoints {
	if a == nil || a.Endpoints == nil {
		return nil
	}

	return a.Endpoints
}

func (a *ArangoRouteSpecDestination) GetProtocol() *ArangoRouteDestinationProtocol {
	if a == nil || a.Schema == nil {
		return nil
	}

	return a.Protocol
}

func (a *ArangoRouteSpecDestination) GetSchema() *ArangoRouteSpecDestinationSchema {
	if a == nil || a.Schema == nil {
		return nil
	}

	return a.Schema
}

func (a *ArangoRouteSpecDestination) GetPath() string {
	if a == nil || a.Path == nil {
		return "/"
	}

	return *a.Path
}

func (a *ArangoRouteSpecDestination) GetTLS() *ArangoRouteSpecDestinationTLS {
	if a == nil || a.TLS == nil {
		return nil
	}

	return a.TLS
}

func (a *ArangoRouteSpecDestination) GetAuthentication() *ArangoRouteSpecDestinationAuthentication {
	if a == nil || a.Authentication == nil {
		return nil
	}

	return a.Authentication
}

func (a *ArangoRouteSpecDestination) Validate() error {
	if a == nil {
		a = &ArangoRouteSpecDestination{}
	}

	if err := shared.WithErrors(
		shared.ValidateExclusiveFields(a, 1, "Service", "Endpoints"),
		shared.ValidateOptionalInterfacePath("service", a.Service),
		shared.ValidateOptionalInterfacePath("endpoints", a.Endpoints),
		shared.ValidateOptionalInterfacePath("schema", a.Schema),
		shared.ValidateOptionalInterfacePath("protocol", a.Protocol),
		shared.ValidateOptionalInterfacePath("tls", a.TLS),
		shared.ValidateOptionalInterfacePath("authentication", a.Authentication),
		shared.PrefixResourceError("path", shared.ValidateAPIPath(a.GetPath())),
	); err != nil {
		return err
	}

	return nil
}
