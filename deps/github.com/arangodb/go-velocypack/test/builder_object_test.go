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
	"encoding/binary"
	"math"
	"testing"

	velocypack "github.com/arangodb/go-velocypack"
)

func TestBuilderEmptyObject(t *testing.T) {
	var b velocypack.Builder
	b.OpenObject()
	b.Close()

	s := mustSlice(b.Slice())
	ASSERT_TRUE(s.IsObject(), t)
	ASSERT_EQ(velocypack.ValueLength(0), mustLength(s.Length()), t)
}

func TestBuilderObjectEmpty(t *testing.T) {
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewObjectValue()))
	must(b.Close())
	l := mustLength(b.Size())
	result := mustBytes(b.Bytes())

	correctResult := []byte{0x0a}

	ASSERT_EQ(velocypack.ValueLength(len(correctResult)), l, t)
	ASSERT_EQ(result, correctResult, t)
}

func TestBuilderObjectEmptyCompact(t *testing.T) {
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewObjectValue(true)))
	must(b.Close())
	l := mustLength(b.Size())
	result := mustBytes(b.Bytes())

	correctResult := []byte{0x0a}

	ASSERT_EQ(velocypack.ValueLength(len(correctResult)), l, t)
	ASSERT_EQ(result, correctResult, t)
}

func TestBuilderObjectSorted(t *testing.T) {
	var b velocypack.Builder
	value := 2.3
	must(b.AddValue(velocypack.NewObjectValue()))
	must(b.AddKeyValue("d", velocypack.NewUIntValue(1200)))
	must(b.AddKeyValue("c", velocypack.NewDoubleValue(value)))
	must(b.AddKeyValue("b", velocypack.NewStringValue("abc")))
	must(b.AddKeyValue("a", velocypack.NewBoolValue(true)))
	must(b.Close())
	l := mustLength(b.Size())
	result := mustBytes(b.Bytes())

	correctResult := []byte{
		0x0b, 0x20, 0x04, 0x41, 0x64, 0x29, 0xb0, 0x04, // "d": uint(1200) =
		// 0x4b0
		0x41, 0x63, 0x1b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// "c": double(2.3)
		0x41, 0x62, 0x43, 0x61, 0x62, 0x63, // "b": "abc"
		0x41, 0x61, 0x1a, // "a": true
		0x19, 0x13, 0x08, 0x03}
	binary.LittleEndian.PutUint64(correctResult[11:], math.Float64bits(value))

	ASSERT_EQ(velocypack.ValueLength(len(correctResult)), l, t)
	ASSERT_EQ(result, correctResult, t)
}

func TestBuilderObjectCompact(t *testing.T) {
	var b velocypack.Builder
	value := 2.3
	must(b.AddValue(velocypack.NewObjectValue(true)))
	must(b.AddKeyValue("d", velocypack.NewUIntValue(1200)))
	must(b.AddKeyValue("c", velocypack.NewDoubleValue(value)))
	must(b.AddKeyValue("b", velocypack.NewStringValue("abc")))
	must(b.AddKeyValue("a", velocypack.NewBoolValue(true)))
	must(b.Close())
	l := mustLength(b.Size())
	result := mustBytes(b.Bytes())

	correctResult := []byte{
		0x14, 0x1c, 0x41, 0x64, 0x29, 0xb0, 0x04, 0x41, 0x63, 0x1b,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // double
		0x41, 0x62, 0x43, 0x61, 0x62, 0x63, 0x41, 0x61, 0x1a, 0x04}
	binary.LittleEndian.PutUint64(correctResult[10:], math.Float64bits(value))

	ASSERT_EQ(velocypack.ValueLength(len(correctResult)), l, t)
	ASSERT_EQ(result, correctResult, t)
}

func TestBuilderObjectValue1(t *testing.T) {
	var b velocypack.Builder
	u := uint64(77)
	b.OpenObject()
	b.AddKeyValue("test", velocypack.NewUIntValue(u))
	b.Close()

	s := mustSlice(b.Slice())
	ASSERT_TRUE(s.IsObject(), t)
	ASSERT_EQ(velocypack.ValueLength(1), mustLength(s.Length()), t)
	ASSERT_EQ(u, mustUInt(mustSlice(s.Get("test")).GetUInt()), t)
}

