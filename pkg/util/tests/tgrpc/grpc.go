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
	"fmt"
	"testing"

	any1 "github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	proto "google.golang.org/protobuf/proto"

	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func NewGRPCClient[T any](t *testing.T, ctx context.Context, in func(cc grpc.ClientConnInterface) T, addr string, opts ...grpc.DialOption) T {
	client, closer, err := ugrpc.NewGRPCClient(ctx, in, addr, opts...)
	require.NoError(t, err)
	go func() {
		<-ctx.Done()

		require.NoError(t, closer.Close())
	}()

	return client
}

type ErrorStatusValidator interface {
	Code(t *testing.T, code codes.Code) ErrorStatusValidator
	Errorf(t *testing.T, msg string, args ...interface{}) ErrorStatusValidator
}

type errorStatusValidator struct {
	st *status.Status
}

func (e errorStatusValidator) Errorf(t *testing.T, msg string, args ...interface{}) ErrorStatusValidator {
	require.Equal(t, e.st.Message(), fmt.Sprintf(msg, args...))
	return e
}

func (e errorStatusValidator) Code(t *testing.T, code codes.Code) ErrorStatusValidator {
	require.Equal(t, code, e.st.Code())
	return e
}

func AsGRPCError(t *testing.T, err error) ErrorStatusValidator {
	v, ok := svc.AsGRPCErrorStatus(err)
	require.True(t, ok)
	st := v.GRPCStatus()
	require.NotNil(t, st)
	return errorStatusValidator{st: st}
}

func GRPCAnyCastAs[T proto.Message](t *testing.T, in *any1.Any, v T) {
	require.NoError(t, ugrpc.GRPCAnyCastAs[T](in, v))
}
