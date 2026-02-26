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

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	pbPongV1 "github.com/arangodb/kube-arangodb/integrations/pong/v1/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
)

func newPongHandler() HandlerGateway {
	return pongService{}
}

type pongService struct {
	pbPongV1.UnimplementedPongV1Server
}

func (p pongService) Gateway(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return pbPongV1.RegisterPongV1Handler(ctx, mux, conn)
}

func (p pongService) Name() string {
	return "pong"
}

func (p pongService) Health(ctx context.Context) HealthState {
	return Healthy
}

func (p pongService) Register(registrar *grpc.Server) {
	pbPongV1.RegisterPongV1Server(registrar, p)
}

func (p pongService) Ping(ctx context.Context, empty *pbSharedV1.Empty) (*pbPongV1.PongV1PingResponse, error) {
	return &pbPongV1.PongV1PingResponse{}, nil
}
