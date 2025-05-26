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

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/arangodb/shared"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type RemoteCacheObject interface {
	SetKey(string)
	GetKey() string
}

func NewRemoteCache[T RemoteCacheObject](collection Object[arangodb.Collection]) RemoteCache[T] {
	r := &remoteCache[T]{
		collection: collection,
	}

	r.cache = NewCache[string, T](r.cacheRead, 15*time.Minute)

	return r
}

type RemoteCache[T RemoteCacheObject] interface {
	Put(ctx context.Context, key string, obj T) error

	Get(ctx context.Context, key string) (T, bool, error)
}

type remoteCache[T RemoteCacheObject] struct {
	collection Object[arangodb.Collection]

	lock sync.RWMutex

	cache Cache[string, T]
}

func (r *remoteCache[T]) Put(ctx context.Context, key string, obj T) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if k := obj.GetKey(); k != key {
		return errors.Errorf("Invalid key in the object. Got %s, expected %s", k, key)
	}

	client, err := r.collection.Get(ctx)
	if err != nil {
		return err
	}

	if _, err := client.UpdateDocumentWithOptions(ctx, key, obj, nil); err != nil {
		if !shared.IsNotFound(err) {
			return err
		}

		if _, err := client.CreateDocumentWithOptions(ctx, obj, nil); err != nil {
			return err
		}
	}

	return nil
}

func (r *remoteCache[T]) cacheRead(ctx context.Context, key string) (T, error) {
	client, err := r.collection.Get(ctx)
	if err != nil {
		return util.Default[T](), err
	}

	var z T
	if _, err := client.ReadDocument(ctx, key, &z); err != nil {
		return util.Default[T](), err
	}

	return z, nil
}

func (r *remoteCache[T]) Get(ctx context.Context, key string) (T, bool, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	obj, err := r.cache.Get(ctx, key)
	if err != nil {
		if shared.IsNotFound(err) {
			return util.Default[T](), false, nil
		}

		return util.Default[T](), false, err
	}

	return obj, true, nil
}
