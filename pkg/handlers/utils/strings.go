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

package utils

import (
	"math/rand"
	"sort"
)

var (
	randomBase = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
)

// StringList extended []string definition
type StringList []string

// Has return true if item is on the list
func (s StringList) Has(i string) bool {
	for _, obj := range s {
		if obj == i {
			return true
		}
	}
	return false
}

// Append add items to the list
func (s StringList) Append(i ...string) StringList {
	return append(s, i...)
}

// Remove items from the list
func (s StringList) Remove(i ...string) StringList {
	r := make(StringList, 0, len(s))

	strings := StringList(i)

	for _, obj := range s {
		if strings.Has(obj) {
			continue
		}

		r = r.Append(obj)
	}

	return r
}

func (s StringList) Sort() StringList {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
	return s
}

func (s StringList) Unique() StringList {
	z := make(StringList, 0, len(s))
	for _, t := range s {
		if !z.Has(t) {
			z = append(z, t)
		}
	}
	return z
}

// RandomString generates random string
func RandomString(n int) string {
	return RandomStringFrom(n, randomBase)
}

// RandomStringFrom generates random string from base runes
func RandomStringFrom(n int, base []rune) string {
	runes := make([]rune, n)
	for id := range runes {
		runes[id] = base[rand.Intn(len(base))]
	}

	return string(runes)
}
