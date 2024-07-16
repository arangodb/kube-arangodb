//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
)

var _ http.Handler = &passthru{}

func NewPassthru(auth Authentication, sslVerify bool, timeout time.Duration, endpoints ...string) (http.Handler, error) {
	return &passthru{
		factory:   newHttpClientFactory(auth, sslVerify, timeout),
		endpoints: endpoints,
	}, nil
}

type httpClientFactory func(endpoint string) (*http.Client, *http.Request, error)

func newHttpClientFactory(auth Authentication, sslVerify bool, timeout time.Duration) httpClientFactory {
	return func(endpoint string) (*http.Client, *http.Request, error) {
		transport := operatorHTTP.Transport(operatorHTTP.WithTransportTLS(util.BoolSwitch(sslVerify, operatorHTTP.Insecure, nil)))

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			return nil, nil, errors.WithStack(err)
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
	endpoints []string
	factory   httpClientFactory
}

func (p passthru) get(endpoint string) (*http.Response, error) {
	c, req, err := p.factory(endpoint)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (p passthru) read(endpoint string) (string, error) {
	data, err := p.get(endpoint)

	if err != nil {
		return "", err
	}

	if data.Body == nil {
		return "", err
	}

	defer data.Body.Close()

	response, err := io.ReadAll(data.Body)
	if err != nil {
		return "", err
	}

	responseStr := string(response)

	// Fix Header response
	return strings.ReplaceAll(responseStr, "guage", "gauge"), nil
}

func (p passthru) getAll() (string, error) {
	errs := make([]error, len(p.endpoints))
	responses := make([]string, len(p.endpoints))

	var wg sync.WaitGroup

	for id := range p.endpoints {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()
			responses[id], errs[id] = p.read(p.endpoints[id])
		}(id)
	}

	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return "", errors.WithStack(err)
		}
	}

	response := strings.Join(responses, "\n")

	// Attach monitor data
	monitorData := currentMembersStatus.Load()
	if monitorData != nil {
		response = response + monitorData.(string)
	}

	return response, nil
}

func (p passthru) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	response, err := p.getAll()

	if err != nil {
		// Ignore error
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(err.Error()))
		return
	}

	resp.WriteHeader(http.StatusOK)
	_, err = resp.Write([]byte(response))
	if err != nil {
		// Ignore error
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte("Unable to write body"))
		return
	}
}
