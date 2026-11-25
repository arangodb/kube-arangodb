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

package resources

import (
	goStrings "strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/logging"
)

// setConfiguredLogLevels drives logging.ConfiguredLevels() through the real --log.level flag binding, so the
// sidecar arg builders observe the given operator-level values. Re-binding on a fresh command resets the
// backing slice to its default before the provided values are parsed.
func setConfiguredLogLevels(t *testing.T, levels ...string) {
	cmd := &cobra.Command{}
	require.NoError(t, logging.Init(cmd))

	args := make([]string, 0, len(levels))
	for _, l := range levels {
		args = append(args, "--log.level="+l)
	}
	require.NoError(t, cmd.ParseFlags(args))
}

// Test_createInternalSidecarArgs_PassSidecarDebugOptions verifies that the internal authorization sidecar
// propagates the operator's --log.level only when the pass-sidecar-debug-options feature is enabled, and that
// every configured level is passed through verbatim.
func Test_createInternalSidecarArgs_PassSidecarDebugOptions(t *testing.T) {
	restore := *features.PassSidecarDebugOptions().EnabledPointer()
	defer func() {
		*features.PassSidecarDebugOptions().EnabledPointer() = restore
		setConfiguredLogLevels(t) // reset the global levels back to the default binding
	}()

	countLevels := func(args []string) int {
		n := 0
		for _, a := range args {
			if goStrings.HasPrefix(a, "--log.level=") {
				n++
			}
		}
		return n
	}
	hasArg := func(args []string, v string) bool {
		for _, a := range args {
			if a == v {
				return true
			}
		}
		return false
	}

	t.Run("disabled - operator level not propagated", func(t *testing.T) {
		*features.PassSidecarDebugOptions().EnabledPointer() = false
		setConfiguredLogLevels(t, "debug")

		args := createInternalSidecarArgs(api.DeploymentSpec{}, api.ServerGroupSpec{}, nil)
		require.Zero(t, countLevels(args), "no --log.level must be emitted while the feature is disabled")
	})

	t.Run("enabled - operator level propagated", func(t *testing.T) {
		*features.PassSidecarDebugOptions().EnabledPointer() = true
		setConfiguredLogLevels(t, "debug")

		args := createInternalSidecarArgs(api.DeploymentSpec{}, api.ServerGroupSpec{}, nil)
		require.True(t, hasArg(args, "--log.level=debug"))
		require.Equal(t, 1, countLevels(args))
	})

	t.Run("enabled - multiple scoped levels propagated verbatim", func(t *testing.T) {
		*features.PassSidecarDebugOptions().EnabledPointer() = true
		setConfiguredLogLevels(t, "info", "grpc-service=debug")

		args := createInternalSidecarArgs(api.DeploymentSpec{}, api.ServerGroupSpec{}, nil)
		require.True(t, hasArg(args, "--log.level=info"))
		require.True(t, hasArg(args, "--log.level=grpc-service=debug"))
		require.Equal(t, 2, countLevels(args))
	})
}
