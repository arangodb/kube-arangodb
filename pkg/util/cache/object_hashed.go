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
	"context"
	"sync"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func NewObjectHash[T any](caller ObjectHashFetcher[T]) Object[T] {
	return &objectHash[T]{
		caller: caller,
	}
}

type ObjectHashFetcher[T any] func(ctx context.Context, hash *string) (T, string, time.Duration, error)

type objectHash[T any] struct {
	lock sync.Mutex

	caller ObjectHashFetcher[T]

	eol  time.Time
	obj  T
	hash *string
}

func (o *objectHash[T]) Init(ctx context.Context) error {
	_, err := o.Get(ctx)
	return err
}

func (o *objectHash[T]) Get(ctx context.Context) (T, error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if time.Now().After(o.eol) || o.eol.IsZero() {
		obj, hash, ttl, err := o.caller(ctx, o.hash)
		if err != nil {
			return util.Default[T](), err
		}

		if ttl <= 0 {
			return obj, nil
		}

		if v := o.hash; v != nil && hash == *v {
			o.eol = time.Now().Add(ttl)
			return o.obj, nil
		}

		o.obj = obj
		o.eol = time.Now().Add(ttl)
		o.hash = &hash
	}

	return o.obj, nil
}
