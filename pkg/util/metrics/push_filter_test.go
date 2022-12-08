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

package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func Test_PushFilter(t *testing.T) {
	m11 := NewDescription("arangodb_a1_metric1", "", []string{}, nil)
	m12 := NewDescription("arangodb_a1_metric2", "", []string{}, nil)
	m21 := NewDescription("arangodb_a2_metric1", "", []string{}, nil)
	m22 := NewDescription("arangodb_a2_metric2", "", []string{}, nil)

	push := func(in PushMetric) {
		in.Push(m11.Gauge(1), m12.Gauge(1), m21.Gauge(1), m22.Gauge(1))
	}

	t.Run("AllAccepted", func(t *testing.T) {
		c := make(chan prometheus.Metric, 1024)

		push(NewPushMetric(c))

		require.Len(t, c, 4)
	})

	t.Run("Filter - AcceptAll", func(t *testing.T) {
		c := make(chan prometheus.Metric, 1024)

		push(NewMetricsPushFilter(NewPushMetric(c), func(m Metric) bool {
			return true
		}))

		require.Len(t, c, 4)
	})

	t.Run("Filter - Prefix - Empty", func(t *testing.T) {
		c := make(chan prometheus.Metric, 1024)

		push(NewMetricsPushFilter(NewPushMetric(c), NewPrefixMetricPushFilter()))

		require.Len(t, c, 0)
	})

	t.Run("Filter - Prefix - Match one", func(t *testing.T) {
		c := make(chan prometheus.Metric, 1024)

		push(NewMetricsPushFilter(NewPushMetric(c), NewPrefixMetricPushFilter("arangodb_a2_metric1")))

		require.Len(t, c, 1)
	})

	t.Run("Filter - Prefix - Match two", func(t *testing.T) {
		c := make(chan prometheus.Metric, 1024)

		push(NewMetricsPushFilter(NewPushMetric(c), NewPrefixMetricPushFilter("arangodb_a2_metric1", "arangodb_a1_metric1")))

		require.Len(t, c, 2)
	})

	t.Run("Filter - Prefix - Match multi", func(t *testing.T) {
		c := make(chan prometheus.Metric, 1024)

		push(NewMetricsPushFilter(NewPushMetric(c), NewPrefixMetricPushFilter("arangodb_a2_")))

		require.Len(t, c, 2)
	})

	t.Run("Filter - Prefix - Match one - Negate", func(t *testing.T) {
		c := make(chan prometheus.Metric, 1024)

		push(NewMetricsPushFilter(NewPushMetric(c), NegateMetricPushFilter(NewPrefixMetricPushFilter("arangodb_a2_metric1"))))

		require.Len(t, c, 3)
	})
}
