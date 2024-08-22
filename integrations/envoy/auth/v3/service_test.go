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

package v3

import (
	"context"
	"testing"

	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Client(t *testing.T, ctx context.Context) pbEnvoyAuthV3.AuthorizationClient {
	local := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
	}, New())

	start := local.Start(ctx)

	client := tgrpc.NewGRPCClient(t, ctx, pbEnvoyAuthV3.NewAuthorizationClient, start.Address())

	return client
}

func Test_AllowAll(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client := Client(t, ctx)

	resp, err := client.Check(ctx, &pbEnvoyAuthV3.CheckRequest{})
	require.NoError(t, err)
	require.NoError(t, resp.Validate())
	require.Nil(t, resp.Status)
	require.Nil(t, resp.HttpResponse)
	require.Nil(t, resp.DynamicMetadata)
}
