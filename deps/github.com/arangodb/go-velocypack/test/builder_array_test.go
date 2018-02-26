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
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	velocypack "github.com/arangodb/go-velocypack"
)

func TestBuilderEmptyArray(t *testing.T) {
	var b velocypack.Builder
	b.OpenArray()
	b.Close()

	s := mustSlice(b.Slice())
	ASSERT_TRUE(s.IsArray(), t)
	ASSERT_EQ(velocypack.ValueLength(0), mustLength(s.Length()), t)
}

func TestBuilderArrayEmpty(t *testing.T) {
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewArrayValue()))
	must(b.Close())
	l := mustLength(b.Size())
	result := mustBytes(b.Bytes())

	correctResult := []byte{0x01}

	ASSERT_EQ(velocypack.ValueLength(len(correctResult)), l, t)
	ASSERT_EQ(result, correctResult, t)
}

func TestBuilderArraySingleEntry(t *testing.T) {
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewArrayValue()))
	must(b.AddValue(velocypack.NewIntValue(1)))
	must(b.Close())
	l := mustLength(b.Size())
	result := mustBytes(b.Bytes())

	correctResult := []byte{0x02, 0x03, 0x31}

	ASSERT_EQ(velocypack.ValueLength(len(correctResult)), l, t)
	ASSERT_EQ(result, correctResult, t)
}

func TestBuilderArraySingleEntryLong(t *testing.T) {
	value := "ngdddddljjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjsdddffffffffffffmmmmmmmmmmmmmmmsf" +
		"dlllllllllllllllllllllllllllllllllllllllllllllllllrjjjjjjsdddddddddddddd" +
		"ddddhhhhhhkkkkkkkksssssssssssssssssssssssssssssssssdddddddddddddddddkkkk" +
		"kkkkkkkkksddddddddddddssssssssssfvvvvvvvvvvvvvvvvvvvvvvvvvvvfvgfff"
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewArrayValue()))
	must(b.AddValue(velocypack.NewStringValue(value)))
	must(b.Close())
	l := mustLength(b.Size())
	result := mustBytes(b.Bytes())

	correctResult := []byte{
		0x03, 0x2c, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xbf, 0x1a, 0x01,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x6e, 0x67, 0x64, 0x64, 0x64, 0x64,
		0x64, 0x6c, 0x6a, 0x6a, 0x6a, 0x6a, 0x6a, 0x6a, 0x6a, 0x6a, 0x6a, 0x6a,
		0x6a, 0x6a, 0x6a, 0x6a, 0x6a, 0x6a, 0x6a, 0x6a, 0x6a, 0x6a, 0x6a, 0x6a,
		0x6a, 0x6a, 0x6a, 0x6a, 0x6a, 0x6a, 0x6a, 0x6a, 0x6a, 0x73, 0x64, 0x64,
		0x64, 0x66, 0x66, 0x66, 0x66, 0x66, 0x66, 0x66, 0x66, 0x66, 0x66, 0x66,
		0x66, 0x6d, 0x6d, 0x6d, 0x6d, 0x6d, 0x6d, 0x6d, 0x6d, 0x6d, 0x6d, 0x6d,
		0x6d, 0x6d, 0x6d, 0x6d, 0x73, 0x66, 0x64, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c,
		0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c,
		0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c,
		0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c,
		0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x6c, 0x72, 0x6a, 0x6a, 0x6a,
		0x6a, 0x6a, 0x6a, 0x73, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64,
		0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x68, 0x68,
		0x68, 0x68, 0x68, 0x68, 0x6b, 0x6b, 0x6b, 0x6b, 0x6b, 0x6b, 0x6b, 0x6b,
		0x73, 0x73, 0x73, 0x73, 0x73, 0x73, 0x73, 0x73, 0x73, 0x73, 0x73, 0x73,
		0x73, 0x73, 0x73, 0x73, 0x73, 0x73, 0x73, 0x73, 0x73, 0x73, 0x73, 0x73,
		0x73, 0x73, 0x73, 0x73, 0x73, 0x73, 0x73, 0x73, 0x73, 0x64, 0x64, 0x64,
		0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64,
		0x64, 0x64, 0x6b, 0x6b, 0x6b, 0x6b, 0x6b, 0x6b, 0x6b, 0x6b, 0x6b, 0x6b,
		0x6b, 0x6b, 0x6b, 0x73, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64,
		0x64, 0x64, 0x64, 0x64, 0x73, 0x73, 0x73, 0x73, 0x73, 0x73, 0x73, 0x73,
		0x73, 0x73, 0x66, 0x76, 0x76, 0x76, 0x76, 0x76, 0x76, 0x76, 0x76, 0x76,
		0x76, 0x76, 0x76, 0x76, 0x76, 0x76, 0x76, 0x76, 0x76, 0x76, 0x76, 0x76,
		0x76, 0x76, 0x76, 0x76, 0x76, 0x76, 0x66, 0x76, 0x67, 0x66, 0x66, 0x66}

	ASSERT_EQ(velocypack.ValueLength(len(correctResult)), l, t)
	ASSERT_EQ(result, correctResult, t)
}

