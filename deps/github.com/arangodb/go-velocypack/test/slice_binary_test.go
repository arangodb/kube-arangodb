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

func TestSliceBinaryEmpty(t *testing.T) {
	slice := velocypack.Slice{0xc0, 0x00}
	assertEqualFromReader(t, slice)

	ASSERT_TRUE(slice.IsBinary(), t)
	ASSERT_EQ([]byte{}, mustBytes(slice.GetBinary()), t)
	ASSERT_EQ(velocypack.ValueLength(0), mustLength(slice.GetBinaryLength()), t)
	ASSERT_EQ(velocypack.ValueLength(len(slice)), mustLength(slice.ByteSize()), t)
}

func TestSliceBinarySomeValue(t *testing.T) {
	slice := velocypack.Slice{0xc0, 0x05, 0xfe, 0xfd, 0xfc, 0xfb, 0xfa}
	assertEqualFromReader(t, slice)

	ASSERT_TRUE(slice.IsBinary(), t)
	ASSERT_EQ([]byte{0xfe, 0xfd, 0xfc, 0xfb, 0xfa}, mustBytes(slice.GetBinary()), t)
	ASSERT_EQ(velocypack.ValueLength(5), mustLength(slice.GetBinaryLength()), t)
	ASSERT_EQ(velocypack.ValueLength(len(slice)), mustLength(slice.ByteSize()), t)
}

func TestSliceBinaryWithNullBytes(t *testing.T) {
	slice := velocypack.Slice{0xc0, 0x05, 0x01, 0x02, 0x00, 0x03, 0x00}
	assertEqualFromReader(t, slice)

	ASSERT_TRUE(slice.IsBinary(), t)
	ASSERT_EQ([]byte{0x01, 0x02, 0x00, 0x03, 0x00}, mustBytes(slice.GetBinary()), t)
	ASSERT_EQ(velocypack.ValueLength(5), mustLength(slice.GetBinaryLength()), t)
	ASSERT_EQ(velocypack.ValueLength(len(slice)), mustLength(slice.ByteSize()), t)
}

func TestSliceBinaryNonBinary(t *testing.T) {
	var slice velocypack.Slice
	assertEqualFromReader(t, slice)

	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsInvalidType, t)(slice.GetBinary())
	ASSERT_VELOCYPACK_EXCEPTION(velocypack.IsInvalidType, t)(slice.GetBinaryLength())
}
