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

func TestDecoderObjectEmpty(t *testing.T) {
	expected := struct{}{}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v struct{}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderObjectEmptyInvalidDestination(t *testing.T) {
	b := velocypack.Builder{}
	must(b.OpenObject())
	must(b.Close())
	s := mustSlice(b.Slice())

	var v int64
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsUnmarshalType, t)(velocypack.Unmarshal(s, &v))
}

func TestDecoderObjectOneField(t *testing.T) {
	expected := struct {
		Name string
	}{
		Name: "Max",
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v struct {
		Name string
	}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderObjectMultipleFields(t *testing.T) {
	expected := struct {
		Name string
		A    bool
		D    float64
		I    int
	}{
		Name: "Max",
		A:    true,
		D:    123.456,
		I:    789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v struct {
		Name string
		A    bool
		D    float64
		I    int
	}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderObjectTagRename(t *testing.T) {
	expected := struct {
		Name string  `json:"name"`
		A    bool    `json:"field9"`
		D    float64 `json:"field7"`
		I    int
	}{
		Name: "Max",
		A:    true,
		D:    123.456,
		I:    789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v struct {
		Name string  `json:"name"`
		A    bool    `json:"field9"`
		D    float64 `json:"field7"`
		I    int
	}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderObjectTagOmitEmptyFull(t *testing.T) {
	expected := struct {
		Name string  `json:"name,omitempty"`
		A    bool    `json:"field9,omitempty"`
		D    float64 `json:"field7,omitempty"`
		I    int     `json:"field8,omitempty"`
	}{
		Name: "Jan",
		A:    true,
		D:    123.456,
		I:    789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v struct {
		Name string  `json:"name,omitempty"`
		A    bool    `json:"field9,omitempty"`
		D    float64 `json:"field7,omitempty"`
		I    int     `json:"field8,omitempty"`
	}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderObjectTagOmitEmptyEmpty(t *testing.T) {
	expected := struct {
		Name string  `json:"name,omitempty"`
		A    bool    `json:"field9,omitempty"`
		D    float64 `json:"field7,omitempty"`
		I    int     `json:"field8,omitempty"`
	}{
		Name: "",
		A:    false,
		D:    0.0,
		I:    0,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v struct {
		Name string  `json:"name,omitempty"`
		A    bool    `json:"field9,omitempty"`
		D    float64 `json:"field7,omitempty"`
		I    int     `json:"field8,omitempty"`
	}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderObjectTagOmitFields(t *testing.T) {
	input := struct {
		Name string  `json:"name,omitempty"`
		A    bool    `json:"field9,omitempty"`
		D    float64 `json:"-"`
		I    int     `json:"-,"`
	}{
		Name: "Jan",
		A:    true,
		D:    123.456,
		I:    789,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)
	expected := input
	expected.D = 0.0

	var v struct {
		Name string  `json:"name,omitempty"`
		A    bool    `json:"field9,omitempty"`
		D    float64 `json:"-"`
		I    int     `json:"-,"`
	}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderObjectNestedStruct(t *testing.T) {
	expected := struct {
		Name   string
		Nested struct {
			Foo int
		}
		A bool
		D float64
		I int
	}{
		Name:   "Jan",
		Nested: struct{ Foo int }{999},
		A:      true,
		D:      123.456,
		I:      789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v struct {
		Name   string
		Nested struct {
			Foo int
		}
		A bool
		D float64
		I int
	}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderObjectNestedStructs(t *testing.T) {
	expected := struct {
		Name   string
		Nested struct {
			Foo    int
			Nested struct {
				Foo bool
			}
		}
		A bool
		D float64
		I int
	}{
		Name: "Jan",
		Nested: struct {
			Foo    int
			Nested struct{ Foo bool }
		}{999, struct{ Foo bool }{true}},
		A: true,
		D: 123.456,
		I: 789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v struct {
		Name   string
		Nested struct {
			Foo    int
			Nested struct {
				Foo bool
			}
		}
		A bool
		D float64
		I int
	}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderObjectNestedStructPtr(t *testing.T) {
	expected := struct {
		Name   string
		Nested *struct {
			Foo int
		}
		A bool
		D float64
		I int
	}{
		Name:   "Jan",
		Nested: &struct{ Foo int }{999},
		A:      true,
		D:      123.456,
		I:      789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v struct {
		Name   string
		Nested *struct {
			Foo int
		}
		A bool
		D float64
		I int
	}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderObjectNestedStructPtrNil(t *testing.T) {
	expected := struct {
		Name   string
		Nested *struct {
			Foo int
		}
		A bool
		D float64
		I int
	}{
		Name:   "Jan",
		Nested: nil,
		A:      true,
		D:      123.456,
		I:      789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v struct {
		Name   string
		Nested *struct {
			Foo int
		}
		A bool
		D float64
		I int
	}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderObjectNestedStructPtrNilOmitEmpty(t *testing.T) {
	expected := struct {
		Name   string
		Nested *struct {
			Foo int
		} `json:",omitempty"`
		A bool
		D float64
		I int
	}{
		Name:   "Jan",
		Nested: nil,
		A:      true,
		D:      123.456,
		I:      789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v struct {
		Name   string
		Nested *struct {
			Foo int
		} `json:",omitempty"`
		A bool
		D float64
		I int
	}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderObjectNestedByteSlice(t *testing.T) {
	expected := struct {
		Name   string
		Nested []byte
		A      bool
		D      float64
		I      int
	}{
		Name:   "Jan",
		Nested: []byte{1, 2, 3, 4, 5},
		A:      true,
		D:      123.456,
		I:      789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v struct {
		Name   string
		Nested []byte
		A      bool
		D      float64
		I      int
	}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderObjectNestedIntSlice(t *testing.T) {
	expected := struct {
		Name   string
		Nested []int
		A      bool
		D      float64
		I      int
	}{
		Name:   "Jan",
		Nested: []int{1, 2, 3, 4, 5},
		A:      true,
		D:      123.456,
		I:      789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v struct {
		Name   string
		Nested []int
		A      bool
		D      float64
		I      int
	}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderObjectNestedStringSlice(t *testing.T) {
	expected := struct {
		Name   string
		Nested []string
		A      bool
		D      float64
		I      int
	}{
		Name:   "Jan",
		Nested: []string{"Aap", "Noot"},
		A:      true,
		D:      123.456,
		I:      789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v struct {
		Name   string
		Nested []string
		A      bool
		D      float64
		I      int
	}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderObjectNestedStringSliceEmpty(t *testing.T) {
	expected := struct {
		Name   string
		Nested []string
		A      bool
		D      float64
		I      int
	}{
		Name:   "Jan",
		Nested: []string{},
		A:      true,
		D:      123.456,
		I:      789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v struct {
		Name   string
		Nested []string
		A      bool
		D      float64
		I      int
	}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderObjectNestedStringSliceNil(t *testing.T) {
	expected := struct {
		Name   string
		Nested []string
		A      bool
		D      float64
		I      int
	}{
		Name:   "Jan",
		Nested: nil,
		A:      true,
		D:      123.456,
		I:      789,
	}
	bytes, err := velocypack.Marshal(expected)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	var v struct {
		Name   string
		Nested []string
		A      bool
		D      float64
		I      int
	}
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

/*
type Struct1 struct {
	Field1 int
	field2 int // Not exposed, must not be exported
}
*/

func TestDecoderObjectStruct1(t *testing.T) {
	input := Struct1{
		Field1: 1,
		field2: 2,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)
	expected := input
	expected.field2 = 0

	var v Struct1
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

/*
type Struct2 struct {
	Field1  bool
	Struct1 // Anonymous struct
}
*/
func TestDecoderObjectStruct2(t *testing.T) {
	input := Struct2{
		Field1: true,
		Struct1: Struct1{
			Field1: 101,
			field2: 102,
		},
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)
	expected := input
	expected.Struct1.Field1 = 0
	expected.Struct1.field2 = 0

	var v Struct2
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

/*
type Struct3 struct {
	Struct1 // Anonymous struct
	Field1  bool
}
*/
func TestDecoderObjectStruct3(t *testing.T) {
	input := Struct3{
		Struct1: Struct1{
			Field1: 101,
			field2: 102,
		},
		Field1: true,
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)
	expected := input
	expected.Struct1.Field1 = 0
	expected.Struct1.field2 = 0

	var v Struct3
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

/*
type Struct4 struct {
	Field4 bool `json:"a"`
	Struct5
}

type Struct5 struct {
	Field5 int `json:"a"`
}
*/
func TestDecoderObjectStruct4(t *testing.T) {
	input := Struct4{
		Field4: true,
		Struct5: Struct5{
			Field5: 5,
		},
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)
	expected := input
	expected.Struct5.Field5 = 0

	var v Struct4
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

/*
type Struct6 struct {
	Field4 bool `json:"a6"`
	Struct5
}
*/
func TestDecoderObjectStruct6(t *testing.T) {
	input := Struct6{
		Field4: true,
		Struct5: Struct5{
			Field5: 5,
		},
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)
	expected := input

	var v Struct6
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

func TestDecoderObjectStructPtr6(t *testing.T) {
	input := &Struct6{
		Field4: true,
		Struct5: Struct5{
			Field5: 5,
		},
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)
	expected := *input

	var v Struct6
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

/*
type Struct7 struct {
	B bool    `json:"b,string"`
	I int     `json:"i,string"`
	U uint    `json:"u,string"`
	F float64 `json:"f,string"`
	S string  `json:"s,string"`
}
*/

func TestDecoderObjectStruct7(t *testing.T) {
	input := Struct7{
		B: true,
		I: -77,
		U: 211,
		F: 3.2,
		S: "Hello world",
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)
	expected := input

	var v Struct7
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}

/*
type Struct8 struct {
	B bool    `json:",string"`
	I int     `json:",string"`
	U uint    `json:",string"`
	F float64 `json:",string"`
	S string  `json:",string"`
}
*/

func TestDecoderObjectStruct8(t *testing.T) {
	input := Struct8{
		B: true,
		I: -77,
		U: 211,
		F: 3.2,
		S: "Hello world",
	}
	bytes, err := velocypack.Marshal(input)
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)
	expected := input

	var v Struct8
	err = velocypack.Unmarshal(s, &v)
	ASSERT_NIL(err, t)
	ASSERT_EQ(v, expected, t)
}