func TestBuilderArraySameSizeEntries(t *testing.T) {
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewArrayValue()))
	must(b.AddValue(velocypack.NewUIntValue(1)))
	must(b.AddValue(velocypack.NewUIntValue(2)))
	must(b.AddValue(velocypack.NewUIntValue(3)))
	must(b.Close())
	l := mustLength(b.Size())
	result := mustBytes(b.Bytes())

	correctResult := []byte{0x02, 0x05, 0x31, 0x32, 0x33}

	ASSERT_EQ(velocypack.ValueLength(len(correctResult)), l, t)
	ASSERT_EQ(result, correctResult, t)
}

func TestBuilderArraySomeEntries(t *testing.T) {
	var b velocypack.Builder
	value := 2.3
	must(b.AddValue(velocypack.NewArrayValue()))
	must(b.AddValue(velocypack.NewUIntValue(1200)))
	must(b.AddValue(velocypack.NewDoubleValue(value)))
	must(b.AddValue(velocypack.NewStringValue("abc")))
	must(b.AddValue(velocypack.NewBoolValue(true)))
	must(b.Close())
	l := mustLength(b.Size())
	result := mustBytes(b.Bytes())

	correctResult := []byte{
		0x06, 0x18, 0x04, 0x29, 0xb0, 0x04, // uint(1200) = 0x4b0
		0x1b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // double(2.3)
		0x43, 0x61, 0x62, 0x63, 0x1a, 0x03, 0x06, 0x0f, 0x13}
	binary.LittleEndian.PutUint64(correctResult[7:], math.Float64bits(value))

	ASSERT_EQ(velocypack.ValueLength(len(correctResult)), l, t)
	ASSERT_EQ(result, correctResult, t)
}

func TestBuilderArrayCompact(t *testing.T) {
	var b velocypack.Builder
	value := 2.3
	must(b.AddValue(velocypack.NewArrayValue(true)))
	must(b.AddValue(velocypack.NewUIntValue(1200)))
	must(b.AddValue(velocypack.NewDoubleValue(value)))
	must(b.AddValue(velocypack.NewStringValue("abc")))
	must(b.AddValue(velocypack.NewBoolValue(true)))
	must(b.Close())
	l := mustLength(b.Size())
	result := mustBytes(b.Bytes())

	correctResult := []byte{
		0x13, 0x14, 0x29, 0xb0, 0x04, 0x1b,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, // double
		0x43, 0x61, 0x62, 0x63, 0x1a, 0x04}
	binary.LittleEndian.PutUint64(correctResult[6:], math.Float64bits(value))

	ASSERT_EQ(velocypack.ValueLength(len(correctResult)), l, t)
	ASSERT_EQ(result, correctResult, t)
}

