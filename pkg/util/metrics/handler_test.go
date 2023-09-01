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

package metrics

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/prom2json"
	"github.com/stretchr/testify/require"

	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
)

func Test_Handler(t *testing.T) {
	m := http.NewServeMux()

	m.HandleFunc("/metrics", Handler())
	m.HandleFunc("/empty", operatorHTTP.WithNoContent(func(writer http.ResponseWriter, request *http.Request) {

	}))

	endpoint, c := StartHTTP(t, m)
	defer c()

	metricsEndpoint := fmt.Sprintf("%s/metrics", endpoint)
	emptyEndpoint := fmt.Sprintf("%s/empty", endpoint)

	t.Run("Get metrics in plain", func(t *testing.T) {
		r, err := http.NewRequest("GET", metricsEndpoint, nil)
		require.NoError(t, err)

		r.Header.Add("Accept-Encoding", "identity")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.True(t, len(data) > 0)
	})

	t.Run("Get metrics in gzip", func(t *testing.T) {
		r, err := http.NewRequest("GET", metricsEndpoint, nil)
		require.NoError(t, err)

		r.Header.Add("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.True(t, len(data) > 0)
	})

	t.Run("Get metrics in default", func(t *testing.T) {
		r, err := http.NewRequest("GET", metricsEndpoint, nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.True(t, len(data) > 0)
	})

	t.Run("Get empty", func(t *testing.T) {
		r, err := http.NewRequest("GET", emptyEndpoint, nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)

		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.True(t, len(data) == 0)
	})

	t.Run("Read metrics", func(t *testing.T) {
		mfChan := make(chan *dto.MetricFamily, 1024*1024)

		r, err := http.NewRequest("GET", metricsEndpoint, nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.NoError(t, prom2json.ParseReader(resp.Body, mfChan))

		metrics := map[string]bool{}

		for mf := range mfChan {
			result := prom2json.NewFamily(mf)
			metrics[result.Name] = true
		}

		require.Contains(t, metrics, "go_info")
	})
}
