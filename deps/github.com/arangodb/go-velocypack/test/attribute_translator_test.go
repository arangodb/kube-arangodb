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

func TestAttributeTranslator1(t *testing.T) {
	tests := map[uint8]string{
		1: "_key",
		2: "_rev",
		3: "_id",
		4: "_from",
		5: "_to",
		6: "6",
	}
	for id, name := range tests {
		// Simple object with only 1 field
		slice := velocypack.Slice{0x0b,
			0x07,           // Bytesize
			0x01,           // NoItems
			0x28, id, 0x1a, // "_xyz": true
			0x03, // Index of "_xyz"
		}

		a := mustSlice(slice.Get(name))
		ASSERT_EQ(velocypack.Bool, a.Type(), t)
		ASSERT_TRUE(mustBool(a.GetBool()), t)
	}
}

func TestAttributeTranslatorObject(t *testing.T) {
	// Normal object with multiple fields
	slice := velocypack.Slice{0x0b,
		0x00,             // Bytesize
		0x05,             // NoItems
		0x28, 0x01, 0x1a, // "_key": true
		0x28, 0x02, 0x19, // "_rev": false
		0x28, 0x03, 0x01, // "_id": []
		0x28, 0x04, 0x18, // "_from": null
		0x28, 0x05, 0x0a, // "_to": {}
		12, 9, 3, 6, 15, // Index of "_from", "_id", "_key", "_rev", "_to"
	}
	slice[1] = byte(len(slice))

	ASSERT_EQ("_key", mustString(mustSlice(slice.KeyAt(2, true)).GetString()), t)
	key := mustSlice(slice.Get("_key"))
	ASSERT_EQ(velocypack.Bool, key.Type(), t)
	ASSERT_TRUE(mustBool(key.GetBool()), t)

	ASSERT_EQ("_rev", mustString(mustSlice(slice.KeyAt(3, true)).GetString()), t)
	rev := mustSlice(slice.Get("_rev"))
	ASSERT_EQ(velocypack.Bool, rev.Type(), t)
	ASSERT_FALSE(mustBool(rev.GetBool()), t)

	ASSERT_EQ("_id", mustString(mustSlice(slice.KeyAt(1, true)).GetString()), t)
	id := mustSlice(slice.Get("_id"))
	ASSERT_EQ(velocypack.Array, id.Type(), t)

	ASSERT_EQ("_from", mustString(mustSlice(slice.KeyAt(0, true)).GetString()), t)
	from := mustSlice(slice.Get("_from"))
	ASSERT_EQ(velocypack.Null, from.Type(), t)

	ASSERT_EQ("_to", mustString(mustSlice(slice.KeyAt(4, true)).GetString()), t)
	to := mustSlice(slice.Get("_to"))
	ASSERT_EQ(velocypack.Object, to.Type(), t)
}

func TestAttributeTranslatorObjectSmallInt(t *testing.T) {
	// Normal object with multiple fields
	slice := velocypack.Slice{0x0b,
		0x00,       // Bytesize
		0x05,       // NoItems
		0x31, 0x1a, // "_key": true
		0x32, 0x19, // "_rev": false
		0x33, 0x01, // "_id": []
		0x34, 0x18, // "_from": null
		0x35, 0x0a, // "_to": {}
		9, 7, 3, 5, 11, // Index of "_from", "_id", "_key", "_rev", "_to"
	}
	slice[1] = byte(len(slice))

	ASSERT_EQ("_key", mustString(mustSlice(slice.KeyAt(2, true)).GetString()), t)
	key := mustSlice(slice.Get("_key"))
	ASSERT_EQ(velocypack.Bool, key.Type(), t)
	ASSERT_TRUE(mustBool(key.GetBool()), t)

	ASSERT_EQ("_rev", mustString(mustSlice(slice.KeyAt(3, true)).GetString()), t)
	rev := mustSlice(slice.Get("_rev"))
	ASSERT_EQ(velocypack.Bool, rev.Type(), t)
	ASSERT_FALSE(mustBool(rev.GetBool()), t)

	ASSERT_EQ("_id", mustString(mustSlice(slice.KeyAt(1, true)).GetString()), t)
	id := mustSlice(slice.Get("_id"))
	ASSERT_EQ(velocypack.Array, id.Type(), t)

	ASSERT_EQ("_from", mustString(mustSlice(slice.KeyAt(0, true)).GetString()), t)
	from := mustSlice(slice.Get("_from"))
	ASSERT_EQ(velocypack.Null, from.Type(), t)

	ASSERT_EQ("_to", mustString(mustSlice(slice.KeyAt(4, true)).GetString()), t)
	to := mustSlice(slice.Get("_to"))
	ASSERT_EQ(velocypack.Object, to.Type(), t)
}

