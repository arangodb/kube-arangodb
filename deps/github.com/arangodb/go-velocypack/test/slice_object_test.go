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
	"strings"
	"testing"

	velocypack "github.com/arangodb/go-velocypack"
)

func TestSliceObjectEmpty(t *testing.T) {
	slice := velocypack.Slice{0x0a}
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.Object, slice.Type(), t)
	ASSERT_TRUE(slice.IsObject(), t)
	ASSERT_TRUE(slice.IsEmptyObject(), t)
	ASSERT_EQ(velocypack.ValueLength(1), mustLength(slice.ByteSize()), t)
	ASSERT_EQ(velocypack.ValueLength(0), mustLength(slice.Length()), t)

	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsIndexOutOfBounds, t)(slice.KeyAt(0))
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsIndexOutOfBounds, t)(slice.ValueAt(0))
}

func TestSliceObjectCases1(t *testing.T) {
	slice := velocypack.Slice{0x0b, 0x00, 0x03, 0x41, 0x61, 0x31, 0x41, 0x62,
		0x32, 0x41, 0x63, 0x33, 0x03, 0x06, 0x09}
	slice[1] = byte(len(slice)) // Set byte length
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.Object, slice.Type(), t)
	ASSERT_TRUE(slice.IsObject(), t)
	ASSERT_FALSE(slice.IsEmptyObject(), t)
	ASSERT_EQ(velocypack.ValueLength(len(slice)), mustLength(slice.ByteSize()), t)
	ASSERT_EQ(velocypack.ValueLength(3), mustLength(slice.Length()), t)
	ss := mustSlice(slice.Get("a"))
	ASSERT_TRUE(ss.IsSmallInt(), t)
	ASSERT_EQ(int64(1), mustInt(ss.GetInt()), t)
	ASSERT_EQ(int64(1), mustInt(mustSlice(slice.ValueAt(0)).GetInt()), t)

	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsIndexOutOfBounds, t)(slice.KeyAt(4))
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsIndexOutOfBounds, t)(slice.ValueAt(4))
}

func TestSliceObjectCases2(t *testing.T) {
	slice := velocypack.Slice{0x0b, 0x00, 0x03, 0x00, 0x00, 0x41, 0x61, 0x31, 0x41,
		0x62, 0x32, 0x41, 0x63, 0x33, 0x05, 0x08, 0x0b}
	slice[1] = byte(len(slice)) // Set byte length
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.Object, slice.Type(), t)
	ASSERT_TRUE(slice.IsObject(), t)
	ASSERT_FALSE(slice.IsEmptyObject(), t)
	ASSERT_EQ(velocypack.ValueLength(len(slice)), mustLength(slice.ByteSize()), t)
	ASSERT_EQ(velocypack.ValueLength(3), mustLength(slice.Length()), t)
	ss := mustSlice(slice.Get("a"))
	ASSERT_TRUE(ss.IsSmallInt(), t)
	ASSERT_EQ(int64(1), mustInt(ss.GetInt()), t)
}

func TestSliceObjectCases3(t *testing.T) {
	slice := velocypack.Slice{0x0b, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x41, 0x61, 0x31, 0x41, 0x62,
		0x32, 0x41, 0x63, 0x33, 0x09, 0x0c, 0x0f}
	slice[1] = byte(len(slice)) // Set byte length
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.Object, slice.Type(), t)
	ASSERT_TRUE(slice.IsObject(), t)
	ASSERT_FALSE(slice.IsEmptyObject(), t)
	ASSERT_EQ(velocypack.ValueLength(len(slice)), mustLength(slice.ByteSize()), t)
	ASSERT_EQ(velocypack.ValueLength(3), mustLength(slice.Length()), t)
	ss := mustSlice(slice.Get("a"))
	ASSERT_TRUE(ss.IsSmallInt(), t)
	ASSERT_EQ(int64(1), mustInt(ss.GetInt()), t)
}

func TestSliceObjectCases7(t *testing.T) {
	slice := velocypack.Slice{0x0c, 0x00, 0x00, 0x03, 0x00, 0x41, 0x61, 0x31, 0x41, 0x62,
		0x32, 0x41, 0x63, 0x33, 0x05, 0x00, 0x08, 0x00, 0x0b, 0x00}
	slice[1] = byte(len(slice)) // Set byte length
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.Object, slice.Type(), t)
	ASSERT_TRUE(slice.IsObject(), t)
	ASSERT_FALSE(slice.IsEmptyObject(), t)
	ASSERT_EQ(velocypack.ValueLength(len(slice)), mustLength(slice.ByteSize()), t)
	ASSERT_EQ(velocypack.ValueLength(3), mustLength(slice.Length()), t)
	ss := mustSlice(slice.Get("a"))
	ASSERT_TRUE(ss.IsSmallInt(), t)
	ASSERT_EQ(int64(1), mustInt(ss.GetInt()), t)
}

