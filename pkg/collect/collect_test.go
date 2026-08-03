//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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

package collect

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

// fakeCollector is a test ECollector pushing a fixed set of metrics, or returning an error.
type fakeCollector struct {
	metrics []Metric
	err     error
}

func (f fakeCollector) CollectEvents(out util.Pusher[Metric]) error {
	if f.err != nil {
		return f.err
	}
	out.Push(f.metrics...)
	return nil
}

func TestRegistry_Collect(t *testing.T) {
	c := NewCollector[Metric]()

	c.Register(fakeCollector{metrics: []Metric{{K: "a", V: 1}}})
	c.Register(fakeCollector{metrics: []Metric{{K: "b", V: 2}, {K: "c", V: 3}}})

	metrics, err := c.Collect()
	require.NoError(t, err)

	values := map[string]float32{}
	for _, m := range metrics {
		values[m.K] = m.V
	}
	require.Equal(t, map[string]float32{"a": 1, "b": 2, "c": 3}, values)
}

func TestRegistry_CollectEmpty(t *testing.T) {
	c := NewCollector[Metric]()

	metrics, err := c.Collect()
	require.NoError(t, err)
	require.Empty(t, metrics)
}

func TestRegistry_CollectError(t *testing.T) {
	c := NewCollector[Metric]()

	c.Register(fakeCollector{metrics: []Metric{{K: "a", V: 1}}})
	c.Register(fakeCollector{err: errBoom})

	metrics, err := c.Collect()
	require.ErrorIs(t, err, errBoom)
	require.Nil(t, metrics)
}

func TestBuildEvent(t *testing.T) {
	created := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)

	event := buildEvent([]Metric{{K: "cpu", V: 4}, {K: "memory", V: 1024}}, "boot-123", created)

	require.Equal(t, eventTypeStartup, event.GetType())
	require.Equal(t, serviceID, event.GetServiceId())
	require.Equal(t, created, event.GetCreated().AsTime())
	require.Equal(t, "boot-123", event.GetDimensions()[dimensionBootID])
	require.Equal(t, map[string]float32{"cpu": 4, "memory": 1024}, event.GetBody())
}

func TestBuildEvent_NoMetrics(t *testing.T) {
	created := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)

	event := buildEvent(nil, "boot-123", created)

	require.Equal(t, eventTypeStartup, event.GetType())
	require.Equal(t, "boot-123", event.GetDimensions()[dimensionBootID])
	require.Empty(t, event.GetBody())
}

func TestResourceCollector(t *testing.T) {
	out := util.NewCollector[Metric]()

	require.NoError(t, resourceCollector{}.CollectEvents(out))
	require.NoError(t, out.Done())

	values := map[string]float32{}
	for _, m := range out.Collect() {
		values[m.K] = m.V
	}

	require.Equal(t, float32(runtime.NumCPU()), values[metricCPU])
	require.Greater(t, values[metricMemory], float32(0))
}

func TestTotalMemory(t *testing.T) {
	mem, err := totalMemory()
	require.NoError(t, err)
	require.Greater(t, mem, uint64(0))
}

var errBoom = boomError("boom")

type boomError string

func (e boomError) Error() string { return string(e) }
