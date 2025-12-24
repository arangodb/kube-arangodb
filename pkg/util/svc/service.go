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
	goHttp "net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

type Service interface {
	Start(ctx context.Context) ServiceStarter

	StartWithHealth(ctx context.Context, health Health) ServiceStarter
}

type service struct {
	server *grpc.Server
	http   *goHttp.Server

	cfg Configuration

	handlers []Handler
}

func (p *service) StartWithHealth(ctx context.Context, health Health) ServiceStarter {
	return newServiceStarter(ctx, p, health)
}

func (p *service) Start(ctx context.Context) ServiceStarter {
	return p.StartWithHealth(ctx, emptyHealth{})
}

func NewService(cfg Configuration, handlers ...Handler) (Service, error) {
	return newService(cfg, handlers...)
}

func newService(cfg Configuration, handlers ...Handler) (*service, error) {
	var q service

	var opts []grpc.ServerOption

	tls, err := cfg.GetTLSOptions(shutdown.Context())
	if err != nil {
		return nil, err
	}

	if tls != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(tls)))
	}

	nopts, err := cfg.RenderOptions()
	if err != nil {
		return nil, err
	}

	opts = append(opts, nopts...)

	q.cfg = cfg
	q.server = grpc.NewServer(opts...)
	q.handlers = handlers

	for _, handler := range q.handlers {
		handler.Register(q.server)
	}

	if gateway := cfg.Gateway; gateway != nil {
		mux := runtime.NewServeMux(cfg.MuxExtensions...)

		for _, handler := range q.handlers {
			if err := handler.Gateway(shutdown.Context(), mux); err != nil {
				return nil, err
			}
		}

		var handler goHttp.Handler = mux

		handler = cfg.Wrap.Wrap(handler)

		q.http = &goHttp.Server{
			Handler:   handler,
			TLSConfig: tls,
		}
	}

	return &q, nil
}
