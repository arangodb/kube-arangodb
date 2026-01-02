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
	"fmt"
	"net"
	goHttp "net/http"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ServiceStarter interface {
	Wait() error

	Address() string
	HTTPAddress() string
}

type serviceStarter struct {
	service *service

	address, httpAddress string

	error error
	done  chan struct{}
}

func (s *serviceStarter) Address() string {
	return s.address
}

func (s *serviceStarter) HTTPAddress() string {
	return s.httpAddress
}

func (s *serviceStarter) Wait() error {
	<-s.done

	return s.error
}

func (s *serviceStarter) run(ctx context.Context, health Health, ln, http net.Listener) {
	defer close(s.done)

	s.error = s.runE(ctx, health, ln, http)
}

func (s *serviceStarter) runE(ctx context.Context, health Health, ln, http net.Listener) error {
	ctx, c := context.WithCancel(ctx)
	defer c()

	var serveError error

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer c()

		go func() {
			<-ctx.Done()

			s.service.server.GracefulStop()
		}()

		if err := s.service.server.Serve(ln); !errors.AnyOf(err, grpc.ErrServerStopped) {
			serveError = err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if s.service.cfg.Gateway == nil {
			return
		}

		defer c()

		go func() {
			<-ctx.Done()

			s.service.http.Close()
		}()

		if s.service.http.TLSConfig == nil {
			if err := s.service.http.Serve(http); !errors.AnyOf(err, goHttp.ErrServerClosed) {
				serveError = err
			}
		} else {
			if err := s.service.http.ServeTLS(http, "", ""); !errors.AnyOf(err, goHttp.ErrServerClosed) {
				serveError = err
			}
		}
	}()

	done := make(chan struct{})

	go func() {
		defer close(done)

		wg.Wait()
	}()

	ticker := time.NewTicker(125 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return serveError
		default:
			for _, h := range s.service.handlers {
				health.Update(h.Name(), h.Health())
			}
			break
		}

		select {
		case <-done:
			return serveError
		case <-ticker.C:
			continue
		}
	}
}

func newServiceStarter(ctx context.Context, service *service, health Health) ServiceStarter {
	st := &serviceStarter{
		service: service,
		error:   nil,
		done:    make(chan struct{}),
	}

	ln, err := net.Listen("tcp", service.cfg.Address)
	if err != nil {
		return serviceError{err}
	}

	pr := ln.Addr().(*net.TCPAddr)
	st.address = fmt.Sprintf("%s:%d", pr.IP.String(), pr.Port)

	var hln net.Listener

	if service.cfg.Gateway != nil {
		httpln, err := net.Listen("tcp", service.cfg.Gateway.Address)
		if err != nil {
			return serviceError{err}
		}

		pr := httpln.Addr().(*net.TCPAddr)
		st.httpAddress = fmt.Sprintf("%s:%d", pr.IP.String(), pr.Port)

		hln = httpln
	}

	go st.run(ctx, health, ln, hln)

	return st
}
