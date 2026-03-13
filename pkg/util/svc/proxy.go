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

package svc

import (
	"context"
	goStrings "strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	sproxy "github.com/siderolabs/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

type GatewayProxyRegisterer func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error

func ProxyGateway(name string, in GatewayProxyRegisterer) HandlerGateway {
	return proxyHandler{in: in, name: name}
}

type proxyHandler struct {
	name string
	in   GatewayProxyRegisterer
}

func (p proxyHandler) Name() string {
	return p.name
}

func (p proxyHandler) Health(ctx context.Context) HealthState {
	return Healthy
}

func (p proxyHandler) Register(registrar *grpc.Server) {

}

func (p proxyHandler) Gateway(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return p.in(ctx, mux, conn)
}

func ProxyClientOpts() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.ForceCodecV2(sproxy.Codec())),
	}
}

func ProxyServer(obj cache.Object[*grpc.ClientConn]) []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.UnknownServiceHandler(sproxy.TransparentHandler(Proxy(obj))),

		grpc.ForceServerCodecV2(sproxy.Codec()),
	}
}

func Proxy(obj cache.Object[*grpc.ClientConn]) sproxy.StreamDirector {
	return func(ctx context.Context, fullMethodName string) (sproxy.Mode, []sproxy.Backend, error) {
		if goStrings.HasPrefix(fullMethodName, "/grpc.health.v1.Health/") {
			return sproxy.One2One, nil, status.Errorf(codes.Unimplemented, "handled locally")
		}

		inMD, _ := metadata.FromIncomingContext(ctx)
		outCtx := metadata.NewOutgoingContext(ctx, inMD)

		logger.Str("method", fullMethodName).Debug("Proxy Request")

		return sproxy.One2One, []sproxy.Backend{
			&sproxy.SingleBackend{
				GetConn: func(ctx context.Context) (context.Context, *grpc.ClientConn, error) {
					conn, err := obj.Get(ctx)
					if err != nil {
						return outCtx, nil, status.Errorf(codes.Unavailable, "Upstream Service not available")
					}
					return outCtx, conn, nil
				},
			},
		}, nil
	}
}
