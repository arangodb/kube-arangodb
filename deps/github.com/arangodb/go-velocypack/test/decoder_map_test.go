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
	"reflect"
	"testing"
	"unsafe"

	velocypack "github.com/arangodb/go-velocypack"
)

func TestDecoderMapEmpty(t *testing.T) {
	expected := map[string]interface{}{}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v map[string]interface{}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderMapOneField(t *testing.T) {
	expected := map[string]string{
		"Name": "Max",
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v map[string]string
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderMapMultipleFields(t *testing.T) {
	expected := map[string]interface{}{
		"Name": "Max",
		"A":    true,
		"D":    123.456,
		"I":    789, // Will be of type int
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v map[string]interface{}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderMapMultipleFieldsInt64(t *testing.T) {
	maxInt32P1 := int64(math.MaxInt32) + 1
	var i interface{}
	if unsafe.Sizeof(int(0)) == 4 {
		i = maxInt32P1
	} else {
		i = int(maxInt32P1)
	}
	expected := map[string]interface{}{
		"Name": "Max",
		"A":    true,
		"D":    123.456,
		"I":    i, // Will be of type int or int64 depending on GOARCH
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v map[string]interface{}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(reflect.ValueOf(v["I"]).Type(), reflect.ValueOf(expected["I"]).Type(), t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderMapMultipleFieldsEmpty(t *testing.T) {
	expected := map[string]interface{}{
		"Name": "",
		"A":    false,
		"D":    0.0,
		"I":    0,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v map[string]interface{}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderMapNestedStruct(t *testing.T) {
	expected := map[string]interface{}{
		"Name": "Jan",
		"Nested": map[string]interface{}{
			"Foo": 999,
		},
		"A": true,
		"D": 123.456,
		"I": 789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v map[string]interface{}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderMapNestedStructs(t *testing.T) {
	expected := map[string]interface{}{
		"Name": "Jan",
		"Nested": map[string]interface{}{
			"Foo": 999,
			"Nested": map[string]interface{}{
				"Foo": true,
			},
		},
		"A": true,
		"D": 123.456,
		"I": 789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v map[string]interface{}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderMapNestedStructPtrNil(t *testing.T) {
	expected := map[string]interface{}{
		"Name":   "Jan",
		"Nested": nil,
		"A":      true,
		"D":      123.456,
		"I":      789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v map[string]interface{}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderMapNestedByteSlice(t *testing.T) {
	expected := map[string]interface{}{
		"Name":   "Jan",
		"Nested": []byte{1, 2, 3, 4, 5, 6},
		"A":      true,
		"D":      123.456,
		"I":      789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v map[string]interface{}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderMapNestedIntSlice(t *testing.T) {
	expected := map[string]interface{}{
		"Name":   "Jan",
		"Nested": []interface{}{1, 2, 3, 4, 5},
		"A":      true,
		"D":      123.456,
		"I":      789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v map[string]interface{}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderMapNestedStringSlice(t *testing.T) {
	expected := map[string]interface{}{
		"Name":   "Jan",
		"Nested": []interface{}{"Aap", "Noot"},
		"A":      true,
		"D":      123.456,
		"I":      789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v map[string]interface{}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderMapNestedStringSliceEmpty(t *testing.T) {
	expected := map[string]interface{}{
		"Name":   "Jan",
		"Nested": []interface{}{},
		"A":      true,
		"D":      123.456,
		"I":      789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v map[string]interface{}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderMapInt8Keys(t *testing.T) {
	expected := map[int8]interface{}{
		0:   "Jan",
		1:   []interface{}{"foo", "monkey"},
		7:   true,
		11:  123.456,
		23:  789,
		-45: false,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v map[int8]interface{}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderMapUInt16Keys(t *testing.T) {
	expected := map[uint16]interface{}{
		0:  "Jan",
		1:  []interface{}{"foo", "monkey"},
		7:  true,
		11: 123.456,
		23: 789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v map[uint16]interface{}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}
