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
			ctx, err := applyClientAuthentication(ctx, auth)
			if err != nil {
				return err
			}

			return invoker(ctx, method, req, reply, cc, opts...)
		}),
		grpc.WithStreamInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			ctx, err := applyClientAuthentication(ctx, auth)
			if err != nil {
				return nil, err
			}

			return streamer(ctx, desc, cc, method, opts...)
		}),
	}
}

// applyClientAuthentication stamps the service token produced by auth onto the outgoing context, unless the
// context already carries an authorization header (e.g. a caller token forwarded by the transparent proxy -
// pkg/util/svc.Proxy). In that case the forwarded token is preserved untouched: the server-side authenticator
// accepts only a single authorization value (two values are rejected as Unauthorized), and overriding it would
// discard the caller identity that the central service's RBAC must evaluate.
func applyClientAuthentication(ctx context.Context, auth Authentication) (context.Context, error) {
	token, change, err := auth.ExtendAuthentication(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, errors.ExtractGRPCCause(err).Error())
	}

	if change && !hasOutgoingAuthorization(ctx) {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)
	}

	return ctx, nil
}

// hasOutgoingAuthorization reports whether the outgoing context already carries an authorization header.
func hasOutgoingAuthorization(ctx context.Context) bool {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return false
	}

	return len(md.Get("authorization")) > 0
}
