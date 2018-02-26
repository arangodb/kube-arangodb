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
	"encoding/json"
	"testing"

	velocypack "github.com/arangodb/go-velocypack"
)

func TestEncoderObjectEmpty(t *testing.T) {
	bytes, err := velocypack.Marshal(struct{}{})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_TRUE(s.IsEmptyObject(), t)
}

func TestEncoderObjectOneField(t *testing.T) {
	bytes, err := velocypack.Marshal(struct {
		Name string
	}{
		Name: "Max",
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"Name":"Max"}`, mustString(s.JSONString()), t)
}

func TestEncoderObjectMultipleFields(t *testing.T) {
	bytes, err := velocypack.Marshal(struct {
		Name string
		A    bool
		D    float64
		I    int
	}{
		Name: "Max",
		A:    true,
		D:    123.456,
		I:    789,
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"A":true,"D":123.456,"I":789,"Name":"Max"}`, mustString(s.JSONString()), t)
}

func TestEncoderObjectTagRename(t *testing.T) {
	bytes, err := velocypack.Marshal(struct {
		Name string  `json:"name"`
		A    bool    `json:"field9"`
		D    float64 `json:"field7"`
		I    int
	}{
		Name: "Max",
		A:    true,
		D:    123.456,
		I:    789,
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"I":789,"field7":123.456,"field9":true,"name":"Max"}`, mustString(s.JSONString()), t)
}

func TestEncoderObjectTagOmitEmptyFull(t *testing.T) {
	bytes, err := velocypack.Marshal(struct {
		Name string  `json:"name,omitempty"`
		A    bool    `json:"field9,omitempty"`
		D    float64 `json:"field7,omitempty"`
		I    int     `json:"field8,omitempty"`
	}{
		Name: "Jan",
		A:    true,
		D:    123.456,
		I:    789,
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"field7":123.456,"field8":789,"field9":true,"name":"Jan"}`, mustString(s.JSONString()), t)
}

