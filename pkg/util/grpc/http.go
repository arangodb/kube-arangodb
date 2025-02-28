//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package grpc

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"google.golang.org/protobuf/proto"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
)

type HTTPResponse[T proto.Message] interface {
	WithCode(codes ...int) HTTPResponse[T]
	Get() (T, error)
}

type httpErrorResponse[T proto.Message] struct {
	err error
}

func (h httpErrorResponse[T]) WithCode(codes ...int) HTTPResponse[T] {
	return h
}

func (h httpErrorResponse[T]) Get() (T, error) {
	return util.Default[T](), h.err
}

type httpResponse[T proto.Message] struct {
	code int

	data []byte
}

func (h httpResponse[T]) WithCode(codes ...int) HTTPResponse[T] {
	for _, code := range codes {
		if h.code == code {
			return h
		}
	}

	return httpErrorResponse[T]{err: errors.Errorf("Unexpected code: %d", h.code)}
}

func (h httpResponse[T]) Get() (T, error) {
	return Unmarshal[T](h.data)
}

func request[T proto.Message](ctx context.Context, client operatorHTTP.HTTPClient, method, url string, body io.Reader, mods ...util.Mod[http.Request]) HTTPResponse[T] {
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return httpErrorResponse[T]{err: err}
	}

	util.ApplyMods(req, mods...)

	resp, err := client.Do(req)
	if err != nil {
		return httpErrorResponse[T]{err: err}
	}

	defer resp.Body.Close()

	nData, err := io.ReadAll(resp.Body)
	if err != nil {
		return httpErrorResponse[T]{err: err}
	}

	return httpResponse[T]{
		code: resp.StatusCode,
		data: nData,
	}
}

func Get[T proto.Message](ctx context.Context, client operatorHTTP.HTTPClient, url string, mods ...util.Mod[http.Request]) HTTPResponse[T] {
	return request[T](ctx, client, http.MethodGet, url, nil, mods...)
}

func Post[IN, T proto.Message](ctx context.Context, client operatorHTTP.HTTPClient, in IN, url string, mods ...util.Mod[http.Request]) HTTPResponse[T] {
	data, err := Marshal(in)
	if err != nil {
		return httpErrorResponse[T]{err: err}
	}

	return request[T](ctx, client, http.MethodPost, url, bytes.NewReader(data), mods...)
}
