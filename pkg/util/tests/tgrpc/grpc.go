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

package tgrpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGRPCClient[T any](t *testing.T, ctx context.Context, in func(cc grpc.ClientConnInterface) T, addr string, opts ...grpc.DialOption) T {
	return in(NewGRPCConn(t, ctx, addr, opts...))
}

func NewGRPCConn(t *testing.T, ctx context.Context, addr string, opts ...grpc.DialOption) *grpc.ClientConn {
	var z []grpc.DialOption

	z = append(z, grpc.WithTransportCredentials(insecure.NewCredentials()))

	z = append(z, opts...)

	conn, err := grpc.DialContext(ctx, addr, z...)
	require.NoError(t, err)

	go func() {
		<-ctx.Done()

		require.NoError(t, conn.Close())
	}()

	return conn
}
