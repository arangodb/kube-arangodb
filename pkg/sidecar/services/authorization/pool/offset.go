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

package pool

import (
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type OffsetItem[T any] struct {
	Item     T
	Name     string
	Sequence uint32
}

type Offset[T any] struct {
	lock sync.RWMutex

	items []OffsetItem[T]
}

func (o *Offset[T]) Pool(start uint32) ([]OffsetItem[T], error) {
	o.lock.RLock()
	defer o.lock.RUnlock()

	if len(o.items) == 0 {
		return nil, nil
	}

	if start < o.items[0].Sequence-1 {
		return nil, PoolOutOfBoundsError{}
	}

	items := util.FilterList(o.items, func(item OffsetItem[T]) bool {
		return item.Sequence > start
	})

	return items, nil
}

func (o *Offset[T]) Trim(size int) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if len(o.items) > size {
		o.items = o.items[len(o.items)-size:]
	}
}

func (o *Offset[T]) Add(id uint32, name string, item T) {
	o.lock.Lock()
	defer o.lock.Unlock()

	o.items = append(o.items, OffsetItem[T]{
		Item:     item,
		Name:     name,
		Sequence: id,
	})
}
