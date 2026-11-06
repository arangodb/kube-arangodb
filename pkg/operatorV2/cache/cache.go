//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package cache

import (
	"sync"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
)

// ObjectCache is a concurrency-safe store of values keyed by operation.ItemName (namespace/name). It lets
// a handler keep a value (for example a chart) across reconciles instead of re-fetching it from the API
// server every time. Entries expire after the cache's TTL, so a value that stops being refreshed is
// eventually dropped rather than served stale forever.
type ObjectCache[V any] interface {
	// Get returns the cached value for name and whether a non-expired entry was present.
	Get(name operation.ItemName) (V, bool)

	// Set stores value for name, replacing any existing entry and resetting its expiry to now+TTL.
	Set(name operation.ItemName, value V)

	// Invalidate drops the cached entry for name.
	Invalidate(name operation.ItemName)

	// List returns all currently cached, non-expired values in an unspecified order.
	List() []V

	// Clear drops all cached entries.
	Clear()
}

type entry[V any] struct {
	value   V
	expires time.Time
}

type cache[V any] struct {
	lock sync.Mutex

	ttl     time.Duration
	entries map[operation.ItemName]entry[V]
}

// NewObjectCache returns an empty ObjectCache whose entries expire ttl after they are Set. A ttl of zero
// or less means entries never expire.
func NewObjectCache[V any](ttl time.Duration) ObjectCache[V] {
	return &cache[V]{
		ttl:     ttl,
		entries: map[operation.ItemName]entry[V]{},
	}
}

func (c *cache[V]) Get(name operation.ItemName) (V, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	e, ok := c.entries[name]
	if !ok {
		var zero V
		return zero, false
	}

	if !e.expires.IsZero() && time.Now().After(e.expires) {
		delete(c.entries, name)
		var zero V
		return zero, false
	}

	return e.value, true
}

func (c *cache[V]) Set(name operation.ItemName, value V) {
	c.lock.Lock()
	defer c.lock.Unlock()

	var expires time.Time
	if c.ttl > 0 {
		expires = time.Now().Add(c.ttl)
	}

	c.entries[name] = entry[V]{
		value:   value,
		expires: expires,
	}
}

func (c *cache[V]) Invalidate(name operation.ItemName) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.entries, name)
}

func (c *cache[V]) List() []V {
	c.lock.Lock()
	defer c.lock.Unlock()

	now := time.Now()

	values := make([]V, 0, len(c.entries))
	for name, e := range c.entries {
		if !e.expires.IsZero() && now.After(e.expires) {
			delete(c.entries, name)
			continue
		}

		values = append(values, e.value)
	}

	return values
}

func (c *cache[V]) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.entries = map[operation.ItemName]entry[V]{}
}
