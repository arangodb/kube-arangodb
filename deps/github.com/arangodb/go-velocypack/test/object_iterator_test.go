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
	"fmt"
	"sort"
	"strings"
	"testing"

	velocypack "github.com/arangodb/go-velocypack"
)

func TestObjectIteratorInvalidSlice(t *testing.T) {
	tests := []velocypack.Slice{
		velocypack.NullSlice(),
		velocypack.TrueSlice(),
		velocypack.FalseSlice(),
		mustSlice(velocypack.ParseJSONFromString("1")),
		mustSlice(velocypack.ParseJSONFromString("7.7")),
		mustSlice(velocypack.ParseJSONFromString("\"foo\"")),
		mustSlice(velocypack.ParseJSONFromString("[]")),
		mustSlice(velocypack.ParseJSONFromString("[]", velocypack.ParserOptions{BuildUnindexedArrays: true})),
	}
	for _, test := range tests {
		ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsInvalidType, t)(velocypack.NewObjectIterator(test))
	}
}

func TestObjectIteratorValues(t *testing.T) {
	tests := []map[string]string{
		map[string]string{},
		map[string]string{"foo": "1"},
	}
	for _, unindexed := range []bool{true, false} {
		for _, test := range tests {
			var keyValuePairs []string
			for k, v := range test {
				keyValuePairs = append(keyValuePairs, fmt.Sprintf(`"%s":%s`, k, v))
			}
			json := "{" + strings.Join(keyValuePairs, ",") + "}"
			sort.Strings(keyValuePairs)
			s := mustSlice(velocypack.ParseJSONFromString(json, velocypack.ParserOptions{BuildUnindexedObjects: unindexed}))
			it, err := velocypack.NewObjectIterator(s)
			if err != nil {
				t.Errorf("Failed to create ObjectIterator for '%s': %v", json, err)
			} else {
				i := 0
				for it.IsValid() {
					k := mustSlice(it.Key(true))
					v := mustSlice(it.Value())
					kv := fmt.Sprintf(`"%s":%s`, mustString(k.GetString()), mustString(v.JSONString()))
					if kv != keyValuePairs[i] {
						t.Errorf("Element %d is invalid; got '%s', expected '%s'", i, kv, keyValuePairs[i])
					}
					must(it.Next())
					i++
				}
			}
		}
	}
}
