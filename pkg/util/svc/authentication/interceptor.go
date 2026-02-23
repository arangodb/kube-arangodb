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

package authentication

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func NewInterceptorClientOptions(auth Authentication) []grpc.DialOption {
	if auth == nil {
		auth = NewEmptyAuthentication()
	}

	return []grpc.DialOption{
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			token, change, err := auth.ExtendAuthentication(ctx)
			if err != nil {
				return status.Error(codes.Unauthenticated, errors.ExtractGRPCCause(err).Error())
			}

			if change {
				ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)
			}

			return invoker(ctx, method, req, reply, cc, opts...)
		}),
		grpc.WithStreamInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			token, change, err := auth.ExtendAuthentication(ctx)
			if err != nil {
				return nil, status.Error(codes.Unauthenticated, errors.ExtractGRPCCause(err).Error())
			}

			if change {
				ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)
			}
			return streamer(ctx, desc, cc, method, opts...)
		}),
	}
}
