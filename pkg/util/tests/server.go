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

package tests

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/go-driver"
	httpdriver "github.com/arangodb/go-driver/http"
)

func NewServer(t *testing.T) Server {
	s := &server{
		t:       t,
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
		started: make(chan struct{}),
		done:    make(chan struct{}),
	}

	go s.run()

	<-s.started

	return s
}

type Server interface {
	NewConnection() driver.Connection
	NewClient() driver.Client

	Handle(f http.HandlerFunc)
	Addr() string
	Stop()
}

type server struct {
	lock sync.Mutex

	t *testing.T

	handlers []http.HandlerFunc

	port int

	stop, stopped, started, done chan struct{}
}

func (s *server) NewClient() driver.Client {
	c, err := driver.NewClient(driver.ClientConfig{
		Connection: s.NewConnection(),
	})
	require.NoError(s.t, err)

	return c
}

func (s *server) NewConnection() driver.Connection {
	c, err := httpdriver.NewConnection(httpdriver.ConnectionConfig{
		Endpoints: []string{
			s.Addr(),
		},
		ContentType: driver.ContentTypeJSON,
	})
	require.NoError(s.t, err)

	return c
}

func (s *server) Handle(f http.HandlerFunc) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.handlers = append(s.handlers, f)
}

func (s *server) Addr() string {
	return fmt.Sprintf("http://127.0.0.1:%d", s.port)
}

func (s *server) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()

	close(s.stop)

	<-s.done

	if q := len(s.handlers); q > 0 {
		require.Failf(s.t, "Pending messages", "Count %d", q)
	}
}

func (s *server) run() {
	defer close(s.done)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(s.t, err)

	s.port = listener.Addr().(*net.TCPAddr).Port

	m := http.NewServeMux()

	m.HandleFunc("/", s.handle)

	server := http.Server{
		Handler: m,
	}

	var serverErr error

	go func() {
		defer close(s.stopped)
		close(s.started)

		go func() {
			<-s.stop
			require.NoError(s.t, server.Close())
		}()

		serverErr = server.Serve(listener)
	}()

	<-s.stopped

	if serverErr != http.ErrServerClosed {
		require.NoError(s.t, serverErr)
	}
}
func (s *server) handle(writer http.ResponseWriter, request *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()

	var handler http.HandlerFunc

	switch len(s.handlers) {
	case 0:
		require.Fail(s.t, "No pending messages")
	case 1:
		handler = s.handlers[0]
		s.handlers = nil
	default:
		handler = s.handlers[0]
		s.handlers = s.handlers[1:]
	}

	handler(writer, request)
}

func NewSimpleHandler(t *testing.T, method string, path string, resp func(t *testing.T) (int, interface{})) http.HandlerFunc {
	return NewCustomRequestHandler(t, method, path, nil, nil, resp)
}

func NewCustomRequestHandler(t *testing.T, method string, path string, reqVerify func(t *testing.T, r *http.Request), respHeaders func(t *testing.T) map[string]string, resp func(t *testing.T) (int, interface{})) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, method, request.Method)
		require.Equal(t, path, request.RequestURI)

		if reqVerify != nil {
			reqVerify(t, request)
		}

		code, r := resp(t)

		writer.Header().Add("content-type", "application/json")
		if respHeaders != nil {
			h := respHeaders(t)

			for k, v := range h {
				writer.Header().Add(k, v)
			}
		}

		writer.WriteHeader(code)

		if r != nil {
			d, err := json.Marshal(r)
			require.NoError(t, err)

			s, err := writer.Write(d)
			require.NoError(t, err)
			require.Equal(t, len(d), s)
		}
	}
}
