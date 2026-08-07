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

package auth_cookie

import (
	"context"
	"fmt"
	"testing"
	"time"

	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/stretchr/testify/require"

	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

func cookieRequest(cookie string, passMode string) *pbEnvoyAuthV3.CheckRequest {
	req := &pbEnvoyAuthV3.CheckRequest{
		Attributes: &pbEnvoyAuthV3.AttributeContext{
			Request: &pbEnvoyAuthV3.AttributeContext_Request{
				Http: &pbEnvoyAuthV3.AttributeContext_HttpRequest{
					Headers: map[string]string{
						"cookie": cookie,
					},
				},
			},
		},
	}

	if passMode != "" {
		req.Attributes.ContextExtensions = map[string]string{
			pbImplEnvoyAuthV3Shared.AuthConfigAuthPassModeKey: passMode,
		}
	}

	return req
}

func newImpl(valid bool) impl {
	return impl{
		cache: cache.NewCache[pbImplEnvoyAuthV3Shared.Token, pbImplEnvoyAuthV3Shared.ResponseAuth](func(ctx context.Context, in pbImplEnvoyAuthV3Shared.Token) (pbImplEnvoyAuthV3Shared.ResponseAuth, time.Time, error) {
			if !valid {
				return pbImplEnvoyAuthV3Shared.ResponseAuth{}, time.Time{}, fmt.Errorf("invalid token")
			}
			return pbImplEnvoyAuthV3Shared.ResponseAuth{User: "root", Groups: []string{"g"}}, time.Now().Add(time.Minute), nil
		}),
	}
}

// preExistingHeader mimics a header injected by an earlier handler (e.g. the request id handler).
func preExistingHeader() *pbEnvoyCoreV3.HeaderValueOption {
	return &pbEnvoyCoreV3.HeaderValueOption{
		Header: &pbEnvoyCoreV3.HeaderValue{Key: "x-request-id", Value: "abc"},
	}
}

func hasHeaderKey(headers []*pbEnvoyCoreV3.HeaderValueOption, key string) bool {
	for _, h := range headers {
		if h.GetHeader().GetKey() == key {
			return true
		}
	}
	return false
}

func Test_Handle_AlreadyAuthenticated(t *testing.T) {
	z := newImpl(true)

	current := &pbImplEnvoyAuthV3Shared.Response{User: &pbImplEnvoyAuthV3Shared.ResponseAuth{User: "existing"}}

	require.NoError(t, z.Handle(context.Background(), cookieRequest(JWTAuthorizationCookieName+"=token", ""), current))
	require.Equal(t, "existing", current.User.User)
}

func Test_Handle_NoMatchingCookie(t *testing.T) {
	z := newImpl(true)

	current := &pbImplEnvoyAuthV3Shared.Response{}

	require.NoError(t, z.Handle(context.Background(), cookieRequest("other=token", ""), current))
	require.Nil(t, current.User)
}

func Test_Handle_InvalidToken(t *testing.T) {
	z := newImpl(false)

	current := &pbImplEnvoyAuthV3Shared.Response{}

	require.NoError(t, z.Handle(context.Background(), cookieRequest(JWTAuthorizationCookieName+"=token", ""), current))
	require.Nil(t, current.User)
}

func Test_Handle_Authenticates(t *testing.T) {
	z := newImpl(true)

	current := &pbImplEnvoyAuthV3Shared.Response{}

	require.NoError(t, z.Handle(context.Background(), cookieRequest(JWTAuthorizationCookieName+"=token", ""), current))
	require.NotNil(t, current.User)
	require.Equal(t, "root", current.User.User)
	require.NotNil(t, current.User.Token)
	require.Equal(t, "token", *current.User.Token)
}

// Regression: authenticating via cookie in the default pass mode must strip the cookie header while
// preserving headers added by earlier handlers.
func Test_Handle_PreservesExistingHeaders(t *testing.T) {
	z := newImpl(true)

	current := &pbImplEnvoyAuthV3Shared.Response{
		Headers: []*pbEnvoyCoreV3.HeaderValueOption{preExistingHeader()},
	}

	require.NoError(t, z.Handle(context.Background(), cookieRequest(JWTAuthorizationCookieName+"=token", ""), current))

	require.True(t, hasHeaderKey(current.Headers, "x-request-id"), "earlier header must survive")
	require.True(t, hasHeaderKey(current.Headers, pbImplEnvoyAuthV3Shared.CookieHeader), "cookie header must be stripped")
}

// In pass mode the cookie header must be kept (not stripped), and earlier headers preserved.
func Test_Handle_PassModeKeepsCookie(t *testing.T) {
	z := newImpl(true)

	current := &pbImplEnvoyAuthV3Shared.Response{
		Headers: []*pbEnvoyCoreV3.HeaderValueOption{preExistingHeader()},
	}

	require.NoError(t, z.Handle(context.Background(), cookieRequest(JWTAuthorizationCookieName+"=token", string(networkingApi.ArangoRouteSpecAuthenticationPassModePass)), current))

	require.NotNil(t, current.User)
	require.True(t, hasHeaderKey(current.Headers, "x-request-id"))
	require.False(t, hasHeaderKey(current.Headers, pbImplEnvoyAuthV3Shared.CookieHeader), "cookie header must be kept in pass mode")
}
