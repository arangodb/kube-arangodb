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

package shared

import (
	"fmt"
	"sort"

	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

type Response struct {
	User *ResponseAuth

	Headers, ResponseHeaders []*pbEnvoyCoreV3.HeaderValueOption
}

type ResponseAuth struct {
	User  string
	Roles []string

	Token *string
}

func (a *ResponseAuth) Hash() string {
	if a == nil {
		return ""
	}

	sort.Strings(a.Roles)

	return util.SHA256FromString(fmt.Sprintf("%s:%s", a.User, util.SHA256FromString(strings.Join(a.Roles, ":"))))
}

func (a Response) GetHeaders() []*pbEnvoyCoreV3.HeaderValueOption {
	return a.Headers
}

func (a Response) Authenticated() bool {
	return a.User != nil
}

func (a Response) AsResponse() *pbEnvoyAuthV3.CheckResponse {
	return &pbEnvoyAuthV3.CheckResponse{
		HttpResponse: &pbEnvoyAuthV3.CheckResponse_OkResponse{
			OkResponse: &pbEnvoyAuthV3.OkHttpResponse{
				Headers:              a.Headers,
				ResponseHeadersToAdd: a.ResponseHeaders,
			},
		},
	}
}
