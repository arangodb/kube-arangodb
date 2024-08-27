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

package svc

import (
	"context"

	"google.golang.org/grpc"
)

type Service interface {
	Start(ctx context.Context) ServiceStarter

	StartWithHealth(ctx context.Context, health Health) ServiceStarter
}

type service struct {
	server *grpc.Server

	cfg Configuration

	handlers []Handler
}

func (p *service) StartWithHealth(ctx context.Context, health Health) ServiceStarter {
	return newServiceStarter(ctx, p, health)
}

func (p *service) Start(ctx context.Context) ServiceStarter {
	return newServiceStarter(ctx, p, emptyHealth{})
}

func NewService(cfg Configuration, handlers ...Handler) Service {
	return newService(cfg, handlers...)
}

func newService(cfg Configuration, handlers ...Handler) *service {
	var q service

	q.cfg = cfg
	q.server = grpc.NewServer(cfg.RenderOptions()...)
	q.handlers = handlers

	for _, handler := range q.handlers {
		handler.Register(q.server)
	}

	return &q
}
