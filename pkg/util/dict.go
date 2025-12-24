//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

import (
	"maps"
	"sort"
)

type KV[K comparable, V any] struct {
	K K
	V V
}

func Extract[K comparable, V any](in map[K]V) []KV[K, V] {
	r := make([]KV[K, V], 0, len(in))

	for k, v := range in {
		r = append(r, KV[K, V]{
			K: k,
			V: v,
		})
	}

	return r
}

func ExtractWithSort[K comparable, V any](in map[K]V, cmp func(i, j K) bool) []KV[K, V] {
	return Sort(Extract(in), func(i, j KV[K, V]) bool {
		return cmp(i.K, j.K)
	})
}

func Sort[IN any](in []IN, cmp func(i, j IN) bool) []IN {
	r := make([]IN, len(in))
	copy(r, in)
	sort.Slice(r, func(i, j int) bool {
		return cmp(r[i], r[j])
	})
	return r
}

func MapValues[K comparable, V any](m map[K]V) []V {
	r := make([]V, 0, len(m))

	for k := range m {
		r = append(r, m[k])
	}

	return r
}

func MapKeys[K comparable, V any](m map[K]V) []K {
	r := make([]K, 0, len(m))

	for k := range m {
		r = append(r, k)
	}

	return r
}

func SortKeys[V any](m map[string]V) []string {
	keys := MapKeys(m)

	sort.Strings(keys)

	return keys
}

func CopyFullMap[K comparable, V any](src map[K]V) map[K]V {
	if src == nil {
		return nil
	}

	r := map[K]V{}

	maps.Copy(r, src)

	return r
}

func MergeMaps[K comparable, V any](override bool, maps ...map[K]V) map[K]V {
	r := map[K]V{}

	for _, m := range maps {
		for k, v := range m {
			if !override {
				if _, ok := r[k]; ok {
					continue
				}
			}

			r[k] = v
		}
	}

	return r
}

func IterateSorted[V any](m map[string]V, cb func(string, V)) {
	for _, k := range SortKeys(m) {
		cb(k, m[k])
	}
}

func Optional[K comparable, V any](m map[K]V, key K, def V) V {
	v, ok := m[key]
	if ok {
		return v
	}

	return def
}

func FormatMap[K comparable, A, B any](in map[K]A, format func(K, A) B) map[K]B {
	var r = make(map[K]B, len(in))

	for k, a := range in {
		r[k] = format(k, a)
	}

	return r
}