func TestBuilderObjectValue2(t *testing.T) {
	var b velocypack.Builder
	u := uint64(77)
	b.OpenObject()
	b.AddKeyValue("test", velocypack.NewUIntValue(u))
	b.AddKeyValue("soup", velocypack.NewUIntValue(u*2))
	b.Close()

	s := mustSlice(b.Slice())
	ASSERT_TRUE(s.IsObject(), t)
	ASSERT_EQ(velocypack.ValueLength(2), mustLength(s.Length()), t)
	ASSERT_EQ(u, mustUInt(mustSlice(s.Get("test")).GetUInt()), t)
	ASSERT_EQ(u*2, mustUInt(mustSlice(s.Get("soup")).GetUInt()), t)
}

func TestBuilderAddObjectIteratorEmpty(t *testing.T) {
	var obj velocypack.Builder
	obj.OpenObject()
	obj.AddKeyValue("1-one", velocypack.NewIntValue(1))
	obj.AddKeyValue("2-two", velocypack.NewIntValue(2))
	obj.AddKeyValue("3-three", velocypack.NewIntValue(3))
	obj.Close()
	objSlice := mustSlice(obj.Slice())

	var b velocypack.Builder
	ASSERT_TRUE(b.IsClosed(), t)
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderNeedOpenObject, t)(b.AddKeyValuesFromIterator(mustObjectIterator(velocypack.NewObjectIterator(objSlice))))
	ASSERT_TRUE(b.IsClosed(), t)
}

func TestBuilderAddObjectIteratorKeyAlreadyWritten(t *testing.T) {
	var obj velocypack.Builder
	obj.OpenObject()
	obj.AddKeyValue("1-one", velocypack.NewIntValue(1))
	obj.AddKeyValue("2-two", velocypack.NewIntValue(2))
	obj.AddKeyValue("3-three", velocypack.NewIntValue(3))
	obj.Close()
	objSlice := mustSlice(obj.Slice())

	var b velocypack.Builder
	ASSERT_TRUE(b.IsClosed(), t)
	must(b.OpenObject())
	must(b.AddValue(velocypack.NewStringValue("foo")))
	ASSERT_FALSE(b.IsClosed(), t)
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderKeyAlreadyWritten, t)(b.AddKeyValuesFromIterator(mustObjectIterator(velocypack.NewObjectIterator(objSlice))))
	ASSERT_FALSE(b.IsClosed(), t)
}

func TestBuilderAddObjectIteratorNonObject(t *testing.T) {
	var obj velocypack.Builder
	obj.OpenObject()
	obj.AddKeyValue("1-one", velocypack.NewIntValue(1))
	obj.AddKeyValue("2-two", velocypack.NewIntValue(2))
	obj.AddKeyValue("3-three", velocypack.NewIntValue(3))
	obj.Close()
	objSlice := mustSlice(obj.Slice())

	var b velocypack.Builder
	must(b.OpenArray())
	ASSERT_FALSE(b.IsClosed(), t)
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderNeedOpenObject, t)(b.AddKeyValuesFromIterator(mustObjectIterator(velocypack.NewObjectIterator(objSlice))))
	ASSERT_FALSE(b.IsClosed(), t)
}

func TestBuilderAddObjectIteratorTop(t *testing.T) {
	var obj velocypack.Builder
	obj.OpenObject()
	obj.AddKeyValue("1-one", velocypack.NewIntValue(1))
	obj.AddKeyValue("2-two", velocypack.NewIntValue(2))
	obj.AddKeyValue("3-three", velocypack.NewIntValue(3))
	obj.Close()
	objSlice := mustSlice(obj.Slice())

	var b velocypack.Builder
	must(b.OpenObject())
	ASSERT_FALSE(b.IsClosed(), t)
	must(b.AddKeyValuesFromIterator(mustObjectIterator(velocypack.NewObjectIterator(objSlice))))
	ASSERT_FALSE(b.IsClosed(), t)
	must(b.Close())
	result := mustSlice(b.Slice())
	ASSERT_TRUE(b.IsClosed(), t)

	ASSERT_EQ("{\"1-one\":1,\"2-two\":2,\"3-three\":3}", mustString(result.JSONString()), t)
}

func TestBuilderAddObjectIteratorReference(t *testing.T) {
	var obj velocypack.Builder
	obj.OpenObject()
	obj.AddKeyValue("1-one", velocypack.NewIntValue(1))
	obj.AddKeyValue("2-two", velocypack.NewIntValue(2))
	obj.AddKeyValue("3-three", velocypack.NewIntValue(3))
	obj.Close()
	objSlice := mustSlice(obj.Slice())

	var b velocypack.Builder
	must(b.OpenObject())
	ASSERT_FALSE(b.IsClosed(), t)
	must(b.Add(mustObjectIterator(velocypack.NewObjectIterator(objSlice))))
	ASSERT_FALSE(b.IsClosed(), t)
	must(b.Close())
	result := mustSlice(b.Slice())
	ASSERT_TRUE(b.IsClosed(), t)

	ASSERT_EQ("{\"1-one\":1,\"2-two\":2,\"3-three\":3}", mustString(result.JSONString()), t)
}

