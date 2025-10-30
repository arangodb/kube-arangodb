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
	"encoding/json"
	goStrings "strings"
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

func NewRemoteCache[T cache.RemoteCacheObject]() cache.RemoteCache[T] {
	return &localRemoteCache[T]{
		objects: map[string]json.RawMessage{},
	}
}

type localRemoteCache[T cache.RemoteCacheObject] struct {
	lock sync.Mutex

	objects map[string]json.RawMessage
}

func (l *localRemoteCache[T]) Init(ctx context.Context) error {
	return nil
}

func (l *localRemoteCache[T]) Put(ctx context.Context, key string, obj T) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	l.objects[key] = data
	return nil
}

func (l *localRemoteCache[T]) Get(ctx context.Context, key string) (T, bool, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	data, ok := l.objects[key]
	if !ok {
		return util.Default[T](), false, nil
	}

	var obj T

	if err := json.Unmarshal(data, &obj); err != nil {
		return obj, false, err
	}

	return obj, true, nil
}

func (l *localRemoteCache[T]) Remove(ctx context.Context, key string) (bool, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if _, ok := l.objects[key]; ok {
		delete(l.objects, key)
		return true, nil
	}

	return false, nil
}

func (l *localRemoteCache[T]) Invalidate(ctx context.Context, key string) {
	l.lock.Lock()
	defer l.lock.Unlock()

	// Nothing to do with reload
}

func (l *localRemoteCache[T]) List(ctx context.Context, size int, prefix string) (util.NextIterator[[]string], error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	keys := util.MapKeys(l.objects)
	keys = util.FilterList(keys, func(key string) bool {
		return goStrings.HasPrefix(key, prefix)
	})

	return util.NewStaticNextIterator(util.BatchList(size, keys)...), nil
}
