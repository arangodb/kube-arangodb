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

func NewHashCache[K Hash, T any](extract HashCacheExtract[K, T]) HashCache[K, T] {
	return &hashCache[K, T]{
		items:   map[string]cacheItem[T]{},
		extract: extract,
	}
}

type Hash interface {
	Hash() string
}

type HashCacheExtract[K Hash, T any] func(ctx context.Context, in K) (T, time.Time, error)

type HashCache[K Hash, T any] interface {
	Get(ctx context.Context, key K) (T, error)
}

type hashCache[K Hash, T any] struct {
	lock sync.Mutex

	items map[string]cacheItem[T]

	extract HashCacheExtract[K, T]
}

func (c *hashCache[K, T]) Get(ctx context.Context, key K) (T, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if v, ok := c.items[key.Hash()]; ok {
		if v.until.After(time.Now()) {
			return v.Object, nil
		}
	}

	el, expires, err := c.extract(ctx, key)
	if err != nil {
		return util.Default[T](), err
	}

	if !expires.After(time.Now()) {
		c.items[key.Hash()] = cacheItem[T]{
			until:  expires,
			Object: el,
		}
	}

	return el, nil
}
