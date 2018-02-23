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

// TestSliceLength checks the Length function of a slice.
func TestSliceLength(t *testing.T) {
	tests := []struct {
		JSON      string
		Length    velocypack.ValueLength
		ErrorType func(error) bool
		Unindexed bool
		Head      byte
	}{
		{"null", velocypack.ValueLength(0), velocypack.IsInvalidType, false, 0x18},
		{"true", velocypack.ValueLength(0), velocypack.IsInvalidType, false, 0x1a},
		{"false", velocypack.ValueLength(0), velocypack.IsInvalidType, false, 0x19},
		{"[]", velocypack.ValueLength(0), nil, false, 0x01},
		{"[1]", velocypack.ValueLength(1), nil, false, 0},
		{"[2,[]]", velocypack.ValueLength(2), nil, false, 0},
		{"[2,{},3]", velocypack.ValueLength(3), nil, false, 0},
		{"[1,2,3,4,5,6,7,8,9,\"ten\"]", velocypack.ValueLength(10), nil, false, 0},
		{"{}", velocypack.ValueLength(0), nil, false, 0x0a},
		{"{\"foo\":1}", velocypack.ValueLength(1), nil, false, 0},
		{"{\"foo\":1,\"bar\":{}}", velocypack.ValueLength(2), nil, false, 0},
		{"{\"a\":1,\"b\":2,\"c\":3,\"d\":4,\"e\":5,\"f\":6,\"g\":7,\"h\":8,\"i\":9,\"j\":10,\"k\":11,\"l\":12}", velocypack.ValueLength(12), nil, false, 0},
		// Unindexed
		{"[]", velocypack.ValueLength(0), nil, true, 0x01},
		{"[1]", velocypack.ValueLength(1), nil, true, 0x13},
		{"[2,[]]", velocypack.ValueLength(2), nil, true, 0x13},
		{"[2,{},3]", velocypack.ValueLength(3), nil, true, 0x13},
		{"[1,2,3,4,5,6,7,8,9,\"ten\"]", velocypack.ValueLength(10), nil, true, 0x13},
		{"{}", velocypack.ValueLength(0), nil, true, 0x0a},
		{"{\"foo\":1}", velocypack.ValueLength(1), nil, true, 0x14},
		{"{\"foo\":1,\"bar\":{}}", velocypack.ValueLength(2), nil, true, 0x14},
		{"{\"a\":1,\"b\":2,\"c\":3,\"d\":4,\"e\":5,\"f\":6,\"g\":7,\"h\":8,\"i\":9,\"j\":10,\"k\":11,\"l\":12}", velocypack.ValueLength(12), nil, true, 0x14},
	}

	for _, test := range tests {
		slice := mustSlice(velocypack.ParseJSONFromString(test.JSON, velocypack.ParserOptions{
			BuildUnindexedArrays:  test.Unindexed,
			BuildUnindexedObjects: test.Unindexed,
		}))
		if test.Head != 0 && slice[0] != test.Head {
			t.Errorf("Invalid Head for '%s': got %02x, expected %02x", test.JSON, slice[0], test.Head)
		}
		l, err := slice.Length()
		if test.ErrorType != nil {
			if !test.ErrorType(err) {
				t.Errorf("Length: invalid error for '%s': got %v", test.JSON, err)
			}
		} else if err != nil {
			t.Errorf("Length failed for '%s': got %v", test.JSON, err)
		} else {
			if l != test.Length {
				t.Errorf("Length returned invalid value for '%s': got %d, expected %d", test.JSON, l, test.Length)
			}
		}
	}
}
