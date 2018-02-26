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

func TestSliceNoneFactory(t *testing.T) {
	slice := velocypack.NoneSlice()
	assertEqualFromReader(t, slice)
	ASSERT_TRUE(slice.IsNone(), t)
}

func TestSliceNullFactory(t *testing.T) {
	slice := velocypack.NullSlice()
	assertEqualFromReader(t, slice)
	ASSERT_TRUE(slice.IsNull(), t)
}

func TestSliceZeroFactory(t *testing.T) {
	slice := velocypack.ZeroSlice()
	assertEqualFromReader(t, slice)
	ASSERT_TRUE(slice.IsSmallInt(), t)
	ASSERT_EQ(int64(0), mustInt(slice.GetSmallInt()), t)
}

func TestSliceIllegalFactory(t *testing.T) {
	slice := velocypack.IllegalSlice()
	assertEqualFromReader(t, slice)
	ASSERT_TRUE(slice.IsIllegal(), t)
}

func TestSliceFalseFactory(t *testing.T) {
	slice := velocypack.FalseSlice()
	assertEqualFromReader(t, slice)
	ASSERT_TRUE(slice.IsBool() && !mustBool(slice.GetBool()), t)
}

func TestSliceTrueFactory(t *testing.T) {
	slice := velocypack.TrueSlice()
	assertEqualFromReader(t, slice)
	ASSERT_TRUE(slice.IsBool() && mustBool(slice.GetBool()), t)
}

func TestSliceEmptyArrayFactory(t *testing.T) {
	slice := velocypack.EmptyArraySlice()
	assertEqualFromReader(t, slice)
	ASSERT_TRUE(slice.IsArray() && mustLength(slice.Length()) == 0, t)
}

func TestSliceEmptyObjectFactory(t *testing.T) {
	slice := velocypack.EmptyObjectSlice()
	assertEqualFromReader(t, slice)
	ASSERT_TRUE(slice.IsObject() && mustLength(slice.Length()) == 0, t)
}

func TestSliceMinKeyFactory(t *testing.T) {
	slice := velocypack.MinKeySlice()
	assertEqualFromReader(t, slice)
	ASSERT_TRUE(slice.IsMinKey(), t)
}

func TestSliceMaxKeyFactory(t *testing.T) {
	slice := velocypack.MaxKeySlice()
	assertEqualFromReader(t, slice)
	ASSERT_TRUE(slice.IsMaxKey(), t)
}

func TestSliceStringFactory(t *testing.T) {
	slice := velocypack.StringSlice("short")
	assertEqualFromReader(t, slice)
	ASSERT_TRUE(slice.IsString(), t)

	slice = velocypack.StringSlice(`long long long long long long long long long long long long long long long long long long 
	long long long long long long long long long long long long long long long long long long long long long long long long 
	long long long long long long long long long long long long long long long long long long long long long long long long 
	long long long long long long long long long long long long long long long long long long long long long long long long 
	long long long long long long long long long long long long long long long long long long long long long long long long 
	long long long long long long long long long long long long long long long long long long long long long long long long `)
	assertEqualFromReader(t, slice)
	ASSERT_TRUE(slice.IsString(), t)
}
