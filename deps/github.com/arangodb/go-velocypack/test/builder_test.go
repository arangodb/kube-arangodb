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

func TestBuilderBytesWithOpenObject(t *testing.T) {
	var b velocypack.Builder
	ASSERT_EQ(0, len(mustBytes(b.Bytes())), t)
	must(b.OpenObject())
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderNotClosed, t)(b.Bytes())
	must(b.Close())
	ASSERT_EQ(1, len(mustBytes(b.Bytes())), t)
}

func TestBuilderSliceWithOpenObject(t *testing.T) {
	var b velocypack.Builder
	ASSERT_EQ(0, len(mustSlice(b.Slice())), t)
	must(b.OpenObject())
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderNotClosed, t)(b.Slice())
	must(b.Close())
	ASSERT_EQ(1, len(mustSlice(b.Slice())), t)
}

func TestBuilderSizeWithOpenObject(t *testing.T) {
	var b velocypack.Builder
	ASSERT_EQ(velocypack.ValueLength(0), mustLength(b.Size()), t)
	must(b.OpenObject())
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderNotClosed, t)(b.Size())
	must(b.Close())
	ASSERT_EQ(velocypack.ValueLength(1), mustLength(b.Size()), t)
}

func TestBuilderIsEmpty(t *testing.T) {
	var b velocypack.Builder
	ASSERT_TRUE(b.IsEmpty(), t)
	must(b.OpenObject())
	ASSERT_FALSE(b.IsEmpty(), t)
}

func TestBuilderIsClosedMixed(t *testing.T) {
	var b velocypack.Builder
	ASSERT_TRUE(b.IsClosed(), t)
	b.AddValue(velocypack.NewNullValue())
	ASSERT_TRUE(b.IsClosed(), t)
	b.AddValue(velocypack.NewBoolValue(true))
	ASSERT_TRUE(b.IsClosed(), t)

	b.AddValue(velocypack.NewArrayValue())
	ASSERT_FALSE(b.IsClosed(), t)

	b.AddValue(velocypack.NewBoolValue(true))
	ASSERT_FALSE(b.IsClosed(), t)
	b.AddValue(velocypack.NewBoolValue(true))
	ASSERT_FALSE(b.IsClosed(), t)

	must(b.Close())
	ASSERT_TRUE(b.IsClosed(), t)

	b.AddValue(velocypack.NewObjectValue())
	ASSERT_FALSE(b.IsClosed(), t)

	b.AddKeyValue("foo", velocypack.NewBoolValue(true))
	ASSERT_FALSE(b.IsClosed(), t)

	b.AddKeyValue("bar", velocypack.NewBoolValue(true))
	ASSERT_FALSE(b.IsClosed(), t)

	b.AddKeyValue("baz", velocypack.NewArrayValue())
	ASSERT_FALSE(b.IsClosed(), t)

	must(b.Close())
	ASSERT_FALSE(b.IsClosed(), t)

	must(b.Close())
	ASSERT_TRUE(b.IsClosed(), t)
}

func TestBuilderIsClosedObject(t *testing.T) {
	var b velocypack.Builder
	ASSERT_TRUE(b.IsClosed(), t)
	must(b.AddValue(velocypack.NewObjectValue()))
	ASSERT_FALSE(b.IsClosed(), t)

	must(b.AddKeyValue("foo", velocypack.NewBoolValue(true)))
	ASSERT_FALSE(b.IsClosed(), t)

	must(b.AddKeyValue("bar", velocypack.NewBoolValue(true)))
	ASSERT_FALSE(b.IsClosed(), t)

	must(b.AddKeyValue("baz", velocypack.NewObjectValue()))
	ASSERT_FALSE(b.IsClosed(), t)

	must(b.Close())
	ASSERT_FALSE(b.IsClosed(), t)

	must(b.Close())
	ASSERT_TRUE(b.IsClosed(), t)
}

func TestBuilderCloseClosed(t *testing.T) {
	var b velocypack.Builder
	ASSERT_TRUE(b.IsClosed(), t)
	must(b.AddValue(velocypack.NewObjectValue()))
	ASSERT_FALSE(b.IsClosed(), t)
	must(b.Close())
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderNeedOpenCompound, t)(b.Close())
}

func TestBuilderRemoveLastNonObject(t *testing.T) {
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewBoolValue(true)))
	must(b.AddValue(velocypack.NewBoolValue(false)))
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderNeedOpenCompound, t)(b.RemoveLast())
}

func TestBuilderRemoveLastSealed(t *testing.T) {
	var b velocypack.Builder
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderNeedOpenCompound, t)(b.RemoveLast())
}

func TestBuilderRemoveLastEmptyObject(t *testing.T) {
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewObjectValue()))
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderNeedSubValue, t)(b.RemoveLast())
}

func TestBuilderRemoveLastObjectInvalid(t *testing.T) {
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewObjectValue()))
	must(b.AddKeyValue("foo", velocypack.NewBoolValue(true)))
	must(b.RemoveLast())
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsBuilderNeedSubValue, t)(b.RemoveLast())
}

func TestBuilderRemoveLastObject(t *testing.T) {
	var b velocypack.Builder
	must(b.AddValue(velocypack.NewObjectValue()))
	must(b.AddKeyValue("foo", velocypack.NewBoolValue(true)))
	must(b.AddKeyValue("bar", velocypack.NewBoolValue(false)))

	must(b.RemoveLast())
	must(b.Close())

	s := mustSlice(b.Slice())
	ASSERT_TRUE(s.IsObject(), t)
	ASSERT_EQ(velocypack.ValueLength(1), mustLength(s.Length()), t)
	ASSERT_TRUE(mustBool(s.HasKey("foo")), t)
	ASSERT_TRUE(mustBool(mustSlice(s.Get("foo")).GetBool()), t)
	ASSERT_FALSE(mustBool(s.HasKey("bar")), t)
}
