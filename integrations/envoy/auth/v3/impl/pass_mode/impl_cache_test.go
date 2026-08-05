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

package pass_mode

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

func implWithToken(token pbImplEnvoyAuthV3Shared.Token) impl {
	return impl{
		cache: cache.NewHashCache[*pbImplEnvoyAuthV3Shared.ResponseAuth, pbImplEnvoyAuthV3Shared.Token](func(ctx context.Context, in *pbImplEnvoyAuthV3Shared.ResponseAuth) (pbImplEnvoyAuthV3Shared.Token, time.Time, error) {
			return token, time.Now().Add(time.Minute), nil
		}),
	}
}

func implWithError() impl {
	return impl{
		cache: cache.NewHashCache[*pbImplEnvoyAuthV3Shared.ResponseAuth, pbImplEnvoyAuthV3Shared.Token](func(ctx context.Context, in *pbImplEnvoyAuthV3Shared.ResponseAuth) (pbImplEnvoyAuthV3Shared.Token, time.Time, error) {
			return "", time.Time{}, fmt.Errorf("boom")
		}),
	}
}

func Test_Handle_Override_RendersToken(t *testing.T) {
	p := implWithToken("rendered")

	current := &pbImplEnvoyAuthV3Shared.Response{User: &pbImplEnvoyAuthV3Shared.ResponseAuth{User: "root"}}

	require.NoError(t, p.Handle(context.Background(), passModeRequest(networkingApi.ArangoRouteSpecAuthenticationPassModeOverride), current))

	h, ok := header(current.Headers, pbImplEnvoyAuthV3Shared.AuthorizationHeader)
	require.True(t, ok)
	require.Equal(t, "bearer rendered", h.GetHeader().GetValue())
}

func Test_Handle_PassWithoutToken_RendersToken(t *testing.T) {
	p := implWithToken("rendered")

	// Authenticated but without an inline token - pass mode must render one via the cache.
	current := &pbImplEnvoyAuthV3Shared.Response{User: &pbImplEnvoyAuthV3Shared.ResponseAuth{User: "root"}}

	require.NoError(t, p.Handle(context.Background(), passModeRequest(networkingApi.ArangoRouteSpecAuthenticationPassModePass), current))

	h, ok := header(current.Headers, pbImplEnvoyAuthV3Shared.AuthorizationHeader)
	require.True(t, ok)
	require.Equal(t, "bearer rendered", h.GetHeader().GetValue())
}

func Test_Handle_Override_TokenError_Denies(t *testing.T) {
	p := implWithError()

	current := &pbImplEnvoyAuthV3Shared.Response{User: &pbImplEnvoyAuthV3Shared.ResponseAuth{User: "root"}}

	err := p.Handle(context.Background(), passModeRequest(networkingApi.ArangoRouteSpecAuthenticationPassModeOverride), current)

	var denied pbImplEnvoyAuthV3Shared.DeniedResponse
	require.ErrorAs(t, err, &denied)
	require.EqualValues(t, 401, denied.Code)
}

func Test_Handle_PassWithoutToken_Error_Denies(t *testing.T) {
	p := implWithError()

	current := &pbImplEnvoyAuthV3Shared.Response{User: &pbImplEnvoyAuthV3Shared.ResponseAuth{User: "root"}}

	err := p.Handle(context.Background(), passModeRequest(networkingApi.ArangoRouteSpecAuthenticationPassModePass), current)

	var denied pbImplEnvoyAuthV3Shared.DeniedResponse
	require.ErrorAs(t, err, &denied)
	require.EqualValues(t, 401, denied.Code)
}
