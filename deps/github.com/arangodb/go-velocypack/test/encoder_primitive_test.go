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

func TestEncoderPrimitiveAddNull(t *testing.T) {
	bytes, err := velocypack.Marshal(nil)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)
	ASSERT_EQ(s.Type(), velocypack.Null, t)
	ASSERT_TRUE(s.IsNull(), t)
}

func TestEncoderPrimitiveAddBool(t *testing.T) {
	tests := []bool{true, false}
	for _, test := range tests {
		bytes, err := velocypack.Marshal(test)
		ASSERT_NIL(err, t)
		s := velocypack.Slice(bytes)

		ASSERT_TRUE(s.IsBool(), t)
		if test {
			ASSERT_TRUE(s.IsTrue(), t)
			ASSERT_FALSE(s.IsFalse(), t)
		} else {
			ASSERT_FALSE(s.IsTrue(), t)
			ASSERT_TRUE(s.IsFalse(), t)
		}
	}
}

func TestEncoderPrimitiveAddDoubleFloat32(t *testing.T) {
	tests := []float32{10.4, -6, 0.0, -999999999, 24643783456252.4545345, math.MaxFloat32, -math.MaxFloat32}
	for _, test := range tests {
		bytes, err := velocypack.Marshal(test)
		ASSERT_NIL(err, t)
		s := velocypack.Slice(bytes)

		ASSERT_TRUE(s.IsDouble(), t)
		ASSERT_DOUBLE_EQ(float64(test), mustDouble(s.GetDouble()), t)
	}
}

func TestEncoderPrimitiveAddDoubleFloat64(t *testing.T) {
	tests := []float64{10.4, -6, 0.0, -999999999, 24643783456252.4545345, math.MaxFloat64, -math.MaxFloat64}
	for _, test := range tests {
		bytes, err := velocypack.Marshal(test)
		ASSERT_NIL(err, t)
		s := velocypack.Slice(bytes)

		ASSERT_TRUE(s.IsDouble(), t)
		ASSERT_DOUBLE_EQ(test, mustDouble(s.GetDouble()), t)
	}
}

func TestEncoderPrimitiveAddInt(t *testing.T) {
	tests := []int{10, -7, -34, 344366, math.MaxInt32, 233224, math.MinInt32}
	for _, test := range tests {
		bytes, err := velocypack.Marshal(test)
		ASSERT_NIL(err, t)
		s := velocypack.Slice(bytes)

		ASSERT_TRUE(s.IsInt(), t)
		ASSERT_EQ(int64(test), mustInt(s.GetInt()), t)
	}
}

func TestEncoderPrimitiveAddInt8(t *testing.T) {
	tests := []int8{10, -7, -34, math.MinInt8, math.MaxInt8}
	for _, test := range tests {
		bytes, err := velocypack.Marshal(test)
		ASSERT_NIL(err, t)
		s := velocypack.Slice(bytes)

		ASSERT_TRUE(s.IsInt(), t)
		ASSERT_EQ(int64(test), mustInt(s.GetInt()), t)
	}
}

func TestEncoderPrimitiveAddInt16(t *testing.T) {
	tests := []int16{10, -7, -34, math.MinInt16, math.MaxInt16}
	for _, test := range tests {
		bytes, err := velocypack.Marshal(test)
		ASSERT_NIL(err, t)
		s := velocypack.Slice(bytes)

		ASSERT_TRUE(s.IsInt(), t)
		ASSERT_EQ(int64(test), mustInt(s.GetInt()), t)
	}
}

func TestEncoderPrimitiveAddInt32(t *testing.T) {
	tests := []int32{10, -7, -34, math.MinInt32, math.MaxInt32}
	for _, test := range tests {
		bytes, err := velocypack.Marshal(test)
		ASSERT_NIL(err, t)
		s := velocypack.Slice(bytes)

		ASSERT_TRUE(s.IsInt(), t)
		ASSERT_EQ(int64(test), mustInt(s.GetInt()), t)
	}
}

