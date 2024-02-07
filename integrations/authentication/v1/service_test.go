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

package v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Handler(t *testing.T, ctx context.Context, mods ...Mod) svc.Handler {
	handler, err := New(ctx, NewConfiguration().With(mods...))
	require.NoError(t, err)

	return handler
}

func Client(t *testing.T, ctx context.Context, mods ...Mod) (pbAuthenticationV1.AuthenticationV1Client, string) {
	directory := t.TempDir()

	var currentMods []Mod

	currentMods = append(currentMods, func(c Configuration) Configuration {
		c.Path = directory
		return c
	})

	currentMods = append(currentMods, mods...)

	local := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
	}, Handler(t, ctx, currentMods...))

	start := local.Start(ctx)

	client := tgrpc.NewGRPCClient(t, ctx, pbAuthenticationV1.NewAuthenticationV1Client, start.Address())

	return client, directory
}

func Test_Service(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client, directory := Client(t, ctx)

	reSaveJWTTokens(t, directory, generateJWTToken())

	token, err := client.CreateToken(ctx, &pbAuthenticationV1.CreateTokenRequest{})
	require.NoError(t, err)

	t.Logf("Token generated for user %s: %s", token.User, token.Token)

	require.EqualValues(t, "root", token.User)

	valid, err := client.Validate(ctx, &pbAuthenticationV1.ValidateRequest{
		Token: token.Token,
	})
	require.NoError(t, err)

	require.True(t, valid.IsValid)
	require.NotNil(t, valid.Details)
	require.EqualValues(t, token.User, valid.Details.User)
}

func Test_Service_DifferentDefaultUser(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client, directory := Client(t, ctx, func(c Configuration) Configuration {
		c.Create.DefaultUser = "different"
		return c
	})

	reSaveJWTTokens(t, directory, generateJWTToken())

	token, err := client.CreateToken(ctx, &pbAuthenticationV1.CreateTokenRequest{})
	require.NoError(t, err)

	t.Logf("Token generated for user %s: %s", token.User, token.Token)

	require.EqualValues(t, "different", token.User)

	valid, err := client.Validate(ctx, &pbAuthenticationV1.ValidateRequest{
		Token: token.Token,
	})
	require.NoError(t, err)

	require.True(t, valid.IsValid)
	require.NotNil(t, valid.Details)
	require.EqualValues(t, token.User, valid.Details.User)
}

func Test_Service_AskForDefaultIfAllowed(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client, directory := Client(t, ctx, func(c Configuration) Configuration {
		c.Create.AllowedUsers = []string{"root"}
		return c
	})

	reSaveJWTTokens(t, directory, generateJWTToken())

	token, err := client.CreateToken(ctx, &pbAuthenticationV1.CreateTokenRequest{})
	require.NoError(t, err)

	t.Logf("Token generated for user %s: %s", token.User, token.Token)

	require.EqualValues(t, "root", token.User)

	valid, err := client.Validate(ctx, &pbAuthenticationV1.ValidateRequest{
		Token: token.Token,
	})
	require.NoError(t, err)

	require.True(t, valid.IsValid)
	require.NotNil(t, valid.Details)
	require.EqualValues(t, token.User, valid.Details.User)
}

func Test_Service_AskForNonDefaultIfAllowed(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client, directory := Client(t, ctx, func(c Configuration) Configuration {
		c.Create.AllowedUsers = []string{"root", "other"}
		return c
	})

	reSaveJWTTokens(t, directory, generateJWTToken())

	token, err := client.CreateToken(ctx, &pbAuthenticationV1.CreateTokenRequest{
		User: util.NewType("other"),
	})
	require.NoError(t, err)

	t.Logf("Token generated for user %s: %s", token.User, token.Token)

	require.EqualValues(t, "other", token.User)

	valid, err := client.Validate(ctx, &pbAuthenticationV1.ValidateRequest{
		Token: token.Token,
	})
	require.NoError(t, err)

	require.True(t, valid.IsValid)
	require.NotNil(t, valid.Details)
	require.EqualValues(t, token.User, valid.Details.User)
}

func Test_Service_AskForDefaultIfBlocked(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client, directory := Client(t, ctx, func(c Configuration) Configuration {
		c.Create.AllowedUsers = []string{"root"}
		return c
	})

	reSaveJWTTokens(t, directory, generateJWTToken())

	_, err := client.CreateToken(ctx, &pbAuthenticationV1.CreateTokenRequest{
		User: util.NewType("blocked"),
	})
	require.EqualError(t, err, "rpc error: code = Unknown desc = User blocked is not allowed")
}
