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

	agencyCache "github.com/arangodb-helper/go-helper/pkg/arangod/agency/cache"
)

func RetryLoader[T interface{}](loader agencyCache.StateLoader[T], retries int) agencyCache.StateLoader[T] {
	if retries <= 0 {
		return loader
	}

	return &retryLoader[T]{
		parent:  loader,
		retries: retries,
	}
}

type retryLoader[T interface{}] struct {
	lock sync.Mutex

	retries int

	parent agencyCache.StateLoader[T]
}

func (i *retryLoader[T]) UpdateTime() time.Time {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.parent.UpdateTime()
}

func (i *retryLoader[T]) Valid() bool {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.parent.Valid()
}

func (i *retryLoader[T]) State() (*T, uint64, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.parent.State()
}

func (i *retryLoader[T]) Invalidate() {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.parent.Invalidate()
}

func (i *retryLoader[T]) Refresh(ctx context.Context, discovery agencyCache.LeaderDiscovery) (err error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	for z := 0; z < i.retries; z++ {
		if err := i.parent.Refresh(ctx, discovery); err != nil {
			logger.Err(err).Debug("Unable to refresh agency while retrying")
			continue
		}

		return nil
	}

	return i.parent.Refresh(ctx, discovery)
}
