//
// DISCLAIMER
//
// Copyright 2023-2024 ArangoDB GmbH, Cologne, Germany
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

package shared

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func Test_ValidateUID(t *testing.T) {
	require.Error(t, ValidateUID(""))
	require.NoError(t, ValidateUID(uuid.NewUUID()))
}

func Test_ValidateAPIPath(t *testing.T) {
	require.NoError(t, ValidateAPIPath(""))
	require.NoError(t, ValidateAPIPath("/"))
	require.Error(t, ValidateAPIPath("//"))
	require.NoError(t, ValidateAPIPath("/api/zz"))
	require.NoError(t, ValidateAPIPath("/api/"))
	require.NoError(t, ValidateAPIPath("/api/test/qw/"))
	require.NoError(t, ValidateAPIPath("/api/test/2/"))
	require.Error(t, ValidateAPIPath("/&/"))
}

func Test_ValidateExclusiveFields(t *testing.T) {
	type z struct {
		A string `json:"a,omitempty"`
		B string `json:"b,omitempty"`
		C string `json:"c,omitempty"`
		D string `json:"d,omitempty"`
	}

	require.EqualError(t, ValidateExclusiveFields(z{}, 1, "A"), "Elements not provided. Expected 1. Possible: a")

	require.NoError(t, ValidateExclusiveFields(z{
		A: "test",
	}, 1, "A"))

	require.EqualError(t, ValidateExclusiveFields(z{
		A: "test",
	}, 2, "A"), "Expected more fields than allowed")

	require.EqualError(t, ValidateExclusiveFields(z{
		A: "test",
	}, 2, "A", "B"), "Not enough elements provided. Expected 2, got 1. Defined: a, Additionally Possible: b")

	require.NoError(t, ValidateExclusiveFields(z{
		A: "test",
		B: "test",
	}, 2, "A", "B"))

	require.EqualError(t, ValidateExclusiveFields(z{
		A: "test",
		B: "test",
	}, 1, "A", "B"), "Too many elements provided. Expected 1, got 2. Defined: a, b")

	require.NoError(t, ValidateExclusiveFields(z{
		A: "test",
		D: "test",
	}, 2, "A", "B", "C", "D"))
}
