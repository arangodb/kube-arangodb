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
)

func Test_Service_Many(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	p := newPluginTest()

	p.Set(t, &pbAuthorizationV1.AuthorizationV1PermissionRequest{
		User:     "admin",
		Action:   "test:Get",
		Resource: "deny",
	}, &pbAuthorizationV1.AuthorizationV1PermissionResponse{
		Message: "Marked",
		Effect:  pbAuthorizationV1.AuthorizationV1Effect_Deny,
	})

	client, _ := Client(t, ctx, Handler(p))

	t.Run("Empty", func(t *testing.T) {
		resp, err := client.EvaluateMany(ctx, &pbAuthorizationV1.AuthorizationV1PermissionManyRequest{
			User: "admin",
		})
		require.NoError(t, err)
		require.EqualValues(t, pbAuthorizationV1.AuthorizationV1Effect_Deny, resp.GetEffect())
		require.Len(t, resp.GetItems(), 0)
	})

	t.Run("Single Allow", func(t *testing.T) {
		resp, err := client.EvaluateMany(ctx, &pbAuthorizationV1.AuthorizationV1PermissionManyRequest{
			User: "admin",
			Items: []*pbAuthorizationV1.AuthorizationV1PermissionManyRequestItem{
				{
					Action:   "test:Get",
					Resource: "test",
				},
			},
		})
		require.NoError(t, err)
		require.EqualValues(t, pbAuthorizationV1.AuthorizationV1Effect_Allow, resp.GetEffect())
		require.Len(t, resp.GetItems(), 1)
		require.EqualValues(t, pbAuthorizationV1.AuthorizationV1Effect_Allow, resp.GetItems()[0].GetEffect())
	})

	t.Run("Multi Allow", func(t *testing.T) {
		resp, err := client.EvaluateMany(ctx, &pbAuthorizationV1.AuthorizationV1PermissionManyRequest{
			User: "admin",
			Items: []*pbAuthorizationV1.AuthorizationV1PermissionManyRequestItem{
				{
					Action:   "test:Get",
					Resource: "test",
				},
				{
					Action:   "test:Get",
					Resource: "test2",
				},
			},
		})
		require.NoError(t, err)
		require.EqualValues(t, pbAuthorizationV1.AuthorizationV1Effect_Allow, resp.GetEffect())
		require.Len(t, resp.GetItems(), 2)
		require.EqualValues(t, pbAuthorizationV1.AuthorizationV1Effect_Allow, resp.GetItems()[0].GetEffect())
		require.EqualValues(t, pbAuthorizationV1.AuthorizationV1Effect_Allow, resp.GetItems()[1].GetEffect())
	})

	t.Run("Multi Allow with deny", func(t *testing.T) {
		resp, err := client.EvaluateMany(ctx, &pbAuthorizationV1.AuthorizationV1PermissionManyRequest{
			User: "admin",
			Items: []*pbAuthorizationV1.AuthorizationV1PermissionManyRequestItem{
				{
					Action:   "test:Get",
					Resource: "test",
				},
				{
					Action:   "test:Get",
					Resource: "test2",
				},
				{
					Action:   "test:Get",
					Resource: "deny",
				},
			},
		})
		require.NoError(t, err)
		require.EqualValues(t, pbAuthorizationV1.AuthorizationV1Effect_Deny, resp.GetEffect())
		require.Len(t, resp.GetItems(), 3)
		require.EqualValues(t, pbAuthorizationV1.AuthorizationV1Effect_Allow, resp.GetItems()[0].GetEffect())
		require.EqualValues(t, pbAuthorizationV1.AuthorizationV1Effect_Allow, resp.GetItems()[1].GetEffect())
		require.EqualValues(t, pbAuthorizationV1.AuthorizationV1Effect_Deny, resp.GetItems()[2].GetEffect())
	})
}