func TestAttributeTranslatorCompactObject(t *testing.T) {
	// Compact object with multiple fields
	slice := velocypack.Slice{0x14,
		0x00,             // Bytesize
		0x28, 0x01, 0x1a, // "_key": true
		0x28, 0x02, 0x19, // "_rev": false
		0x28, 0x03, 0x01, // "_id": []
		0x28, 0x04, 0x18, // "_from": null
		0x28, 0x05, 0x0a, // "_to": {}
		0x05, // NoItems
	}
	slice[1] = byte(len(slice))

	ASSERT_EQ("_key", mustString(mustSlice(slice.KeyAt(0, true)).GetString()), t)
	key := mustSlice(slice.Get("_key"))
	ASSERT_EQ(velocypack.Bool, key.Type(), t)
	ASSERT_TRUE(mustBool(key.GetBool()), t)

	ASSERT_EQ("_rev", mustString(mustSlice(slice.KeyAt(1, true)).GetString()), t)
	rev := mustSlice(slice.Get("_rev"))
	ASSERT_EQ(velocypack.Bool, rev.Type(), t)
	ASSERT_FALSE(mustBool(rev.GetBool()), t)

	ASSERT_EQ("_id", mustString(mustSlice(slice.KeyAt(2, true)).GetString()), t)
	id := mustSlice(slice.Get("_id"))
	ASSERT_EQ(velocypack.Array, id.Type(), t)

	ASSERT_EQ("_from", mustString(mustSlice(slice.KeyAt(3, true)).GetString()), t)
	from := mustSlice(slice.Get("_from"))
	ASSERT_EQ(velocypack.Null, from.Type(), t)

	ASSERT_EQ("_to", mustString(mustSlice(slice.KeyAt(4, true)).GetString()), t)
	to := mustSlice(slice.Get("_to"))
	ASSERT_EQ(velocypack.Object, to.Type(), t)
}

func TestAttributeTranslatorCompactObjectSmallInt(t *testing.T) {
	// Compact object with multiple fields
	slice := velocypack.Slice{0x14,
		0x00,       // Bytesize
		0x31, 0x1a, // "_key": true
		0x32, 0x19, // "_rev": false
		0x33, 0x01, // "_id": []
		0x34, 0x18, // "_from": null
		0x35, 0x0a, // "_to": {}
		0x05, // NoItems
	}
	slice[1] = byte(len(slice))

	ASSERT_EQ("_key", mustString(mustSlice(slice.KeyAt(0, true)).GetString()), t)
	key := mustSlice(slice.Get("_key"))
	ASSERT_EQ(velocypack.Bool, key.Type(), t)
	ASSERT_TRUE(mustBool(key.GetBool()), t)

	ASSERT_EQ("_rev", mustString(mustSlice(slice.KeyAt(1, true)).GetString()), t)
	rev := mustSlice(slice.Get("_rev"))
	ASSERT_EQ(velocypack.Bool, rev.Type(), t)
	ASSERT_FALSE(mustBool(rev.GetBool()), t)

	ASSERT_EQ("_id", mustString(mustSlice(slice.KeyAt(2, true)).GetString()), t)
	id := mustSlice(slice.Get("_id"))
	ASSERT_EQ(velocypack.Array, id.Type(), t)

	ASSERT_EQ("_from", mustString(mustSlice(slice.KeyAt(3, true)).GetString()), t)
	from := mustSlice(slice.Get("_from"))
	ASSERT_EQ(velocypack.Null, from.Type(), t)

	ASSERT_EQ("_to", mustString(mustSlice(slice.KeyAt(4, true)).GetString()), t)
	to := mustSlice(slice.Get("_to"))
	ASSERT_EQ(velocypack.Object, to.Type(), t)
}
