//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package helm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func limitArray[T any](in []T, max int) []T {
	if len(in) <= max {
		return in
	}

	return in[:max]
}

func Test_Manager(t *testing.T) {
	mgr, err := NewChartManager(context.Background(), nil, "https://arangodb-platform-dev-chart-registry.s3.amazonaws.com")
	require.NoError(t, err)

	for _, repo := range limitArray(mgr.Repositories(), 5) {
		t.Run(repo, func(t *testing.T) {
			t.Run("Latest", func(t *testing.T) {
				v, ok := mgr.Latest(repo)
				require.True(t, ok)

				vdata, err := mgr.Chart(context.Background(), repo, v)
				require.NoError(t, err)

				vchart, err := vdata.Get()
				require.NoError(t, err)
				require.NotNil(t, vchart.Metadata)

				data, err := mgr.Chart(context.Background(), repo, "latest")
				require.NoError(t, err)

				chart, err := data.Get()
				require.NoError(t, err)
				require.NotNil(t, chart.Metadata)

				require.EqualValues(t, v, vchart.Metadata.Version)
				require.EqualValues(t, v, chart.Metadata.Version)
			})
			t.Run("ByVersion", func(t *testing.T) {
				for _, version := range limitArray(mgr.Versions(repo), 5) {
					t.Run(version, func(t *testing.T) {
						data, err := mgr.Chart(context.Background(), repo, version)
						require.NoError(t, err)

						c, err := data.Get()
						require.NoError(t, err)

						require.NotNil(t, c.Metadata)
						require.EqualValues(t, version, c.Metadata.Version)
					})
				}
			})
		})
	}

	require.NoError(t, mgr.Reload(context.Background()))
}
