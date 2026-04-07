//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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

package arangod

import (
	"context"
	"fmt"
	"io"
	goHttp "net/http"
	"path"
	goStrings "strings"
	"sync"
	"time"

	adbDriverV2Connection "github.com/arangodb/go-driver/v2/connection"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Response[OUT any] interface {
	AcceptCode(codes ...int) Response[OUT]

	HTTPResponse() (*goHttp.Response, error)

	Response() (OUT, error)

	Code() int

	Evaluate() error
}

type request[OUT any] struct {
	lock sync.Mutex

	conn    adbDriverV2Connection.Connection
	request adbDriverV2Connection.Request

	// timeout, if non-zero, is applied to the context passed to Do.
	timeout time.Duration
}

func (r *request[OUT]) Query(key, value string) Request[OUT] {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.request.AddQuery(key, value)

	return r
}

func (r *request[OUT]) Do(ctx context.Context) Response[OUT] {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
		defer cancel()
	}

	resp, closer, err := r.conn.Stream(ctx, r.request)
	if err != nil {
		return NewResponseError[OUT](err)
	}

	defer closer.Close()

	data, err := io.ReadAll(closer)
	if err != nil {
		return NewResponseError[OUT](err)
	}

	return response[OUT]{resp: resp, data: data}
}

type Request[OUT any] interface {
	Query(key, value string) Request[OUT]

	Do(ctx context.Context) Response[OUT]
}

func NewResponseError[OUT any](err error) Response[OUT] {
	return responseError[OUT]{err: err}
}
func NewRequestError[OUT any](err error) Request[OUT] {
	return responseError[OUT]{err: err}
}

type responseError[OUT any] struct {
	err error
}

func (r responseError[OUT]) Query(key, value string) Request[OUT] {
	return r
}

func (r responseError[OUT]) Do(ctx context.Context) Response[OUT] {
	return r
}

func (r responseError[OUT]) HTTPResponse() (*goHttp.Response, error) {
	return nil, r.err
}

func (r responseError[OUT]) Code() int {
	return goHttp.StatusInternalServerError
}

func (r responseError[OUT]) Response() (OUT, error) {
	return util.Default[OUT](), r.err
}

func (r responseError[OUT]) Evaluate() error {
	return r.err
}

func (r responseError[OUT]) AcceptCode(codes ...int) Response[OUT] {
	return r
}

type response[OUT any] struct {
	resp adbDriverV2Connection.Response
	data []byte
}

func (r response[OUT]) HTTPResponse() (*goHttp.Response, error) {
	if u := r.resp.RawResponse(); u != nil {
		return u, nil
	}

	return nil, errors.Errorf("Missing HTTP Response Body")
}

func (r response[OUT]) Code() int {
	return r.resp.Code()
}

func (r response[OUT]) AcceptCode(codes ...int) Response[OUT] {
	if err := EvaluateCode(r.resp.Code(), codes...); err != nil {
		return NewResponseError[OUT](err)
	}

	return r
}

func (r response[OUT]) Response() (OUT, error) {
	return util.Unmarshal[OUT](r.data)
}

func (r response[OUT]) Evaluate() error {
	return nil
}

func newPath(p ...string) string {
	return path.Clean(fmt.Sprintf("/%s", goStrings.Join(p, "/")))
}

func NewRequest[IN, OUT any](ctx context.Context, conn adbDriverV2Connection.Connection, method string, body IN, path ...string) Request[OUT] {
	req, err := conn.NewRequest(method, newPath(path...))
	if err != nil {
		return NewRequestError[OUT](err)
	}

	if !util.IsDefault(body) {
		if err := req.SetBody(body); err != nil {
			return NewRequestError[OUT](err)
		}
	}

	return &request[OUT]{request: req, conn: conn}
}

func withTimeout[OUT any](r Request[OUT], timeout time.Duration) Request[OUT] {
	if rq, ok := r.(*request[OUT]); ok {
		rq.timeout = timeout
	}
	return r
}

func GetRequestWithTimeout[OUT any](ctx context.Context, timeout time.Duration, conn adbDriverV2Connection.Connection, path ...string) Request[OUT] {
	return withTimeout(GetRequest[OUT](ctx, conn, path...), timeout)
}

func GetRequest[OUT any](ctx context.Context, conn adbDriverV2Connection.Connection, path ...string) Request[OUT] {
	return NewRequest[any, OUT](ctx, conn, goHttp.MethodGet, nil, path...)
}

func DeleteRequestWithTimeout[OUT any](ctx context.Context, timeout time.Duration, conn adbDriverV2Connection.Connection, path ...string) Request[OUT] {
	return withTimeout(DeleteRequest[OUT](ctx, conn, path...), timeout)
}

func DeleteRequest[OUT any](ctx context.Context, conn adbDriverV2Connection.Connection, path ...string) Request[OUT] {
	return NewRequest[any, OUT](ctx, conn, goHttp.MethodDelete, nil, path...)
}

func PostRequestWithTimeout[IN, OUT any](ctx context.Context, timeout time.Duration, conn adbDriverV2Connection.Connection, body IN, path ...string) Request[OUT] {
	return withTimeout(PostRequest[IN, OUT](ctx, conn, body, path...), timeout)
}

func PostRequest[IN, OUT any](ctx context.Context, conn adbDriverV2Connection.Connection, body IN, path ...string) Request[OUT] {
	return NewRequest[IN, OUT](ctx, conn, goHttp.MethodPost, body, path...)
}

func PutRequestWithTimeout[IN, OUT any](ctx context.Context, timeout time.Duration, conn adbDriverV2Connection.Connection, body IN, path ...string) Request[OUT] {
	return withTimeout(PutRequest[IN, OUT](ctx, conn, body, path...), timeout)
}

func PutRequest[IN, OUT any](ctx context.Context, conn adbDriverV2Connection.Connection, body IN, path ...string) Request[OUT] {
	return NewRequest[IN, OUT](ctx, conn, goHttp.MethodPut, body, path...)
}