func TestBuilderArrayCompactBytesizeBelowThreshold(t *testing.T) {
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewArrayValue(true)))
	for i := uint64(0); i < 124; i++ {
		must(b.AddValue(velocypack.NewUIntValue(i % 10)))
	}
	must(b.Close())
	l := mustLength(b.Size())
	result := mustBytes(b.Bytes())

	ASSERT_EQ(velocypack.ValueLength(127), l, t)
	ASSERT_EQ(byte(0x13), result[0], t)
	ASSERT_EQ(byte(0x7f), result[1], t)
	for i := uint64(0); i < 124; i++ {
		ASSERT_EQ(byte(0x30+(i%10)), result[2+i], t)
	}
	ASSERT_EQ(byte(0x7c), result[126], t)
}

func TestBuilderArrayCompactBytesizeAboveThreshold(t *testing.T) {
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewArrayValue(true)))
	for i := uint64(0); i < 125; i++ {
		must(b.AddValue(velocypack.NewUIntValue(i % 10)))
	}
	must(b.Close())
	l := mustLength(b.Size())
	result := mustBytes(b.Bytes())

	ASSERT_EQ(velocypack.ValueLength(129), l, t)
	ASSERT_EQ(byte(0x13), result[0], t)
	ASSERT_EQ(byte(0x81), result[1], t)
	ASSERT_EQ(byte(0x01), result[2], t)
	for i := uint64(0); i < 125; i++ {
		ASSERT_EQ(byte(0x30+(i%10)), result[3+i], t)
	}
	ASSERT_EQ(byte(0x7d), result[128], t)
}

func TestBuilderArrayCompactLengthBelowThreshold(t *testing.T) {
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewArrayValue(true)))
	for i := uint64(0); i < 127; i++ {
		must(b.AddValue(velocypack.NewStringValue("aaa")))
	}
	must(b.Close())
	l := mustLength(b.Size())
	result := mustBytes(b.Bytes())

	ASSERT_EQ(velocypack.ValueLength(512), l, t)
	ASSERT_EQ(byte(0x13), result[0], t)
	ASSERT_EQ(byte(0x80), result[1], t)
	ASSERT_EQ(byte(0x04), result[2], t)
	for i := uint64(0); i < 127; i++ {
		ASSERT_EQ(byte(0x43), result[3+i*4], t)
	}
	ASSERT_EQ(byte(0x7f), result[511], t)
}

func TestBuilderArrayCompactLengthAboveThreshold(t *testing.T) {
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewArrayValue(true)))
	for i := uint64(0); i < 128; i++ {
		must(b.AddValue(velocypack.NewStringValue("aaa")))
	}
	must(b.Close())
	l := mustLength(b.Size())
	result := mustBytes(b.Bytes())

	ASSERT_EQ(velocypack.ValueLength(517), l, t)
	ASSERT_EQ(byte(0x13), result[0], t)
	ASSERT_EQ(byte(0x85), result[1], t)
	ASSERT_EQ(byte(0x04), result[2], t)
	for i := uint64(0); i < 128; i++ {
		ASSERT_EQ(byte(0x43), result[3+i*4], t)
	}
	ASSERT_EQ(byte(0x01), result[515], t)
	ASSERT_EQ(byte(0x80), result[516], t)
}

func TestBuilderAddObjectInArray(t *testing.T) {
	var b velocypack.Builder
	b.OpenArray()
	b.OpenObject()
	b.Close()
	b.Close()

	s := mustSlice(b.Slice())
	ASSERT_TRUE(s.IsArray(), t)
	ASSERT_EQ(velocypack.ValueLength(1), mustLength(s.Length()), t)
	ss := mustSlice(s.At(0))
	ASSERT_TRUE(ss.IsObject(), t)
	ASSERT_EQ(velocypack.ValueLength(0), mustLength(ss.Length()), t)
}

