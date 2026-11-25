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
	goStrings "strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// authenticationValidateFullMethod is the gRPC full-method name of the authentication token-check
// (AuthenticationV1.Validate). It is hard-coded here - rather than imported from the authentication
// definition - to keep this generic authenticator package free of a dependency on a concrete service,
// matching how isHealthCheck hard-codes the health full-method prefix.
const authenticationValidateFullMethod = "/authentication.AuthenticationV1/Validate"

// isHealthCheck returns true for gRPC health check methods which must be
// accessible without authentication (used by proxies, load balancers, and
// Kubernetes probes).
func isHealthCheck(fullMethod string) bool {
	return goStrings.HasPrefix(fullMethod, "/grpc.health.v1.Health/")
}

// isPublicMethod returns true for methods that must be reachable without authentication: gRPC health
// checks, and the authentication token-check (AuthenticationV1.Validate). Validate only verifies a token
// against the public signing-key pool and returns its validity - it is the call a client makes before it
// is authenticated, so requiring credentials would be circular. Privileged operations (e.g. CreateToken)
// are not listed here and stay gated.
func isPublicMethod(fullMethod string) bool {
	return isHealthCheck(fullMethod) || fullMethod == authenticationValidateFullMethod
}

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

func (s *serverStream) Context() context.Context {
	return s.ctx
}

func NewInterceptorOptions(auth Authenticator) []grpc.ServerOption {
	if auth == nil {
		auth = NewAlwaysAuthenticator()
	}

	return []grpc.ServerOption{
		grpc.StreamInterceptor(func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			if isPublicMethod(info.FullMethod) {
				return handler(srv, ss)
			}

			identity, err := auth.ValidateGRPC(ss.Context())
			if err != nil {
				return status.Error(codes.Unauthenticated, errors.ExtractGRPCCause(err).Error())
			}

			return handler(srv, serverStreamWithAuth(ss, identity))
		}),
		grpc.UnaryInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			if isPublicMethod(info.FullMethod) {
				return handler(ctx, req)
			}

			identity, err := auth.ValidateGRPC(ctx)
			if err != nil {
				return nil, status.Error(codes.Unauthenticated, errors.ExtractGRPCCause(err).Error())
			}

			if identity != nil {
				ctx = context.WithValue(ctx, identityContextKey, identity)
			}

			return handler(ctx, req)
		}),
	}
}
