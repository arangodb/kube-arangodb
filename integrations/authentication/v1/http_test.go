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

package v1

import (
	"context"
	"fmt"
	goHttp "net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
)

func Test_Authentication_HTTP(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	directory, server := Server(t, ctx)

	token1 := generateJWTToken()

	reSaveJWTTokens(t, directory, token1)

	client := operatorHTTP.NewHTTPClient()

	t.Run("Without header", func(t *testing.T) {
		resp := ugrpc.Get[*pbAuthenticationV1.IdentityResponse](ctx, client, fmt.Sprintf("http://%s/_integration/authn/v1/identity", server.HTTPAddress()))

		resp.WithCode(goHttp.StatusUnauthorized)
	})

	t.Run("With invalid header", func(t *testing.T) {
		resp := ugrpc.Get[*pbAuthenticationV1.IdentityResponse](ctx, client, fmt.Sprintf("http://%s/_integration/authn/v1/identity", server.HTTPAddress()), func(in *goHttp.Request) {
			in.Header.Add("invalid", "")
		})

		resp.WithCode(goHttp.StatusUnauthorized)
	})

	t.Run("With empty header", func(t *testing.T) {
		resp := ugrpc.Get[*pbAuthenticationV1.IdentityResponse](ctx, client, fmt.Sprintf("http://%s/_integration/authn/v1/identity", server.HTTPAddress()), func(in *goHttp.Request) {
			in.Header.Add("Authorization", "")
		})

		resp.WithCode(goHttp.StatusUnauthorized)
	})

	t.Run("With missing prefix header", func(t *testing.T) {
		token, err := ugrpc.Post[*pbAuthenticationV1.CreateTokenRequest, *pbAuthenticationV1.CreateTokenResponse](
			ctx,
			client,
			&pbAuthenticationV1.CreateTokenRequest{
				Lifetime: durationpb.New(time.Minute),
				User:     util.NewType(DefaultUser),
			},
			fmt.Sprintf("http://%s/_integration/authn/v1/createToken", server.HTTPAddress()),
		).
			WithCode(goHttp.StatusOK).
			Get()
		require.NoError(t, err)

		resp := ugrpc.Get[*pbAuthenticationV1.IdentityResponse](ctx, client, fmt.Sprintf("http://%s/_integration/authn/v1/identity", server.HTTPAddress()), func(in *goHttp.Request) {
			in.Header.Add("Authorization", token.Token)
		})

		resp.WithCode(goHttp.StatusUnauthorized)
	})

	t.Run("With header", func(t *testing.T) {
		token, err := ugrpc.Post[*pbAuthenticationV1.CreateTokenRequest, *pbAuthenticationV1.CreateTokenResponse](
			ctx,
			client,
			&pbAuthenticationV1.CreateTokenRequest{
				Lifetime: durationpb.New(time.Minute),
				User:     util.NewType(DefaultUser),
			},
			fmt.Sprintf("http://%s/_integration/authn/v1/createToken", server.HTTPAddress()),
		).
			WithCode(goHttp.StatusOK).
			Get()
		require.NoError(t, err)

		resp := ugrpc.Get[*pbAuthenticationV1.IdentityResponse](ctx, client, fmt.Sprintf("http://%s/_integration/authn/v1/identity", server.HTTPAddress()), func(in *goHttp.Request) {
			in.Header.Add("Authorization", fmt.Sprintf("bearer %s", token.Token))
		})

		data, err := resp.WithCode(goHttp.StatusOK).Get()
		require.NoError(t, err)

		require.EqualValues(t, DefaultUser, data.GetUser())
	})

	t.Run("With multi header", func(t *testing.T) {
		token, err := ugrpc.Post[*pbAuthenticationV1.CreateTokenRequest, *pbAuthenticationV1.CreateTokenResponse](
			ctx,
			client,
			&pbAuthenticationV1.CreateTokenRequest{
				Lifetime: durationpb.New(time.Minute),
				User:     util.NewType(DefaultUser),
			},
			fmt.Sprintf("http://%s/_integration/authn/v1/createToken", server.HTTPAddress()),
		).
			WithCode(goHttp.StatusOK).
			Get()
		require.NoError(t, err)

		resp := ugrpc.Get[*pbAuthenticationV1.IdentityResponse](ctx, client, fmt.Sprintf("http://%s/_integration/authn/v1/identity", server.HTTPAddress()), func(in *goHttp.Request) {
			in.Header.Add("Authorization", fmt.Sprintf("bearer %s", token.Token))
			in.Header.Add("Authorization", fmt.Sprintf("bearer %s", token.Token))
		})

		resp.WithCode(goHttp.StatusUnauthorized)
	})

	t.Run("Validate", func(t *testing.T) {
		token, err := ugrpc.Post[*pbAuthenticationV1.CreateTokenRequest, *pbAuthenticationV1.CreateTokenResponse](
			ctx,
			client,
			&pbAuthenticationV1.CreateTokenRequest{
				Lifetime: durationpb.New(time.Minute),
				User:     util.NewType(DefaultUser),
			},
			fmt.Sprintf("http://%s/_integration/authn/v1/createToken", server.HTTPAddress()),
		).
			WithCode(goHttp.StatusOK).
			Get()
		require.NoError(t, err)

		validate, err := ugrpc.Post[*pbAuthenticationV1.ValidateRequest, *pbAuthenticationV1.ValidateResponse](
			ctx,
			client,
			&pbAuthenticationV1.ValidateRequest{
				Token: token.Token,
			},
			fmt.Sprintf("http://%s/_integration/authn/v1/validate", server.HTTPAddress()),
		).
			WithCode(goHttp.StatusOK).
			Get()
		require.NoError(t, err)

		require.True(t, validate.GetIsValid())
	})
}
