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

package tgrpc

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	proto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
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

func NewExecutor[IN, OUT proto.Message](t *testing.T, caller func(ctx context.Context, in IN, opts ...grpc.CallOption) (OUT, error), in IN, opts ...grpc.CallOption) Executor[OUT] {
	resp, err := caller(t.Context(), in, opts...)
	return executor[OUT]{
		ErrorStatusValidator: AsGRPCError(t, err),
		resp:                 resp,
	}
}

type Executor[T proto.Message] interface {
	ErrorStatusValidator

	Get(t *testing.T) T
}

type executor[T proto.Message] struct {
	ErrorStatusValidator

	resp T
}

func (e executor[T]) Get(t *testing.T) T {
	e.Code(t, codes.OK)
	return e.resp
}

type ErrorStatusValidator interface {
	Code(t *testing.T, code codes.Code) ErrorStatusValidator
	Errorf(t *testing.T, msg string, args ...interface{}) ErrorStatusValidator
}

type noErrorValidator struct {
}

func (n noErrorValidator) Code(t *testing.T, code codes.Code) ErrorStatusValidator {
	require.Equal(t, codes.OK, code, "code should be OK when no error provided")
	return n
}

func (n noErrorValidator) Errorf(t *testing.T, msg string, args ...interface{}) ErrorStatusValidator {
	require.Fail(t, "no error provided")
	return n
}

type errorStatusValidator struct {
	st *status.Status
}

func (e errorStatusValidator) Errorf(t *testing.T, msg string, args ...interface{}) ErrorStatusValidator {
	require.Equal(t, e.st.Message(), fmt.Sprintf(msg, args...))
	return e
}

func (e errorStatusValidator) Code(t *testing.T, code codes.Code) ErrorStatusValidator {
	require.Equal(t, code, e.st.Code(), e.st.Code().String())
	return e
}

func AsGRPCError(t *testing.T, err error) ErrorStatusValidator {
	if err == nil {
		return noErrorValidator{}
	}

	v, ok := errors.AsGRPCErrorStatus(err)
	require.True(t, ok)
	st := v.GRPCStatus()
	require.NotNil(t, st)
	return errorStatusValidator{st: st}
}

func GRPCAnyCastAs[T proto.Message](t *testing.T, in *anypb.Any, v T) {
	require.NoError(t, ugrpc.GRPCAnyCastAs[T](in, v))
}
