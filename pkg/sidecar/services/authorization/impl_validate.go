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

package authorization

import (
	"context"

	sidecarSvcAuthzDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

func (a *implementation) ValidateSelfPermission(ctx context.Context, request *sidecarSvcAuthzDefinition.AuthorizationAPIValidateSelfRequest) (*sidecarSvcAuthzDefinition.AuthorizationAPIValidateResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.Plugin(), request.GetAction(), request.GetResource()); err != nil {
		return &sidecarSvcAuthzDefinition.AuthorizationAPIValidateResponse{
			Message: err.Error(),
			Effect:  sidecarSvcAuthzTypes.Effect_Deny,
		}, nil
	}

	return &sidecarSvcAuthzDefinition.AuthorizationAPIValidateResponse{
		Effect: sidecarSvcAuthzTypes.Effect_Allow,
	}, nil
}
