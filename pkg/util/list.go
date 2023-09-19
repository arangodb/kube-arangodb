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

func (l List[T]) Sort(fn func(T, T) bool) List[T] {
	clone := l
	sort.Slice(clone, func(i, j int) bool {
		return fn(clone[i], clone[j])
	})
	return clone
}

func MapList[T, V any](in List[T], fn func(T) V) List[V] {
	if in == nil {
		return nil
	}
	result := make(List[V], 0, len(in))
	for _, em := range in {
		result = append(result, fn(em))
	}
	return result
}
