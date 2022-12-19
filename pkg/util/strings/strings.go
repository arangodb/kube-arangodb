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

package strings

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

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

// DiffStringsOneWay returns the elements in `compareWhat` that are not in `compareTo`.
func DiffStringsOneWay(compareWhat, compareTo []string) []string {
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

// DiffStrings returns the elements in `compareWhat` that are not in `compareTo` and
// elements in `compareTo` that are not in `compareFrom`.
func DiffStrings(compareWhat, compareTo []string) []string {
	diff := DiffStringsOneWay(compareWhat, compareTo)

	return append(diff, DiffStringsOneWay(compareTo, compareWhat)...)
}

// Title returns string in Title format
func Title(in string) string {
	return cases.Title(language.English).String(in)
}

// Join concatenates the elements of its first argument to create a single string. The separator
// string sep is placed between elements in the resulting string.
func Join(elems []string, sep string) string {
	return strings.Join(elems, sep)
}

// Split slices s into all substrings separated by sep and returns a slice of
// the substrings between those separators.
//
// If s does not contain sep and sep is not empty, Split returns a
// slice of length 1 whose only element is s.
//
// If sep is empty, Split splits after each UTF-8 sequence. If both s
// and sep are empty, Split returns an empty slice.
//
// It is equivalent to SplitN with a count of -1.
//
// To split around the first instance of a separator, see Cut.
func Split(s, sep string) []string {
	return strings.Split(s, sep)
}

// ToLower returns s with all Unicode letters mapped to their lower case.
func ToLower(s string) string {
	return strings.ToLower(s)
}