func TestBuilderAddObjectIteratorSub(t *testing.T) {
	var obj velocypack.Builder
	obj.OpenObject()
	obj.AddKeyValue("1-one", velocypack.NewIntValue(1))
	obj.AddKeyValue("2-two", velocypack.NewIntValue(2))
	obj.AddKeyValue("3-three", velocypack.NewIntValue(3))
	obj.Close()
	objSlice := mustSlice(obj.Slice())

	var b velocypack.Builder
	must(b.OpenObject())
	must(b.AddKeyValue("1-something", velocypack.NewStringValue("tennis")))
	must(b.AddValue(velocypack.NewStringValue("2-values")))
	must(b.OpenObject())
	must(b.Add(mustObjectIterator(velocypack.NewObjectIterator(objSlice))))
	ASSERT_FALSE(b.IsClosed(), t)
	must(b.Close()) // close one level
	must(b.AddKeyValue("3-bark", velocypack.NewStringValue("qux")))
	ASSERT_FALSE(b.IsClosed(), t)
	must(b.Close())
	result := mustSlice(b.Slice())
	ASSERT_TRUE(b.IsClosed(), t)

	ASSERT_EQ("{\"1-something\":\"tennis\",\"2-values\":{\"1-one\":1,\"2-two\":2,\"3-three\":3},\"3-bark\":\"qux\"}", mustString(result.JSONString()), t)
}

func TestBuilderAddAndOpenObject(t *testing.T) {
	var b1 velocypack.Builder
	ASSERT_TRUE(b1.IsClosed(), t)
	must(b1.OpenObject())
	ASSERT_FALSE(b1.IsClosed(), t)
	must(b1.AddKeyValue("foo", velocypack.NewStringValue("bar")))
	must(b1.Close())
	ASSERT_TRUE(b1.IsClosed(), t)
	ASSERT_EQ(byte(0x14), mustSlice(b1.Slice())[0], t)
	ASSERT_EQ(velocypack.ValueLength(1), mustLength(mustSlice(b1.Slice()).Length()), t)

	var b2 velocypack.Builder
	ASSERT_TRUE(b2.IsClosed(), t)
	must(b2.OpenObject())
	ASSERT_FALSE(b2.IsClosed(), t)
	must(b2.AddKeyValue("foo", velocypack.NewStringValue("bar")))
	must(b2.Close())
	ASSERT_TRUE(b2.IsClosed(), t)
	ASSERT_EQ(byte(0x14), mustSlice(b2.Slice())[0], t)
	ASSERT_EQ(velocypack.ValueLength(1), mustLength(mustSlice(b2.Slice()).Length()), t)
}

func TestBuilderAddOnNonObject(t *testing.T) {
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewArrayValue()))
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderNeedOpenObject, t)(b.AddKeyValue("foo", velocypack.NewBoolValue(true)))
}

func TestBuilderIsOpenObject(t *testing.T) {
	var b velocypack.Builder
	ASSERT_FALSE(b.IsOpenObject(), t)
	must(b.OpenObject())
	ASSERT_TRUE(b.IsOpenObject(), t)
	must(b.Close())
	ASSERT_FALSE(b.IsOpenObject(), t)
}

func TestBuilderHasKeyNonObject(t *testing.T) {
	var b velocypack.Builder
	b.AddValue(velocypack.NewIntValue(1))
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderNeedOpenObject, t)(b.HasKey("foo"))
}

func TestBuilderHasKeyArray(t *testing.T) {
	var b velocypack.Builder
	b.AddValue(velocypack.NewArrayValue())
	b.AddValue(velocypack.NewIntValue(1))
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderNeedOpenObject, t)(b.HasKey("foo"))
}

func TestBuilderHasKeyEmptyObject(t *testing.T) {
	var b velocypack.Builder
	b.AddValue(velocypack.NewObjectValue())
	ASSERT_FALSE(mustBool(b.HasKey("foo")), t)
	ASSERT_FALSE(mustBool(b.HasKey("bar")), t)
	ASSERT_FALSE(mustBool(b.HasKey("baz")), t)
	ASSERT_FALSE(mustBool(b.HasKey("quetzalcoatl")), t)
	b.Close()
}

