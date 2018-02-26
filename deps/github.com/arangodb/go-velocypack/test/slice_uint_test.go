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
	"math"
	"testing"

	velocypack "github.com/arangodb/go-velocypack"
)

func TestSliceUInt1(t *testing.T) {
	slice := velocypack.Slice{0x28, 0x33}
	value := uint64(0x33)
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.UInt, slice.Type(), t)
	ASSERT_TRUE(slice.IsUInt(), t)
	ASSERT_EQ(velocypack.ValueLength(2), mustLength(slice.ByteSize()), t)

	ASSERT_EQ(value, mustUInt(slice.GetUInt()), t)
	ASSERT_EQ(int64(value), mustInt(slice.GetInt()), t)
}

func TestSliceUInt2(t *testing.T) {
	slice := velocypack.Slice{0x29, 0x23, 0x42}
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.UInt, slice.Type(), t)
	ASSERT_TRUE(slice.IsUInt(), t)
	ASSERT_EQ(velocypack.ValueLength(3), mustLength(slice.ByteSize()), t)

	ASSERT_EQ(uint64(0x4223), mustUInt(slice.GetUInt()), t)
}

func TestSliceUInt3(t *testing.T) {
	slice := velocypack.Slice{0x2a, 0x23, 0x42, 0x66}
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.UInt, slice.Type(), t)
	ASSERT_TRUE(slice.IsUInt(), t)
	ASSERT_EQ(velocypack.ValueLength(4), mustLength(slice.ByteSize()), t)

	ASSERT_EQ(uint64(0x664223), mustUInt(slice.GetUInt()), t)
}

func TestSliceUInt4(t *testing.T) {
	slice := velocypack.Slice{0x2b, 0x23, 0x42, 0x66, 0x7c}
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.UInt, slice.Type(), t)
	ASSERT_TRUE(slice.IsUInt(), t)
	ASSERT_EQ(velocypack.ValueLength(5), mustLength(slice.ByteSize()), t)

	ASSERT_EQ(uint64(0x7c664223), mustUInt(slice.GetUInt()), t)
}

func TestSliceUInt5(t *testing.T) {
	slice := velocypack.Slice{0x2c, 0x23, 0x42, 0x66, 0xac, 0x6f}
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.UInt, slice.Type(), t)
	ASSERT_TRUE(slice.IsUInt(), t)
	ASSERT_EQ(velocypack.ValueLength(6), mustLength(slice.ByteSize()), t)

	ASSERT_EQ(uint64(0x6fac664223), mustUInt(slice.GetUInt()), t)
}

func TestSliceUInt6(t *testing.T) {
	slice := velocypack.Slice{0x2d, 0x23, 0x42, 0x66, 0xac, 0xff, 0x3f}
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.UInt, slice.Type(), t)
	ASSERT_TRUE(slice.IsUInt(), t)
	ASSERT_EQ(velocypack.ValueLength(7), mustLength(slice.ByteSize()), t)

	ASSERT_EQ(uint64(0x3fffac664223), mustUInt(slice.GetUInt()), t)
}

func TestSliceUInt7(t *testing.T) {
	slice := velocypack.Slice{0x2e, 0x23, 0x42, 0x66, 0xac, 0xff, 0x3f, 0x5a}
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.UInt, slice.Type(), t)
	ASSERT_TRUE(slice.IsUInt(), t)
	ASSERT_EQ(velocypack.ValueLength(8), mustLength(slice.ByteSize()), t)

	ASSERT_EQ(uint64(0x5a3fffac664223), mustUInt(slice.GetUInt()), t)
}

func TestSliceUInt8(t *testing.T) {
	slice := velocypack.Slice{0x2f, 0x23, 0x42, 0x66, 0xac, 0xff, 0x3f, 0xfa, 0x6f}
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.UInt, slice.Type(), t)
	ASSERT_TRUE(slice.IsUInt(), t)
	ASSERT_EQ(velocypack.ValueLength(9), mustLength(slice.ByteSize()), t)

	ASSERT_EQ(uint64(0x6ffa3fffac664223), mustUInt(slice.GetUInt()), t)
}

func TestSliceUIntMax(t *testing.T) {
	b := velocypack.Builder{}
	must(b.AddValue(velocypack.NewUIntValue(math.MaxUint64)))
	slice := mustSlice(b.Slice())

	ASSERT_EQ(velocypack.UInt, slice.Type(), t)
	ASSERT_TRUE(slice.IsUInt(), t)
	ASSERT_EQ(velocypack.ValueLength(9), mustLength(slice.ByteSize()), t)

	ASSERT_EQ(uint64(math.MaxUint64), mustUInt(slice.GetUInt()), t)
}

func TestSliceUIntNegative1(t *testing.T) {
	b := velocypack.Builder{}
	must(b.AddValue(velocypack.NewIntValue(-3))) // SmallInt
	slice := mustSlice(b.Slice())

	ASSERT_EQ(velocypack.SmallInt, slice.Type(), t)
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsNumberOutOfRange, t)(slice.GetUInt())
}

func TestSliceUIntNegative2(t *testing.T) {
	b := velocypack.Builder{}
	must(b.AddValue(velocypack.NewIntValue(-300))) // Int
	slice := mustSlice(b.Slice())

	ASSERT_EQ(velocypack.Int, slice.Type(), t)
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsNumberOutOfRange, t)(slice.GetUInt())
}