func TestSliceObjectCases8(t *testing.T) {
	slice := velocypack.Slice{0x0c, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x41, 0x61, 0x31, 0x41, 0x62, 0x32, 0x41,
		0x63, 0x33, 0x09, 0x00, 0x0c, 0x00, 0x0f, 0x00}
	slice[1] = byte(len(slice)) // Set byte length
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.Object, slice.Type(), t)
	ASSERT_TRUE(slice.IsObject(), t)
	ASSERT_FALSE(slice.IsEmptyObject(), t)
	ASSERT_EQ(velocypack.ValueLength(len(slice)), mustLength(slice.ByteSize()), t)
	ASSERT_EQ(velocypack.ValueLength(3), mustLength(slice.Length()), t)
	ss := mustSlice(slice.Get("a"))
	ASSERT_TRUE(ss.IsSmallInt(), t)
	ASSERT_EQ(int64(1), mustInt(ss.GetInt()), t)
}

func TestSliceObjectCases11(t *testing.T) {
	slice := velocypack.Slice{0x0d, 0x00, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x41,
		0x61, 0x31, 0x41, 0x62, 0x32, 0x41, 0x63, 0x33, 0x09, 0x00,
		0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x0f, 0x00, 0x00, 0x00}
	slice[1] = byte(len(slice)) // Set byte length
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.Object, slice.Type(), t)
	ASSERT_TRUE(slice.IsObject(), t)
	ASSERT_FALSE(slice.IsEmptyObject(), t)
	ASSERT_EQ(velocypack.ValueLength(len(slice)), mustLength(slice.ByteSize()), t)
	ASSERT_EQ(velocypack.ValueLength(3), mustLength(slice.Length()), t)
	ss := mustSlice(slice.Get("a"))
	ASSERT_TRUE(ss.IsSmallInt(), t)
	ASSERT_EQ(int64(1), mustInt(ss.GetInt()), t)
}

func TestSliceObjectCases13(t *testing.T) {
	slice := velocypack.Slice{0x0e, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x41,
		0x61, 0x31, 0x41, 0x62, 0x32, 0x41, 0x63, 0x33, 0x09, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	slice[1] = byte(len(slice)) // Set byte length
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.Object, slice.Type(), t)
	ASSERT_TRUE(slice.IsObject(), t)
	ASSERT_FALSE(slice.IsEmptyObject(), t)
	ASSERT_EQ(velocypack.ValueLength(len(slice)), mustLength(slice.ByteSize()), t)
	ASSERT_EQ(velocypack.ValueLength(3), mustLength(slice.Length()), t)
	ss := mustSlice(slice.Get("a"))
	ASSERT_TRUE(ss.IsSmallInt(), t)
	ASSERT_EQ(int64(1), mustInt(ss.GetInt()), t)
}

func TestSliceObjectCompact(t *testing.T) {
	slice := velocypack.Slice{0x14, 0x0f, 0x41, 0x61, 0x30, 0x41, 0x62, 0x31,
		0x41, 0x63, 0x32, 0x41, 0x64, 0x33, 0x04}
	slice[1] = byte(len(slice)) // Set byte length
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.Object, slice.Type(), t)
	ASSERT_TRUE(slice.IsObject(), t)
	ASSERT_FALSE(slice.IsEmptyObject(), t)
	ASSERT_EQ(velocypack.ValueLength(len(slice)), mustLength(slice.ByteSize()), t)
	ASSERT_EQ(velocypack.ValueLength(4), mustLength(slice.Length()), t)
	ss := mustSlice(slice.Get("a"))
	ASSERT_TRUE(ss.IsSmallInt(), t)
	ASSERT_EQ(int64(0), mustInt(ss.GetInt()), t)

	ss = mustSlice(slice.Get("b"))
	ASSERT_TRUE(ss.IsSmallInt(), t)
	ASSERT_EQ(int64(1), mustInt(ss.GetInt()), t)

	ss = mustSlice(slice.Get("d"))
	ASSERT_TRUE(ss.IsSmallInt(), t)
	ASSERT_EQ(int64(3), mustInt(ss.GetInt()), t)

	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsIndexOutOfBounds, t)(slice.KeyAt(4))
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsIndexOutOfBounds, t)(slice.ValueAt(4))
}

