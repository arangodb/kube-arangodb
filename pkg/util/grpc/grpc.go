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

package grpc

import (
	"context"
	"crypto/tls"
	"io"

	any1 "github.com/golang/protobuf/ptypes/any"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	proto "google.golang.org/protobuf/proto"

	pbPongV1 "github.com/arangodb/kube-arangodb/integrations/pong/v1/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

const AuthorizationGRPCHeader = "adb-authorization"

func NewGRPCClient[T any](ctx context.Context, in func(cc grpc.ClientConnInterface) T, addr string, opts ...grpc.DialOption) (T, io.Closer, error) {
	con, err := NewGRPCConn(addr, opts...)
	if err != nil {
		return util.Default[T](), nil, err
	}

	return in(con), con, nil
}

func NewOptionalTLSGRPCClient[T any](ctx context.Context, in func(cc grpc.ClientConnInterface) T, addr string, tls *tls.Config, opts ...grpc.DialOption) (T, io.Closer, error) {
	con, err := NewOptionalTLSGRPCConn(ctx, addr, tls, opts...)
	if err != nil {
		return util.Default[T](), nil, err
	}

	return in(con), con, nil
}

func NewOptionalTLSGRPCConn(ctx context.Context, addr string, tls *tls.Config, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if tls != nil {
		// Try with TLS
		tlsOpts := ClientTLS(tls)
		newOpts := make([]grpc.DialOption, len(opts)+len(tlsOpts))
		copy(newOpts, opts)
		copy(newOpts[len(opts):], tlsOpts)

		// Create conn
		conn, err := newGRPCConn(addr, tlsOpts...)
		if err != nil {
			return nil, err
		}

		if _, err := pbPongV1.NewPongV1Client(conn).Ping(ctx, &pbSharedV1.Empty{}); err != nil {
			if v, ok := svc.AsGRPCErrorStatus(err); !ok {
				return nil, err
			} else {
				if status := v.GRPCStatus(); status == nil {
					return nil, err
				} else {
					if status.Code() != codes.Unavailable {
						return nil, err
					}
				}
			}
		} else {
			return conn, nil
		}

		if err := conn.Close(); err != nil {
			return nil, err
		}
	}

	return newGRPCConn(addr, opts...)
}

func newGRPCConn(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	var z []grpc.DialOption

	z = append(z, grpc.WithTransportCredentials(insecure.NewCredentials()))

	z = append(z, opts...)

	conn, err := grpc.NewClient(addr, z...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func NewGRPCConn(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return newGRPCConn(addr, opts...)
}

func ClientTLS(config *tls.Config) []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(config)),
	}
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

func GRPCAnyCastAs[T proto.Message](in *any1.Any, v T) error {
	if err := in.UnmarshalTo(v); err != nil {
		return err
	}

	return nil
}
