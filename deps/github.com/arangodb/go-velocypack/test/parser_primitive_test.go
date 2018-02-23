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
	"math"
	"strconv"
	"testing"

	velocypack "github.com/arangodb/go-velocypack"
)

func TestParserNull(t *testing.T) {
	s := mustSlice(velocypack.ParseJSONFromString("null"))

	ASSERT_TRUE(s.IsNull(), t)
}

func TestParserWhitespace(t *testing.T) {
	s := mustSlice(velocypack.ParseJSONFromString(" "))

	ASSERT_TRUE(s.IsNone(), t)
}

func TestParserFalse(t *testing.T) {
	s := mustSlice(velocypack.ParseJSONFromString("false"))

	ASSERT_TRUE(s.IsBool(), t)
	ASSERT_EQ(false, mustBool(s.GetBool()), t)
}

func TestParserTrue(t *testing.T) {
	s := mustSlice(velocypack.ParseJSONFromString("true"))

	ASSERT_TRUE(s.IsBool(), t)
	ASSERT_EQ(true, mustBool(s.GetBool()), t)
}

func TestParserSmallInt(t *testing.T) {
	tests := []int{-6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for _, test := range tests {
		s := mustSlice(velocypack.ParseJSONFromString(strconv.Itoa(test)))

		ASSERT_EQ(velocypack.SmallInt, s.Type(), t)
		ASSERT_EQ(int64(test), mustInt(s.GetInt()), t)
	}
}

func TestParserInt(t *testing.T) {
	tests := []int{-7, -10, -23, -456, math.MinInt32}
	for _, test := range tests {
		s := mustSlice(velocypack.ParseJSONFromString(strconv.Itoa(test)))

		ASSERT_EQ(velocypack.Int, s.Type(), t)
		ASSERT_EQ(int64(test), mustInt(s.GetInt()), t)
	}
}

func TestParserUInt(t *testing.T) {
	tests := []int{10, 23, 456, math.MaxInt32}
	for _, test := range tests {
		s := mustSlice(velocypack.ParseJSONFromString(strconv.Itoa(test)))

		ASSERT_EQ(velocypack.UInt, s.Type(), t)
		ASSERT_EQ(uint64(test), mustUInt(s.GetUInt()), t)
	}
}

func TestParserDouble(t *testing.T) {
	tests := []float64{10.77, 23.88, 456.01, 10e45, -9223372036854775809 /*MinInt64-1*/, 18446744073709551616 /*MaxUint64+1*/}
	jsons := []string{"10.77", "23.88", "456.01", "10e45", "-9223372036854775809", "18446744073709551616"}
	for i, test := range tests {
		s := mustSlice(velocypack.ParseJSONFromString(jsons[i]))

		ASSERT_EQ(velocypack.Double, s.Type(), t)
		ASSERT_DOUBLE_EQ(test, mustDouble(s.GetDouble()), t)
	}
}

func TestParserString(t *testing.T) {
	tests := []string{
		`foo`,
		`'quoted "foo"'`,
		``,
	}
	for _, test := range tests {
		j, err := json.Marshal(test)
		ASSERT_NIL(err, t)
		s := mustSlice(velocypack.ParseJSONFromString(string(j)))

		ASSERT_EQ(velocypack.String, s.Type(), t)
		ASSERT_EQ(test, mustString(s.GetString()), t)
	}
}
