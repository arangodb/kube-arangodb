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

func TestDecoderPrimitiveAddNull(t *testing.T) {
	var v interface{}
	err := velocypack.Unmarshal(velocypack.NullSlice(), &v)
	ASSERT_NIL(err, t)
	ASSERT_NIL(v, t)
}

func TestDecoderPrimitiveAddBool(t *testing.T) {
	var v bool
	err := velocypack.Unmarshal(velocypack.TrueSlice(), &v)
	ASSERT_NIL(err, t)
	ASSERT_TRUE(v, t)

	err = velocypack.Unmarshal(velocypack.FalseSlice(), &v)
	ASSERT_NIL(err, t)
	ASSERT_FALSE(v, t)
}

func TestDecoderPrimitiveAddDoubleFloat32(t *testing.T) {
	tests := []float32{10.4, -6, 0.0, -999999999, 24643783456252.4545345, math.MaxFloat32, -math.MaxFloat32}
	for _, test := range tests {
		b := velocypack.Builder{}
		b.AddValue(velocypack.NewDoubleValue(float64(test)))
		s, err := b.Slice()
		ASSERT_NIL(err, t)

		var v float32
		err = velocypack.Unmarshal(s, &v)
		ASSERT_NIL(err, t)
		ASSERT_DOUBLE_EQ(float64(v), float64(test), t)
	}
}

func TestDecoderPrimitiveAddDoubleFloat64(t *testing.T) {
	tests := []float64{10.4, -6, 0.0, -999999999, 24643783456252.4545345, math.MaxFloat64, -math.MaxFloat64}
	for _, test := range tests {
		b := velocypack.Builder{}
		b.AddValue(velocypack.NewDoubleValue(test))
		s, err := b.Slice()
		ASSERT_NIL(err, t)

		var v float64
		err = velocypack.Unmarshal(s, &v)
		ASSERT_NIL(err, t)
		ASSERT_DOUBLE_EQ(v, test, t)
	}
}

func TestDecoderPrimitiveAddInt(t *testing.T) {
	tests := []int{10, -7, -34, 344366, math.MaxInt32, 233224, math.MinInt32}
	for _, test := range tests {
		b := velocypack.Builder{}
		b.AddValue(velocypack.NewIntValue(int64(test)))
		s, err := b.Slice()
		ASSERT_NIL(err, t)

		var v int
		err = velocypack.Unmarshal(s, &v)
		ASSERT_NIL(err, t)
		ASSERT_EQ(v, test, t)
	}
}

func TestDecoderPrimitiveAddInt8(t *testing.T) {
	tests := []int8{10, -7, -34, math.MinInt8, math.MaxInt8}
	for _, test := range tests {
		b := velocypack.Builder{}
		b.AddValue(velocypack.NewIntValue(int64(test)))
		s, err := b.Slice()
		ASSERT_NIL(err, t)

		var v int8
		err = velocypack.Unmarshal(s, &v)
		ASSERT_NIL(err, t)
		ASSERT_EQ(v, test, t)
	}
}

func TestDecoderPrimitiveAddInt16(t *testing.T) {
	tests := []int16{10, -7, -34, math.MinInt16, math.MaxInt16}
	for _, test := range tests {
		b := velocypack.Builder{}
		b.AddValue(velocypack.NewIntValue(int64(test)))
		s, err := b.Slice()
		ASSERT_NIL(err, t)

		var v int16
		err = velocypack.Unmarshal(s, &v)
		ASSERT_NIL(err, t)
		ASSERT_EQ(v, test, t)
	}
}

func TestDecoderPrimitiveAddInt32(t *testing.T) {
	tests := []int32{10, -7, -34, math.MinInt32, math.MaxInt32}
	for _, test := range tests {
		b := velocypack.Builder{}
		b.AddValue(velocypack.NewIntValue(int64(test)))
		s, err := b.Slice()
		ASSERT_NIL(err, t)

		var v int32
		err = velocypack.Unmarshal(s, &v)
		ASSERT_NIL(err, t)
		ASSERT_EQ(v, test, t)
	}
}

func TestDecoderPrimitiveAddInt64(t *testing.T) {
	tests := []int64{10, -7, -34, math.MinInt64, math.MaxInt64}
	for _, test := range tests {
		b := velocypack.Builder{}
		b.AddValue(velocypack.NewIntValue(test))
		s, err := b.Slice()
		ASSERT_NIL(err, t)

		var v int64
		err = velocypack.Unmarshal(s, &v)
		ASSERT_NIL(err, t)
		ASSERT_EQ(v, test, t)
	}
}

