//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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
			r, ok := mgr.Get(repo)
			require.True(t, ok)

			t.Run("Latest", func(t *testing.T) {
				v, ok := r.Latest()
				require.True(t, ok)

				vdata, err := v.Get(context.Background())
				require.NoError(t, err)

				t.Logf("Chart %s, Version %s, Release %s", repo, v.Chart().Version, v.Chart().Created.String())

				vchart, err := vdata.Get()
				require.NoError(t, err)
				require.NotNil(t, vchart.Chart().Metadata)

				vl, ok := r.Get("latest")
				require.True(t, ok)

				vldata, err := vl.Get(context.Background())
				require.NoError(t, err)

				chart, err := vldata.Get()
				require.NoError(t, err)
				require.NotNil(t, chart.Chart().Metadata)

				require.EqualValues(t, v.Chart().Version, vchart.Chart().Metadata.Version)
				require.EqualValues(t, v.Chart().Version, chart.Chart().Metadata.Version)
			})
			t.Run("ByVersion", func(t *testing.T) {
				for _, version := range limitArray(r.Versions(), 5) {
					t.Run(version, func(t *testing.T) {
						v, ok := r.Get(version)
						require.True(t, ok)

						data, err := v.Get(context.Background())
						require.NoError(t, err)

						c, err := data.Get()
						require.NoError(t, err)

						require.NotNil(t, c.Chart().Metadata)
						require.EqualValues(t, version, c.Chart().Metadata.Version)
					})
				}
			})
		})
	}

	require.NoError(t, mgr.Reload(context.Background()))
}