func TestBuilderHasKeySubObject(t *testing.T) {
	var b velocypack.Builder
	b.AddValue(velocypack.NewObjectValue())
	must(b.AddKeyValue("foo", velocypack.NewIntValue(1)))
	must(b.AddKeyValue("bar", velocypack.NewBoolValue(true)))
	ASSERT_TRUE(mustBool(b.HasKey("foo")), t)
	ASSERT_TRUE(mustBool(b.HasKey("bar")), t)
	ASSERT_FALSE(mustBool(b.HasKey("baz")), t)

	must(b.AddKeyValue("bark", velocypack.NewObjectValue()))
	ASSERT_FALSE(mustBool(b.HasKey("bark")), t)
	ASSERT_FALSE(mustBool(b.HasKey("foo")), t)
	ASSERT_FALSE(mustBool(b.HasKey("bar")), t)
	ASSERT_FALSE(mustBool(b.HasKey("baz")), t)
	must(b.Close())

	ASSERT_TRUE(mustBool(b.HasKey("foo")), t)
	ASSERT_TRUE(mustBool(b.HasKey("bar")), t)
	ASSERT_TRUE(mustBool(b.HasKey("bark")), t)
	ASSERT_FALSE(mustBool(b.HasKey("baz")), t)

	must(b.AddKeyValue("baz", velocypack.NewIntValue(42)))
	ASSERT_TRUE(mustBool(b.HasKey("foo")), t)
	ASSERT_TRUE(mustBool(b.HasKey("bar")), t)
	ASSERT_TRUE(mustBool(b.HasKey("bark")), t)
	ASSERT_TRUE(mustBool(b.HasKey("baz")), t)
	b.Close()
}

func TestBuilderHasKeyCompact(t *testing.T) {
	var b velocypack.Builder
	b.AddValue(velocypack.NewObjectValue(true))
	must(b.AddKeyValue("foo", velocypack.NewIntValue(1)))
	must(b.AddKeyValue("bar", velocypack.NewBoolValue(true)))
	ASSERT_TRUE(mustBool(b.HasKey("foo")), t)
	ASSERT_TRUE(mustBool(b.HasKey("bar")), t)
	ASSERT_FALSE(mustBool(b.HasKey("baz")), t)

	must(b.AddKeyValue("bark", velocypack.NewObjectValue()))
	ASSERT_FALSE(mustBool(b.HasKey("bark")), t)
	ASSERT_FALSE(mustBool(b.HasKey("foo")), t)
	ASSERT_FALSE(mustBool(b.HasKey("bar")), t)
	ASSERT_FALSE(mustBool(b.HasKey("baz")), t)
	must(b.Close())

	ASSERT_TRUE(mustBool(b.HasKey("foo")), t)
	ASSERT_TRUE(mustBool(b.HasKey("bar")), t)
	ASSERT_TRUE(mustBool(b.HasKey("bark")), t)
	ASSERT_FALSE(mustBool(b.HasKey("baz")), t)

	must(b.AddKeyValue("baz", velocypack.NewIntValue(42)))
	ASSERT_TRUE(mustBool(b.HasKey("foo")), t)
	ASSERT_TRUE(mustBool(b.HasKey("bar")), t)
	ASSERT_TRUE(mustBool(b.HasKey("bark")), t)
	ASSERT_TRUE(mustBool(b.HasKey("baz")), t)
	b.Close()
}

func TestBuilderGetKeyNonObject(t *testing.T) {
	var b velocypack.Builder
	b.AddValue(velocypack.NewIntValue(1))
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderNeedOpenObject, t)(b.GetKey("foo"))
}

func TestBuilderGetKeyArray(t *testing.T) {
	var b velocypack.Builder
	b.AddValue(velocypack.NewArrayValue())
	b.AddValue(velocypack.NewIntValue(1))
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderNeedOpenObject, t)(b.GetKey("foo"))
}

func TestBuilderGetKeyEmptyObject(t *testing.T) {
	var b velocypack.Builder
	b.AddValue(velocypack.NewObjectValue())
	ASSERT_TRUE(mustSlice(b.GetKey("foo")).IsNone(), t)
	ASSERT_TRUE(mustSlice(b.GetKey("bar")).IsNone(), t)
	ASSERT_TRUE(mustSlice(b.GetKey("baz")).IsNone(), t)
	ASSERT_TRUE(mustSlice(b.GetKey("quetzalcoatl")).IsNone(), t)
	b.Close()
}