func TestEncoderPrimitiveAddInt64(t *testing.T) {
	tests := []int64{10, -7, -34, math.MinInt64, math.MaxInt64}
	for _, test := range tests {
		bytes, err := velocypack.Marshal(test)
		ASSERT_NIL(err, t)
		s := velocypack.Slice(bytes)

		ASSERT_TRUE(s.IsInt(), t)
		ASSERT_EQ(int64(test), mustInt(s.GetInt()), t)
	}
}

func TestEncoderPrimitiveAddUInt(t *testing.T) {
	tests := []uint{10, 34, math.MaxUint32}
	for _, test := range tests {
		bytes, err := velocypack.Marshal(test)
		ASSERT_NIL(err, t)
		s := velocypack.Slice(bytes)

		ASSERT_TRUE(s.IsUInt(), t)
		ASSERT_EQ(uint64(test), mustUInt(s.GetUInt()), t)
	}
}

func TestEncoderPrimitiveAddUInt8(t *testing.T) {
	tests := []uint8{10, 34, math.MaxUint8}
	for _, test := range tests {
		bytes, err := velocypack.Marshal(test)
		ASSERT_NIL(err, t)
		s := velocypack.Slice(bytes)

		ASSERT_TRUE(s.IsUInt(), t)
		ASSERT_EQ(uint64(test), mustUInt(s.GetUInt()), t)
	}
}

func TestEncoderPrimitiveAddUInt16(t *testing.T) {
	tests := []uint16{10, 34, math.MaxUint16}
	for _, test := range tests {
		bytes, err := velocypack.Marshal(test)
		ASSERT_NIL(err, t)
		s := velocypack.Slice(bytes)

		ASSERT_TRUE(s.IsUInt(), t)
		ASSERT_EQ(uint64(test), mustUInt(s.GetUInt()), t)
	}
}

func TestEncoderPrimitiveAddUInt32(t *testing.T) {
	tests := []uint32{10, 34, 56345344, math.MaxUint32}
	for _, test := range tests {
		bytes, err := velocypack.Marshal(test)
		ASSERT_NIL(err, t)
		s := velocypack.Slice(bytes)

		ASSERT_TRUE(s.IsUInt(), t)
		ASSERT_EQ(uint64(test), mustUInt(s.GetUInt()), t)
	}
}

func TestEncoderPrimitiveAddUInt64(t *testing.T) {
	tests := []uint64{10, 34, 636346346345342355, 0, math.MaxUint64}
	for _, test := range tests {
		bytes, err := velocypack.Marshal(test)
		ASSERT_NIL(err, t)
		s := velocypack.Slice(bytes)

		if test == 0 {
			ASSERT_TRUE(s.IsSmallInt(), t)
		} else {
			ASSERT_TRUE(s.IsUInt(), t)
		}
		ASSERT_EQ(uint64(test), mustUInt(s.GetUInt()), t)
	}
}

func TestEncoderPrimitiveAddSmallInt(t *testing.T) {
	tests := []int{-6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 9}
	for _, test := range tests {
		bytes, err := velocypack.Marshal(test)
		ASSERT_NIL(err, t)
		s := velocypack.Slice(bytes)

		ASSERT_TRUE(s.IsSmallInt(), t)
		ASSERT_EQ(int64(test), mustInt(s.GetInt()), t)
	}
}

func TestEncoderPrimitiveAddString(t *testing.T) {
	tests := []string{"", "foo", "你好，世界", "\t\n\x00", "Some space and stuff"}
	for _, test := range tests {
		bytes, err := velocypack.Marshal(test)
		ASSERT_NIL(err, t)
		s := velocypack.Slice(bytes)

		ASSERT_TRUE(s.IsString(), t)
		ASSERT_EQ(test, mustString(s.GetString()), t)
	}
}

func TestEncoderPrimitiveAddBinary(t *testing.T) {
	tests := [][]byte{[]byte{1, 2, 3}, []byte{}, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 12, 13, 14, 15, 16, 17, 18, 19, 20}}
	for _, test := range tests {
		bytes, err := velocypack.Marshal(test)
		ASSERT_NIL(err, t)
		s := velocypack.Slice(bytes)

		ASSERT_EQ(s.Type(), velocypack.Binary, t)
		ASSERT_TRUE(s.IsBinary(), t)
		ASSERT_EQ(test, mustBytes(s.GetBinary()), t)
	}
}
