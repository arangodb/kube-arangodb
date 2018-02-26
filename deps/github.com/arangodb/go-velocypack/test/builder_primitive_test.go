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
	"time"

	velocypack "github.com/arangodb/go-velocypack"
)

func TestBuilderPrimitiveAddNone(t *testing.T) {
	var b velocypack.Builder
	s := velocypack.NoneSlice()
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderUnexpectedType, t)(b.Add(s))
}

func TestBuilderPrimitiveAddNull(t *testing.T) {
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewNullValue()))
	s := mustSlice(b.Slice())
	ASSERT_TRUE(s.IsNull(), t)
}

func TestBuilderPrimitiveAddMinKey(t *testing.T) {
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewMinKeyValue()))
	s := mustSlice(b.Slice())
	ASSERT_TRUE(s.IsMinKey(), t)
}

func TestBuilderPrimitiveAddMaxKey(t *testing.T) {
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewMaxKeyValue()))
	s := mustSlice(b.Slice())
	ASSERT_TRUE(s.IsMaxKey(), t)
}

func TestBuilderPrimitiveAddBool(t *testing.T) {
	tests := []bool{true, false}
	for _, test := range tests {
		var b velocypack.Builder
		b.Add(test)

		s := mustSlice(b.Slice())
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

func TestBuilderPrimitiveAddDoubleFloat32(t *testing.T) {
	tests := []float32{10.4, -6, 0.0, -999999999, 24643783456252.4545345, math.MaxFloat32, -math.MaxFloat32}
	for _, test := range tests {
		var b velocypack.Builder
		b.Add(test)

		s := mustSlice(b.Slice())
		ASSERT_TRUE(s.IsDouble(), t)
		ASSERT_DOUBLE_EQ(float64(test), mustDouble(s.GetDouble()), t)
	}
}

func TestBuilderPrimitiveAddDoubleFloat64(t *testing.T) {
	tests := []float64{10.4, -6, 0.0, -999999999, 24643783456252.4545345, math.MaxFloat64, -math.MaxFloat64}
	for _, test := range tests {
		var b velocypack.Builder
		b.Add(test)

		s := mustSlice(b.Slice())
		ASSERT_TRUE(s.IsDouble(), t)
		ASSERT_DOUBLE_EQ(test, mustDouble(s.GetDouble()), t)
	}
}

func TestBuilderPrimitiveAddInt(t *testing.T) {
	tests := []int{10, -7, -34, 344366, math.MaxInt32, 233224, math.MinInt32}
	for _, test := range tests {
		var b velocypack.Builder
		b.Add(test)

		s := mustSlice(b.Slice())
		ASSERT_TRUE(s.IsInt(), t)
		ASSERT_EQ(int64(test), mustInt(s.GetInt()), t)
	}
}

func TestBuilderPrimitiveAddInt8(t *testing.T) {
	tests := []int8{10, -7, -34, math.MinInt8, math.MaxInt8}
	for _, test := range tests {
		var b velocypack.Builder
		b.Add(test)

		s := mustSlice(b.Slice())
		ASSERT_TRUE(s.IsInt(), t)
		ASSERT_EQ(int64(test), mustInt(s.GetInt()), t)
	}
}

func TestBuilderPrimitiveAddInt16(t *testing.T) {
	tests := []int16{10, -7, -34, math.MinInt16, math.MaxInt16}
	for _, test := range tests {
		var b velocypack.Builder
		b.Add(test)

		s := mustSlice(b.Slice())
		ASSERT_TRUE(s.IsInt(), t)
		ASSERT_EQ(int64(test), mustInt(s.GetInt()), t)
	}
}

func TestBuilderPrimitiveAddInt32(t *testing.T) {
	tests := []int32{10, -7, -34, math.MinInt32, math.MaxInt32}
	for _, test := range tests {
		var b velocypack.Builder
		b.Add(test)

		s := mustSlice(b.Slice())
		ASSERT_TRUE(s.IsInt(), t)
		ASSERT_EQ(int64(test), mustInt(s.GetInt()), t)
	}
}

func TestBuilderPrimitiveAddInt64(t *testing.T) {
	tests := []int64{10, -7, -34, math.MinInt64, math.MaxInt64}
	for _, test := range tests {
		var b velocypack.Builder
		b.Add(test)

		s := mustSlice(b.Slice())
		ASSERT_TRUE(s.IsInt(), t)
		ASSERT_EQ(int64(test), mustInt(s.GetInt()), t)
	}
}

func TestBuilderPrimitiveAddUInt(t *testing.T) {
	tests := []uint{10, 34, math.MaxUint32}
	for _, test := range tests {
		var b velocypack.Builder
		b.Add(test)

		s := mustSlice(b.Slice())
		ASSERT_TRUE(s.IsUInt(), t)
		ASSERT_EQ(uint64(test), mustUInt(s.GetUInt()), t)
	}
}

func TestBuilderPrimitiveAddUInt8(t *testing.T) {
	tests := []uint8{10, 34, math.MaxUint8}
	for _, test := range tests {
		var b velocypack.Builder
		b.Add(test)

		s := mustSlice(b.Slice())
		ASSERT_TRUE(s.IsUInt(), t)
		ASSERT_EQ(uint64(test), mustUInt(s.GetUInt()), t)
	}
}

func TestBuilderPrimitiveAddUInt16(t *testing.T) {
	tests := []uint16{10, 34, math.MaxUint16}
	for _, test := range tests {
		var b velocypack.Builder
		b.Add(test)

		s := mustSlice(b.Slice())
		ASSERT_TRUE(s.IsUInt(), t)
		ASSERT_EQ(uint64(test), mustUInt(s.GetUInt()), t)
	}
}

func TestBuilderPrimitiveAddUInt32(t *testing.T) {
	tests := []uint32{10, 34, 56345344, math.MaxUint32}
	for _, test := range tests {
		var b velocypack.Builder
		b.Add(test)

		s := mustSlice(b.Slice())
		ASSERT_TRUE(s.IsUInt(), t)
		ASSERT_EQ(uint64(test), mustUInt(s.GetUInt()), t)
	}
}

func TestBuilderPrimitiveAddUInt64(t *testing.T) {
	tests := []uint64{10, 34, 636346346345342355, math.MaxUint64}
	for _, test := range tests {
		var b velocypack.Builder
		b.Add(test)

		s := mustSlice(b.Slice())
		ASSERT_TRUE(s.IsUInt(), t)
		ASSERT_EQ(uint64(test), mustUInt(s.GetUInt()), t)
	}
}

func TestBuilderPrimitiveAddSmallInt(t *testing.T) {
	tests := []int{-6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 9}
	for _, test := range tests {
		var b velocypack.Builder
		b.Add(test)

		s := mustSlice(b.Slice())
		ASSERT_TRUE(s.IsSmallInt(), t)
		ASSERT_EQ(int64(test), mustInt(s.GetInt()), t)
	}
}

func TestBuilderPrimitiveAddString(t *testing.T) {
	tests := []string{"", "foo", "你好，世界", "\t\n\x00", "Some space and stuff"}
	for _, test := range tests {
		var b velocypack.Builder
		b.Add(test)

		s := mustSlice(b.Slice())
		ASSERT_TRUE(s.IsString(), t)
		ASSERT_EQ(test, mustString(s.GetString()), t)
	}
}

func TestBuilderPrimitiveAddUTCDate(t *testing.T) {
	cet, err := time.LoadLocation("CET")
	ASSERT_NIL(err, t)
	tests := []time.Time{
		time.Date(2585, time.January, 12, 1, 2, 3, 0, time.UTC),
		time.Date(2123, time.October, 9, 1, 2, 3, 0, time.UTC),
		time.Date(2001, time.September, 11, 1, 2, 3, 0, time.UTC),
		time.Date(1985, time.July, 4, 10, 22, 0, 0, cet),
		time.Date(1950, time.December, 5, 20, 10, 50, 0, cet),
	}
	for _, test := range tests {
		var b velocypack.Builder
		b.Add(test)

		s := mustSlice(b.Slice())
		ASSERT_TRUE(s.IsUTCDate(), t)
		ASSERT_EQ(test.UTC(), mustTime(s.GetUTCDate()), t)
	}
}

func TestBuilderPrimitiveAddBinary(t *testing.T) {
	tests := [][]byte{[]byte{1, 2, 3}, []byte{}, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 12, 13, 14, 15, 16, 17, 18, 19, 20}}
	for _, test := range tests {
		var b velocypack.Builder
		b.Add(test)

		s := mustSlice(b.Slice())
		ASSERT_EQ(s.Type(), velocypack.Binary, t)
		ASSERT_TRUE(s.IsBinary(), t)
		ASSERT_EQ(test, mustBytes(s.GetBinary()), t)
	}
}
