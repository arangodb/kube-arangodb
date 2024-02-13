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

package svc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pbHealth "google.golang.org/grpc/health/grpc_health_v1"
)

func Test_Service(t *testing.T) {
	h := NewHealthService(Configuration{
		Address: "127.0.0.1:0",
	}, Readiness)

	other := NewService(Configuration{
		Address: "127.0.0.1:0",
	})

	ctx, c := context.WithCancel(context.Background())
	defer c()

	st := h.Start(ctx)

	othStart := other.StartWithHealth(ctx, h)

	healthConn, err := grpc.DialContext(ctx, st.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	defer healthConn.Close()

	otherConn, err := grpc.DialContext(ctx, othStart.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	defer otherConn.Close()

	cl := pbHealth.NewHealthClient(healthConn)

	_, err = cl.Check(context.Background(), &pbHealth.HealthCheckRequest{})
	require.NoError(t, err)

	ol := pbHealth.NewHealthClient(otherConn)

	_, err = ol.Check(context.Background(), &pbHealth.HealthCheckRequest{})
	require.Error(t, err)

	c()

	require.NoError(t, st.Wait())
}
