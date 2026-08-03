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

package util

import (
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCollector_Empty(t *testing.T) {
	c := NewCollector[int]()

	require.NoError(t, c.Done())
	require.Empty(t, c.Collect())
}

func TestCollector_DirectPush(t *testing.T) {
	c := NewCollector[int]()

	c.Push(1)
	c.Push(2, 3)

	require.NoError(t, c.Done())

	got := c.Collect()
	sort.Ints(got)
	require.Equal(t, []int{1, 2, 3}, got)
}

func TestCollector_Run(t *testing.T) {
	c := NewCollector[int]()

	require.NoError(t, c.Run(func(p Pusher[int]) error {
		p.Push(1, 2)
		return nil
	}))
	require.NoError(t, c.Run(func(p Pusher[int]) error {
		p.Push(3, 4)
		return nil
	}))

	require.NoError(t, c.Done())

	got := c.Collect()
	sort.Ints(got)
	require.Equal(t, []int{1, 2, 3, 4}, got)
}

func TestCollector_RunConcurrent(t *testing.T) {
	const producers = 16
	const perProducer = 64

	c := NewCollector[int]()

	for i := 0; i < producers; i++ {
		require.NoError(t, c.Run(func(p Pusher[int]) error {
			for j := 0; j < perProducer; j++ {
				p.Push(1)
			}
			return nil
		}))
	}

	require.NoError(t, c.Done())

	got := c.Collect()
	require.Len(t, got, producers*perProducer)

	sum := 0
	for _, v := range got {
		sum += v
	}
	require.Equal(t, producers*perProducer, sum)
}

func TestCollector_Done_ReturnsFirstError(t *testing.T) {
	c := NewCollector[int]()

	require.NoError(t, c.Run(func(p Pusher[int]) error {
		p.Push(1)
		return nil
	}))
	require.NoError(t, c.Run(func(p Pusher[int]) error {
		return errTest
	}))

	require.ErrorIs(t, c.Done(), errTest)

	// Even on error, the events that were pushed are still collected.
	require.Equal(t, []int{1}, c.Collect())
}

func TestCollector_Done_Idempotent(t *testing.T) {
	c := NewCollector[int]()

	require.NoError(t, c.Run(func(p Pusher[int]) error {
		return errTest
	}))

	// Repeated Done calls return the same error and do not panic (double close).
	require.ErrorIs(t, c.Done(), errTest)
	require.ErrorIs(t, c.Done(), errTest)
	require.ErrorIs(t, c.Done(), errTest)
}

func TestCollector_RunAfterDone(t *testing.T) {
	c := NewCollector[int]()

	require.NoError(t, c.Done())

	err := c.Run(func(p Pusher[int]) error {
		p.Push(1)
		return nil
	})
	require.ErrorIs(t, err, ErrCollectorDone)

	// The rejected producer did not push anything.
	require.Empty(t, c.Collect())
}

func TestCollector_CollectBlocksUntilDone(t *testing.T) {
	c := NewCollector[int]()

	require.NoError(t, c.Run(func(p Pusher[int]) error {
		p.Push(1, 2)
		return nil
	}))

	collected := make(chan []int, 1)
	go func() {
		collected <- c.Collect()
	}()

	// Collect must not return before Done is called.
	select {
	case <-collected:
		require.Fail(t, "Collect returned before Done")
	case <-time.After(50 * time.Millisecond):
	}

	require.NoError(t, c.Done())

	select {
	case got := <-collected:
		sort.Ints(got)
		require.Equal(t, []int{1, 2}, got)
	case <-time.After(time.Second):
		require.Fail(t, "Collect did not return after Done")
	}
}

func TestCollector_DoneConcurrent(t *testing.T) {
	c := NewCollector[int]()

	require.NoError(t, c.Run(func(p Pusher[int]) error {
		p.Push(1)
		return nil
	}))

	// Many concurrent Done calls must all return without panicking on a double close.
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			require.NoError(t, c.Done())
		}()
	}
	wg.Wait()

	require.Equal(t, []int{1}, c.Collect())
}

var errTest = testError("test error")

type testError string

func (e testError) Error() string { return string(e) }
