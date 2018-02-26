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

func TestParserGarbage(t *testing.T) {
	tests := map[string]func(error) bool{
		`foo`:            velocypack.IsParse,
		`'quoted "foo"'`: velocypack.IsParse,
		`x`:              velocypack.IsParse,
		`!`:              velocypack.IsParse,
		`/`:              velocypack.IsParse,
		`-`:              velocypack.IsParse,
		`--11`:           velocypack.IsParse,
		`[[}`:            velocypack.IsParse,
		`5.6.7`:          velocypack.IsParse,
		`[`:              velocypack.IsBuilderNotClosed,
		`{`:              velocypack.IsBuilderNotClosed,
	}
	for test, errFunc := range tests {
		ASSERT_VELOCYPACK_EXCEPTION(errFunc, t)(velocypack.ParseJSONFromString(test))
	}
}
