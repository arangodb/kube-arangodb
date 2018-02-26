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
	"encoding/base64"
	"encoding/json"
	"testing"

	velocypack "github.com/arangodb/go-velocypack"
)

func TestDecoderConvertFloat32Int(t *testing.T) {
	input := struct {
		A float32
	}{
		A: -345.0,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A int
	}
	expected.A = -345
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertFloat32UInt(t *testing.T) {
	input := struct {
		A float32
	}{
		A: 333.0,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A uint
	}
	expected.A = 333
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertFloat32Number(t *testing.T) {
	input := struct {
		A float32
	}{
		A: 333.5,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A json.Number
	}
	expected.A = "333.5"
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertFloat32Interface(t *testing.T) {
	input := struct {
		A float32
	}{
		A: 333.5,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A interface{}
	}
	expected.A = 333.5
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertFloat64Int(t *testing.T) {
	input := struct {
		A float64
	}{
		A: -345.0,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A int
	}
	expected.A = -345
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertFloat64UInt(t *testing.T) {
	input := struct {
		A float64
	}{
		A: 333.0,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A uint
	}
	expected.A = 333
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertFloat64Number(t *testing.T) {
	input := struct {
		A float64
	}{
		A: 333.7,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A json.Number
	}
	expected.A = "333.7"
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertFloat64Interface(t *testing.T) {
	input := struct {
		A float64
	}{
		A: 333.7,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A interface{}
	}
	expected.A = 333.7
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertIntFloat32(t *testing.T) {
	input := struct {
		A int
	}{
		A: -123,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A float32
	}
	expected.A = -123.0
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertIntFloat64(t *testing.T) {
	input := struct {
		A int
	}{
		A: -12345,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A float64
	}
	expected.A = -12345.0
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertIntUInt(t *testing.T) {
	input := struct {
		A int
	}{
		A: 12345,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A uint
	}
	expected.A = 12345
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertIntNumber(t *testing.T) {
	input := struct {
		A int
	}{
		A: -12345,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A json.Number
	}
	expected.A = "-12345"
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertIntInterface(t *testing.T) {
	input := struct {
		A int
	}{
		A: -12345,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A interface{}
	}
	expected.A = -12345
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertUIntFloat32(t *testing.T) {
	input := struct {
		A uint
	}{
		A: 123,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A float32
	}
	expected.A = 123.0
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertUIntFloat64(t *testing.T) {
	input := struct {
		A uint
	}{
		A: 12345,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A float64
	}
	expected.A = 12345.0
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertUIntInt(t *testing.T) {
	input := struct {
		A uint
	}{
		A: 12345,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A int
	}
	expected.A = 12345
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertUIntNumber(t *testing.T) {
	input := struct {
		A uint
	}{
		A: 12345,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A json.Number
	}
	expected.A = "12345"
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertUIntInterface(t *testing.T) {
	input := struct {
		A uint
	}{
		A: 12345,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A interface{}
	}
	expected.A = uint64(12345)
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertStringByteSlice(t *testing.T) {
	expectedBytes := []byte{5, 6, 7, 8, 9}
	input := struct {
		A string
	}{
		A: base64.StdEncoding.EncodeToString(expectedBytes),
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A []byte
	}
	expected.A = expectedBytes
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderConvertNil(t *testing.T) {
	input := struct {
		A *Struct1
	}{
		A: nil,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var expected, v struct {
		A interface{}
	}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}
