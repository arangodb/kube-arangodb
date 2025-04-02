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

func NewCache[K comparable, T any](extract CacheExtract[K, T], ttl time.Duration) Cache[K, T] {
	return &cache[K, T]{
		ttl:     ttl,
		items:   map[K]cacheItem[T]{},
		extract: extract,
	}
}

type CacheExtract[K comparable, T any] func(ctx context.Context, in K) (T, error)

type Cache[K comparable, T any] interface {
	Get(ctx context.Context, key K) (T, error)
}

type cache[K comparable, T any] struct {
	lock sync.Mutex

	ttl time.Duration

	items map[K]cacheItem[T]

	extract CacheExtract[K, T]
}

func (c *cache[K, T]) Get(ctx context.Context, key K) (T, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if v, ok := c.items[key]; ok {
		if time.Since(v.created) <= c.ttl {
			return v.Object, nil
		}
	}

	el, err := c.extract(ctx, key)
	if err != nil {
		return util.Default[T](), err
	}

	c.items[key] = cacheItem[T]{
		created: time.Now(),
		Object:  el,
	}

	return el, nil
}

type cacheItem[T any] struct {
	created time.Time

	Object T
}
