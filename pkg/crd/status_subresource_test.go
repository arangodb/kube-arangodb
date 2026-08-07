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

package crd

import (
	"testing"

	"github.com/stretchr/testify/require"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/arangodb/kube-arangodb/pkg/crd/crds"
)

func v1StatusSubresource(t *testing.T, spec *apiextensions.CustomResourceDefinitionSpec) *apiextensions.CustomResourceSubresourceStatus {
	t.Helper()

	status, ok := statusSubresourceForVersion(spec, "v1")
	require.True(t, ok, "v1 version must exist")
	return status
}

// Test_MirrorStorageStatusSubresource covers the disabled-feature behaviour: the operator must not
// touch the status subresource on the v1 (storage) version — neither adding it when the live CRD
// lacks it, nor removing it when the chart already ships it.
func Test_MirrorStorageStatusSubresource(t *testing.T) {
	t.Run("Disabled keeps the subresource when the live CRD already has it (chart-shipped)", func(t *testing.T) {
		// Desired spec renders with the feature disabled; the chart ships the subresource on v1.
		desired := crds.DatabaseDeploymentWithOptions().Spec
		// The live CRD already serves it (installed from the chart).
		live := crds.DatabaseDeploymentWithOptions(crds.WithStatusSubresource()).Spec

		require.NotNil(t, v1StatusSubresource(t, &live), "precondition: live CRD serves the v1 status subresource")

		mirrorStorageStatusSubresource(&desired, &live)

		// Not removed: the subresource stays.
		require.NotNil(t, v1StatusSubresource(t, &desired))
		require.NotNil(t, statusSubresourceOnVersion(t, &desired, "v2alpha1"))
	})

	t.Run("Disabled does not add the subresource when the live CRD lacks it", func(t *testing.T) {
		// Desired spec (chart-shipped) has the v1 subresource, but the live CRD does not.
		desired := crds.DatabaseDeploymentWithOptions().Spec
		live := crds.DatabaseDeploymentWithOptions().Spec
		clearV1StatusSubresource(&live)

		require.Nil(t, v1StatusSubresource(t, &live), "precondition: live CRD lacks the v1 status subresource")
		require.NotNil(t, v1StatusSubresource(t, &desired), "precondition: desired (chart) spec has the v1 status subresource")

		mirrorStorageStatusSubresource(&desired, &live)

		// Not added: the desired spec is brought down to the live state, so applying it is a no-op.
		require.Nil(t, v1StatusSubresource(t, &desired))
		// v2alpha1 is not a storage version, so it is left as the chart ships it.
		require.NotNil(t, statusSubresourceOnVersion(t, &desired, "v2alpha1"))
	})
}

func statusSubresourceOnVersion(t *testing.T, spec *apiextensions.CustomResourceDefinitionSpec, version string) *apiextensions.CustomResourceSubresourceStatus {
	t.Helper()

	status, ok := statusSubresourceForVersion(spec, version)
	require.True(t, ok, "version %s must exist", version)
	return status
}

func clearV1StatusSubresource(spec *apiextensions.CustomResourceDefinitionSpec) {
	for i := range spec.Versions {
		if spec.Versions[i].Name == "v1" {
			spec.Versions[i].Subresources = nil
		}
	}
}
