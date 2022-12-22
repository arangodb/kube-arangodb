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

package exporter

import (
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

var _ http.Handler = &passthru{}

func NewPassthru(arangodbEndpoint string, auth Authentication, sslVerify bool, timeout time.Duration) (http.Handler, error) {
	return &passthru{
		factory: newHttpClientFactory(arangodbEndpoint, auth, sslVerify, timeout),
	}, nil
}

type httpClientFactory func() (*http.Client, *http.Request, error)

func newHttpClientFactory(arangodbEndpoint string, auth Authentication, sslVerify bool, timeout time.Duration) httpClientFactory {
	return func() (*http.Client, *http.Request, error) {
		transport := &http.Transport{}

		req, err := http.NewRequest("GET", arangodbEndpoint, nil)
		if err != nil {
			return nil, nil, errors.WithStack(err)
		}

		if !sslVerify {
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}

		jwt, err := auth()
		if err != nil {
			return nil, nil, err
		}

		if jwt != "" {
			hdr, err := CreateArangodJwtAuthorizationHeader(jwt)
			if err != nil {
				return nil, nil, errors.WithStack(err)
			}
			req.Header.Add("Authorization", hdr)
		}

		req.Header.Add("x-arango-allow-dirty-read", "true") // Allow read from follower in AF mode

		client := &http.Client{
			Transport: transport,
			Timeout:   timeout,
		}

		return client, req, nil
	}
}

type passthru struct {
	factory httpClientFactory
}

func (p passthru) get() (*http.Response, error) {
	c, req, err := p.factory()
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (p passthru) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	data, err := p.get()

	if err != nil {
		// Ignore error
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(err.Error()))
		return
	}

	if data.Body == nil {
		// Ignore error
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte("Body is empty"))
		return
	}

	defer data.Body.Close()

	response, err := io.ReadAll(data.Body)
	if err != nil {
		// Ignore error
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(err.Error()))
		return
	}

	responseStr := string(response)

	// Fix Header response
	responseStr = strings.ReplaceAll(responseStr, "guage", "gauge")

	// Attach monitor data
	monitorData := currentMembersStatus.Load()
	if monitorData != nil {
		responseStr = responseStr + monitorData.(string)
	}

	resp.WriteHeader(data.StatusCode)
	_, err = resp.Write([]byte(responseStr))
	if err != nil {
		// Ignore error
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte("Unable to write body"))
		return
	}
}
