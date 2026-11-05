//
// DISCLAIMER
//
// Copyright 2023-2026 ArangoDB GmbH, Cologne, Germany
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
	"compress/gzip"
	"fmt"
	"io"
	goHttp "net/http"
	"testing"

	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"

	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
)

// metricFamilyNames parses the Prometheus text-format metrics from in and returns the set of metric
// family names.
func metricFamilyNames(t *testing.T, in io.Reader) map[string]bool {
	parser := expfmt.NewTextParser(model.UTF8Validation)
	mfs, err := parser.TextToMetricFamilies(in)
	require.NoError(t, err)

	names := map[string]bool{}
	for name := range mfs {
		names[name] = true
	}
	return names
}

func Test_Handler(t *testing.T) {
	m := goHttp.NewServeMux()

	m.HandleFunc("/metrics", Handler())
	m.HandleFunc("/empty", operatorHTTP.WithNoContent(func(writer goHttp.ResponseWriter, request *goHttp.Request) {

	}))

	endpoint, c := StartHTTP(t, m)
	defer c()

	metricsEndpoint := fmt.Sprintf("%s/metrics", endpoint)
	emptyEndpoint := fmt.Sprintf("%s/empty", endpoint)

	t.Run("Get metrics in plain", func(t *testing.T) {
		r, err := goHttp.NewRequest("GET", metricsEndpoint, nil)
		require.NoError(t, err)

		r.Header.Add("Accept-Encoding", "identity")

		resp, err := goHttp.DefaultClient.Do(r)
		require.NoError(t, err)

		require.Equal(t, goHttp.StatusOK, resp.StatusCode)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.True(t, len(data) > 0)
	})

	t.Run("Get metrics in gzip", func(t *testing.T) {
		r, err := goHttp.NewRequest("GET", metricsEndpoint, nil)
		require.NoError(t, err)

		r.Header.Add("Accept-Encoding", "gzip")

		resp, err := goHttp.DefaultClient.Do(r)
		require.NoError(t, err)

		require.Equal(t, goHttp.StatusOK, resp.StatusCode)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.True(t, len(data) > 0)
	})

	t.Run("Get metrics in default", func(t *testing.T) {
		r, err := goHttp.NewRequest("GET", metricsEndpoint, nil)
		require.NoError(t, err)

		resp, err := goHttp.DefaultClient.Do(r)
		require.NoError(t, err)

		require.Equal(t, goHttp.StatusOK, resp.StatusCode)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.True(t, len(data) > 0)
	})

	t.Run("Get empty", func(t *testing.T) {
		r, err := goHttp.NewRequest("GET", emptyEndpoint, nil)
		require.NoError(t, err)

		resp, err := goHttp.DefaultClient.Do(r)
		require.NoError(t, err)

		require.Equal(t, goHttp.StatusNoContent, resp.StatusCode)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.True(t, len(data) == 0)
	})

	t.Run("Read metrics - plain", func(t *testing.T) {
		r, err := goHttp.NewRequest("GET", metricsEndpoint, nil)
		require.NoError(t, err)

		r.Header.Add("Accept-Encoding", "identity")

		resp, err := goHttp.DefaultClient.Do(r)
		require.NoError(t, err)

		require.Equal(t, goHttp.StatusOK, resp.StatusCode)

		metrics := metricFamilyNames(t, resp.Body)

		require.Contains(t, metrics, "go_info")
	})

	t.Run("Read metrics - gzip", func(t *testing.T) {
		r, err := goHttp.NewRequest("GET", metricsEndpoint, nil)
		require.NoError(t, err)

		r.Header.Add("Accept-Encoding", "gzip")

		resp, err := goHttp.DefaultClient.Do(r)
		require.NoError(t, err)

		require.Equal(t, goHttp.StatusOK, resp.StatusCode)

		reader, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		metrics := metricFamilyNames(t, reader)

		require.Contains(t, metrics, "go_info")
	})

	t.Run("Read metrics - default", func(t *testing.T) {
		r, err := goHttp.NewRequest("GET", metricsEndpoint, nil)
		require.NoError(t, err)

		resp, err := goHttp.DefaultClient.Do(r)
		require.NoError(t, err)

		require.Equal(t, goHttp.StatusOK, resp.StatusCode)

		metrics := metricFamilyNames(t, resp.Body)

		require.Contains(t, metrics, "go_info")
	})
}
