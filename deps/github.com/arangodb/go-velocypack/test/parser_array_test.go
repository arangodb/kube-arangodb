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
	"testing"

	velocypack "github.com/arangodb/go-velocypack"
)

func TestParserArray(t *testing.T) {
	tests := map[string][]interface{}{
		`[]`:                []interface{}{},
		`[1,2,3]`:           []interface{}{1, 2, 3},
		`[1,[2,"foo"],3]`:   []interface{}{1, []interface{}{2, "foo"}, 3},
		`[1,[[],[]],[[3]]]`: []interface{}{1, []interface{}{[]interface{}{}, []interface{}{}}, []interface{}{[]interface{}{3}}},
	}
	for test, expected := range tests {
		slice := mustSlice(velocypack.ParseJSONFromString(test))

		var v interface{}
		must(velocypack.Unmarshal(slice, &v))
		ASSERT_EQ(v, expected, t)
	}
}
