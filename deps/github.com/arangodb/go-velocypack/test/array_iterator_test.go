//
// DISCLAIMER
//
// Copyright 2017 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package test

import (
	"strings"
	"testing"

	velocypack "github.com/arangodb/go-velocypack"
)

func TestArrayIteratorInvalidSlice(t *testing.T) {
	tests := []velocypack.Slice{
		velocypack.NullSlice(),
		velocypack.TrueSlice(),
		velocypack.FalseSlice(),
		mustSlice(velocypack.ParseJSONFromString("1")),
		mustSlice(velocypack.ParseJSONFromString("7.7")),
		mustSlice(velocypack.ParseJSONFromString("\"foo\"")),
		mustSlice(velocypack.ParseJSONFromString("{}")),
		mustSlice(velocypack.ParseJSONFromString("{}", velocypack.ParserOptions{BuildUnindexedObjects: true})),
	}
	for _, test := range tests {
		ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsInvalidType, t)(velocypack.NewArrayIterator(test))
	}
}

func TestArrayIteratorValues(t *testing.T) {
	tests := [][]string{
		[]string{},
		[]string{"1", "2", "true", "null", "false", "{}"},
	}
	for _, unindexed := range []bool{true, false} {
		for _, test := range tests {
			json := "[" + strings.Join(test, ",") + "]"
			s := mustSlice(velocypack.ParseJSONFromString(json, velocypack.ParserOptions{BuildUnindexedArrays: unindexed}))
			it, err := velocypack.NewArrayIterator(s)
			if err != nil {
				t.Errorf("Failed to create ArrayIterator for '%s': %v", json, err)
			} else {
				i := 0
				for it.IsValid() {
					v := mustSlice(it.Value())
					if mustString(v.JSONString()) != test[i] {
						t.Errorf("Element %d is invalid; got '%s', expected '%s'", i, mustString(v.JSONString()), test[i])
					}
					must(it.Next())
					i++
				}
			}
		}
	}
}
