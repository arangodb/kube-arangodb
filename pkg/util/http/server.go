//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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
	"context"
	"crypto/tls"
	"net"
	goHttp "net/http"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func DefaultHTTPServerSettings(in *goHttp.Server, _ context.Context) error {
	in.ReadTimeout = time.Second * 30
	in.ReadHeaderTimeout = time.Second * 15
	in.WriteTimeout = time.Second * 30
	in.TLSNextProto = make(map[string]func(*goHttp.Server, *tls.Conn, goHttp.Handler))

	return nil
}

func WithTLSConfigFetcherGen(gen func() util.TLSConfigFetcher) util.ModEP1[goHttp.Server, context.Context] {
	return WithTLSConfigFetcher(gen())
}

func WithTLSConfigFetcher(fetcher util.TLSConfigFetcher) util.ModEP1[goHttp.Server, context.Context] {
	return func(in *goHttp.Server, p1 context.Context) error {
		v, err := fetcher.Eval(p1)
		if err != nil {
			return err
		}

		in.TLSConfig = v

		return nil
	}
}

func WithServeMux(mods ...util.Mod[goHttp.ServeMux]) util.ModEP1[goHttp.Server, context.Context] {
	return func(in *goHttp.Server, p1 context.Context) error {
		mux := goHttp.NewServeMux()

		util.ApplyMods(mux, mods...)

		in.Handler = mux

		return nil
	}
}

func NewServer(ctx context.Context, mods ...util.ModEP1[goHttp.Server, context.Context]) (Server, error) {
	var sv goHttp.Server

	if err := util.ApplyModsEP1(&sv, ctx, mods...); err != nil {
		return nil, err
	}

	return &server{
		server: &sv,
	}, nil
}

type Server interface {
	AsyncAddr(ctx context.Context, addr string) func() error
	Async(ctx context.Context, ln net.Listener) func() error

	StartAddr(ctx context.Context, addr string) error
	Start(ctx context.Context, ln net.Listener) error
}

type server struct {
	server *goHttp.Server
}

func (s *server) AsyncAddr(ctx context.Context, addr string) func() error {
	var err error

	done := make(chan any)

	go func() {
		defer close(done)

		err = s.StartAddr(ctx, addr)
	}()

	return func() error {
		<-done

		return err
	}
}

func (s *server) Async(ctx context.Context, ln net.Listener) func() error {
	var err error

	done := make(chan any)

	go func() {
		defer close(done)

		err = s.Start(ctx, ln)
	}()

	return func() error {
		<-done

		return err
	}
}

func (s *server) StartAddr(ctx context.Context, addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return s.Start(ctx, ln)
}

func (s *server) Start(ctx context.Context, ln net.Listener) error {
	go func() {
		<-ctx.Done()

		if err := s.server.Close(); err != nil {
			if !errors.Is(err, goHttp.ErrServerClosed) {
				logger.Err(err).Warn("Unable to close server")
			}
		}
	}()

	if s.server.TLSConfig == nil {
		if err := s.server.Serve(ln); err != nil {
			if !errors.Is(err, goHttp.ErrServerClosed) {
				return err
			}
		}
	} else {
		if err := s.server.ServeTLS(ln, "", ""); err != nil {
			if !errors.Is(err, goHttp.ErrServerClosed) {
				return err
			}
		}
	}

	return nil
}
