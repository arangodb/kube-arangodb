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
	"strings"
	"testing"

	velocypack "github.com/arangodb/go-velocypack"
)

func TestBuilderObjectLarge(t *testing.T) {
	var obj velocypack.Builder
	max := math.MaxInt16 * 2
	expected := make([]string, max)
	must(obj.OpenObject())
	for i := 0; i < max; i++ {
		must(obj.AddValue(velocypack.NewStringValue(fmt.Sprintf("x%06d", i))))
		must(obj.AddValue(velocypack.NewIntValue(int64(i))))
		expected[i] = fmt.Sprintf(`"x%06d":%d`, i, i)
	}
	must(obj.Close())
	objSlice := mustSlice(obj.Slice())

	expectedJSON := "{" + strings.Join(expected, ",") + "}"
	ASSERT_EQ(expectedJSON, mustString(objSlice.JSONString()), t)
}
