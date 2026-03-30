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

func NewResponseError[OUT any](err error) Response[OUT] {
	return responseError[OUT]{err: err}
}

type responseError[OUT any] struct {
	err error
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

func Request[IN, OUT any](ctx context.Context, conn adbDriverV2Connection.Connection, method string, body IN, path ...string) Response[OUT] {
	req, err := conn.NewRequest(method, newPath(path...))
	if err != nil {
		return NewResponseError[OUT](err)
	}

	if !util.IsDefault(body) {
		if err := req.SetBody(body); err != nil {
			return NewResponseError[OUT](err)
		}
	}

	resp, closer, err := conn.Stream(ctx, req)
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

func GetRequestWithTimeout[OUT any](ctx context.Context, timeout time.Duration, conn adbDriverV2Connection.Connection, path ...string) Response[OUT] {
	nctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return GetRequest[OUT](nctx, conn, path...)
}

func GetRequest[OUT any](ctx context.Context, conn adbDriverV2Connection.Connection, path ...string) Response[OUT] {
	return Request[any, OUT](ctx, conn, goHttp.MethodGet, nil, path...)
}

func DeleteRequestWithTimeout[OUT any](ctx context.Context, timeout time.Duration, conn adbDriverV2Connection.Connection, path ...string) Response[OUT] {
	nctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return DeleteRequest[OUT](nctx, conn, path...)
}

func DeleteRequest[OUT any](ctx context.Context, conn adbDriverV2Connection.Connection, path ...string) Response[OUT] {
	return Request[any, OUT](ctx, conn, goHttp.MethodDelete, nil, path...)
}

func PostRequestWithTimeout[IN, OUT any](ctx context.Context, timeout time.Duration, conn adbDriverV2Connection.Connection, body IN, path ...string) Response[OUT] {
	nctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return PostRequest[IN, OUT](nctx, conn, body, path...)
}

func PostRequest[IN, OUT any](ctx context.Context, conn adbDriverV2Connection.Connection, body IN, path ...string) Response[OUT] {
	return Request[IN, OUT](ctx, conn, goHttp.MethodPost, body, path...)
}

func PutRequestWithTimeout[IN, OUT any](ctx context.Context, timeout time.Duration, conn adbDriverV2Connection.Connection, body IN, path ...string) Response[OUT] {
	nctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return PutRequest[IN, OUT](nctx, conn, body, path...)
}

func PutRequest[IN, OUT any](ctx context.Context, conn adbDriverV2Connection.Connection, body IN, path ...string) Response[OUT] {
	return Request[IN, OUT](ctx, conn, goHttp.MethodPut, body, path...)
}
