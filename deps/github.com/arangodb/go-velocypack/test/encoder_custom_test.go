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
	"fmt"
	"testing"

	velocypack "github.com/arangodb/go-velocypack"
)

type CustomStruct1 struct {
	Field1 int
}

func (cs *CustomStruct1) MarshalVPack() (velocypack.Slice, error) {
	var b velocypack.Builder
	if err := b.AddValue(velocypack.NewStringValue("Hello world")); err != nil {
		return nil, err
	}
	return b.Slice()
}

func TestEncoderCustomStruct1(t *testing.T) {
	bytes, err := velocypack.Marshal(&CustomStruct1{
		Field1: 999,
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.String, t)
	ASSERT_EQ(`"Hello world"`, mustString(s.JSONString()), t)
}

type CustomStruct1Object struct {
	Field1 int
}

func (cs *CustomStruct1Object) MarshalVPack() (velocypack.Slice, error) {
	var b velocypack.Builder
	if err := b.OpenObject(); err != nil {
		return nil, err
	}
	if err := b.AddKeyValue("foo", velocypack.NewStringValue("Hello world")); err != nil {
		return nil, err
	}
	if err := b.Close(); err != nil {
		return nil, err
	}
	return b.Slice()
}

func TestEncoderCustomCustomStruct1Object(t *testing.T) {
	bytes, err := velocypack.Marshal(&CustomStruct1Object{
		Field1: 999,
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_EQ(velocypack.ValueLength(1), mustLength(s.Length()), t)

	ss := mustSlice(s.Get("foo"))
	ASSERT_EQ(ss.Type(), velocypack.String, t)
	ASSERT_EQ(`"Hello world"`, mustString(ss.JSONString()), t)
}

type CustomStruct1Array struct {
	Field1 int
}

func (cs *CustomStruct1Array) MarshalVPack() (velocypack.Slice, error) {
	var b velocypack.Builder
	if err := b.OpenArray(); err != nil {
		return nil, err
	}
	if err := b.AddValue(velocypack.NewStringValue("Hello world Array")); err != nil {
		return nil, err
	}
	if err := b.Close(); err != nil {
		return nil, err
	}
	return b.Slice()
}

func TestEncoderCustomCustomStruct1Array(t *testing.T) {
	bytes, err := velocypack.Marshal(&CustomStruct1Array{
		Field1: 999,
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Array, t)
	ASSERT_EQ(velocypack.ValueLength(1), mustLength(s.Length()), t)

	ss := mustSlice(s.At(0))
	ASSERT_EQ(ss.Type(), velocypack.String, t)
	ASSERT_EQ(`"Hello world Array"`, mustString(ss.JSONString()), t)
}

type CustomStruct2 struct {
	Field CustomStruct1
}

func TestEncoderCustomStruct2(t *testing.T) {
	bytes, err := velocypack.Marshal(CustomStruct2{
		Field: CustomStruct1{
			Field1: 999222,
		},
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_EQ(`{"Field":{"Field1":999222}}`, mustString(s.JSONString()), t)
}

type CustomStruct3 struct {
	Field *CustomStruct1
}

func TestEncoderCustomStruct3(t *testing.T) {
	bytes, err := velocypack.Marshal(CustomStruct3{
		Field: &CustomStruct1{
			Field1: 999222,
		},
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_EQ(`{"Field":"Hello world"}`, mustString(s.JSONString()), t)
}

type CustomText1 struct {
	I int
}

func (ct CustomText1) MarshalText() ([]byte, error) {
	key := fmt.Sprintf("key%d", ct.I)
	return []byte(key), nil
}

func TestEncoderCustomText1(t *testing.T) {
	bytes, err := velocypack.Marshal(map[CustomText1]bool{
		CustomText1{7}: true,
		CustomText1{2}: false,
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_EQ(`{"key2":false,"key7":true}`, mustString(s.JSONString()), t)
}

type CustomJSONStruct1 struct {
	Field1 int
}

func (cs *CustomJSONStruct1) MarshalJSON() ([]byte, error) {
	return json.Marshal("Hello JSON")
}

func TestEncoderCustomJSONStruct1(t *testing.T) {
	bytes, err := velocypack.Marshal(&CustomJSONStruct1{
		Field1: 999,
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.String, t)
	ASSERT_EQ(`"Hello JSON"`, mustString(s.JSONString()), t)
}

type CustomJSONStruct1Object struct {
	Field1 int
}

func (cs *CustomJSONStruct1Object) MarshalJSON() ([]byte, error) {
	return []byte(`{"foo":"Hello JSON Object"}`), nil
}

func TestEncoderCustomJSONStruct1Object(t *testing.T) {
	bytes, err := velocypack.Marshal(&CustomJSONStruct1Object{
		Field1: 999,
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Object, t)
	ASSERT_EQ(velocypack.ValueLength(1), mustLength(s.Length()), t)

	ss := mustSlice(s.Get("foo"))
	ASSERT_EQ(ss.Type(), velocypack.String, t)
	ASSERT_EQ(`"Hello JSON Object"`, mustString(ss.JSONString()), t)
}

type CustomJSONStruct1Array struct {
	Field1 int
}

func (cs *CustomJSONStruct1Array) MarshalJSON() ([]byte, error) {
	return []byte(`["Hello JSON Array"]`), nil
}

func TestEncoderCustomJSONStruct1Array(t *testing.T) {
	bytes, err := velocypack.Marshal(&CustomJSONStruct1Array{
		Field1: 999,
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.Array, t)
	ASSERT_EQ(velocypack.ValueLength(1), mustLength(s.Length()), t)

	ss := mustSlice(s.At(0))
	ASSERT_EQ(ss.Type(), velocypack.String, t)
	ASSERT_EQ(`"Hello JSON Array"`, mustString(ss.JSONString()), t)
}

type CustomJSONVPACKStruct1 struct {
	Field1 int
}

func (cs *CustomJSONVPACKStruct1) MarshalVPack() (velocypack.Slice, error) {
	var b velocypack.Builder
	if err := b.AddValue(velocypack.NewStringValue("Hello VPACK, goodbye JSON")); err != nil {
		return nil, err
	}
	return b.Slice()
}

func (cs *CustomJSONVPACKStruct1) MarshalJSON() ([]byte, error) {
	return json.Marshal("Hello JSON, goodbye VPACK")
}

func TestEncoderCustomJSONVPACKStruct1(t *testing.T) {
	// MarshalVPack is preferred over MarshalJSON
	bytes, err := velocypack.Marshal(&CustomJSONVPACKStruct1{
		Field1: 999,
	})
	ASSERT_NIL(err, t)
	s := velocypack.Slice(bytes)

	ASSERT_EQ(s.Type(), velocypack.String, t)
	ASSERT_EQ(`"Hello VPACK, goodbye JSON"`, mustString(s.JSONString()), t)
}