func TestBuilderAddNonEmptyObjectsInArray(t *testing.T) {
	var b velocypack.Builder
	must(b.OpenArray())
	for i := 0; i < 5; i++ {
		must(b.OpenObject())
		must(b.AddKeyValue("Field1", velocypack.NewIntValue(int64(i+1))))
		must(b.Close())
	}
	must(b.Close())

	s := mustSlice(b.Slice())
	ASSERT_TRUE(s.IsArray(), t)
	ASSERT_EQ(velocypack.ValueLength(5), mustLength(s.Length()), t)
	ss := mustSlice(s.At(0))
	ASSERT_TRUE(ss.IsObject(), t)
	ASSERT_EQ(velocypack.ValueLength(1), mustLength(ss.Length()), t)
	ASSERT_EQ(int64(1), mustInt(mustSlice(ss.Get("Field1")).GetInt()), t)

	it := mustArrayIterator(velocypack.NewArrayIterator(s))
	i := 1
	for it.IsValid() {
		ss := mustSlice(it.Value())
		ASSERT_TRUE(ss.IsObject(), t)
		ASSERT_EQ(velocypack.ValueLength(1), mustLength(ss.Length()), t)
		ASSERT_EQ(int64(i), mustInt(mustSlice(ss.Get("Field1")).GetInt()), t)
		it.Next()
		i++
	}
}

func TestBuilderAddArrayIteratorEmpty(t *testing.T) {
	var obj velocypack.Builder
	must(obj.OpenArray())
	must(obj.AddValue(velocypack.NewIntValue(1)))
	must(obj.AddValue(velocypack.NewIntValue(2)))
	must(obj.AddValue(velocypack.NewIntValue(3)))
	must(obj.Close())
	objSlice := mustSlice(obj.Slice())

	var b velocypack.Builder
	ASSERT_TRUE(b.IsClosed(), t)
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderNeedOpenArray, t)(b.AddValuesFromIterator(mustArrayIterator(velocypack.NewArrayIterator(objSlice))))
	ASSERT_TRUE(b.IsClosed(), t)
}

func TestBuilderAddArrayIteratorNonArray(t *testing.T) {
	var obj velocypack.Builder
	must(obj.OpenArray())
	must(obj.AddValue(velocypack.NewIntValue(1)))
	must(obj.AddValue(velocypack.NewIntValue(2)))
	must(obj.AddValue(velocypack.NewIntValue(3)))
	must(obj.Close())
	objSlice := mustSlice(obj.Slice())

	var b velocypack.Builder
	must(b.OpenObject())
	ASSERT_FALSE(b.IsClosed(), t)
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderNeedOpenArray, t)(b.AddValuesFromIterator(mustArrayIterator(velocypack.NewArrayIterator(objSlice))))
	ASSERT_FALSE(b.IsClosed(), t)
}

func TestBuilderAddArrayIteratorTop(t *testing.T) {
	var obj velocypack.Builder
	must(obj.OpenArray())
	must(obj.AddValue(velocypack.NewIntValue(1)))
	must(obj.AddValue(velocypack.NewIntValue(2)))
	must(obj.AddValue(velocypack.NewIntValue(3)))
	must(obj.Close())
	objSlice := mustSlice(obj.Slice())

	var b velocypack.Builder
	must(b.OpenArray())
	ASSERT_FALSE(b.IsClosed(), t)
	must(b.AddValuesFromIterator(mustArrayIterator(velocypack.NewArrayIterator(objSlice))))
	ASSERT_FALSE(b.IsClosed(), t)
	must(b.Close())
	result := mustSlice(b.Slice())

	ASSERT_EQ("[1,2,3]", mustString(result.JSONString()), t)
}

