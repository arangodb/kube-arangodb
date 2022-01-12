//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package http

import (
	"crypto/tls"
	"errors"
	"net/http"
	"sync"

	"github.com/arangodb-helper/go-certificates"
)

func NewServer(server *http.Server) PlainServer {
	return &plainServer{server: server}
}

type ServerRunner interface {
	Stop()
	Wait() error
}

type PlainServer interface {
	Server
	WithSSL(key, cert string) (Server, error)
	WithKeyfile(keyfile string) (Server, error)
}

type Server interface {
	Start() (ServerRunner, error)
}

var _ Server = &tlsServer{}

type tlsServer struct {
	server *http.Server
}

func (t *tlsServer) Start() (ServerRunner, error) {
	i := serverRunner{
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}

	go i.run(t.server, func(s *http.Server) error {
		return s.ListenAndServeTLS("", "")
	})

	return &i, nil
}

var _ PlainServer = &plainServer{}

type plainServer struct {
	server *http.Server
}

func (p *plainServer) WithKeyfile(keyfile string) (Server, error) {
	certificate, err := certificates.LoadKeyFile(keyfile)
	if err != nil {
		return nil, err
	}

	s := p.server
	s.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}

	return &tlsServer{server: s}, nil
}

func (p *plainServer) Start() (ServerRunner, error) {
	i := serverRunner{
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}

	go i.run(p.server, func(s *http.Server) error {
		return s.ListenAndServe()
	})

	return &i, nil
}

func (p *plainServer) WithSSL(key, cert string) (Server, error) {
	certificate, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	s := p.server
	s.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}

	return &tlsServer{server: s}, nil
}

var _ ServerRunner = &serverRunner{}

type serverRunner struct {
	lock sync.Mutex

	stopCh chan struct{}
	doneCh chan struct{}
	err    error
}

func (s *serverRunner) run(server *http.Server, f func(s *http.Server) error) {
	go func() {
		defer close(s.doneCh)
		if err := f(server); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				s.err = err
			}
		}
	}()

	<-s.stopCh

	server.Close()

	<-s.doneCh
}

func (s *serverRunner) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()
	select {
	case <-s.stopCh:
		return
	default:
		close(s.stopCh)
	}
}

func (s *serverRunner) Wait() error {
	<-s.doneCh

	return s.err
}
