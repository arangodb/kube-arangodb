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

func TestDecoderArrayEmpty(t *testing.T) {
	b := velocypack.Builder{}
	must(b.OpenArray())
	must(b.Close())
	s := mustSlice(b.Slice())

	var v []struct{}
	err := velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(0, len(v), t)
}

func TestDecoderArrayByteSlice(t *testing.T) {
	expected := []byte{1, 2, 3, 4, 5}
	b := velocypack.Builder{}
	must(b.AddValue(velocypack.NewBinaryValue(expected)))
	s := mustSlice(b.Slice())

	var v []byte
	err := velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderArrayBoolSlice(t *testing.T) {
	expected := []bool{true, false, false, true}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v []bool
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderArrayIntSlice(t *testing.T) {
	expected := []int{1, 2, 3, -4, 5, 6, 100000}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v []int
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderArrayUIntSlice(t *testing.T) {
	expected := []uint{1, 2, 3, 4, 5, 6, 100000}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v []uint
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderArrayFloat32Slice(t *testing.T) {
	expected := []float32{0.0, -1.5, 66, 45}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v []float32
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderArrayFloat64Slice(t *testing.T) {
	expected := []float64{0.0, -1.5, 6.23, 45e+10}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v []float64
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderArrayStructSlice(t *testing.T) {
	input := []Struct1{
		Struct1{Field1: 1, field2: 2},
		Struct1{Field1: 10, field2: 200},
		Struct1{Field1: 100, field2: 200},
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)
	expected := input
	expected[0].field2 = 0
	expected[1].field2 = 0
	expected[2].field2 = 0

	var v []Struct1
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderArrayStructPtrSlice(t *testing.T) {
	input := []*Struct1{
		&Struct1{Field1: 1, field2: 2},
		nil,
		&Struct1{Field1: 10, field2: 200},
		&Struct1{Field1: 100, field2: 200},
		nil,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)
	expected := input
	expected[0].field2 = 0
	expected[2].field2 = 0
	expected[3].field2 = 0

	var v []*Struct1
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderArrayNestedArray(t *testing.T) {
	input := [][]Struct1{
		[]Struct1{Struct1{Field1: 1, field2: 2}, Struct1{Field1: 3, field2: 4}},
		[]Struct1{Struct1{Field1: 10, field2: 200}},
		[]Struct1{Struct1{Field1: 100, field2: 200}},
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)
	expected := input
	expected[0][0].field2 = 0
	expected[0][1].field2 = 0
	expected[1][0].field2 = 0
	expected[2][0].field2 = 0

	var v [][]Struct1
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderArrayExtraLengthInSlice(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6, 7, 8}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)
	expected := input

	v := make([]int, 16)
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderArrayExtraLengthInArray(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6, 7, 8}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)
	expected := [16]int{}
	copy(expected[:], input)

	v := [16]int{}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderArrayInterface(t *testing.T) {
	input := []interface{}{1, false, Struct1{}, "foo", []byte{1, 2, 3, 4, 5}}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)
	expected := input
	expected[2] = map[string]interface{}{
		"Field1": 0,
	}

	var v []interface{}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}
