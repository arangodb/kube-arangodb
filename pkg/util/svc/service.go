//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

type Service interface {
	Start(ctx context.Context) ServiceStarter

	StartWithHealth(ctx context.Context, health Health) ServiceStarter
}

type service struct {
	server *grpc.Server
	http   *http.Server

	cfg Configuration

	handlers []Handler
}

func (p *service) StartWithHealth(ctx context.Context, health Health) ServiceStarter {
	return newServiceStarter(ctx, p, health)
}

func (p *service) Start(ctx context.Context) ServiceStarter {
	return newServiceStarter(ctx, p, emptyHealth{})
}

func NewService(cfg Configuration, handlers ...Handler) (Service, error) {
	return newService(cfg, handlers...)
}

func newService(cfg Configuration, handlers ...Handler) (*service, error) {
	var q service

	q.cfg = cfg
	q.server = grpc.NewServer(cfg.RenderOptions()...)
	q.handlers = handlers

	for _, handler := range q.handlers {
		handler.Register(q.server)
	}

	if gateway := cfg.Gateway; gateway != nil {
		mux := runtime.NewServeMux()

		for _, handler := range q.handlers {
			if err := handler.Gateway(shutdown.Context(), mux); err != nil {
				return nil, err
			}
		}

		q.http = &http.Server{
			Handler:   mux,
			TLSConfig: cfg.TLSOptions,
		}
	}

	return &q, nil
}
