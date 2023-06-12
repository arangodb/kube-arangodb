//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package conn

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"

	"github.com/arangodb/kube-arangodb/pkg/util/metrics/nctx"
)

func NewExecutor[IN, OUT interface{}](conn Connection) Executor[IN, OUT] {
	return executor[IN, OUT]{
		conn: conn,
	}
}

type executor[IN, OUT interface{}] struct {
	conn Connection
}

func (e executor[IN, OUT]) ExecuteGet(ctx context.Context, endpoint string) (*OUT, int, error) {
	var t IN
	return e.Execute(ctx, http.MethodGet, endpoint, t)
}

func (e executor[IN, OUT]) Execute(ctx context.Context, method string, endpoint string, in IN) (*OUT, int, error) {
	var reader io.Reader
	if q := reflect.ValueOf(in); q.IsValid() && !q.IsZero() && !q.IsNil() {
		data, err := json.Marshal(in)
		if err != nil {
			return nil, 0, err
		}

		reader = bytes.NewReader(data)
	}

	resp, code, err := e.conn.Execute(ctx, method, endpoint, reader)
	if err != nil {
		return nil, 0, err
	}

	if resp == nil {
		return nil, code, nil
	}

	defer resp.Close()

	var out OUT

	if err := json.NewDecoder(nctx.WithRequestReadBytes(ctx, resp)).Decode(&out); err != nil {
		return nil, 0, err
	}

	return &out, code, err
}

type Executor[IN, OUT interface{}] interface {
	ExecuteGet(ctx context.Context, endpoint string) (*OUT, int, error)
	Execute(ctx context.Context, method string, endpoint string, in IN) (*OUT, int, error)
}