func TestBuilderGetKeySubObject(t *testing.T) {
	var b velocypack.Builder
	b.AddValue(velocypack.NewObjectValue())
	must(b.AddKeyValue("foo", velocypack.NewIntValue(1)))
	must(b.AddKeyValue("bar", velocypack.NewBoolValue(true)))
	ASSERT_EQ(uint64(1), mustUInt(mustSlice(b.GetKey("foo")).GetUInt()), t)
	ASSERT_TRUE(mustSlice(b.GetKey("bar")).IsBool(), t)
	ASSERT_TRUE(mustSlice(b.GetKey("baz")).IsNone(), t)

	must(b.AddKeyValue("bark", velocypack.NewObjectValue()))
	ASSERT_TRUE(mustSlice(b.GetKey("bark")).IsNone(), t)
	ASSERT_TRUE(mustSlice(b.GetKey("foo")).IsNone(), t)
	ASSERT_TRUE(mustSlice(b.GetKey("bar")).IsNone(), t)
	ASSERT_TRUE(mustSlice(b.GetKey("baz")).IsNone(), t)
	must(b.Close())

	ASSERT_EQ(uint64(1), mustUInt(mustSlice(b.GetKey("foo")).GetUInt()), t)
	ASSERT_TRUE(mustSlice(b.GetKey("bar")).IsBool(), t)
	ASSERT_TRUE(mustSlice(b.GetKey("baz")).IsNone(), t)
	ASSERT_TRUE(mustSlice(b.GetKey("bark")).IsObject(), t)

	must(b.AddKeyValue("baz", velocypack.NewIntValue(42)))
	ASSERT_EQ(uint64(1), mustUInt(mustSlice(b.GetKey("foo")).GetUInt()), t)
	ASSERT_TRUE(mustSlice(b.GetKey("bar")).IsBool(), t)
	ASSERT_EQ(uint64(42), mustUInt(mustSlice(b.GetKey("baz")).GetUInt()), t)
	ASSERT_TRUE(mustSlice(b.GetKey("bark")).IsObject(), t)
	b.Close()
}

func TestBuilderGetKeyCompact(t *testing.T) {
	var b velocypack.Builder
	b.AddValue(velocypack.NewObjectValue(true))
	must(b.AddKeyValue("foo", velocypack.NewIntValue(1)))
	must(b.AddKeyValue("bar", velocypack.NewBoolValue(true)))
	ASSERT_EQ(uint64(1), mustUInt(mustSlice(b.GetKey("foo")).GetUInt()), t)
	ASSERT_TRUE(mustSlice(b.GetKey("bar")).IsBool(), t)
	ASSERT_TRUE(mustSlice(b.GetKey("baz")).IsNone(), t)

	must(b.AddKeyValue("bark", velocypack.NewObjectValue()))
	ASSERT_TRUE(mustSlice(b.GetKey("bark")).IsNone(), t)
	ASSERT_TRUE(mustSlice(b.GetKey("foo")).IsNone(), t)
	ASSERT_TRUE(mustSlice(b.GetKey("bar")).IsNone(), t)
	ASSERT_TRUE(mustSlice(b.GetKey("baz")).IsNone(), t)
	must(b.Close())

	ASSERT_EQ(uint64(1), mustUInt(mustSlice(b.GetKey("foo")).GetUInt()), t)
	ASSERT_TRUE(mustSlice(b.GetKey("bar")).IsBool(), t)
	ASSERT_TRUE(mustSlice(b.GetKey("baz")).IsNone(), t)
	ASSERT_TRUE(mustSlice(b.GetKey("bark")).IsObject(), t)

	must(b.AddKeyValue("baz", velocypack.NewIntValue(42)))
	ASSERT_EQ(uint64(1), mustUInt(mustSlice(b.GetKey("foo")).GetUInt()), t)
	ASSERT_TRUE(mustSlice(b.GetKey("bar")).IsBool(), t)
	ASSERT_EQ(uint64(42), mustUInt(mustSlice(b.GetKey("baz")).GetUInt()), t)
	ASSERT_TRUE(mustSlice(b.GetKey("bark")).IsObject(), t)
	b.Close()
}

func TestBuilderAddKeysSeparately1(t *testing.T) {
	var b velocypack.Builder
	must(b.OpenObject())
	must(b.AddValue(velocypack.NewStringValue("name")))
	must(b.AddValue(velocypack.NewStringValue("Neunhoeffer")))
	must(b.AddValue(velocypack.NewStringValue("firstName")))
	must(b.AddValue(velocypack.NewStringValue("Max")))
	must(b.Close())

	ASSERT_EQ(`{"firstName":"Max","name":"Neunhoeffer"}`, mustString(mustSlice(b.Slice()).JSONString()), t)
}