func TestBuilderAddArrayIteratorReference(t *testing.T) {
	var obj velocypack.Builder
	must(obj.OpenArray())
	must(obj.AddValue(velocypack.NewIntValue(1)))
	must(obj.AddValue(velocypack.NewIntValue(2)))
	must(obj.AddValue(velocypack.NewIntValue(3)))
	must(obj.Close())
	objSlice := mustSlice(obj.Slice())

	var b velocypack.Builder
	must(b.OpenArray())
	ASSERT_FALSE(b.IsClosed(), t)
	must(b.Add(mustArrayIterator(velocypack.NewArrayIterator(objSlice))))
	ASSERT_FALSE(b.IsClosed(), t)
	must(b.Close())
	result := mustSlice(b.Slice())

	ASSERT_EQ("[1,2,3]", mustString(result.JSONString()), t)
}

func TestBuilderAddArrayIteratorSub(t *testing.T) {
	var obj velocypack.Builder
	must(obj.OpenArray())
	must(obj.AddValue(velocypack.NewIntValue(1)))
	must(obj.AddValue(velocypack.NewIntValue(2)))
	must(obj.AddValue(velocypack.NewIntValue(3)))
	must(obj.Close())
	objSlice := mustSlice(obj.Slice())

	var b velocypack.Builder
	must(b.OpenArray())
	must(b.AddValue(velocypack.NewStringValue("tennis")))
	must(b.OpenArray())
	must(b.Add(mustArrayIterator(velocypack.NewArrayIterator(objSlice))))
	ASSERT_FALSE(b.IsClosed(), t)
	must(b.Close()) // close one level
	must(b.AddValue(velocypack.NewStringValue("qux")))
	ASSERT_FALSE(b.IsClosed(), t)
	must(b.Close())
	result := mustSlice(b.Slice())
	ASSERT_TRUE(b.IsClosed(), t)

	ASSERT_EQ("[\"tennis\",[1,2,3],\"qux\"]", mustString(result.JSONString()), t)
}

func TestBuilderAddAndOpenArray(t *testing.T) {
	var b1 velocypack.Builder
	ASSERT_TRUE(b1.IsClosed(), t)
	must(b1.OpenArray())
	ASSERT_FALSE(b1.IsClosed(), t)
	must(b1.AddValue(velocypack.NewStringValue("bar")))
	must(b1.Close())
	ASSERT_TRUE(b1.IsClosed(), t)
	ASSERT_EQ(byte(0x02), mustSlice(b1.Slice())[0], t)

	var b2 velocypack.Builder
	ASSERT_TRUE(b2.IsClosed(), t)
	must(b2.OpenArray())
	ASSERT_FALSE(b2.IsClosed(), t)
	must(b2.AddValue(velocypack.NewStringValue("bar")))
	must(b2.Close())
	ASSERT_TRUE(b2.IsClosed(), t)
	ASSERT_EQ(byte(0x02), mustSlice(b2.Slice())[0], t)
}

func TestBuilderAddOnNonArray(t *testing.T) {
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewObjectValue()))
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderKeyMustBeString, t)(b.AddValue(velocypack.NewBoolValue(true)))
}

func TestBuilderIsOpenArray(t *testing.T) {
	var b velocypack.Builder
	ASSERT_FALSE(b.IsOpenArray(), t)
	must(b.OpenArray())
	ASSERT_TRUE(b.IsOpenArray(), t)
	must(b.Close())
	ASSERT_FALSE(b.IsOpenArray(), t)
}

func TestBuilderWriteTo(t *testing.T) {
	var b velocypack.Builder
	must(b.OpenArray())
	must(b.Close())
	var buf bytes.Buffer
	_, err := b.WriteTo(&buf)
	ASSERT_NIL(err, t)
}

func TestBuilderWriteToNotClosed(t *testing.T) {
	var b velocypack.Builder
	must(b.OpenArray())
	var buf bytes.Buffer
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderNotClosed, t)(b.WriteTo(&buf))
}

func TestBuilderClear(t *testing.T) {
	var b velocypack.Builder
	must(b.OpenArray())
	ASSERT_FALSE(b.IsClosed(), t)
	b.Clear()
	ASSERT_TRUE(b.IsClosed(), t)
	ASSERT_EQ(0, len(mustBytes(b.Bytes())), t)
}
