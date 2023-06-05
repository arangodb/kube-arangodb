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
	"context"
	"fmt"
	"io"
	"net/http"
)

type Connection interface {
	Execute(ctx context.Context, method string, endpoint string, body io.Reader) (io.ReadCloser, int, error)
}

type connection struct {
	client *http.Client

	auth *string

	host string
}

func (c connection) Execute(ctx context.Context, method string, endpoint string, body io.Reader) (io.ReadCloser, int, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.host, endpoint), body)
	if err != nil {
		return nil, 0, err
	}

	req = req.WithContext(ctx)

	if a := c.auth; a != nil {
		req.Header.Add("Authorization", *a)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	if b := resp.Body; b != nil {
		return b, resp.StatusCode, nil
	}

	return nil, resp.StatusCode, nil
}
