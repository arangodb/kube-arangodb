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
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Test_Basic(t *testing.T) {
	directory := t.TempDir()

	reSaveJWTTokens(t, directory, generateJWTToken(), generateJWTToken(), generateJWTToken(), generateJWTToken(), generateJWTToken(), generateJWTToken())

	ctx, c := context.WithCancel(context.Background())
	defer c()

	s, err := newInternal(ctx, Configuration{
		Path: directory,
		TTL:  time.Duration(0),
	})
	require.NoError(t, err)

	// Create token
	tokenResponse, err := s.CreateToken(context.Background(), &pbAuthenticationV1.CreateTokenRequest{
		Lifetime: durationpb.New(time.Minute),
		User:     util.NewType(DefaultUser),
	})
	require.NoError(t, err)

	validateResponse, err := s.Validate(context.Background(), &pbAuthenticationV1.ValidateRequest{
		Token: tokenResponse.Token,
	})

	require.NoError(t, err)

	require.True(t, validateResponse.IsValid)
	require.NotNil(t, validateResponse.Details)
	require.EqualValues(t, DefaultUser, validateResponse.Details.User)
}

func Test_Flow_WithoutTTL(t *testing.T) {
	directory := t.TempDir()

	ctx, c := context.WithCancel(context.Background())
	defer c()

	cfg := NewConfiguration()

	cfg.Path = directory
	cfg.TTL = time.Duration(0)

	s, err := newInternal(ctx, cfg)
	require.NoError(t, err)

	secret1 := generateJWTToken()
	secret2 := generateJWTToken()

	var token1, token2 string

	t.Run("Ensure we cant work without secrets", func(t *testing.T) {
		_, err := s.CreateToken(context.Background(), &pbAuthenticationV1.CreateTokenRequest{})
		require.EqualError(t, err, "unexpected EOF")
	})

	t.Run("Save secret1", func(t *testing.T) {
		reSaveJWTTokens(t, directory, secret1)
	})

	t.Run("Create token1", func(t *testing.T) {
		response, err := s.CreateToken(context.Background(), &pbAuthenticationV1.CreateTokenRequest{})
		require.NoError(t, err)
		require.NotNil(t, response)
		token1 = response.Token
	})

	t.Run("Validate token1", func(t *testing.T) {
		response, err := s.Validate(context.Background(), &pbAuthenticationV1.ValidateRequest{
			Token: token1,
		})
		require.NoError(t, err)
		require.NotNil(t, response)
		require.True(t, response.IsValid)
		require.EqualValues(t, cfg.Create.DefaultUser, response.Details.User)
	})

	t.Run("Save secret2", func(t *testing.T) {
		reSaveJWTTokens(t, directory, secret2)
	})

	t.Run("Create token2", func(t *testing.T) {
		response, err := s.CreateToken(context.Background(), &pbAuthenticationV1.CreateTokenRequest{})
		require.NoError(t, err)
		require.NotNil(t, response)
		token2 = response.Token
	})

	t.Run("Validate token2", func(t *testing.T) {
		response, err := s.Validate(context.Background(), &pbAuthenticationV1.ValidateRequest{
			Token: token2,
		})
		require.NoError(t, err)
		require.NotNil(t, response)
		require.True(t, response.IsValid)
		require.EqualValues(t, cfg.Create.DefaultUser, response.Details.User)
	})

	t.Run("Save secret1", func(t *testing.T) {
		reSaveJWTTokens(t, directory, secret1)
	})

	t.Run("Validate token2", func(t *testing.T) {
		response, err := s.Validate(context.Background(), &pbAuthenticationV1.ValidateRequest{
			Token: token2,
		})
		require.NoError(t, err)
		require.NotNil(t, response)
		require.False(t, response.IsValid)
		require.EqualValues(t, "signature is invalid", response.Message)
	})

	t.Run("Save secret2", func(t *testing.T) {
		reSaveJWTTokens(t, directory, secret2)
	})

	t.Run("Validate token1", func(t *testing.T) {
		response, err := s.Validate(context.Background(), &pbAuthenticationV1.ValidateRequest{
			Token: token1,
		})
		require.NoError(t, err)
		require.NotNil(t, response)
		require.False(t, response.IsValid)
		require.EqualValues(t, "signature is invalid", response.Message)
	})

	t.Run("Save secret1 & secret2", func(t *testing.T) {
		reSaveJWTTokens(t, directory, secret1, secret2)
	})

	t.Run("Validate token1", func(t *testing.T) {
		response, err := s.Validate(context.Background(), &pbAuthenticationV1.ValidateRequest{
			Token: token1,
		})
		require.NoError(t, err)
		require.NotNil(t, response)
		require.True(t, response.IsValid)
		require.EqualValues(t, cfg.Create.DefaultUser, response.Details.User)
	})

	t.Run("Validate token2", func(t *testing.T) {
		response, err := s.Validate(context.Background(), &pbAuthenticationV1.ValidateRequest{
			Token: token2,
		})
		require.NoError(t, err)
		require.NotNil(t, response)
		require.True(t, response.IsValid)
		require.EqualValues(t, cfg.Create.DefaultUser, response.Details.User)
	})
}
