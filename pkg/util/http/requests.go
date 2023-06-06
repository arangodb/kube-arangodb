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

package http

import (
	"encoding/json"
	"net/http"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type RequestInvoker interface {
	Do(req *http.Request) (*http.Response, error)
}

func RequestInvoke[T interface{}](invoker RequestInvoker, request *http.Request) (*T, int, error) {
	resp, err := invoker.Do(request)
	if err != nil {
		return nil, 0, err
	}

	if body := resp.Body; body != nil {
		c := util.CloseOnce(body)
		defer c.Close()

		var obj T

		decoder := json.NewDecoder(body)

		if err := decoder.Decode(&obj); err != nil {
			return nil, 0, errors.Wrapf(err, "Unable to decode object")
		}

		if err := c.Close(); err != nil {
			return nil, 0, err
		}

		return &obj, resp.StatusCode, nil
	}

	return nil, resp.StatusCode, nil
}
