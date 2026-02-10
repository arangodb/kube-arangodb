//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package authenticator

import (
	"context"

	"google.golang.org/grpc"
)

type identityContext string

const identityContextKey identityContext = "identity_context_key"

func serverStreamWithAuth(ss grpc.ServerStream, auth *Identity) grpc.ServerStream {
	if auth == nil {
		return ss
	}

	return &serverStream{
		ServerStream: ss,
		ctx:          context.WithValue(ss.Context(), identityContextKey, auth),
	}
}

type serverStream struct {
	grpc.ServerStream
	ctx context.Context
}

func NewInterceptorOptions(auth Authenticator) []grpc.ServerOption {
	if auth == nil {
		auth = NewAlwaysAuthenticator()
	}

	return []grpc.ServerOption{
		grpc.StreamInterceptor(func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			identity, err := auth.ValidateGRPC(ss.Context())
			if err != nil {
				return err
			}

			return handler(srv, serverStreamWithAuth(ss, identity))
		}),
		grpc.UnaryInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			identity, err := auth.ValidateGRPC(ctx)
			if err != nil {
				return nil, err
			}

			if identity != nil {
				ctx = context.WithValue(context.Background(), identityContextKey, identity)
			}

			return handler(ctx, req)
		}),
	}
}
