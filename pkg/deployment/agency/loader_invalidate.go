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
)

func InvalidateOnErrorLoader[T interface{}](loader StateLoader[T]) StateLoader[T] {
	return &invalidateOnErrorLoader[T]{
		StateLoader: loader,
	}
}

type invalidateOnErrorLoader[T interface{}] struct {
	lock sync.Mutex

	StateLoader[T]
}

func (i *invalidateOnErrorLoader[T]) Refresh(ctx context.Context, discovery LeaderDiscovery) (err error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	defer func() {
		if err != nil {
			i.StateLoader.Invalidate()
		}
	}()

	defer func() {
		if p := recover(); p != nil {
			i.StateLoader.Invalidate()
			panic(p)
		}
	}()

	return i.StateLoader.Refresh(ctx, discovery)
}