func TestEncoderObjectTagOmitEmptyEmpty(t *testing.T) {
	bytes, err := velocypack.Marshal(struct {
		Name string  `json:"name,omitempty"`
		A    bool    `json:"field9,omitempty"`
		D    float64 `json:"field7,omitempty"`
		I    int     `json:"field8,omitempty"`
	}{
		Name: "",
		A:    false,
		D:    0.0,
		I:    0,
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_TRUE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{}`, mustString(s.JSONString()), t)
}

func TestEncoderObjectTagOmitFields(t *testing.T) {
	bytes, err := velocypack.Marshal(struct {
		Name string  `json:"name,omitempty"`
		A    bool    `json:"field9,omitempty"`
		D    float64 `json:"-"`
		I    int     `json:"-,"`
	}{
		Name: "Jan",
		A:    true,
		D:    123.456,
		I:    789,
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"-":789,"field9":true,"name":"Jan"}`, mustString(s.JSONString()), t)
}

func TestEncoderObjectNestedStruct(t *testing.T) {
	bytes, err := velocypack.Marshal(struct {
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
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"A":true,"D":123.456,"I":789,"Name":"Jan","Nested":{"Foo":999}}`, mustString(s.JSONString()), t)
}

func TestEncoderObjectNestedStructs(t *testing.T) {
	bytes, err := velocypack.Marshal(struct {
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
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"A":true,"D":123.456,"I":789,"Name":"Jan","Nested":{"Foo":999,"Nested":{"Foo":true}}}`, mustString(s.JSONString()), t)
}

func TestEncoderObjectNestedStructPtr(t *testing.T) {
	bytes, err := velocypack.Marshal(struct {
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
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"A":true,"D":123.456,"I":789,"Name":"Jan","Nested":{"Foo":999}}`, mustString(s.JSONString()), t)
}

func TestEncoderObjectNestedStructPtrNil(t *testing.T) {
	bytes, err := velocypack.Marshal(struct {
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
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"A":true,"D":123.456,"I":789,"Name":"Jan","Nested":null}`, mustString(s.JSONString()), t)
}

func TestEncoderObjectNestedStructPtrNilOmitEmpty(t *testing.T) {
	bytes, err := velocypack.Marshal(struct {
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
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"A":true,"D":123.456,"I":789,"Name":"Jan"}`, mustString(s.JSONString()), t)
}

func TestEncoderObjectNestedByteSlice(t *testing.T) {
	bytes, err := velocypack.Marshal(struct {
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
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"A":true,"D":123.456,"I":789,"Name":"Jan","Nested":"(non-representable type Binary)"}`, mustString(s.JSONString(velocypack.DumperOptions{UnsupportedTypeBehavior: velocypack.ConvertUnsupportedType})), t)
}

func TestEncoderObjectNestedIntSlice(t *testing.T) {
	bytes, err := velocypack.Marshal(struct {
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
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"A":true,"D":123.456,"I":789,"Name":"Jan","Nested":[1,2,3,4,5]}`, mustString(s.JSONString()), t)
}

func TestEncoderObjectNestedStringSlice(t *testing.T) {
	bytes, err := velocypack.Marshal(struct {
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
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"A":true,"D":123.456,"I":789,"Name":"Jan","Nested":["Aap","Noot"]}`, mustString(s.JSONString()), t)
}

func TestEncoderObjectNestedStringSliceEmpty(t *testing.T) {
	bytes, err := velocypack.Marshal(struct {
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
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"A":true,"D":123.456,"I":789,"Name":"Jan","Nested":[]}`, mustString(s.JSONString()), t)
}

func TestEncoderObjectNestedStringSliceNil(t *testing.T) {
	bytes, err := velocypack.Marshal(struct {
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
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"A":true,"D":123.456,"I":789,"Name":"Jan","Nested":null}`, mustString(s.JSONString()), t)
}

type Struct1 struct {
	Field1 int
	field2 int // Not exposed, must not be exported
}

func TestEncoderObjectStruct1(t *testing.T) {
	bytes, err := velocypack.Marshal(Struct1{
		Field1: 1,
		field2: 2,
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"Field1":1}`, mustString(s.JSONString()), t)
}

type Struct2 struct {
	Field1  bool
	Struct1 // Anonymous struct
}

func TestEncoderObjectStruct2(t *testing.T) {
	bytes, err := velocypack.Marshal(Struct2{
		Field1: true,
		Struct1: Struct1{
			Field1: 101,
			field2: 102,
		},
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"Field1":true}`, mustString(s.JSONString()), t)
}

type Struct3 struct {
	Struct1 // Anonymous struct
	Field1  bool
}

func TestEncoderObjectStruct3(t *testing.T) {
	bytes, err := velocypack.Marshal(Struct3{
		Struct1: Struct1{
			Field1: 101,
			field2: 102,
		},
		Field1: true,
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"Field1":true}`, mustString(s.JSONString()), t)
}

type Struct4 struct {
	Field4 bool `json:"a"`
	Struct5
}

type Struct5 struct {
	Field5 int `json:"a"`
}

func TestEncoderObjectStruct4(t *testing.T) {
	bytes, err := velocypack.Marshal(Struct4{
		Field4: true,
		Struct5: Struct5{
			Field5: 5,
		},
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"a":true}`, mustString(s.JSONString()), t)
}

type Struct6 struct {
	Field4 bool `json:"a6"`
	Struct5
}

func TestEncoderObjectStruct6(t *testing.T) {
	bytes, err := velocypack.Marshal(Struct6{
		Field4: true,
		Struct5: Struct5{
			Field5: 5,
		},
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"a":5,"a6":true}`, mustString(s.JSONString()), t)
}

func TestEncoderObjectStructPtr6(t *testing.T) {
	bytes, err := velocypack.Marshal(&Struct6{
		Field4: true,
		Struct5: Struct5{
			Field5: 5,
		},
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"a":5,"a6":true}`, mustString(s.JSONString()), t)
}

type Struct7 struct {
	B bool    `json:"b,string"`
	I int     `json:"i,string"`
	U uint    `json:"u,string"`
	F float64 `json:"f,string"`
	S string  `json:"s,string"`
}

func TestEncoderObjectStruct7(t *testing.T) {
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

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"b":"true","f":"3.2","i":"-77","s":"\"Hello world\"","u":"211"}`, mustString(s.JSONString()), t)

	goJSON, err := json.Marshal(input)
	ASSERT_NIL(err, t)
	ASSERT_EQ(`{"b":"true","i":"-77","u":"211","f":"3.2","s":"\"Hello world\""}`, string(goJSON), t)
}

type Struct8 struct {
	B bool    `json:",string"`
	I int     `json:",string"`
	U uint    `json:",string"`
	F float64 `json:",string"`
	S string  `json:",string"`
}

func TestEncoderObjectStruct8(t *testing.T) {
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

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_FALSE(s.IsEmptyObject(), t)
	ASSERT_EQ(`{"B":"true","F":"3.2","I":"-77","S":"\"Hello world\"","U":"211"}`, mustString(s.JSONString()), t)

	goJSON, err := json.Marshal(input)
	ASSERT_NIL(err, t)
	ASSERT_EQ(`{"B":"true","I":"-77","U":"211","F":"3.2","S":"\"Hello world\""}`, string(goJSON), t)
}
