//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package auth_required

import (
	"context"
	goHttp "net/http"

	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"

	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func New(configuration pbImplEnvoyAuthV3Shared.Configuration) (pbImplEnvoyAuthV3Shared.AuthHandler, bool) {
	return impl{}, true
}

type impl struct {
}

func (a impl) Handle(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest, current *pbImplEnvoyAuthV3Shared.Response) error {
	ext := request.GetAttributes().GetContextExtensions()

	if util.Optional(ext, pbImplEnvoyAuthV3Shared.AuthConfigAuthRequiredKey, pbImplEnvoyAuthV3Shared.AuthConfigKeywordFalse) == pbImplEnvoyAuthV3Shared.AuthConfigKeywordTrue {
		if !current.Authenticated() {
			return pbImplEnvoyAuthV3Shared.DeniedResponse{
				Code: goHttp.StatusUnauthorized,
				Message: &pbImplEnvoyAuthV3Shared.DeniedMessage{
					Message: "Unauthorized",
				},
			}
		}
	}

	return nil
}
