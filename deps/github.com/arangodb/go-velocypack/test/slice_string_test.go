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

func TestSliceStringNoString(t *testing.T) {
	slice := velocypack.Slice{}
	assertEqualFromReader(t, slice)

	ASSERT_FALSE(slice.IsString(), t)
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsInvalidType, t)(slice.GetString())
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsInvalidType, t)(slice.GetStringLength())
}

func TestSliceStringEmpty(t *testing.T) {
	slice := velocypack.Slice{0x40}
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.String, slice.Type(), t)
	ASSERT_TRUE(slice.IsString(), t)
	ASSERT_EQ(velocypack.ValueLength(1), mustLength(slice.ByteSize()), t)
	ASSERT_EQ("", mustString(slice.GetString()), t)
	ASSERT_EQ(velocypack.ValueLength(0), mustLength(slice.GetStringLength()), t)
	ASSERT_EQ(0, mustGoInt(slice.CompareString("")), t)
}

func TestSliceStringLengths(t *testing.T) {
	for i := 0; i < 255; i++ {
		builder := velocypack.Builder{}
		temp := ""
		for j := 0; j < i; j++ {
			temp = temp + "x"
		}
		must(builder.AddValue(velocypack.NewStringValue(temp)))
		slice := mustSlice(builder.Slice())

		ASSERT_TRUE(slice.IsString(), t)
		ASSERT_EQ(velocypack.String, slice.Type(), t)
		ASSERT_EQ(0, mustGoInt(slice.CompareString(temp)), t)
		ASSERT_EQ(temp, mustString(slice.GetString()), t)

		ASSERT_EQ(velocypack.ValueLength(i), mustLength(slice.GetStringLength()), t)

		if i <= 126 {
			ASSERT_EQ(velocypack.ValueLength(i+1), mustLength(slice.ByteSize()), t)
		} else {
			ASSERT_EQ(velocypack.ValueLength(i+9), mustLength(slice.ByteSize()), t)
		}
	}
}

func TestSliceString1(t *testing.T) {
	value := "foobar"
	slice := velocypack.Slice(append([]byte{byte(0x40 + len(value))}, value...))
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.String, slice.Type(), t)
	ASSERT_TRUE(slice.IsString(), t)
	ASSERT_EQ(velocypack.ValueLength(7), mustLength(slice.ByteSize()), t)
	ASSERT_EQ(value, mustString(slice.GetString()), t)
	ASSERT_EQ(velocypack.ValueLength(len(value)), mustLength(slice.GetStringLength()), t)
}

func TestSliceString2(t *testing.T) {
	slice := velocypack.Slice{0x48, '1', '2', '3', 'f', '\r', '\t', '\n', 'x'}
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.String, slice.Type(), t)
	ASSERT_TRUE(slice.IsString(), t)
	ASSERT_EQ(velocypack.ValueLength(9), mustLength(slice.ByteSize()), t)
	ASSERT_EQ("123f\r\t\nx", mustString(slice.GetString()), t)
	ASSERT_EQ(velocypack.ValueLength(8), mustLength(slice.GetStringLength()), t)
}

func TestSliceStringNullBytes(t *testing.T) {
	slice := velocypack.Slice{0x48, 0, '1', '2', 0, '3', '4', 0, 'x'}
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.String, slice.Type(), t)
	ASSERT_TRUE(slice.IsString(), t)
	ASSERT_EQ(velocypack.ValueLength(9), mustLength(slice.ByteSize()), t)
	ASSERT_EQ("\x0012\x0034\x00x", mustString(slice.GetString()), t)
	ASSERT_EQ(velocypack.ValueLength(8), mustLength(slice.GetStringLength()), t)
}

func TestSliceStringLong(t *testing.T) {
	slice := velocypack.Slice{0xbf, 6, 0, 0, 0, 0, 0, 0, 0, 'f', 'o', 'o', 'b', 'a', 'r'}
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.String, slice.Type(), t)
	ASSERT_TRUE(slice.IsString(), t)
	ASSERT_EQ(velocypack.ValueLength(15), mustLength(slice.ByteSize()), t)
	ASSERT_EQ("foobar", mustString(slice.GetString()), t)
	ASSERT_EQ(velocypack.ValueLength(6), mustLength(slice.GetStringLength()), t)
}

func TestSliceStringToStringNull(t *testing.T) {
	slice := velocypack.NullSlice()
	assertEqualFromReader(t, slice)

	ASSERT_EQ("null", mustString(slice.JSONString()), t)
}
