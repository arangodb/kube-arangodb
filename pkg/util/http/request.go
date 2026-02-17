//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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
	"bytes"
	"context"
	"encoding/json"
	"io"
	goHttp "net/http"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Response[T any] interface {
	WithCode(codes ...int) Response[T]
	Data() ([]byte, error)
	Get() (T, error)
	Validate() error
}

type httpErrorResponse[T any] struct {
	err error
}

func (h httpErrorResponse[T]) Data() ([]byte, error) {
	return nil, h.err
}

func (h httpErrorResponse[T]) WithCode(codes ...int) Response[T] {
	return h
}

func (h httpErrorResponse[T]) Get() (T, error) {
	return util.Default[T](), h.err
}

func (h httpErrorResponse[T]) Validate() error {
	return h.err
}

func request[T any, ERR error](ctx context.Context, client HTTPClient, method, url string, body io.Reader, mods ...util.Mod[goHttp.Request]) Response[T] {
	if client == nil {
		client = goHttp.DefaultClient
	}

	req, err := goHttp.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return newResponseError[T, ERR](err, nil)
	}

	util.ApplyMods(req, mods...)

	resp, err := client.Do(req)
	if err != nil {
		return newResponseError[T, ERR](err, nil)
	}

	defer resp.Body.Close()

	nData, err := io.ReadAll(resp.Body)
	if err != nil {
		return newResponseError[T, ERR](err, nil)
	}

	return httpResponse[T, ERR]{
		code: resp.StatusCode,
		data: nData,
	}
}

type httpResponse[T any, ERR error] struct {
	code int

	data []byte
}

func (h httpResponse[T, ERR]) Data() ([]byte, error) {
	return h.data, nil
}

func (h httpResponse[T, ERR]) Validate() error {
	return nil
}

func (h httpResponse[T, ERR]) WithCode(codes ...int) Response[T] {
	for _, code := range codes {
		if h.code == code {
			return h
		}
	}

	return newResponseError[T, ERR](errors.Errorf("Unexpected code: %d", h.code), h.data)
}

func (h httpResponse[T, ERR]) Get() (T, error) {
	return util.JsonOrYamlUnmarshal[T](h.data)
}

func newResponseError[T any, ERR error](err error, data []byte) Response[T] {
	if len(data) == 0 {
		return httpErrorResponse[T]{err: err}
	}

	var d ERR

	if nerr := json.Unmarshal(data, &d); nerr != nil {
		return httpErrorResponse[T]{err: errors.Wrapf(errors.Errorf("failed to parse body: %v", nerr), "%v", err)}
	}

	return httpErrorResponse[T]{errors.Wrapf(d, "%v", err)}
}

type DataError []byte

func (d *DataError) UnmarshalJSON(i []byte) error {
	q := make([]byte, len(i))

	copy(q, i)

	*d = q

	return nil
}

func (d *DataError) Error() string {
	return string(*d)
}

func Get[T any, ERR error](ctx context.Context, client HTTPClient, url string, mods ...util.Mod[goHttp.Request]) Response[T] {
	return request[T, ERR](ctx, client, goHttp.MethodGet, url, nil, mods...)
}

func Post[IN, T any, ERR error](ctx context.Context, client HTTPClient, in IN, url string, mods ...util.Mod[goHttp.Request]) Response[T] {
	data, err := json.Marshal(in)
	if err != nil {
		return newResponseError[T, ERR](err, nil)
	}

	return request[T, ERR](ctx, client, goHttp.MethodPost, url, bytes.NewReader(data), mods...)
}
