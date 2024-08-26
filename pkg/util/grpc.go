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

package util

import (
	"context"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const AuthorizationGRPCHeader = "adb-authorization"

func NewGRPCClient[T any](ctx context.Context, in func(cc grpc.ClientConnInterface) T, addr string, opts ...grpc.DialOption) (T, io.Closer, error) {
	con, err := NewGRPCConn(ctx, addr, opts...)
	if err != nil {
		return Default[T](), nil, err
	}

	return in(con), con, nil
}

func NewGRPCConn(ctx context.Context, addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	var z []grpc.DialOption

	z = append(z, grpc.WithTransportCredentials(insecure.NewCredentials()))

	z = append(z, opts...)

	conn, err := grpc.DialContext(ctx, addr, z...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func TokenAuthInterceptors(token string) []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return invoker(attachTokenAuthToInterceptors(ctx, token), method, req, reply, cc, opts...)
		}),
		grpc.WithStreamInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return streamer(attachTokenAuthToInterceptors(ctx, token), desc, cc, method, opts...)
		}),
	}
}

func attachTokenAuthToInterceptors(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, AuthorizationGRPCHeader, token)
}
