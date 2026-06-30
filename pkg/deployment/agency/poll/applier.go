//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package poll

import (
	"encoding/json"
	"reflect"
	goStrings "strings"
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/transaction"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ApplierConfig struct {
	AllowUnsupportedOperations bool
}

func NewApplier[T interface{}](cfg ApplierConfig) Applier[T] {
	return &applier[T]{
		cfg: cfg,
	}
}

type Applier[T interface{}] interface {
	ApplyItemSet(items ItemSet) error
	ApplyItem(key string, item Item) error
	Get() *T
	Set(*T)
}
type applier[T interface{}] struct {
	lock sync.Mutex
	obj  T
	cfg  ApplierConfig
}

func (a *applier[T]) Set(t *T) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if t == nil {
		return
	}
	a.obj = *t
}
func (a *applier[T]) ApplyItemSet(items ItemSet) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	return a.applyItemSet(&a.obj, items)
}
func (a *applier[T]) ApplyItem(key string, item Item) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	return a.applyItem(&a.obj, key, item)
}
func (a *applier[T]) Get() *T {
	a.lock.Lock()
	defer a.lock.Unlock()
	return &a.obj
}
func (a *applier[T]) applyItemSet(value interface{}, items ItemSet) error {
	for k, v := range items {
		if err := a.applyItem(value, k, v); err != nil {
			return err
		}
	}
	return nil
}
func (a *applier[T]) applyItem(value interface{}, key string, item Item) error {
	parts := prepareKey(key)
	wrap := func(err error) error {
		return errors.WithMessagef(err, "key %s, val: %+v, item: %+v (op %s)", key, value, item, item.Operation.Get())
	}
	switch op := item.Operation.Get(); op {
	case transaction.OperationSet, transaction.OperationReplace:
		err := extract(reflect.ValueOf(value), func(out reflect.Value) error {
			z := reflect.New(out.Type())
			if err := json.Unmarshal(item.GetData(), z.Interface()); err != nil {
				return err
			}
			out.Set(z.Elem())
			return nil
		}, parts...)
		return wrap(errors.WithMessage(err, "extract failed"))
	case transaction.OperationDelete:
		err := remove(reflect.ValueOf(value), parts...)
		return wrap(errors.WithMessage(err, "remove failed"))
	case transaction.OperationIncrement:
		size := 1
		if len(item.GetData()) > 0 {
			if err := json.Unmarshal(item.GetData(), &size); err != nil {
				return wrap(err)
			}
		}
		err := extract(reflect.ValueOf(value), func(out reflect.Value) error {
			return changeInteger(&out, size)
		}, parts...)
		return wrap(errors.WithMessage(err, "extract failed"))
	case transaction.OperationDecrement:
		size := 1
		if len(item.GetData()) > 0 {
			if err := json.Unmarshal(item.GetData(), &size); err != nil {
				return wrap(err)
			}
		}
		err := extract(reflect.ValueOf(value), func(out reflect.Value) error {
			return changeInteger(&out, -size)
		}, parts...)
		return wrap(errors.WithMessage(err, "extract failed"))
	case transaction.OperationPush:
		err := extract(reflect.ValueOf(value), func(out reflect.Value) error {
			if out.Kind() != reflect.Slice && out.Kind() != reflect.Array {
				return errors.Errorf("Only Slice or Array are supported")
			}
			z := reflect.New(out.Type().Elem())
			if err := json.Unmarshal(item.GetData(), z.Interface()); err != nil {
				return err
			}
			ret := reflect.MakeSlice(out.Type(), out.Len()+1, out.Len()+1)
			for id := 0; id < out.Len(); id++ {
				ret.Index(id).Set(out.Index(id))
			}
			ret.Index(out.Len()).Set(z.Elem())
			out.Set(ret)
			return nil
		}, parts...)
		return wrap(errors.WithMessage(err, "extract failed"))
	case transaction.OperationPop:
		err := extract(reflect.ValueOf(value), func(out reflect.Value) error {
			if out.Kind() != reflect.Slice && out.Kind() != reflect.Array {
				return errors.Errorf("Only Slice or Array are supported")
			}
			if out.Len() == 0 {
				return errors.Errorf("Slice is already empty")
			}
			out.SetLen(out.Len() - 1)
			return nil
		}, parts...)
		return wrap(errors.WithMessage(err, "extract failed"))
	case transaction.OperationErase:
		err := extract(reflect.ValueOf(value), func(out reflect.Value) error {
			if out.Kind() != reflect.Slice {
				return errors.Errorf("Only Slice is supported")
			}
			if out.Len() == 0 {
				return errors.Errorf("Slice is already empty")
			}
			if posBytes := item.GetPosition(); len(posBytes) > 0 {
				var pos int
				if err := json.Unmarshal(item.GetPosition(), &pos); err != nil {
					return err
				}
				if pos < 0 || pos >= out.Len() {
					return errors.Errorf("pos field is not valid: value len is %d but pos is %d", out.Len(), pos)
				}
				// copy everything except element at pos
				ret := reflect.MakeSlice(out.Type(), out.Len()-1, out.Len()-1)
				retIndex := 0
				for id := 0; id < out.Len(); id++ {
					if id != pos {
						ret.Index(retIndex).Set(out.Index(id))
						retIndex++
					}
				}
				out.Set(ret)
				return nil
			} else if valBytes := item.GetValue(); len(valBytes) > 0 {
				val := reflect.New(out.Type().Elem())
				if err := json.Unmarshal(item.GetValue(), val.Interface()); err != nil {
					return err
				}
				// copy everything except elements which equal val
				var newValues = make([]reflect.Value, 0, out.Len())
				for id := 0; id < out.Len(); id++ {
					right := val
					if right.Kind() == reflect.Interface || right.Kind() == reflect.Pointer {
						right = val.Elem()
					}
					left := out.Index(id)
					if left.Kind() == reflect.Interface || left.Kind() == reflect.Pointer {
						left = left.Elem()
					}
					if !left.Equal(right) {
						newValues = append(newValues, out.Index(id))
					}
				}
				ret := reflect.MakeSlice(out.Type(), len(newValues), len(newValues))
				for id, v := range newValues {
					ret.Index(id).Set(v)
				}
				out.Set(ret)
				return nil
			}
			return errors.Errorf("Erase operation should have pos or val fields non-empty")
		}, parts...)
		return wrap(errors.WithMessage(err, "extract failed"))
	case "observe", "unobserve":
		// known but deprecated operations, has no effect
		return nil
	default:
		if a.cfg.AllowUnsupportedOperations {
			return nil
		}
		return wrap(errors.Errorf("Unknown operation \"%s\"", op))
	}
}
func prepareKey(in string) []string {
	for {
		v := goStrings.ReplaceAll(in, "//", "/")
		if v == in {
			break
		}
		in = v
	}
	return goStrings.Split(goStrings.TrimPrefix(in, "/"), "/")
}
