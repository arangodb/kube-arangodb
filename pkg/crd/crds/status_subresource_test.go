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

package crds

import (
	"testing"

	"github.com/stretchr/testify/require"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func statusSubresource(t *testing.T, crd *apiextensions.CustomResourceDefinition, version string) *apiextensions.CustomResourceSubresourceStatus {
	t.Helper()

	for _, v := range crd.Spec.Versions {
		if v.Name == version {
			if v.Subresources == nil {
				return nil
			}
			return v.Subresources.Status
		}
	}

	t.Fatalf("version %s not found", version)
	return nil
}

func Test_WithStatusSubresource_ArangoDeployment(t *testing.T) {
	t.Run("The chart ships the status subresource on v1 and v2alpha1", func(t *testing.T) {
		// The raw (ungated) CRD, as shipped in the chart, carries the status subresource on both the
		// v1 (storage) version and v2alpha1.
		raw := DatabaseDeploymentDefinitionData().definitionLoader().MustGet()

		require.NotNil(t, statusSubresource(t, &raw, "v1"))
		require.NotNil(t, statusSubresource(t, &raw, "v2alpha1"))
	})

	t.Run("Feature enabled keeps the shipped v1 status subresource", func(t *testing.T) {
		crd := DatabaseDeploymentWithOptions(WithStatusSubresource())

		require.NotNil(t, statusSubresource(t, crd, "v1"))
		require.NotNil(t, statusSubresource(t, crd, "v2alpha1"))
	})

	t.Run("Feature disabled leaves the shipped spec untouched", func(t *testing.T) {
		// The rendered spec is not stripped when the feature is off — it stays as the chart ships it.
		// Whether the live CRD keeps the subresource is decided at apply time against the live CRD, so
		// that a disabled operator neither adds nor removes it.
		crd := DatabaseDeploymentWithOptions()

		require.NotNil(t, statusSubresource(t, crd, "v1"))
		require.NotNil(t, statusSubresource(t, crd, "v2alpha1"))
	})
}
