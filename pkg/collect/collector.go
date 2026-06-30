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
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

// Metric is a single event body value: a key with a float value. Collectors push metrics, which the
// collector assembles into the body of the emitted event.
type Metric = util.KV[string, float32]

// ECollector is a single source of values of type T. It pushes every value it produces to the
// provided pusher and returns an error if it failed. Returning an error fails the whole collection,
// which is then retried on the next interval - so a collector should only return an error when it
// could not collect, not when it simply has nothing to push.
type ECollector[T any] interface {
	CollectEvents(out util.Pusher[T]) error
}

// Collector is a registry of ECollector[T]. Collectors register once (typically from an init
// function) and Collect fans them into a single list.
type Collector[T any] interface {
	// Register adds a collector to the registry.
	Register(c ECollector[T])

	// Collect runs every registered collector and returns the aggregated values, or the first
	// collector error.
	Collect() ([]T, error)
}

// NewCollector creates an empty Collector registry.
func NewCollector[T any]() Collector[T] {
	return &collector[T]{}
}

var registry = NewCollector[Metric]()

// GetCollector returns the global registry of event metric collectors.
func GetCollector() Collector[Metric] {
	return registry
}

type collector[T any] struct {
	lock sync.Mutex

	collectors []ECollector[T]
}

func (c *collector[T]) Register(e ECollector[T]) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.collectors = append(c.collectors, e)
}

// Collect runs every registered collector concurrently in the background, each pushing its values
// into a shared collector. It waits until all of them have completed, then returns the aggregated
// list (or the first collector error).
func (c *collector[T]) Collect() ([]T, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	out := util.NewCollector[T]()

	for _, e := range c.collectors {
		e := e
		if err := out.Run(func(p util.Pusher[T]) error {
			return e.CollectEvents(p)
		}); err != nil {
			return nil, err
		}
	}

	if err := out.Done(); err != nil {
		return nil, err
	}

	return out.Collect(), nil
}
