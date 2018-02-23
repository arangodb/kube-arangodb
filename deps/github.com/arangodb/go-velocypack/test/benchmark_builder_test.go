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

func BenchmarkBuilderString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		builder := velocypack.NewBuilder(64)
		builder.AddValue(velocypack.NewStringValue("Some string"))
		if _, err := builder.Slice(); err != nil {
			b.Errorf("Slice failed: %v", err)
		}
	}
}

func BenchmarkBuilderObject1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		builder := velocypack.Builder{}
		builder.OpenObject()
		builder.AddKeyValue("Name", velocypack.NewStringValue("John Doe"))
		builder.AddKeyValue("Age", velocypack.NewIntValue(42))
		builder.Close()
		if _, err := builder.Slice(); err != nil {
			b.Errorf("Slice failed: %v", err)
		}
	}
}

func BenchmarkBuilderObject2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		builder := velocypack.Builder{}
		builder.OpenObject()
		builder.AddKeyValue("Name", velocypack.NewStringValue("John Doe"))
		builder.AddKeyValue("FirstName", velocypack.NewStringValue("John"))
		builder.AddKeyValue("LastName", velocypack.NewStringValue("Doe"))
		builder.AddKeyValue("Age", velocypack.NewIntValue(42))
		builder.AddValue(velocypack.NewStringValue("Address"))
		builder.OpenArray()
		builder.AddValue(velocypack.NewStringValue("Some street"))
		builder.AddValue(velocypack.NewStringValue("Block  123"))
		builder.AddValue(velocypack.NewStringValue("South"))
		builder.Close()
		builder.Close()
		if _, err := builder.Slice(); err != nil {
			b.Errorf("Slice failed: %v", err)
		}
	}
}
