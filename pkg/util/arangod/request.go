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

package arangod

import (
	"context"
	"encoding/json"
	"fmt"
	goHttp "net/http"
	goStrings "strings"
	"time"

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Response[OUT any] interface {
	AcceptCode(codes ...int) Response[OUT]

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
	resp driver.Response
}

func (r response[OUT]) Code() int {
	return r.resp.StatusCode()
}

func (r response[OUT]) AcceptCode(codes ...int) Response[OUT] {
	for _, code := range codes {
		if r.resp.StatusCode() == code {
			return r
		}
	}

	var data string
	var obj = map[string]interface{}{}
	if err := r.resp.ParseBody("", &obj); err != nil {
		data = fmt.Sprintf("Error: %s", err.Error())
	} else {
		if dz, err := json.Marshal(obj); err != nil {
			data = fmt.Sprintf("Error: %s", err.Error())
		} else {
			data = fmt.Sprintf("Data: %s", string(dz))
		}
	}

	return NewResponseError[OUT](errors.Errorf("Code %d not allowed in expected status codes: %s. Body: %s", r.resp.StatusCode(), goStrings.Join(util.FormatList(codes, func(a int) string {
		return fmt.Sprintf("%d", a)
	}), ", "), data))
}

func (r response[OUT]) Response() (OUT, error) {
	var d OUT

	if err := r.resp.ParseBody("", &d); err != nil {
		return util.Default[OUT](), err
	}

	return d, nil
}

func (r response[OUT]) Evaluate() error {
	return nil
}

func newPath(path ...string) string {
	return fmt.Sprintf("/%s", goStrings.Join(path, "/"))
}

func GetRequestWithTimeout[OUT any](ctx context.Context, timeout time.Duration, conn driver.Connection, path ...string) Response[OUT] {
	nctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return GetRequest[OUT](nctx, conn, path...)
}

func GetRequest[OUT any](ctx context.Context, conn driver.Connection, path ...string) Response[OUT] {
	req, err := conn.NewRequest(goHttp.MethodGet, newPath(path...))
	if err != nil {
		return NewResponseError[OUT](err)
	}

	resp, err := conn.Do(ctx, req)
	if err != nil {
		return NewResponseError[OUT](err)
	}

	return response[OUT]{resp: resp}
}

func PostRequestWithTimeout[IN, OUT any](ctx context.Context, timeout time.Duration, conn driver.Connection, body IN, path ...string) Response[OUT] {
	nctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return PostRequest[IN, OUT](nctx, conn, body, path...)
}

func PostRequest[IN, OUT any](ctx context.Context, conn driver.Connection, body IN, path ...string) Response[OUT] {
	req, err := conn.NewRequest(goHttp.MethodPost, newPath(path...))
	if err != nil {
		return NewResponseError[OUT](err)
	}

	if r, err := req.SetBody(body); err != nil {
		return NewResponseError[OUT](err)
	} else {
		req = r
	}

	resp, err := conn.Do(ctx, req)
	if err != nil {
		return NewResponseError[OUT](err)
	}

	return response[OUT]{resp: resp}
}
