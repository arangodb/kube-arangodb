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

package auth_custom

import (
	"context"

	"github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/impl/auth_custom/openid"
	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
)

func New(ctx context.Context, configuration pbImplEnvoyAuthV3Shared.Configuration) (pbImplEnvoyAuthV3Shared.AuthHandler, bool) {
	if !configuration.Enabled {
		return nil, false
	}

	if !configuration.Auth.Enabled {
		return nil, false
	}

	switch configuration.Auth.Type {
	case "OpenID":
		return openid.New(ctx, configuration)
	}

	return nil, false
}
