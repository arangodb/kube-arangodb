//
// DISCLAIMER
//
// Copyright 2023-2024 ArangoDB GmbH, Cologne, Germany
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

package kubernetes

import (
	"context"
	"sort"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func Extract[T1, T2 any](in List[T1], ex func(in T1) T2) List[T2] {
	r := make(List[T2], len(in))

	for id := range r {
		r[id] = ex(in[id])
	}

	return r
}

type List[T any] []T

func (l List[T]) Sort(pred func(a, b T) bool) List[T] {
	sort.Slice(l, func(i, j int) bool {
		return pred(l[i], l[j])
	})

	return l
}

func (l List[T]) Append(obj ...T) List[T] {
	r := make(List[T], 0, len(l)+len(obj))

	r = append(r, l...)
	r = append(r, obj...)

	return r
}

func (l List[T]) Filter(f func(a T) bool) List[T] {
	r := make(List[T], 0, len(l))

	for _, o := range l {
		if !f(o) {
			continue
		}

		r = append(r, o)
	}

	return r
}

func (l List[T]) Contains(f func(a T) bool) bool {
	for _, o := range l {
		if f(o) {
			return true
		}
	}

	return false
}

func (l List[T]) Unique(f func(existing List[T], a T) bool) List[T] {
	r := make(List[T], 0, len(l))

	for _, o := range l {
		if f(r, o) {
			continue
		}

		r = append(r, o)
	}

	return r
}

type ObjectList[T meta.Object] map[types.UID]T

func (d ObjectList[T]) ByName(name string) (T, bool) {
	for _, obj := range d {
		if obj.GetName() == name {
			return obj, true
		}
	}

	return util.Default[T](), false
}

func (d ObjectList[T]) AsList() List[T] {
	list := make([]T, 0, len(d))
	for _, p := range d {
		list = append(list, p)
	}

	return list
}

func MapObjects[L k8sutil.ListContinue, T meta.Object](ctx context.Context, k k8sutil.ListAPI[L], extract func(result L) []T) (ObjectList[T], error) {
	objects := ObjectList[T]{}

	if err := k8sutil.APIList[L](ctx, k, meta.ListOptions{}, func(result L, err error) error {
		if err != nil {
			return err
		}
		for _, obj := range extract(result) {
			obj.SetManagedFields(nil)

			objects[obj.GetUID()] = obj
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return objects, nil
}

func ListObjects[L k8sutil.ListContinue, T meta.Object](ctx context.Context, k k8sutil.ListAPI[L], extract func(result L) []T) ([]T, error) {
	objects, err := MapObjects[L, T](ctx, k, extract)
	if err != nil {
		return nil, err
	}

	return objects.AsList(), nil
}
