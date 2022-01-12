//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package v1

import "sort"

type List []string

func (l List) Equal(b List) bool {
	if len(l) != len(b) {
		return false
	}

	for i := range l {
		if l[i] != b[i] {
			return false
		}
	}

	return true
}

func (l List) Contains(v string) bool {
	for _, z := range l {
		if z == v {
			return true
		}
	}

	return false
}

func (l List) Sort() List {
	z := l.DeepCopy()
	sort.Strings(z)
	return z
}

func (l List) Unique() List {
	var m List

	for _, k := range l {
		if m.Contains(k) {
			continue
		}

		m = m.Add(k)
	}

	return m
}

func (l List) Remove(values ...string) List {
	var m List

	toRemove := List(values)

	for _, v := range l {
		if toRemove.Contains(v) {
			continue
		}

		m = append(m, v)
	}

	return m
}

func (l List) Add(values ...string) List {
	var m List

	m = append(m, l...)
	m = append(m, values...)

	return m
}