func TestSliceObjectNestedGet1(t *testing.T) {
	slice := mustSlice(velocypack.ParseJSONFromString(`{"a":{"b":{"c":55},"d":true}}`))

	a := mustSlice(slice.Get("a"))
	ASSERT_EQ(velocypack.Object, a.Type(), t)
	ASSERT_EQ(velocypack.ValueLength(2), mustLength(a.Length()), t)

	b := mustSlice(slice.Get("a", "b"))
	ASSERT_EQ(velocypack.Object, a.Type(), t)
	ASSERT_EQ(velocypack.ValueLength(1), mustLength(b.Length()), t)

	c := mustSlice(slice.Get("a", "b", "c"))
	ASSERT_EQ(velocypack.UInt, c.Type(), t)
	ASSERT_EQ(int64(55), mustInt(c.GetInt()), t)

	d := mustSlice(slice.Get("a", "d"))
	ASSERT_EQ(velocypack.Bool, d.Type(), t)
	ASSERT_TRUE(mustBool(d.GetBool()), t)

	// Not found
	ASSERT_EQ(velocypack.None, mustSlice(slice.Get("a", "e")).Type(), t)
	ASSERT_EQ(velocypack.None, mustSlice(slice.Get("a", "b", "f")).Type(), t)
	ASSERT_EQ(velocypack.None, mustSlice(slice.Get("g")).Type(), t)

	// Special: no path
	ASSERT_EQ(slice, mustSlice(slice.Get()), t)
}

func TestSliceObjectGetLength1(t *testing.T) {
	// Test fast path with single object field
	slice := velocypack.Slice{0x0b,
		0x07,             // Bytesize
		0x01,             // NoItems
		0x41, 0x61, 0x1a, // "a": true
		0x03, // Index of "a"
	}

	a := mustSlice(slice.Get("a"))
	ASSERT_EQ(velocypack.Bool, a.Type(), t)
	ASSERT_TRUE(mustBool(a.GetBool()), t)
}

func TestSliceObjectGetLength0(t *testing.T) {
	slice := velocypack.Slice{0x0a} // empty object

	a := mustSlice(slice.Get("a"))
	ASSERT_EQ(velocypack.None, a.Type(), t)
}

func TestSliceObjectGetLength2(t *testing.T) {
	// Test fast path with two fields (linear search of fields kicks in from 2..3 fields)
	slice := velocypack.Slice{0x0b,
		0,                // Bytesize
		0x02,             // NoItems
		0x41, 0x61, 0x1a, // "a": true
		0x32, 0x19, // "_rev": false
		0x03, 0x06, // Index of "a", "_rev"
	}
	slice[1] = byte(len(slice))

	a := mustSlice(slice.Get("a"))
	ASSERT_EQ(velocypack.Bool, a.Type(), t)
	ASSERT_TRUE(mustBool(a.GetBool()), t)

	b := mustSlice(slice.Get("_rev"))
	ASSERT_EQ(velocypack.Bool, b.Type(), t)
	ASSERT_FALSE(mustBool(b.GetBool()), t)
}

func TestSliceObjectGetLengthMany(t *testing.T) {
	fields := []string{}
	for i := 0; i <= 255; i++ {
		fields = append(fields, fmt.Sprintf(`"f%d":%d`, i, i+10))
	}
	json := fmt.Sprintf(`{%s}`, strings.Join(fields, ","))
	slice := mustSlice(velocypack.ParseJSONFromString(json))

	for i := 255; i >= 0; i-- {
		value := mustSlice(slice.Get(fmt.Sprintf("f%d", i)))
		ASSERT_EQ(velocypack.UInt, value.Type(), t)
		ASSERT_EQ(uint64(i)+10, mustUInt(value.GetUInt()), t)
	}
}

func TestSliceObjectNestedHasKey(t *testing.T) {
	slice := mustSlice(velocypack.ParseJSONFromString(`{"a":{"b":{"c":55},"d":true}}`))

	ASSERT_TRUE(mustBool(slice.HasKey("a")), t)
	ASSERT_TRUE(mustBool(slice.HasKey("a", "b")), t)
	ASSERT_TRUE(mustBool(slice.HasKey("a", "b", "c")), t)
	ASSERT_TRUE(mustBool(slice.HasKey("a", "d")), t)

	// Not found
	ASSERT_FALSE(mustBool(slice.HasKey("a", "e")), t)
	ASSERT_FALSE(mustBool(slice.HasKey("a", "b", "f")), t)
	ASSERT_FALSE(mustBool(slice.HasKey("g")), t)

	// Special: no path
	ASSERT_TRUE(mustBool(slice.HasKey()), t)
}

func TestSliceObjectKeyValueAtInvalidType(t *testing.T) {
	slice := mustSlice(velocypack.ParseJSONFromString(`77`))

	ASSERT_EQ(velocypack.UInt, slice.Type(), t)
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsInvalidType, t)(slice.KeyAt(0))
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsInvalidType, t)(slice.ValueAt(0))
}

func TestSliceObjectGetInvalidType(t *testing.T) {
	slice := mustSlice(velocypack.ParseJSONFromString(`77`))

	ASSERT_EQ(velocypack.UInt, slice.Type(), t)
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsInvalidType, t)(slice.Get("foo"))
}
