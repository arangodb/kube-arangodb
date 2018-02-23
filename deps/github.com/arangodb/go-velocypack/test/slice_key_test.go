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

func TestSliceMinKey(t *testing.T) {
	slice := velocypack.Slice{0x1e}
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.MinKey, slice.Type(), t)
	ASSERT_TRUE(slice.IsMinKey(), t)
	ASSERT_EQ(velocypack.ValueLength(1), mustLength(slice.ByteSize()), t)
}

func TestSliceMaxKey(t *testing.T) {
	slice := velocypack.Slice{0x1f}
	assertEqualFromReader(t, slice)

	ASSERT_EQ(velocypack.MaxKey, slice.Type(), t)
	ASSERT_TRUE(slice.IsMaxKey(), t)
	ASSERT_EQ(velocypack.ValueLength(1), mustLength(slice.ByteSize()), t)
}
