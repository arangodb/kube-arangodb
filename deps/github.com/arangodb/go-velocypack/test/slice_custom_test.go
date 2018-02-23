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

func TestSliceCustomTypeByteSize(t *testing.T) {
	tests := []velocypack.Slice{
		velocypack.Slice([]byte{0xf0, 0x00}),
		velocypack.Slice([]byte{0xf1, 0x00, 0x00}),
		velocypack.Slice([]byte{0xf2, 0x00, 0x00, 0x00, 0x00}),
		velocypack.Slice([]byte{0xf3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}),
		velocypack.Slice([]byte{0xf4, 0x03, 0x00, 0x00, 0x00}),
		velocypack.Slice([]byte{0xf5, 0x02, 0x00, 0x00}),
		velocypack.Slice([]byte{0xf6, 0x01, 0x00}),
		velocypack.Slice([]byte{0xf7, 0x01, 0x00, 0x00}),
		velocypack.Slice([]byte{0xf8, 0x02, 0x00, 0x00, 0x00}),
		velocypack.Slice([]byte{0xf9, 0x03, 0x00, 0x00, 0x00, 0x00}),
		velocypack.Slice([]byte{0xfa, 0x01, 0x00, 0x00, 0x00, 0x00}),
		velocypack.Slice([]byte{0xfb, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00}),
		velocypack.Slice([]byte{0xfc, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}),
		velocypack.Slice([]byte{0xfd, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}),
		velocypack.Slice([]byte{0xfe, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}),
		velocypack.Slice([]byte{0xff, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}),
	}

	for _, test := range tests {
		assertEqualFromReader(t, test)
		sz := mustLength(test.ByteSize())
		if sz != velocypack.ValueLength(len(test)) {
			t.Errorf("Invalid ByteSize in '%s', expected %d, got %d", test, len(test), sz)
		}
	}
}
