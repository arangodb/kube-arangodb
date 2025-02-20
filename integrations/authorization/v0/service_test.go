//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package v0

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	pbAuthorizationV0 "github.com/arangodb/kube-arangodb/integrations/authorization/v0/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Client(t *testing.T, ctx context.Context) pbAuthorizationV0.AuthorizationV0Client {
	local, err := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
	}, New())
	require.NoError(t, err)

	start := local.Start(ctx)

	client := tgrpc.NewGRPCClient(t, ctx, pbAuthorizationV0.NewAuthorizationV0Client, start.Address())

	return client
}

func Test_AllowAll(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client := Client(t, ctx)

	resp, err := client.Can(ctx, &pbAuthorizationV0.CanRequest{})
	require.NoError(t, err)
	require.True(t, resp.Allowed)
}
