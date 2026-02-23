//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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
	"crypto/tls"
	"fmt"
	"net"
	goHttp "net/http"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

type Service interface {
	Start(ctx context.Context) ServiceStarter

	StartWithHealth(ctx context.Context, health Health) ServiceStarter

	GetHandler(name string) (Handler, bool)

	Dial() (grpc.ClientConnInterface, error)
}

type service struct {
	lock sync.Mutex

	grpc serviceGRPC
	http *serviceHTTP

	cfg Configuration

	handlers []Handler

	starter ServiceStarter

	tls bool
}

type serviceGRPC struct {
	network *grpc.Server
	unix    *grpc.Server
}

type serviceHTTP struct {
	network *goHttp.Server
	unix    *goHttp.Server
}

func (p *service) Dial() (grpc.ClientConnInterface, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.starter == nil {
		return nil, errors.Errorf("server not initialized")
	}

	return p.dial(p.starter.Address(), p.starter.Unix())
}

func (p *service) dial(address string, unix string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	if p.tls {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if unix != "" {
		logger.Str("addr", fmt.Sprintf("unix://%s", unix)).Info("Connecting via UNIX Socket")
		opts = append(opts, grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			// ignore addr; dial the unix socket directly
			return (&net.Dialer{}).DialContext(ctx, "unix", unix)
		}))
	} else {
		logger.Str("addr", fmt.Sprintf("http://%s", address)).Info("Connecting via Socket")
	}

	return grpc.NewClient(address, opts...)
}

func (p *service) StartWithHealth(ctx context.Context, health Health) ServiceStarter {
	return p.start(ctx, health)
}

func (p *service) start(ctx context.Context, health Health) ServiceStarter {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.starter != nil {
		return serviceError{errors.Errorf("SErvice already started")}
	}

	p.starter = newServiceStarter(ctx, p, health)
	return p.starter
}

func (p *service) Start(ctx context.Context) ServiceStarter {
	return p.start(ctx, emptyHealth{})
}

func NewService(cfg Configuration, handlers ...Handler) (Service, error) {
	return newService(cfg, handlers...)
}

func newService(cfg Configuration, handlers ...Handler) (*service, error) {
	var q service

	var opts []grpc.ServerOption

	for _, handler := range handlers {
		if o, ok := handler.(HandlerInitService); ok {
			if err := o.InitService(&q); err != nil {
				return nil, errors.Wrapf(err, "Unable to init handler: %s", handler.Name())
			}
		}
	}

	tls, err := cfg.GetTLSOptions(shutdown.Context())
	if err != nil {
		return nil, err
	}

	if tls != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(tls)))
		q.tls = true
	}

	nopts, err := cfg.RenderOptions()
	if err != nil {
		return nil, err
	}

	opts = append(opts, nopts...)

	opts = append(opts, authenticator.NewInterceptorOptions(cfg.Authenticator)...)

	q.cfg = cfg
	q.grpc.network = grpc.NewServer(opts...)
	if unix := cfg.Unix; unix != "" {
		q.grpc.unix = grpc.NewServer(opts...)
	}

	q.handlers = handlers

	for _, handler := range q.handlers {
		handler.Register(q.grpc.network)
		if q.grpc.unix != nil {
			handler.Register(q.grpc.unix)
		}
	}

	if gateway := cfg.Gateway; gateway != nil {
		var http serviceHTTP
		http.network = &goHttp.Server{
			TLSConfig: tls,
		}

		if gateway.Unix != "" {
			http.unix = &goHttp.Server{}
		}

		q.http = &http
	}

	return &q, nil
}

func (p *service) GetHandler(name string) (Handler, bool) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, v := range p.handlers {
		if v.Name() == name {
			return v, true
		}
	}

	return nil, false
}
