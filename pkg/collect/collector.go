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

	pbEventsV1 "github.com/arangodb/kube-arangodb/integrations/events/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

const (
	// serviceID identifies the collector as the source of the emitted events.
	serviceID = "collector"

	// dimensionBootID is the event dimension carrying the unique boot identifier.
	dimensionBootID = "bootID"
)

// Event is a single event produced by a Collector. It is the events integration Event message.
type Event = pbEventsV1.Event

// ECollector is a single source of events. It pushes every event it produces to the provided pusher
// and returns an error if it failed. Returning an error fails the whole collection, which is then
// retried on the next interval - so a collector should only return an error when it could not
// collect, not when it simply has nothing to push.
type ECollector interface {
	CollectEvents(out util.Pusher[*Event]) error
}

// Collector is the registry of all event collectors. Collectors register once (typically from an
// init function) and Collect fans them into a single list.
type Collector interface {
	// Register adds a collector to the registry.
	Register(c ECollector)

	// Collect runs every registered collector and returns the aggregated events, or the first
	// collector error.
	Collect() ([]*Event, error)
}

var collectorObject = &collector{}

// GetCollector returns the global collector registry.
func GetCollector() Collector {
	return collectorObject
}

type collector struct {
	lock sync.Mutex

	collectors []ECollector
}

func (c *collector) Register(e ECollector) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.collectors = append(c.collectors, e)
}

// Collect runs every registered collector concurrently in the background, each pushing its events
// into a shared collector. It waits until all of them have completed, then returns the aggregated
// list (or the first collector error).
func (c *collector) Collect() ([]*Event, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	out := util.NewCollector[*Event]()

	for _, e := range c.collectors {
		e := e
		if err := out.Run(func(p util.Pusher[*Event]) error {
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
