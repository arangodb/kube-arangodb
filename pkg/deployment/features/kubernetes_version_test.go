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

// Test_KubernetesVersion_Getter verifies the KubernetesVersion accessor returns the declared constraint
// and that an undeclared constraint is empty (rendered as "Any" in the docs).
func Test_KubernetesVersion_Getter(t *testing.T) {
	require.True(t, (&feature{}).KubernetesVersion().Empty())
	require.EqualValues(t, ">= 1.29.0", (&feature{kubernetesVersion: ">= 1.29.0"}).KubernetesVersion())
}

// Test_SupportedKubernetesVersion verifies the dedicated Kubernetes version gate: it defaults to any
// (a feature with no constraint is supported on every Kubernetes version, including an unknown one),
// and a declared constraint is enforced and fails closed on an unknown version.
func Test_SupportedKubernetesVersion(t *testing.T) {
	t.Run("No constraint defaults to any", func(t *testing.T) {
		f := &feature{name: "test-no-k8s"}

		require.True(t, f.SupportedKubernetesVersion("1.29.0"))
		require.True(t, f.SupportedKubernetesVersion("1.21.0"))
		// Even an unknown (empty) Kubernetes version is accepted when no constraint is declared, so the
		// not-yet-discovered cluster version never disables an unconstrained feature.
		require.True(t, f.SupportedKubernetesVersion(""))
	})

	t.Run("Declared constraint gates support", func(t *testing.T) {
		f := &feature{name: "test-k8s", kubernetesVersion: ">= 1.29.0"}

		require.True(t, f.SupportedKubernetesVersion("1.29.0"))
		require.True(t, f.SupportedKubernetesVersion("1.29.4"))
		require.True(t, f.SupportedKubernetesVersion("1.30.5"))

		require.False(t, f.SupportedKubernetesVersion("1.28.9"))
		require.False(t, f.SupportedKubernetesVersion("1.21.0"))
	})

	t.Run("Unknown version fails closed when a constraint is declared", func(t *testing.T) {
		f := &feature{name: "test-k8s-unknown", kubernetesVersion: ">= 1.29.0"}

		// The cluster Kubernetes version is not yet discovered (empty); a Kubernetes-gated feature must
		// not be treated as supported.
		require.False(t, f.SupportedKubernetesVersion(""))
	})

	t.Run("Accessor and package function agree", func(t *testing.T) {
		f := &feature{name: "test-k8s-agree", kubernetesVersion: ">= 1.29.0"}

		require.Equal(t, f.SupportedKubernetesVersion("1.30.0"), SupportedKubernetesVersion(f, "1.30.0"))
		require.Equal(t, f.SupportedKubernetesVersion("1.20.0"), SupportedKubernetesVersion(f, "1.20.0"))
	})

	t.Run("Kubernetes gate is independent of the ArangoDB version gate", func(t *testing.T) {
		// SupportedKubernetesVersion evaluates only the Kubernetes constraint, regardless of the declared
		// ArangoDB version requirement or enabled state.
		f := &feature{name: "test-k8s-independent", version: ">= 3.12.0", kubernetesVersion: ">= 1.29.0"}

		require.True(t, f.SupportedKubernetesVersion("1.29.0"))
		require.False(t, f.SupportedKubernetesVersion("1.28.0"))
	})
}
