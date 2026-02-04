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

package v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

func Test_Service_Token(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	directory, authn := NewAuthenticationHandler(t)

	secret1 := tests.GenerateJWTToken()

	directory.Set(t, secret1)

	p := newPluginTest()

	client, _ := Client(t, ctx, Handler(p), authn)

	p.Set(t, &pbAuthorizationV1.AuthorizationV1PermissionRequest{
		User:     "admin",
		Roles:    []string{"x"},
		Action:   "test:Get",
		Resource: "test",
	}, &pbAuthorizationV1.AuthorizationV1PermissionResponse{
		Message: "Marked",
		Effect:  pbAuthorizationV1.AuthorizationV1Effect_Deny,
	})

	t.Run("Without Roles", func(t *testing.T) {
		token := directory.Sign(t, utilToken.NewClaims().With(utilToken.WithDefaultClaims(), utilToken.WithUsername("admin")))

		resp, err := client.EvaluateToken(ctx, &pbAuthorizationV1.AuthorizationV1PermissionTokenRequest{
			Token:    token,
			Action:   "test:Get",
			Resource: "test",
		})
		require.NoError(t, err)
		require.EqualValues(t, pbAuthorizationV1.AuthorizationV1Effect_Allow, resp.GetEffect())
	})

	t.Run("With Roles", func(t *testing.T) {
		token := directory.Sign(t, utilToken.NewClaims().With(utilToken.WithDefaultClaims(), utilToken.WithUsername("admin"), utilToken.WithRoles("x")))

		resp, err := client.EvaluateToken(ctx, &pbAuthorizationV1.AuthorizationV1PermissionTokenRequest{
			Token:    token,
			Action:   "test:Get",
			Resource: "test",
		})
		require.NoError(t, err)
		require.EqualValues(t, pbAuthorizationV1.AuthorizationV1Effect_Deny, resp.GetEffect())
	})
}
