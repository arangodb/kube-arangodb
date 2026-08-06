//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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

package v3

import (
	"context"
	goHttp "net/http"
	"testing"

	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/stretchr/testify/require"

	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Client(t *testing.T, ctx context.Context) pbEnvoyAuthV3.AuthorizationClient {
	handler, err := New(ctx, pbImplEnvoyAuthV3Shared.Configuration{})
	require.NoError(t, err)
	local, err := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
	}, handler)
	require.NoError(t, err)

	start := local.Start(ctx)

	client := tgrpc.NewGRPCClient(t, ctx, pbEnvoyAuthV3.NewAuthorizationClient, start.Address())

	return client
}

func Test_DenyHeader(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client := Client(t, ctx)

	resp, err := client.Check(ctx, &pbEnvoyAuthV3.CheckRequest{})
	require.NoError(t, err)
	require.NoError(t, resp.Validate())
	require.NotNil(t, resp.Status)
	require.NotNil(t, resp.HttpResponse)
	require.NotNil(t, tests.CastAs[*pbEnvoyAuthV3.CheckResponse_DeniedResponse](t, resp.GetHttpResponse()).DeniedResponse)
	require.EqualValues(t, goHttp.StatusBadRequest, tests.CastAs[*pbEnvoyAuthV3.CheckResponse_DeniedResponse](t, resp.GetHttpResponse()).DeniedResponse.GetStatus().GetCode())
}

func Test_AllowAll(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client := Client(t, ctx)

	resp, err := client.Check(ctx, &pbEnvoyAuthV3.CheckRequest{
		Attributes: &pbEnvoyAuthV3.AttributeContext{
			ContextExtensions: map[string]string{
				pbImplEnvoyAuthV3Shared.AuthConfigTypeKey: pbImplEnvoyAuthV3Shared.AuthConfigTypeValue,
			},
		},
	})
	require.NoError(t, err)
	require.NoError(t, resp.Validate())
	require.Nil(t, resp.Status)
	require.NotNil(t, resp.HttpResponse)
	require.NotNil(t, tests.CastAs[*pbEnvoyAuthV3.CheckResponse_OkResponse](t, resp.GetHttpResponse()).OkResponse)
}

func Test_AllowAll_AddsRequestID(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client := Client(t, ctx)

	resp, err := client.Check(ctx, &pbEnvoyAuthV3.CheckRequest{
		Attributes: &pbEnvoyAuthV3.AttributeContext{
			ContextExtensions: map[string]string{
				pbImplEnvoyAuthV3Shared.AuthConfigTypeKey: pbImplEnvoyAuthV3Shared.AuthConfigTypeValue,
			},
		},
	})
	require.NoError(t, err)

	ok := tests.CastAs[*pbEnvoyAuthV3.CheckResponse_OkResponse](t, resp.GetHttpResponse()).OkResponse

	require.True(t, hasRequestID(ok.GetHeaders()), "request headers must carry the request id")
	require.True(t, hasRequestID(ok.GetResponseHeadersToAdd()), "response headers must carry the request id")
}

func hasRequestID(headers []*pbEnvoyCoreV3.HeaderValueOption) bool {
	for _, h := range headers {
		if h.GetHeader().GetKey() == utilConstants.EnvoyRequestIDHeader && h.GetHeader().GetValue() != "" {
			return true
		}
	}
	return false
}
