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
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ServiceStarter interface {
	Wait() error

	Address() string
}

type serviceStarter struct {
	service *service

	address string

	error error
	done  chan struct{}
}

func (s *serviceStarter) Address() string {
	return s.address
}

func (s *serviceStarter) Wait() error {
	<-s.done

	return s.error
}

func (s *serviceStarter) run(ctx context.Context, health Health, ln net.Listener) {
	defer close(s.done)

	s.error = s.runE(ctx, health, ln)
}

func (s *serviceStarter) runE(ctx context.Context, health Health, ln net.Listener) error {
	pr := ln.Addr().(*net.TCPAddr)
	s.address = fmt.Sprintf("%s:%d", pr.IP.String(), pr.Port)

	var serveError error

	done := make(chan struct{})
	go func() {
		defer close(done)

		if err := s.service.server.Serve(ln); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			serveError = err
		}
	}()

	go func() {
		<-ctx.Done()

		s.service.server.GracefulStop()
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

	go st.run(ctx, health, ln)

	return st
}
