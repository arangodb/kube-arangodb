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
	"fmt"
	"net/http"
	"testing"

	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
)

func NewAsyncHandler(t *testing.T, s Server, method string, path string, retCode int, ret interface{}) AsyncHandler {
	return &asyncHandler{
		t:       t,
		s:       s,
		ret:     ret,
		retCode: retCode,
		method:  method,
		path:    path,
		id:      uniuri.NewLen(32),
	}
}

type AsyncHandler interface {
	ID() string

	Start()
	InProgress()
	Missing()
	Done()
}

type asyncHandler struct {
	t *testing.T
	s Server

	ret     interface{}
	retCode int

	method, path string

	id string
}

func (a *asyncHandler) Missing() {
	p := fmt.Sprintf("/_api/job/%s", a.id)

	a.s.Handle(NewCustomRequestHandler(a.t, http.MethodPut, p, func(t *testing.T, r *http.Request) {
		v := r.Header.Get(constants.ArangoHeaderAsyncKey)
		require.Equal(t, "", v)
	}, nil, func(t *testing.T) (int, interface{}) {
		return http.StatusNotFound, nil
	}))
}

func (a *asyncHandler) Start() {
	a.s.Handle(NewCustomRequestHandler(a.t, a.method, a.path, func(t *testing.T, r *http.Request) {
		v := r.Header.Get(constants.ArangoHeaderAsyncKey)
		require.Equal(t, constants.ArangoHeaderAsyncValue, v)
	}, func(t *testing.T) map[string]string {
		return map[string]string{
			constants.ArangoHeaderAsyncIDKey: a.id,
		}
	}, func(t *testing.T) (int, interface{}) {
		return http.StatusAccepted, nil
	}))
}

func (a *asyncHandler) InProgress() {
	p := fmt.Sprintf("/_api/job/%s", a.id)

	a.s.Handle(NewCustomRequestHandler(a.t, http.MethodPut, p, func(t *testing.T, r *http.Request) {
		v := r.Header.Get(constants.ArangoHeaderAsyncKey)
		require.Equal(t, "", v)
	}, nil, func(t *testing.T) (int, interface{}) {
		return http.StatusNoContent, nil
	}))
}

func (a *asyncHandler) Done() {
	p := fmt.Sprintf("/_api/job/%s", a.id)

	a.s.Handle(NewCustomRequestHandler(a.t, http.MethodPut, p, func(t *testing.T, r *http.Request) {
		v := r.Header.Get(constants.ArangoHeaderAsyncKey)
		require.Equal(t, "", v)
	}, nil, func(t *testing.T) (int, interface{}) {
		return a.retCode, a.ret
	}))
}

func (a *asyncHandler) ID() string {
	return a.id
}
