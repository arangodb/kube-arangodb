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

package v1alpha1

import (
	"fmt"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type ArangoRouteStatusTarget struct {
	// Destinations keeps target destinations
	Destinations ArangoRouteStatusTargetDestinations `json:"destinations,omitempty"`

	// Type define destination type
	Type ArangoRouteStatusTargetType `json:"type,omitempty"`

	// TLS Keeps target TLS Settings (if not nil, TLS is enabled)
	TLS *ArangoRouteStatusTargetTLS `json:"tls,omitempty"`

	// Protocol defines http protocol used for the route
	Protocol ArangoRouteDestinationProtocol `json:"protocol,omitempty"`

	// Authentication specifies the authentication details
	Authentication ArangoRouteStatusTargetAuthentication `json:"authentication,omitempty"`

	// Options defines connection upgrade options
	Options *ArangoRouteStatusTargetOptions `json:"options,omitempty"`

	// Path specifies request path override
	Path string `json:"path,omitempty"`

	// Route defines the route status
	Route ArangoRouteStatusTargetRoute `json:"route,omitempty"`

	// Timeout specify the upstream request timeout
	Timeout meta.Duration `json:"timeout,omitempty"`
}

func (a *ArangoRouteStatusTarget) RenderURLs() []string {
	if a == nil {
		return nil
	}

	var urls = make([]string, len(a.Destinations))

	proto := "http"

	if a.TLS != nil {
		proto = "https"
	}

	for id, dest := range a.Destinations {
		urls[id] = fmt.Sprintf("%s://%s:%d%s", proto, dest.Host, dest.Port, a.Path)
	}

	return urls
}

func (a *ArangoRouteStatusTarget) Hash() string {
	if a == nil {
		return ""
	}
	return util.SHA256FromStringArray(a.Destinations.Hash(), a.Type.Hash(), a.TLS.Hash(), a.Protocol.String(), a.Path, a.Authentication.Hash(), a.Options.Hash(), a.Timeout.String(), a.Route.Hash())
}
