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

package authenticator

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

type Identity struct {
	User  *string
	Roles []string
}

func (i *Identity) EvaluatePermission(ctx context.Context, client cache.Object[pbAuthorizationV1.AuthorizationV1Client], action, resource string) error {
	if i == nil {
		return status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	c, err := client.Get(ctx)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	resp, err := c.Evaluate(ctx, &pbAuthorizationV1.AuthorizationV1PermissionRequest{
		User:     i.User,
		Roles:    i.Roles,
		Action:   action,
		Resource: resource,
	})
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if resp.GetEffect() == pbAuthorizationV1.AuthorizationV1Effect_Allow {
		return nil
	}

	return status.Error(codes.PermissionDenied, resp.GetMessage())
}

func GetIdentity(ctx context.Context) *Identity {
	v := ctx.Value(identityContextKey)
	if v == nil {
		return nil
	}

	z, ok := v.(*Identity)
	if !ok {
		return nil
	}

	return z
}
