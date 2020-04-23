//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package tls

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/arangodb/go-driver"
)

type KeyFile struct {
	PrivateKeyChecksum string `json:"privateKeySHA256,omitempty"`
	Checksum string `json:"SHA256,omitempty"`
	Certificates []string `json:"certificates,omitempty"`
}

type DetailsResult struct {
	KeyFile KeyFile `json:"keyfile,omitempty"`
}

type Details struct {
	Result DetailsResult `json:"result,omitempty"`
}

func NewClient(c driver.Connection) Client {
	return &client{
		c: c,
	}
}

type Client interface {
	GetTLS(ctx context.Context) (Details, string, error)
}

type client struct {
	c driver.Connection
}

func (c *client) parseResponse(response driver.Response) (Details, string, error) {
	if err := response.CheckStatus(http.StatusOK); err != nil {
		return Details{}, "", err
	}

	var result map[string]interface{}

	if err := response.ParseBody("", &result); err != nil {
		return Details{}, "", err
	}

	resultData, err := json.Marshal(result)
	if err != nil {
		return Details{}, "", err
	}

	var d Details
	if err := json.Unmarshal(resultData, &d); err != nil {
		return Details{}, "", err
	}

	return d, string(resultData), nil
}

func (c *client) GetTLS(ctx context.Context) (Details, string, error) {
	r, err := c.c.NewRequest(http.MethodGet, "/_admin/server/tls")
	if err != nil {
		return Details{}, "", err
	}

	response, err := c.c.Do(ctx, r)
	if err != nil {
		return Details{}, "", err
	}

	d, j, err := c.parseResponse(response)
	if err != nil {
		return Details{}, "", err
	}

	return d, j, nil
}
