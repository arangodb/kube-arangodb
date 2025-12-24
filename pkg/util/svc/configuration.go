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
	"crypto/tls"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

// RequestWrap if returns true execution is stopped
type RequestWrap func(w http.ResponseWriter, r *http.Request) bool

func (r RequestWrap) Wrap(handler http.Handler) http.Handler {
	if r == nil {
		return handler
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if r(w, req) {
			return
		}

		handler.ServeHTTP(w, req)
	})
}

type RequestWraps []RequestWrap

func (r RequestWraps) Wrap(handler http.Handler) http.Handler {
	for id := len(r) - 1; id >= 0; id-- {
		handler = r[id].Wrap(handler)
	}
	return handler
}

type Configuration struct {
	Address string

	TLSOptions util.TLSConfigFetcher

	Options []grpc.ServerOption

	Wrap RequestWraps

	MuxExtensions []runtime.ServeMuxOption

	Gateway *ConfigurationGateway
}

type ConfigurationGateway struct {
	Address string
}

func (c *Configuration) GetTLSOptions(ctx context.Context) (*tls.Config, error) {
	if z := c.TLSOptions; z != nil {
		if tls, err := z.Eval(ctx); err != nil {
			return nil, err
		} else if tls != nil {
			return tls, nil
		}
	}
	return nil, nil
}

func (c *Configuration) RenderOptions() ([]grpc.ServerOption, error) {
	if c == nil {
		return nil, nil
	}

	ret := make([]grpc.ServerOption, len(c.Options))
	copy(ret, c.Options)

	return ret, nil
}
