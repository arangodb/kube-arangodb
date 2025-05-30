//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

func NewFilter[T any](in []T) Filter[T] {
	return filterList[T](in)
}

type Filter[T any] interface {
	Filter(predicate func(in T) bool) Filter[T]
	Get() []T
}

type filterList[T any] []T

func (f filterList[T]) Filter(predicate func(in T) bool) Filter[T] {
	n := make(filterList[T], 0, len(f))

	for _, el := range f {
		if predicate(el) {
			n = append(n, el)
		}
	}

	return n
}

func (f filterList[T]) Get() []T {
	return f
}
