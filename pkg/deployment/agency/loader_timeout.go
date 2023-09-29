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
	"time"

	agencyCache "github.com/arangodb-helper/go-helper/pkg/arangod/agency/cache"
)

func TimeoutLoader[T interface{}](loader agencyCache.StateLoader[T], timeout time.Duration) agencyCache.StateLoader[T] {
	if timeout <= 0 {
		return loader
	}

	return &timeoutLoader[T]{
		parent:  loader,
		timeout: timeout,
	}
}

type timeoutLoader[T interface{}] struct {
	parent agencyCache.StateLoader[T]

	timeout time.Duration
}

func (i *timeoutLoader[T]) UpdateTime() time.Time {
	return i.parent.UpdateTime()
}

func (i *timeoutLoader[T]) Valid() bool {
	return i.parent.Valid()
}

func (i *timeoutLoader[T]) State() (*T, uint64, bool) {
	return i.parent.State()
}

func (i *timeoutLoader[T]) Invalidate() {
	i.parent.Invalidate()
}

func (i *timeoutLoader[T]) Refresh(ctx context.Context, discovery agencyCache.LeaderDiscovery) error {
	nctx, c := context.WithTimeout(ctx, i.timeout)
	defer c()

	return i.parent.Refresh(nctx, discovery)
}
