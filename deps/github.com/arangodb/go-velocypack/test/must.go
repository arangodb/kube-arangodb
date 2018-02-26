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
	"time"

	velocypack "github.com/arangodb/go-velocypack"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func mustArrayIterator(v *velocypack.ArrayIterator, err error) *velocypack.ArrayIterator {
	if err != nil {
		panic(err)
	}
	return v
}

func mustBool(v bool, err error) bool {
	if err != nil {
		panic(err)
	}
	return v
}

func mustBytes(v []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return v
}

func mustDouble(v float64, err error) float64 {
	if err != nil {
		panic(err)
	}
	return v
}

func mustInt(v int64, err error) int64 {
	if err != nil {
		panic(err)
	}
	return v
}

func mustGoInt(v int, err error) int {
	if err != nil {
		panic(err)
	}
	return v
}

func mustLength(v velocypack.ValueLength, err error) velocypack.ValueLength {
	if err != nil {
		panic(err)
	}
	return v
}

func mustObjectIterator(v *velocypack.ObjectIterator, err error) *velocypack.ObjectIterator {
	if err != nil {
		panic(err)
	}
	return v
}

func mustSlice(v velocypack.Slice, err error) velocypack.Slice {
	if err != nil {
		panic(err)
	}
	return v
}

func mustString(v string, err error) string {
	if err != nil {
		panic(err)
	}
	return v
}

func mustTime(v time.Time, err error) time.Time {
	if err != nil {
		panic(err)
	}
	return v
}

func mustUInt(v uint64, err error) uint64 {
	if err != nil {
		panic(err)
	}
	return v
}
