//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	"context"
	"net/http"
	"path"

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ConnectionWrap func(c driver.Connection) driver.Connection

func NewAsyncConnection(c driver.Connection) driver.Connection {
	return async{
		connectionPass: connectionPass{
			c:    c,
			wrap: asyncConnectionWrap,
		},
	}
}

func asyncConnectionWrap(c driver.Connection) (driver.Connection, error) {
	return NewAsyncConnection(c), nil
}

type async struct {
	connectionPass
}

func (a async) isAsyncIDSet(ctx context.Context) (string, bool) {
	if ctx != nil {
		if q := ctx.Value(asyncOperatorContextKey); q != nil {
			if v, ok := q.(string); ok {
				return v, true
			}
		}
	}

	return "", false
}

func (a async) Do(ctx context.Context, req driver.Request) (driver.Response, error) {
	if id, ok := a.isAsyncIDSet(ctx); ok {
		// We have ID Set, request should be done to fetch job id
		req, err := a.c.NewRequest(http.MethodPut, path.Join("/_api/job", id))
		if err != nil {
			return nil, err
		}

		resp, err := a.c.Do(ctx, req)
		if err != nil {
			return nil, err
		}

		switch resp.StatusCode() {
		case http.StatusNotFound:
			return nil, newAsyncErrorNotFound(id)
		case http.StatusNoContent:
			asyncID := resp.Header(constants.ArangoHeaderAsyncIDKey)
			if asyncID == id {
				// Job is done
				return resp, nil
			}

			// Job is in progress
			return nil, newAsyncJobInProgress(id)
		default:
			return resp, nil
		}
	} else {
		req.SetHeader(constants.ArangoHeaderAsyncKey, constants.ArangoHeaderAsyncValue)

		resp, err := a.c.Do(ctx, req)
		if err != nil {
			return nil, err
		}

		switch resp.StatusCode() {
		case http.StatusAccepted:
			if v := resp.Header(constants.ArangoHeaderAsyncIDKey); len(v) == 0 {
				return nil, errors.Newf("Missing async key response")
			} else {
				return nil, newAsyncJobInProgress(v)
			}
		default:
			return nil, resp.CheckStatus(http.StatusAccepted)
		}
	}
}
