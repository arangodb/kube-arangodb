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

// TestSliceMerge checks the Merge.
func TestSliceMerge(t *testing.T) {
	tests := []struct {
		InputJSONs []string
		OutputJSON string
	}{
		{
			InputJSONs: []string{
				`{"a":1,"b":2}`,
				`{"a":7}`,
			},
			OutputJSON: `{"a":1,"b":2}`,
		},
		{
			InputJSONs: []string{
				`{"a":1,"b":2}`,
				`{"a":7,"d":true}`,
			},
			OutputJSON: `{"a":1,"b":2,"d":true}`,
		},
		{
			InputJSONs: []string{
				`{"a":1,"b":{"c":"foo"},"d":[5,6,7]}`,
				`{"a":7,"b":[1,2,3,4]}`,
			},
			OutputJSON: `{"a":1,"b":{"c":"foo"},"d":[5,6,7]}`,
		},
		{
			InputJSONs: []string{
				`{"a":1,"b":{"c":"foo"},"d":[5,6,7]}`,
				`{"A":7,"B":[1,2,3,4]}`,
			},
			OutputJSON: `{"A":7,"B":[1,2,3,4],"a":1,"b":{"c":"foo"},"d":[5,6,7]}`,
		},
	}

	for testIndex, test := range tests {
		slices := make([]velocypack.Slice, len(test.InputJSONs))
		for i, inp := range test.InputJSONs {
			var err error
			slices[i], err = velocypack.ParseJSONFromString(inp)
			if err != nil {
				t.Fatalf("Failed to parse '%s': %#v", inp, err)
			}
		}
		result, err := velocypack.Merge(slices...)
		if err != nil {
			t.Fatalf("Failed to Merge test %d: %#v", testIndex, err)
		}
		output, err := result.JSONString()
		if err != nil {
			t.Fatalf("Failed to Dump result of test %d: %#v", testIndex, err)
		}
		if output != test.OutputJSON {
			t.Errorf("Unexpected result in test %d\nExpected: %s\nGot: %s", testIndex, test.OutputJSON, output)
		}
	}
}

// TestSliceMergeNonObject checks the Merge with invalid input.
func TestSliceMergeNonObject(t *testing.T) {
	if _, err := velocypack.Merge(velocypack.NullSlice()); !velocypack.IsInvalidType(err) {
		t.Errorf("Expected InvalidTypeError, got %#v", err)
	}
}
