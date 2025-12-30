//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package impl

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	"github.com/arangodb/kube-arangodb/pkg/api/server"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func New(cfg Configuration) svc.Handler {
	return newInternal(cfg)
}

func newInternal(cfg Configuration) *implementation {
	return &implementation{
		cfg: cfg,
	}
}

var _ server.OperatorServer = &implementation{}
var _ svc.Handler = &implementation{}

type implementation struct {
	server.UnimplementedOperatorServer

	cfg Configuration
}

func (i *implementation) Name() string {
	return "operator"
}

func (i *implementation) Health() svc.HealthState {
	return svc.Healthy
}

func (i *implementation) Register(registrar *grpc.Server) {
	server.RegisterOperatorServer(registrar, i)
}

func (i *implementation) Gateway(ctx context.Context, mux *runtime.ServeMux) error {
	return server.RegisterOperatorHandlerServer(ctx, mux, i)
}
func (i *implementation) authenticate(ctx context.Context) error {
	if i.cfg.Authenticator == nil {
		return nil
	}

	return i.cfg.Authenticator.ValidateGRPC(ctx)
}
