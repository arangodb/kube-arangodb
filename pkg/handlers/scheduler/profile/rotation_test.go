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

package profile

import (
	"testing"

	"github.com/stretchr/testify/require"

	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
)

func Test_profileChecksumMismatch(t *testing.T) {
	applied := utilConstants.ProfilesAnnotationApplied
	checksum := utilConstants.ProfilesAnnotationChecksum

	t.Run("Not profile-injected is never a mismatch", func(t *testing.T) {
		_, changed := profileChecksumMismatch(map[string]string{checksum: "abc"}, "def")
		require.False(t, changed, "without applied=true the pod must be left alone")
	})

	t.Run("Applied but no stamped checksum is not a mismatch", func(t *testing.T) {
		_, changed := profileChecksumMismatch(map[string]string{applied: "true"}, "def")
		require.False(t, changed)
	})

	t.Run("Equal checksum is not a mismatch", func(t *testing.T) {
		_, changed := profileChecksumMismatch(map[string]string{applied: "true", checksum: "abc"}, "abc")
		require.False(t, changed)
	})

	t.Run("Different checksum is a mismatch with a reason", func(t *testing.T) {
		reason, changed := profileChecksumMismatch(map[string]string{applied: "true", checksum: "abc"}, "def")
		require.True(t, changed)
		require.Contains(t, reason, "abc")
		require.Contains(t, reason, "def")
	})
}

func Test_profilesForPod(t *testing.T) {
	list := utilConstants.ProfilesList

	t.Run("Empty when no profiles are referenced", func(t *testing.T) {
		require.Empty(t, profilesForPod(map[string]string{}, map[string]string{}))
	})

	t.Run("Merges label and annotation, trims and drops empties", func(t *testing.T) {
		got := profilesForPod(
			map[string]string{list: " a , b "},
			map[string]string{list: "c,,"},
		)
		require.ElementsMatch(t, []string{"a", "b", "c"}, got)
	})
}
