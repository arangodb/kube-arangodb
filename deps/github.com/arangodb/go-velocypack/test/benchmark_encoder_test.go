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
	"encoding/json"
	"testing"

	velocypack "github.com/arangodb/go-velocypack"
)

type (
	benchmarkObjectType struct {
		Name      string
		FirstName string
		LastName  string
		Age       int
		Address   []string
	}
)

var (
	benchmarkObjectInput = benchmarkObjectType{
		Name:      "John Doe",
		FirstName: "John",
		LastName:  "Doe",
		Age:       42,
		Address: []string{
			"Some street",
			"Block 123",
			"South",
		},
	}
)

func BenchmarkVPackEncoderObject(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := velocypack.Marshal(benchmarkObjectInput); err != nil {
			b.Errorf("Marshal failed: %v", err)
		}
	}
}

func BenchmarkJSONEncoderObject(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(benchmarkObjectInput); err != nil {
			b.Errorf("Marshal failed: %v", err)
		}
	}
}
