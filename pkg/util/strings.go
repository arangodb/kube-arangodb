//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package util

import "fmt"

func CompareStrings(a, b string) bool {
	return a == b
}

func CompareStringPointers(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return CompareStrings(*a, *b)
}

func CompareStringArray(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for id, ak := range a {
		if ak != b[id] {
			return false
		}
	}

	return true
}

func PrefixStringArray(a []string, prefix string) []string {
	b := make([]string, len(a))

	for id, element := range a {
		b[id] = fmt.Sprintf("%s%s", prefix, element)
	}

	return b
}

// Diff returns the elements in `compareWhat` that are not in `compareTo`.
func Diff(compareWhat, compareTo []string) []string {
	compareToMap := make(map[string]struct{}, len(compareTo))
	for _, x := range compareTo {
		compareToMap[x] = struct{}{}
	}

	var diff []string
	for _, x := range compareWhat {
		if _, found := compareToMap[x]; !found {
			diff = append(diff, x)
		}
	}

	return diff
}
