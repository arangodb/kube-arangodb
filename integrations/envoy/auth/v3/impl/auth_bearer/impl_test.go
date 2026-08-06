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

package auth_bearer

import (
	"context"
	"fmt"
	"testing"
	"time"

	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/stretchr/testify/require"

	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

func bearerRequest(authorization string) *pbEnvoyAuthV3.CheckRequest {
	headers := map[string]string{}
	if authorization != "" {
		headers["authorization"] = authorization
	}
	return &pbEnvoyAuthV3.CheckRequest{
		Attributes: &pbEnvoyAuthV3.AttributeContext{
			Request: &pbEnvoyAuthV3.AttributeContext_Request{
				Http: &pbEnvoyAuthV3.AttributeContext_HttpRequest{
					Headers: headers,
				},
			},
		},
	}
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

func Test_Handle_AlreadyAuthenticated(t *testing.T) {
	z := newImpl(true)
	current := &pbImplEnvoyAuthV3Shared.Response{User: &pbImplEnvoyAuthV3Shared.ResponseAuth{User: "existing"}}
	require.NoError(t, z.Handle(context.Background(), bearerRequest("bearer token"), current))
	require.Equal(t, "existing", current.User.User)
}

func Test_Handle_NoAuthorizationHeader(t *testing.T) {
	z := newImpl(true)
	current := &pbImplEnvoyAuthV3Shared.Response{}
	require.NoError(t, z.Handle(context.Background(), bearerRequest(""), current))
	require.Nil(t, current.User)
}

func Test_Handle_NonBearerScheme(t *testing.T) {
	z := newImpl(true)
	current := &pbImplEnvoyAuthV3Shared.Response{}
	require.NoError(t, z.Handle(context.Background(), bearerRequest("Basic dXNlcjpwYXNz"), current))
	require.Nil(t, current.User)
}

func Test_Handle_MalformedHeader(t *testing.T) {
	z := newImpl(true)
	current := &pbImplEnvoyAuthV3Shared.Response{}
	require.NoError(t, z.Handle(context.Background(), bearerRequest("bearer"), current))
	require.Nil(t, current.User)
}

func Test_Handle_InvalidToken(t *testing.T) {
	z := newImpl(false)
	current := &pbImplEnvoyAuthV3Shared.Response{}
	require.NoError(t, z.Handle(context.Background(), bearerRequest("bearer token"), current))
	require.Nil(t, current.User)
}

func Test_Handle_Authenticates(t *testing.T) {
	z := newImpl(true)
	current := &pbImplEnvoyAuthV3Shared.Response{}
	require.NoError(t, z.Handle(context.Background(), bearerRequest("bearer mytoken"), current))
	require.NotNil(t, current.User)
	require.Equal(t, "root", current.User.User)
	require.NotNil(t, current.User.Token)
	require.Equal(t, "mytoken", *current.User.Token)
}

func Test_Handle_SchemeCaseInsensitive(t *testing.T) {
	z := newImpl(true)
	current := &pbImplEnvoyAuthV3Shared.Response{}
	require.NoError(t, z.Handle(context.Background(), bearerRequest("Bearer mytoken"), current))
	require.NotNil(t, current.User)
	require.Equal(t, "root", current.User.User)
}
