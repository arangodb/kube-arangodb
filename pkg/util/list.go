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

package util

import "sort"

type List[T any] []T

func (l List[T]) Filter(fn func(T) bool) List[T] {
	if l == nil {
		return nil
	}
	result := make([]T, 0)
	for _, item := range l {
		if fn(item) {
			result = append(result, item)
		}
	}
	return result
}

func (l List[T]) Count(fn func(T) bool) int {
	return len(l.Filter(fn))
}

func (l List[T]) Append(items ...T) List[T] {
	return append(l, items...)
}

func (l List[T]) List() []T {
	return l
}

func (l List[T]) Sort(fn func(T, T) bool) List[T] {
	clone := l
	sort.Slice(clone, func(i, j int) bool {
		return fn(clone[i], clone[j])
	})
	return clone
}

func (l List[T]) Contains(fn func(T) bool) bool {
	for _, e := range l {
		if fn(e) {
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

func ListAsMap[K comparable, V any](in []V, extract func(in V) K) map[K]V {
	ret := make(map[K]V, len(in))

	for _, el := range in {
		ret[extract(el)] = el
	}

	return ret
}

func PickFromList[V any](in []V, q func(v V) bool) (V, bool) {
	for _, v := range in {
		if q(v) {
			return v, true
		}
	}

	return Default[V](), false
}

func MapList[T, V comparable](in List[T], fn func(T) V) List[V] {
	if in == nil {
		return nil
	}
	result := make(List[V], 0, len(in))
	for _, em := range in {
		result = append(result, fn(em))
	}
	return result
}

func FormatList[A, B any](in []A, format func(A) B) []B {
	var r = make([]B, len(in))

	for id := range in {
		r[id] = format(in[id])
	}

	return r
}

func ContainsList[A comparable](in []A, item A) bool {
	for _, el := range in {
		if el == item {
			return true
		}
	}

	return false
}

func UniqueList[A comparable](in []A) []A {
	var r = make([]A, 0, len(in))

	for _, el := range in {
		if !ContainsList(r, el) {
			r = append(r, el)
		}
	}

	return r
}

func FormatListErr[A, B any](in []A, format func(A) (B, error)) ([]B, error) {
	var r = make([]B, len(in))

	for id := range in {
		if o, err := format(in[id]); err != nil {
			return nil, err
		} else {
			r[id] = o
		}
	}

	return r, nil
}

func CopyList[A any](in []A) []A {
	ret := make([]A, len(in))
	copy(ret, in)
	return ret
}

func FlattenList[A any](in [][]A) []A {
	count := 0

	for _, v := range in {
		count += len(v)
	}

	res := make([]A, 0, count)

	for _, v := range in {
		res = append(res, v...)
	}

	return res
}
