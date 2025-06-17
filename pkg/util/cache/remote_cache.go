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

	Expires() time.Time
}

type RemoteCacheObjectRev interface {
	RemoteCacheObject

	GetRev() string
}

func GetRemoteCacheObjectRev(in RemoteCacheObject) string {
	if v, ok := in.(RemoteCacheObjectRev); ok {
		return v.GetRev()
	}

	return ""
}

func NewRemoteCacheWithTTL[T RemoteCacheObject](collection Object[arangodb.Collection], ttl time.Duration) RemoteCache[T] {
	r := &remoteCache[T]{
		collection: collection,
	}

	r.cache = NewCacheWithTTL[string, T](r.cacheRead, ttl)

	return r
}

func NewRemoteCache[T RemoteCacheObject](collection Object[arangodb.Collection]) RemoteCache[T] {
	return NewRemoteCacheWithTTL[T](collection, time.Minute)
}

type RemoteCache[T RemoteCacheObject] interface {
	// Put puts the key in the cache
	Put(ctx context.Context, key string, obj T) error

	// Get gets the key from the cache
	// Returns T, Exists, Error
	Get(ctx context.Context, key string) (T, bool, error)

	// Remove removed the key from the cache
	// Returns Removed, Error
	Remove(ctx context.Context, key string) (bool, error)

	// Invalidate invalidates internal cache
	Invalidate(ctx context.Context, key string)
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

	if _, err := client.UpdateDocumentWithOptions(ctx, key, obj, &arangodb.CollectionDocumentUpdateOptions{
		// Ignore the revision if it is not set
		IgnoreRevs: util.NewType(GetRemoteCacheObjectRev(obj) == ""),
	}); err != nil {
		if !shared.IsNotFound(err) {
			return err
		}

		if _, err := client.CreateDocumentWithOptions(ctx, obj, nil); err != nil {
			return err
		}
	}

	return nil
}

func (r *remoteCache[T]) cacheRead(ctx context.Context, key string) (T, time.Time, error) {
	client, err := r.collection.Get(ctx)
	if err != nil {
		return util.Default[T](), util.Default[time.Time](), err
	}

	var z T
	if _, err := client.ReadDocument(ctx, key, &z); err != nil {
		return util.Default[T](), util.Default[time.Time](), err
	}

	return z, z.Expires(), nil
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

func (r *remoteCache[T]) Invalidate(ctx context.Context, key string) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	r.cache.Invalidate(key)
}

func (r *remoteCache[T]) Remove(ctx context.Context, key string) (bool, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	client, err := r.collection.Get(ctx)
	if err != nil {
		return false, err
	}

	if _, err := client.DeleteDocumentWithOptions(ctx, key, &arangodb.CollectionDocumentDeleteOptions{}); err != nil {
		if !shared.IsNotFound(err) {
			return false, err
		}

		return false, nil
	}

	r.cache.Invalidate(key)

	return true, nil
}
