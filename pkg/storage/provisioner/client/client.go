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

package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// New creates a new client for the provisioner API.
func New(endpoint string) (provisioner.API, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	u.Path = ""
	return &client{
		endpoint: *u,
	}, nil
}

type client struct {
	endpoint url.URL
}

const (
	defaultHTTPTimeout = time.Minute * 2
)

var (
	httpClient = &http.Client{
		Timeout: defaultHTTPTimeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 90 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
)

// GetNodeInfo fetches information from the current node.
func (c *client) GetNodeInfo(ctx context.Context) (provisioner.NodeInfo, error) {
	req, err := c.newRequest("GET", "/nodeinfo", nil)
	if err != nil {
		return provisioner.NodeInfo{}, errors.WithStack(err)
	}
	var result provisioner.NodeInfo
	if err := c.do(ctx, req, &result); err != nil {
		return provisioner.NodeInfo{}, errors.WithStack(err)
	}
	return result, nil
}

// GetInfo fetches information from the filesystem containing
// the given local path.
func (c *client) GetInfo(ctx context.Context, localPath string) (provisioner.Info, error) {
	input := provisioner.Request{
		LocalPath: localPath,
	}
	req, err := c.newRequest("POST", "/info", input)
	if err != nil {
		return provisioner.Info{}, errors.WithStack(err)
	}
	var result provisioner.Info
	if err := c.do(ctx, req, &result); err != nil {
		return provisioner.Info{}, errors.WithStack(err)
	}
	return result, nil
}

// Prepare a volume at the given local path
func (c *client) Prepare(ctx context.Context, localPath string) error {
	input := provisioner.Request{
		LocalPath: localPath,
	}
	req, err := c.newRequest("POST", "/prepare", input)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := c.do(ctx, req, nil); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Remove a volume with the given local path
func (c *client) Remove(ctx context.Context, localPath string) error {
	input := provisioner.Request{
		LocalPath: localPath,
	}
	req, err := c.newRequest("POST", "/remove", input)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := c.do(ctx, req, nil); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// newRequest creates a new request with optional body and context
// Returns: request, cancel, error
func (c *client) newRequest(method string, localPath string, body interface{}) (*http.Request, error) {
	var encoded []byte
	if body != nil {
		var err error
		encoded, err = json.Marshal(body)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	var bodyRd io.Reader
	if encoded != nil {
		bodyRd = bytes.NewReader(encoded)
	}
	url := c.endpoint
	url.Path = localPath
	req, err := http.NewRequest(method, url.String(), bodyRd)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return req, nil
}

// do performs the given request and parses the result.
func (c *client) do(ctx context.Context, req *http.Request, result interface{}) error {
	req = req.WithContext(ctx)
	resp, err := httpClient.Do(req)
	if err != nil {
		// Request failed
		return errors.WithStack(err)
	}

	// Read content
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.WithStack(err)
	}

	// Check status
	statusCode := resp.StatusCode
	if statusCode != 200 {
		if err := provisioner.ParseResponseError(resp, body); err != nil {
			return errors.WithStack(err)
		}
		return errors.WithStack(errors.Newf("Invalid status %d", statusCode))
	}

	// Got a success status
	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			method := resp.Request.Method
			url := resp.Request.URL.String()
			return errors.Wrapf(err, "Failed decoding response data from %s request to %s: %v", method, url, err)
		}
	}
	return nil
}
