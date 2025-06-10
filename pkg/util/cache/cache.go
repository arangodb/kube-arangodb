//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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
	"context"
	"sync"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func NewCacheWithTTL[K comparable, T any](extract CacheExtract[K, T], maxTTL time.Duration) Cache[K, T] {
	return &cache[K, T]{
		items:   map[K]cacheItem[T]{},
		extract: extract,
		maxTTL:  maxTTL,
	}
}

func NewCache[K comparable, T any](extract CacheExtract[K, T]) Cache[K, T] {
	return NewCacheWithTTL[K, T](extract, 0)
}

type CacheExtract[K comparable, T any] func(ctx context.Context, in K) (T, time.Time, error)

type Cache[K comparable, T any] interface {
	Get(ctx context.Context, key K) (T, error)
	Invalidate(key K)
}

type cache[K comparable, T any] struct {
	lock sync.Mutex

	items map[K]cacheItem[T]

	extract CacheExtract[K, T]

	maxTTL time.Duration
}

func (c *cache[K, T]) Get(ctx context.Context, key K) (T, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if v, ok := c.items[key]; ok {
		if v.until.After(time.Now()) {
			return v.Object, nil
		}
	}

	el, expires, err := c.extract(ctx, key)
	if err != nil {
		return util.Default[T](), err
	}

	if c.maxTTL > 0 {
		if time.Until(expires) > c.maxTTL {
			expires = time.Now().Add(c.maxTTL)
		}
	}

	if expires.After(time.Now()) {
		c.items[key] = cacheItem[T]{
			until:  expires,
			Object: el,
		}
	}

	return el, nil
}

func (c *cache[K, T]) Invalidate(key K) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.items, key)
}

type cacheItem[T any] struct {
	until time.Time

	Object T
}
