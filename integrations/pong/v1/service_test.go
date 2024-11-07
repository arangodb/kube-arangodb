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

	pbPongV1 "github.com/arangodb/kube-arangodb/integrations/pong/v1/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Client(t *testing.T, ctx context.Context, services ...Service) pbPongV1.PongV1Client {
	h, err := New(services...)
	require.NoError(t, err)

	local := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
	}, h)

	start := local.Start(ctx)

	client := tgrpc.NewGRPCClient(t, ctx, pbPongV1.NewPongV1Client, start.Address())

	return client
}

func Test_Ping(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client := Client(t, ctx)

	r1, err := client.Ping(ctx, &pbSharedV1.Empty{})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	r2, err := client.Ping(ctx, &pbSharedV1.Empty{})
	require.NoError(t, err)

	require.True(t, r2.GetTime().AsTime().After(r1.GetTime().AsTime()))
}

func Test_Services(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	v2 := Service{
		Name:    "A",
		Version: "V2",
		Enabled: true,
	}

	v3 := Service{
		Name:    "A",
		Version: "V3",
		Enabled: false,
	}

	client := Client(t, ctx,
		v3,
		v2,
	)

	r1, err := client.Services(ctx, &pbSharedV1.Empty{})
	require.NoError(t, err)

	require.Len(t, r1.GetServices(), 2)

	require.NotNil(t, r1.GetServices()[0])
	require.NotNil(t, r1.GetServices()[1])

	require.EqualValues(t, v2.asService(), r1.GetServices()[0])
	require.EqualValues(t, v3.asService(), r1.GetServices()[1])
}
