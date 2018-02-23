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

func TestSliceSmallInt(t *testing.T) {
	expected := []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, -6, -5, -4, -3, -2, -1}

	for i := 0; i < 16; i++ {
		slice := velocypack.Slice{byte(0x30 + i)}
		assertEqualFromReader(t, slice)

		ASSERT_EQ(velocypack.SmallInt, slice.Type(), t)
		ASSERT_TRUE(slice.IsSmallInt(), t)
		ASSERT_EQ(velocypack.ValueLength(1), mustLength(slice.ByteSize()), t)

		ASSERT_EQ(expected[i], mustInt(slice.GetSmallInt()), t)
		ASSERT_EQ(expected[i], mustInt(slice.GetInt()), t)
	}
}
