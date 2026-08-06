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
	"testing"

	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/stretchr/testify/require"

	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func passModeRequest(mode networkingApi.ArangoRouteSpecAuthenticationPassMode) *pbEnvoyAuthV3.CheckRequest {
	return &pbEnvoyAuthV3.CheckRequest{
		Attributes: &pbEnvoyAuthV3.AttributeContext{
			ContextExtensions: map[string]string{
				pbImplEnvoyAuthV3Shared.AuthConfigAuthPassModeKey: string(mode),
			},
		},
	}
}

func header(headers []*pbEnvoyCoreV3.HeaderValueOption, key string) (*pbEnvoyCoreV3.HeaderValueOption, bool) {
	for _, h := range headers {
		if h.GetHeader().GetKey() == key {
			return h, true
		}
	}
	return nil, false
}

func Test_Handle_NotAuthenticated(t *testing.T) {
	var p impl

	current := &pbImplEnvoyAuthV3Shared.Response{}

	require.NoError(t, p.Handle(context.Background(), passModeRequest(networkingApi.ArangoRouteSpecAuthenticationPassModePass), current))

	h, ok := header(current.Headers, pbImplEnvoyAuthV3Shared.AuthAuthenticatedHeader)
	require.True(t, ok)
	require.Equal(t, "false", h.GetHeader().GetValue())
}

func Test_Handle_PassWithExistingToken(t *testing.T) {
	var p impl

	current := &pbImplEnvoyAuthV3Shared.Response{
		User: &pbImplEnvoyAuthV3Shared.ResponseAuth{User: "root", Token: util.NewType("mytoken")},
	}

	require.NoError(t, p.Handle(context.Background(), passModeRequest(networkingApi.ArangoRouteSpecAuthenticationPassModePass), current))

	h, ok := header(current.Headers, pbImplEnvoyAuthV3Shared.AuthorizationHeader)
	require.True(t, ok)
	require.Equal(t, "bearer mytoken", h.GetHeader().GetValue())
}

func Test_Handle_Remove(t *testing.T) {
	var p impl

	current := &pbImplEnvoyAuthV3Shared.Response{
		User: &pbImplEnvoyAuthV3Shared.ResponseAuth{User: "root", Token: util.NewType("mytoken")},
	}

	require.NoError(t, p.Handle(context.Background(), passModeRequest(networkingApi.ArangoRouteSpecAuthenticationPassModeRemove), current))

	h, ok := header(current.Headers, pbImplEnvoyAuthV3Shared.AuthorizationHeader)
	require.True(t, ok)
	require.Equal(t, "", h.GetHeader().GetValue())
}

func Test_Handle_NoPassMode(t *testing.T) {
	var p impl

	current := &pbImplEnvoyAuthV3Shared.Response{
		User: &pbImplEnvoyAuthV3Shared.ResponseAuth{User: "root", Token: util.NewType("mytoken")},
	}

	require.NoError(t, p.Handle(context.Background(), &pbEnvoyAuthV3.CheckRequest{}, current))
	require.Empty(t, current.Headers)
}