func TestDecoderPrimitiveAddUInt(t *testing.T) {
	tests := []uint{10, 34, math.MaxUint32}
	for _, test := range tests {
		b := velocypack.Builder{}
		b.AddValue(velocypack.NewUIntValue(uint64(test)))
		s, err := b.Slice()
		ASSERT_NIL(err, t)

		var v uint
		err = velocypack.Unmarshal(s, &v)
		ASSERT_NIL(err, t)
		ASSERT_EQ(v, test, t)
	}
}

func TestDecoderPrimitiveAddUInt8(t *testing.T) {
	tests := []uint8{10, 34, math.MaxUint8}
	for _, test := range tests {
		b := velocypack.Builder{}
		b.AddValue(velocypack.NewUIntValue(uint64(test)))
		s, err := b.Slice()
		ASSERT_NIL(err, t)

		var v uint8
		err = velocypack.Unmarshal(s, &v)
		ASSERT_NIL(err, t)
		ASSERT_EQ(v, test, t)
	}
}

func TestDecoderPrimitiveAddUInt16(t *testing.T) {
	tests := []uint16{10, 34, math.MaxUint16}
	for _, test := range tests {
		b := velocypack.Builder{}
		b.AddValue(velocypack.NewUIntValue(uint64(test)))
		s, err := b.Slice()
		ASSERT_NIL(err, t)

		var v uint16
		err = velocypack.Unmarshal(s, &v)
		ASSERT_NIL(err, t)
		ASSERT_EQ(v, test, t)
	}
}

func TestDecoderPrimitiveAddUInt32(t *testing.T) {
	tests := []uint32{10, 34, 56345344, math.MaxUint32}
	for _, test := range tests {
		b := velocypack.Builder{}
		b.AddValue(velocypack.NewUIntValue(uint64(test)))
		s, err := b.Slice()
		ASSERT_NIL(err, t)

		var v uint32
		err = velocypack.Unmarshal(s, &v)
		ASSERT_NIL(err, t)
		ASSERT_EQ(v, test, t)
	}
}

func TestDecoderPrimitiveAddUInt64(t *testing.T) {
	tests := []uint64{10, 34, 636346346345342355, 0, math.MaxUint64}
	for _, test := range tests {
		b := velocypack.Builder{}
		b.AddValue(velocypack.NewUIntValue(test))
		s, err := b.Slice()
		ASSERT_NIL(err, t)

		var v uint64
		err = velocypack.Unmarshal(s, &v)
		ASSERT_NIL(err, t)
		ASSERT_EQ(v, test, t)
	}
}

func TestDecoderPrimitiveAddSmallInt(t *testing.T) {
	tests := []int{-6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 9}
	for _, test := range tests {
		b := velocypack.Builder{}
		b.AddValue(velocypack.NewIntValue(int64(test)))
		s, err := b.Slice()
		ASSERT_NIL(err, t)

		var v int
		err = velocypack.Unmarshal(s, &v)
		ASSERT_NIL(err, t)
		ASSERT_EQ(v, test, t)
	}
}

func TestDecoderPrimitiveAddString(t *testing.T) {
	tests := []string{"", "foo", "你好，世界", "\t\n\x00", "Some space and stuff"}
	for _, test := range tests {
		b := velocypack.Builder{}
		b.AddValue(velocypack.NewStringValue(test))
		s, err := b.Slice()
		ASSERT_NIL(err, t)

		var v string
		err = velocypack.Unmarshal(s, &v)
		ASSERT_NIL(err, t)
		ASSERT_EQ(v, test, t)
	}
}

func TestDecoderPrimitiveAddBinary(t *testing.T) {
	tests := [][]byte{[]byte{1, 2, 3}, []byte{}, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 12, 13, 14, 15, 16, 17, 18, 19, 20}}
	for _, test := range tests {
		b := velocypack.Builder{}
		b.AddValue(velocypack.NewBinaryValue(test))
		s, err := b.Slice()
		ASSERT_NIL(err, t)

		var v []byte
		err = velocypack.Unmarshal(s, &v)
		ASSERT_NIL(err, t)
		ASSERT_EQ(v, test, t)
	}
}
