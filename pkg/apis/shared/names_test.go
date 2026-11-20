//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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
	"k8s.io/apimachinery/pkg/util/validation"
)

func Test_Names(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		require.EqualError(t, ValidateResourceName(""), "Name '' is not a valid resource name")
	})
	t.Run("Pod name is valid", func(t *testing.T) {
		name := CreatePodHostName("the-matrix-db", "arangodb-coordinator", "CRDN-549cznuy")
		require.Empty(t, validation.IsQualifiedName(name))

		name = CreatePodHostName("the-matrix-application-db-instance", "arangodb-coordinator", "CRDN-549cznuy")
		require.Empty(t, validation.IsQualifiedName(name))
	})
}

func Test_FixupResourceName(t *testing.T) {
	t.Run("Valid name is unchanged", func(t *testing.T) {
		require.Equal(t, "valid-name-123", FixupResourceName("valid-name-123"))
	})

	t.Run("Invalid characters trigger a 6-hex hash suffix", func(t *testing.T) {
		// An uppercase character makes the name non-DNS-valid, so a hash is appended.
		out := FixupResourceName("Invalid_Name")
		require.Regexp(t, `-[0-9a-f]{6}$`, out)
		require.Empty(t, validation.IsQualifiedName(out))

		// Deterministic.
		require.Equal(t, out, FixupResourceName("Invalid_Name"))
	})

	t.Run("Too long name is truncated to the length limit and hashed", func(t *testing.T) {
		long := ""
		for i := 0; i < 100; i++ {
			long += "a"
		}
		out := FixupResourceName(long)
		require.LessOrEqual(t, len(out), 63)
		require.Regexp(t, `-[0-9a-f]{6}$`, out)
		require.Equal(t, out, FixupResourceName(long))
	})
}
