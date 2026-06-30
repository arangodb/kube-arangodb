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
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

// fakeCollector is a test ECollector pushing a fixed set of events, or returning an error.
type fakeCollector struct {
	events []*Event
	err    error
}

func (f fakeCollector) CollectEvents(out util.Pusher[*Event]) error {
	if f.err != nil {
		return f.err
	}
	out.Push(f.events...)
	return nil
}

func TestBootCollector(t *testing.T) {
	out := util.NewCollector[*Event]()

	require.NoError(t, bootCollector{}.CollectEvents(out))
	require.NoError(t, out.Done())

	events := out.Collect()
	require.Len(t, events, 1)
	require.Equal(t, eventTypeBoot, events[0].GetType())
	require.Equal(t, serviceID, events[0].GetServiceId())
	// The boot id and timestamp are stamped centrally, not by the collector.
	require.Nil(t, events[0].GetCreated())
	require.Empty(t, events[0].GetDimensions())
}

func TestRegistry_Collect(t *testing.T) {
	c := &collector{}

	c.Register(fakeCollector{events: []*Event{{Type: "a"}}})
	c.Register(fakeCollector{events: []*Event{{Type: "b"}, {Type: "c"}}})

	events, err := c.Collect()
	require.NoError(t, err)

	types := make([]string, 0, len(events))
	for _, e := range events {
		types = append(types, e.GetType())
	}
	sort.Strings(types)
	require.Equal(t, []string{"a", "b", "c"}, types)
}

func TestRegistry_CollectEmpty(t *testing.T) {
	c := &collector{}

	events, err := c.Collect()
	require.NoError(t, err)
	require.Empty(t, events)
}

func TestRegistry_CollectError(t *testing.T) {
	c := &collector{}

	c.Register(fakeCollector{events: []*Event{{Type: "a"}}})
	c.Register(fakeCollector{err: errBoom})

	events, err := c.Collect()
	require.ErrorIs(t, err, errBoom)
	require.Nil(t, events)
}

func TestStamp(t *testing.T) {
	created := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)

	events := []*Event{
		{Type: "a"},
		{Type: "b", Dimensions: map[string]string{"existing": "value"}},
	}

	stamp(events, "boot-123", created)

	for _, e := range events {
		require.NotNil(t, e.GetCreated())
		require.Equal(t, created, e.GetCreated().AsTime())
		require.Equal(t, "boot-123", e.GetDimensions()[dimensionBootID])
	}

	// Pre-existing dimensions are preserved.
	require.Equal(t, "value", events[1].GetDimensions()["existing"])
}

var errBoom = boomError("boom")

type boomError string

func (e boomError) Error() string { return string(e) }
