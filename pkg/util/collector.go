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
	"errors"
	"sync"

	"golang.org/x/sync/errgroup"
)

// ErrCollectorDone is returned by Collector.Run when the collector has already been closed via Done.
var ErrCollectorDone = errors.New("collector is done")

// Pusher is the producer side of a Collector: it accepts items pushed from one or more producers.
type Pusher[T any] interface {
	// Push appends items to the collector. It is safe to call from multiple goroutines.
	Push(items ...T)
}

// Collector aggregates items pushed from one or more producers into a single list.
//
// Producers either Push items directly or are launched in the background with Run. Done waits for
// every background producer to finish and returns the first error, after which Collect returns the
// aggregated list.
//
// Once Done has been called the collector is closed: Run reports ErrCollectorDone and no further
// producers can be started. Done and Collect block until the producers have actually finished, so
// they never return a partial result.
type Collector[T any] interface {
	Pusher[T]

	// Run launches fn in the background, passing the collector as its pusher. The error returned by fn
	// is reported by Done. Run returns ErrCollectorDone if the collector has already been closed.
	Run(fn func(p Pusher[T]) error) error

	// Done closes the collector, waits for every background producer to finish and returns the first
	// error. It is idempotent - it only ever closes once and repeated calls return the same error.
	Done() error

	// Collect returns all pushed items. It blocks until Done has completed.
	Collect() []T
}

// NewCollector creates a Collector. A background goroutine drains pushed items until Done is called,
// so Push never blocks and is safe to call concurrently.
func NewCollector[T any]() Collector[T] {
	c := &collector[T]{
		in:      make(chan T),
		drained: make(chan struct{}),
	}

	go func() {
		for v := range c.in {
			c.items = append(c.items, v)
		}
		close(c.drained)
	}()

	return c
}

type collector[T any] struct {
	in      chan T
	drained chan struct{}

	g errgroup.Group

	lock   sync.Mutex
	closed bool

	once sync.Once
	err  error

	// items is only ever touched by the draining goroutine, and read by Collect after drained is closed.
	items []T
}

func (c *collector[T]) Push(items ...T) {
	for _, i := range items {
		c.in <- i
	}
}

func (c *collector[T]) Run(fn func(p Pusher[T]) error) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.closed {
		return ErrCollectorDone
	}

	c.g.Go(func() error {
		return fn(c)
	})

	return nil
}

func (c *collector[T]) Done() error {
	c.once.Do(func() {
		// Close the collector so no further producers can be started.
		c.lock.Lock()
		c.closed = true
		c.lock.Unlock()

		// Wait for every background producer to finish pushing.
		c.err = c.g.Wait()

		// No more items will be pushed - close the input and wait for the drain to complete.
		close(c.in)
		<-c.drained
	})

	return c.err
}

func (c *collector[T]) Collect() []T {
	// Block until the drain has completed, so a partial result is never returned.
	<-c.drained
	return c.items
}
