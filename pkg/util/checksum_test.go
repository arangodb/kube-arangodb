//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SHA224(t *testing.T) {
	// Known FIPS 180-4 vector for SHA-224 of the empty input.
	t.Run("EmptyVector", func(t *testing.T) {
		require.Equal(t, "d14a028c2a3a2bc9476102bb288234c415a2b01f828ea62ac5b3e42f", SHA224(nil))
		require.Equal(t, "d14a028c2a3a2bc9476102bb288234c415a2b01f828ea62ac5b3e42f", SHA224FromString(""))
	})

	t.Run("Length", func(t *testing.T) {
		// SHA-224 is 56 hex characters, which fits the 63-character Kubernetes label value limit.
		require.Len(t, SHA224FromString("arangodb"), 56)
	})

	t.Run("Deterministic", func(t *testing.T) {
		require.Equal(t, SHA224FromString("arangodb"), SHA224FromString("arangodb"))
		require.Equal(t, SHA224([]byte("arangodb")), SHA224FromString("arangodb"))
	})

	t.Run("Distinct", func(t *testing.T) {
		require.NotEqual(t, SHA224FromString("arangodb"), SHA224FromString("arangodc"))
	})
}

func Test_SHA256(t *testing.T) {
	// Known FIPS 180-4 vector for SHA-256 of the empty input.
	t.Run("EmptyVector", func(t *testing.T) {
		require.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", SHA256(nil))
	})

	t.Run("Length", func(t *testing.T) {
		require.Len(t, SHA256FromString("arangodb"), 64)
	})

	t.Run("Deterministic", func(t *testing.T) {
		require.Equal(t, SHA256FromString("arangodb"), SHA256FromString("arangodb"))
	})
}
