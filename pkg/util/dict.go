//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	"reflect"
	"sort"
)

func SortKeys(m interface{}) []string {
	if m == nil {
		return []string{}
	}

	q := reflect.ValueOf(m).MapKeys()

	r := make([]string, len(q))

	for id, v := range q {
		r[id] = v.String()
	}

	sort.Strings(r)

	return r
}

func CopyMap[K comparable, V any](dst, src map[K]V) {
	// TODO: replace with maps.Copy when switching to go1.21
	for k, v := range src {
		dst[k] = v
	}
}

func IterateSorted[V any](m map[string]V, cb func(string, V)) {
	for _, k := range SortKeys(m) {
		cb(k, m[k])
	}
}
