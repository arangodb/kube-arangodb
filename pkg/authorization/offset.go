package authorization

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

	if start < o.items[0].Sequence {
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
