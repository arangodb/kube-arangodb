//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package agency

import (
	"context"
	"sync"
	"time"

	agencyCache "github.com/arangodb/kube-arangodb/pkg/deployment/agency/cache"
)

func InvalidateOnErrorLoader[T interface{}](loader agencyCache.StateLoader[T]) agencyCache.StateLoader[T] {
	return &invalidateOnErrorLoader[T]{
		parent: loader,
	}
}

type invalidateOnErrorLoader[T interface{}] struct {
	lock sync.Mutex

	parent agencyCache.StateLoader[T]
}

func (i *invalidateOnErrorLoader[T]) UpdateTime() time.Time {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.parent.UpdateTime()
}

func (i *invalidateOnErrorLoader[T]) Valid() bool {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.parent.Valid()
}

func (i *invalidateOnErrorLoader[T]) State() (*T, uint64, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.parent.State()
}

func (i *invalidateOnErrorLoader[T]) Invalidate() {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.parent.Invalidate()
}

func (i *invalidateOnErrorLoader[T]) Refresh(ctx context.Context, discovery agencyCache.LeaderDiscovery) (err error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	defer func() {
		if err != nil {
			i.parent.Invalidate()
		}
	}()

	defer func() {
		if p := recover(); p != nil {
			i.parent.Invalidate()
			panic(p)
		}
	}()

	return i.parent.Refresh(ctx, discovery)
}
