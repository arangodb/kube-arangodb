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

func TestEncoderArrayEmptySlice(t *testing.T) {
	bytes, err := velocypack.Marshal([]struct{}{})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Array, t)
	ASSERT_TRUE(s.IsEmptyArray(), t)
	ASSERT_EQ(`[]`, mustString(s.JSONString()), t)
}

func TestEncoderArrayByteSlice(t *testing.T) {
	bytes, err := velocypack.Marshal([]byte{1, 2, 3, 4, 5})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Binary, t)
	ASSERT_TRUE(s.IsBinary(), t)
	ASSERT_EQ(`null`, mustString(s.JSONString()), t) // Dumper does not support Binary data
	ASSERT_EQ(velocypack.ValueLength(5), mustLength(s.GetBinaryLength()), t)
}

func TestEncoderArrayBoolSlice(t *testing.T) {
	bytes, err := velocypack.Marshal([]bool{true, false, false, true})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Array, t)
	ASSERT_TRUE(s.IsArray(), t)
	ASSERT_EQ(`[true,false,false,true]`, mustString(s.JSONString()), t)
}

func TestEncoderArrayIntSlice(t *testing.T) {
	bytes, err := velocypack.Marshal([]int{1, 2, 3, -4, 5, 6, 100000})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Array, t)
	ASSERT_TRUE(s.IsArray(), t)
	ASSERT_EQ(`[1,2,3,-4,5,6,100000]`, mustString(s.JSONString()), t)
}

func TestEncoderArrayUIntSlice(t *testing.T) {
	bytes, err := velocypack.Marshal([]uint{1, 2, 3, 4, 5, 6, 100000})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Array, t)
	ASSERT_TRUE(s.IsArray(), t)
	ASSERT_EQ(`[1,2,3,4,5,6,100000]`, mustString(s.JSONString()), t)
}

func TestEncoderArrayFloat32Slice(t *testing.T) {
	bytes, err := velocypack.Marshal([]float32{0.0, -1.5, 66, 45})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Array, t)
	ASSERT_TRUE(s.IsArray(), t)
	ASSERT_EQ(`[0,-1.5,66,45]`, mustString(s.JSONString()), t)
}

func TestEncoderArrayFloat64Slice(t *testing.T) {
	bytes, err := velocypack.Marshal([]float64{0.0, -1.5, 6.23, 45e+10})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Array, t)
	ASSERT_TRUE(s.IsArray(), t)
	ASSERT_EQ(`[0,-1.5,6.23,4.5e+11]`, mustString(s.JSONString()), t)
}

func TestEncoderArrayStructSlice(t *testing.T) {
	bytes, err := velocypack.Marshal([]Struct1{
		Struct1{Field1: 1, field2: 2},
		Struct1{Field1: 10, field2: 200},
		Struct1{Field1: 100, field2: 200},
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	t.Log(s.String())
	ASSERT_EQ(s.Type(), velocypack.Array, t)
	ASSERT_TRUE(s.IsArray(), t)
	ASSERT_EQ(`[{"Field1":1},{"Field1":10},{"Field1":100}]`, mustString(s.JSONString()), t)
}

func TestEncoderArrayStructPtrSlice(t *testing.T) {
	bytes, err := velocypack.Marshal([]*Struct1{
		&Struct1{Field1: 1, field2: 2},
		nil,
		&Struct1{Field1: 10, field2: 200},
		&Struct1{Field1: 100, field2: 200},
		nil,
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	t.Log(s.String())
	ASSERT_EQ(s.Type(), velocypack.Array, t)
	ASSERT_TRUE(s.IsArray(), t)
	ASSERT_EQ(`[{"Field1":1},null,{"Field1":10},{"Field1":100},null]`, mustString(s.JSONString()), t)
}

func TestEncoderArrayNestedSlice(t *testing.T) {
	bytes, err := velocypack.Marshal([][]Struct1{
		[]Struct1{Struct1{Field1: 1, field2: 2}, Struct1{Field1: 3, field2: 4}},
		[]Struct1{Struct1{Field1: 10, field2: 200}},
		[]Struct1{Struct1{Field1: 100, field2: 200}},
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	t.Log(s.String())
	ASSERT_EQ(s.Type(), velocypack.Array, t)
	ASSERT_TRUE(s.IsArray(), t)
	ASSERT_EQ(`[[{"Field1":1},{"Field1":3}],[{"Field1":10}],[{"Field1":100}]]`, mustString(s.JSONString()), t)
}
