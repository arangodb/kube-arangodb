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

package shared

import (
	"context"

	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/logging"
)

func Permissive(parent Plugin, log logging.Logger) Plugin {
	return permissive{
		log:    log,
		parent: parent,
	}
}

type permissive struct {
	log    logging.Logger
	parent Plugin
}

func (p permissive) Ready(ctx context.Context) error {
	return p.parent.Ready(ctx)
}

func (p permissive) Revision() uint64 {
	return p.parent.Revision()
}

func (p permissive) Evaluate(ctx context.Context, req *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
	resp, err := p.parent.Evaluate(ctx, req)
	if err != nil {
		return nil, err
	}

	log := p.log.
		Str("user", req.GetUser()).
		Str("action", req.GetAction()).
		Str("resource", req.GetResource()).
		Strs("roles", req.GetRoles()...).
		JSON("context", req.GetContext()).
		Str("message", resp.GetMessage())

	switch resp.GetEffect() {
	case pbAuthorizationV1.AuthorizationV1Effect_Allow:
		log.Info("Access Granted")
		return &pbAuthorizationV1.AuthorizationV1PermissionResponse{
			Message: "Access granted",
			Effect:  pbAuthorizationV1.AuthorizationV1Effect_Allow,
		}, nil
	case pbAuthorizationV1.AuthorizationV1Effect_Deny:
		log.Info("Access Denied")
		return &pbAuthorizationV1.AuthorizationV1PermissionResponse{
			Message: "Access granted due to the Permissive mode",
			Effect:  pbAuthorizationV1.AuthorizationV1Effect_Allow,
		}, nil
	default:
		log.Info("Unknown Effect")
		return &pbAuthorizationV1.AuthorizationV1PermissionResponse{
			Message: "Access granted due to the Permissive mode",
			Effect:  pbAuthorizationV1.AuthorizationV1Effect_Allow,
		}, nil
	}

}
