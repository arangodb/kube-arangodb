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
	"encoding/binary"
	"math"
	"testing"

	velocypack "github.com/arangodb/go-velocypack"
)

func TestSliceDouble(t *testing.T) {
	slice := velocypack.Slice{0x1b, 1, 2, 3, 4, 5, 6, 7, 8}
	value := 23.5
	binary.LittleEndian.PutUint64(slice[1:], math.Float64bits(value))
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.Double, slice.Type(), t)
	ASSERT_TRUE(slice.IsDouble(), t)
	ASSERT_EQ(velocypack.ValueLength(9), mustLength(slice.ByteSize()), t)
	ASSERT_DOUBLE_EQ(value, mustDouble(slice.GetDouble()), t)
}

func TestSliceDoubleNegative(t *testing.T) {
	slice := velocypack.Slice{0x1b, 1, 2, 3, 4, 5, 6, 7, 8}
	value := -999.91355
	binary.LittleEndian.PutUint64(slice[1:], math.Float64bits(value))
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.Double, slice.Type(), t)
	ASSERT_TRUE(slice.IsDouble(), t)
	ASSERT_EQ(velocypack.ValueLength(9), mustLength(slice.ByteSize()), t)
	ASSERT_DOUBLE_EQ(value, mustDouble(slice.GetDouble()), t)
}
