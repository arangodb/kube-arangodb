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
	"bytes"
	"testing"

	velocypack "github.com/arangodb/go-velocypack"
)

func TestDecoderReaderSmall(t *testing.T) {
	s := append(velocypack.NullSlice(), velocypack.TrueSlice()...)
	r := bytes.NewReader(s)
	d := velocypack.NewDecoder(r)

	var v1 interface{}
	must(d.Decode(&v1))

	var v2 bool
	must(d.Decode(&v2))

	ASSERT_EQ(v1, nil, t)
	ASSERT_EQ(v2, true, t)
}

func TestDecoderReaderLarge(t *testing.T) {
	testX := func(x int) string {
		result := ""
		for i := 0; i < x; i++ {
			result = result + "-foo-"
		}
		return result
	}
	b := velocypack.Builder{}
	for i := 0; i < 1000; i++ {
		must(b.AddValue(velocypack.NewStringValue(testX(i))))
	}
	s := mustSlice(b.Slice())
	r := bytes.NewReader(s)
	d := velocypack.NewDecoder(r)

	for i := 0; i < 1000; i++ {
		var v string
		must(d.Decode(&v))
		ASSERT_EQ(v, testX(i), t)
	}
}

func TestDecoderReaderStruct1(t *testing.T) {
	var s velocypack.Slice
	for i := 0; i < 1000; i++ {
		input := Struct1{
			Field1: i,
		}
		s = append(s, mustSlice(velocypack.Marshal(input))...)
	}
	r := bytes.NewReader(s)
	d := velocypack.NewDecoder(r)

	for i := 0; i < 1000; i++ {
		var v Struct1
		must(d.Decode(&v))
		expected := Struct1{
			Field1: i,
		}
		ASSERT_EQ(v, expected, t)
	}
}
