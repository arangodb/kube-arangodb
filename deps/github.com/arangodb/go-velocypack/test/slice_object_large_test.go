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

// +build !nolarge

package test

import (
	"fmt"
	"math"
	"testing"

	velocypack "github.com/arangodb/go-velocypack"
)

func TestSliceObjectGetLengthMany2(t *testing.T) {
	max := math.MaxUint16
	var builder velocypack.Builder
	must(builder.OpenObject())
	for i := 1; i <= max; i++ {
		key := fmt.Sprintf("f%d", i)
		must(builder.AddKeyValue(key, velocypack.NewUIntValue(uint64(i)+10)))
	}
	must(builder.Close())
	slice := mustSlice(builder.Slice())

	for i := max; i >= 1; i-- {
		value := mustSlice(slice.Get(fmt.Sprintf("f%d", i)))
		ASSERT_EQ(velocypack.UInt, value.Type(), t)
		ASSERT_EQ(uint64(i)+10, mustUInt(value.GetUInt()), t)
	}
}
