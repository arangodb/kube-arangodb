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
	"context"
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/arangodb/shared"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/operations"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func NewPooler[T PoolerObject](connection cache.Object[arangodb.Collection]) Pooler[T] {
	return &pooler[T]{
		connection: connection,
		state:      make(map[string]Document[T]),
		index:      1,
	}
}

type PoolerObject interface {
	proto.Message

	Hash() string

	Clean() error
	Validate() error
}

type Pooler[T PoolerObject] interface {
	Refresh(ctx context.Context) error

	Create(ctx context.Context, name string, obj T) (T, uint32, error)
	Update(ctx context.Context, name string, obj T) (T, uint32, error)
	Delete(ctx context.Context, name string) (uint32, error)

	Item(name string) (T, uint32, bool)

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

func (p *pooler[T]) Item(name string) (T, uint32, bool) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	v, ok := p.state[name]
	if !ok {
		return util.Default[T](), 0, false
	}

	return v.Spec, v.Sequence, true
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

func (p *pooler[T]) run(ctx context.Context, name string, action DocumentAction, obj T, validate func(p *pooler[T], ctx context.Context, name string, obj T) error) (T, uint32, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if err := obj.Validate(); err != nil {
		return util.Default[T](), 0, err
	}

	if err := obj.Clean(); err != nil {
		return util.Default[T](), 0, err
	}

	col, err := p.connection.Get(ctx)
	if err != nil {
		return util.Default[T](), 0, err
	}

	res, err := operations.WithTransaction[Document[T]](ctx, col.Database(), arangodb.TransactionCollections{
		Read:  []string{col.Name()},
		Write: []string{col.Name()},
	}, &arangodb.BeginTransactionOptions{}, operations.WithLock(col.Name(), func(ctx context.Context, c arangodb.Transaction, lock *operations.LockDocument) (Document[T], error) {
		col, err := c.GetCollection(ctx, col.Name(), &arangodb.GetCollectionOptions{SkipExistCheck: true})
		if err != nil {
			return util.Default[Document[T]](), err
		}

		if err := p.refresh(ctx, col); err != nil {
			return util.Default[Document[T]](), err
		}

		if p.index != lock.CurrentSequence {
			return util.Default[Document[T]](), errors.Errorf("Sequence changed")
		}

		if err := validate(p, ctx, name, obj); err != nil {
			return util.Default[Document[T]](), err
		}

		doc := Document[T]{
			Key:      fmt.Sprintf("%09d", lock.CurrentSequence),
			Name:     name,
			Sequence: lock.CurrentSequence,
			Created:  meta.Now(),
			Action:   action,
			Spec:     obj,
		}

		_, err = col.CreateDocument(ctx, doc)
		if err != nil {
			return util.Default[Document[T]](), err
		}

		lock.CurrentSequence += 1

		if old, ok := p.state[name]; ok {
			// Delete old document
			old.Deleted = doc.Created

			if _, err := col.UpdateDocument(ctx, old.Key, old); err != nil {
				return util.Default[Document[T]](), err
			}
		}

		return doc, nil
	}))
	if err != nil {
		return util.Default[T](), 0, err
	}

	return res.Spec, res.Sequence, p.refresh(ctx, col)
}

func (p *pooler[T]) Update(ctx context.Context, name string, obj T) (T, uint32, error) {
	return p.run(ctx, name, DocumentUpdateAction, obj, func(p *pooler[T], ctx context.Context, name string, obj T) error {
		if _, v := p.state[name]; !v {
			return PoolNotFound{}
		}

		return nil
	})
}

func (p *pooler[T]) Delete(ctx context.Context, name string) (uint32, error) {
	_, index, err := p.run(ctx, name, DocumentDeleteAction, util.Default[T](), func(p *pooler[T], ctx context.Context, name string, obj T) error {
		if _, v := p.state[name]; !v {
			return PoolNotFound{}
		}

		return nil
	})
	return index, err
}

func (p *pooler[T]) Create(ctx context.Context, name string, obj T) (T, uint32, error) {
	return p.run(ctx, name, DocumentCreateAction, obj, func(p *pooler[T], ctx context.Context, name string, obj T) error {
		if _, v := p.state[name]; v {
			return PoolAlreadyExists{}
		}

		return nil
	})
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

		if doc.Action == DocumentDeleteAction {
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
		BindVars:  map[string]interface{}{"start": start, "key": operations.LockDocumentID},
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
