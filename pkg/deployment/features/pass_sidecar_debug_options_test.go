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

package features

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Test_PassSidecarDebugOptions verifies the pass-sidecar-debug-options feature definition: it is a hidden,
// disabled-by-default diagnostic flag, its accessor returns the registered feature, and toggling it is
// reflected by Enabled().
func Test_PassSidecarDebugOptions(t *testing.T) {
	f := PassSidecarDebugOptions()
	require.NotNil(t, f)

	t.Run("Definition", func(t *testing.T) {
		require.Equal(t, "pass-sidecar-debug-options", f.Name())
		require.True(t, f.Hidden(), "diagnostic flag must stay hidden")
		require.False(t, f.EnabledByDefault(), "must be off by default")
		require.False(t, f.EnterpriseRequired())
		require.False(t, f.OperatorEnterpriseRequired())
		require.Empty(t, f.GetDependencies())
		require.True(t, f.Version().Empty(), "no ArangoDB version constraint")
	})

	t.Run("Accessor is the registered feature", func(t *testing.T) {
		// The exported accessor and the package-level var are the same instance.
		require.Same(t, passSidecarDebugOptions, f)
	})

	t.Run("Toggle", func(t *testing.T) {
		restore := *passSidecarDebugOptions.EnabledPointer()
		defer func() { *passSidecarDebugOptions.EnabledPointer() = restore }()

		*passSidecarDebugOptions.EnabledPointer() = false
		require.False(t, PassSidecarDebugOptions().Enabled())

		*passSidecarDebugOptions.EnabledPointer() = true
		require.True(t, PassSidecarDebugOptions().Enabled())
	})
}
