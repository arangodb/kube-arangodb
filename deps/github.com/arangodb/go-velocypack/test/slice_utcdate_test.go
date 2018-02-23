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
	"testing"
	"time"

	velocypack "github.com/arangodb/go-velocypack"
)

func TestSliceUTCDate1(t *testing.T) {
	slice := velocypack.Slice{0x1c, 0, 0, 0, 0, 0, 0, 0, 0}
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.UTCDate, slice.Type(), t)
	ASSERT_TRUE(slice.IsUTCDate(), t)
	ASSERT_EQ(velocypack.ValueLength(9), mustLength(slice.ByteSize()), t)
	ASSERT_EQ(time.Unix(0, 0).UTC(), mustTime(slice.GetUTCDate()), t)
}

func TestSliceUTCDate2(t *testing.T) {
	msec := 1234567
	slice := velocypack.Slice{0x1c, 0, 0, 0, 0, 0, 0, 0, 0}
	binary.LittleEndian.PutUint64(slice[1:], uint64(msec))
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.UTCDate, slice.Type(), t)
	ASSERT_TRUE(slice.IsUTCDate(), t)
	ASSERT_EQ(velocypack.ValueLength(9), mustLength(slice.ByteSize()), t)
	ASSERT_EQ(time.Unix(0, 0).UTC().Add(time.Millisecond*time.Duration(msec)), mustTime(slice.GetUTCDate()), t)
}
