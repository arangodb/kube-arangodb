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

package authorization

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/arangodb/shared"

	"github.com/arangodb/kube-arangodb/pkg/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func NewPooler[T PoolerObject](connection cache.Object[arangodb.Collection]) Pooler[T] {
	return &pooler[T]{
		connection: connection,
		state:      make(map[string]Document[T]),
	}
}

type PoolerObject interface {
	proto.Message

	Deleted() bool

	Clean() error
	Validate() error
}

type Pooler[T PoolerObject] interface {
	Refresh(ctx context.Context) error

	Update(ctx context.Context, name string, obj T) error

	Index() uint32

	Pool(start uint32) ([]OffsetItem[T], error)
	Get() []OffsetItem[T]
}

type pooler[T PoolerObject] struct {
	lock sync.RWMutex

	index uint32

	state map[string]Document[T]

	connection cache.Object[arangodb.Collection]

	offset Offset[T]
}

func (p *pooler[T]) Get() []OffsetItem[T] {
	p.lock.RLock()
	defer p.lock.RUnlock()

	r := util.ExtractMap(p.state, func(k string, a Document[T]) Document[T] {
		return a
	})

	util.Sort(r, func(i, j Document[T]) bool {
		return i.Sequence < j.Sequence
	})

	return util.FormatList(r, func(a Document[T]) OffsetItem[T] {
		return OffsetItem[T]{
			Item:     a.Spec,
			Sequence: a.Sequence,
		}
	})
}

func (p *pooler[T]) Pool(start uint32) ([]OffsetItem[T], error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.offset.Pool(start)
}

func (p *pooler[T]) Index() uint32 {
	return p.index
}

func (p *pooler[T]) Update(ctx context.Context, name string, obj T) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	col, err := p.connection.Get(ctx)
	if err != nil {
		return err
	}

	_, err = WithTransaction[T](ctx, col.Database(), arangodb.TransactionCollections{
		Read:  []string{col.Name()},
		Write: []string{col.Name()},
	}, &arangodb.BeginTransactionOptions{}, WithLock(col.Name(), func(ctx context.Context, c arangodb.Transaction, lock *types.LockDocument) (T, error) {
		col, err := c.GetCollection(ctx, col.Name(), &arangodb.GetCollectionOptions{SkipExistCheck: true})
		if err != nil {
			return util.Default[T](), err
		}

		if err := p.refresh(ctx, col); err != nil {
			return util.Default[T](), err
		}

		if p.index != lock.CurrentSequence {
			return util.Default[T](), errors.Errorf("Sequence changed")
		}

		doc := Document[T]{
			Key:      fmt.Sprintf("%09d", lock.CurrentSequence),
			Name:     name,
			Sequence: lock.CurrentSequence,
			Created:  meta.Now(),
			Spec:     obj,
		}

		_, err = col.CreateDocument(ctx, doc)
		if err != nil {
			return util.Default[T](), err
		}

		lock.CurrentSequence += 1

		if old, ok := p.state[name]; ok {
			// Delete old document
			old.Deleted = doc.Created

			if _, err := col.UpdateDocument(ctx, old.Key, old); err != nil {
				return util.Default[T](), err
			}
		}

		return obj, nil
	}))
	if err != nil {
		return err
	}

	return p.refresh(ctx, col)
}

func (p *pooler[T]) Refresh(ctx context.Context) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	col, err := p.connection.Get(ctx)
	if err != nil {
		return err
	}

	return p.refresh(ctx, col)
}

func (p *pooler[T]) refresh(ctx context.Context, col arangodb.Collection) error {
	changes, err := PoolChanges[T](ctx, col.Database(), col.Name(), p.index)
	if err != nil {
		return err
	}

	for _, doc := range changes {
		p.index = doc.Sequence + 1

		p.offset.Add(doc.Sequence, doc.Name, doc.Spec)

		if doc.Spec.Deleted() {
			// Deleted
			delete(p.state, doc.Name)
		} else {
			p.state[doc.Name] = doc
		}
	}

	p.offset.Trim(1024)

	return nil
}

// PoolChanges pools the changes from registry. If no documents found EOF is returned
func PoolChanges[T proto.Message](ctx context.Context, db arangodb.DatabaseQuery, col string, start uint32) ([]Document[T], error) {
	query := fmt.Sprintf("FOR doc IN %s FILTER doc.sequence >= @start FILTER doc._key != @key SORT doc.sequence ASC RETURN doc", col)

	result, err := db.Query(ctx, query, &arangodb.QueryOptions{
		BatchSize: 1024,
		BindVars:  map[string]interface{}{"start": start, "key": types.LockDocumentID},
	})
	if err != nil {
		return nil, err
	}

	var ret []Document[T]

	for {
		var d Document[T]

		if _, err := result.ReadDocument(ctx, &d); err != nil {
			if shared.IsEOF(err) {
				break
			}

			return nil, err
		}

		ret = append(ret, d)
	}

	if len(ret) == 0 {
		return nil, nil
	}

	return ret, nil
}
