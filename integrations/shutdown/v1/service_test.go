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

package v1

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	pbShutdownV1 "github.com/arangodb/kube-arangodb/integrations/shutdown/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/closer"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Client(t *testing.T, ctx context.Context, c context.CancelFunc) pbShutdownV1.ShutdownV1Client {
	local, err := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
	}, New(Configuration{}, c))
	require.NoError(t, err)

	start := local.Start(ctx)

	client := tgrpc.NewGRPCClient(t, ctx, pbShutdownV1.NewShutdownV1Client, start.Address())

	return client
}

func Test_AllowAll(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	nctx, nc := context.WithCancel(ctx)

	client := Client(t, ctx, nc)

	require.False(t, closer.IsChannelClosed(nctx.Done()))

	_, err := client.Shutdown(ctx, &pbSharedV1.Empty{})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	require.True(t, closer.IsChannelClosed(nctx.Done()))

}
