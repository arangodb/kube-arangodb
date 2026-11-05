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

package sidecar

import (
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// enableFeatureWithDependencies enables f and all of its transitive dependencies for the duration of
// the test, restoring the previous state afterwards.
func enableFeatureWithDependencies(t *testing.T, f features.Feature) {
	seen := map[string]bool{}

	var enable func(x features.Feature)
	enable = func(x features.Feature) {
		if seen[x.Name()] {
			return
		}
		seen[x.Name()] = true

		p := x.EnabledPointer()
		old := *p
		*p = true
		t.Cleanup(func() { *p = old })

		for _, d := range x.Dependencies() {
			enable(d)
		}
	}

	enable(f)
}

func Test_securityModeEnvs(t *testing.T) {
	modes := func(spec api.DeploymentSpec, status api.DeploymentStatus) map[string]string {
		m := map[string]string{}
		for _, e := range securityModeEnvs(spec, status) {
			m[e.Name] = e.Value
		}
		return m
	}

	t.Run("Authentication disabled", func(t *testing.T) {
		e := modes(api.DeploymentSpec{Authentication: api.AuthenticationSpec{JWTSecretName: util.NewType("None")}}, api.DeploymentStatus{})
		require.Equal(t, "None", e["INTEGRATION_AUTHENTICATION_MODE"])
		require.Equal(t, "None", e["INTEGRATION_AUTHORIZATION_MODE"])
		require.Equal(t, "None", e["INTEGRATION_AUTHORIZATION_MODE_COREDB"])
	})

	t.Run("Native", func(t *testing.T) {
		e := modes(api.DeploymentSpec{}, api.DeploymentStatus{})
		require.Equal(t, "Native", e["INTEGRATION_AUTHENTICATION_MODE"])
		require.Equal(t, "Native", e["INTEGRATION_AUTHORIZATION_MODE"])
		require.Equal(t, "Native", e["INTEGRATION_AUTHORIZATION_MODE_COREDB"])
	})

	t.Run("Platform RBAC, CoreDB Native (rbac-coredb disabled)", func(t *testing.T) {
		var status api.DeploymentStatus
		status.Conditions.Update(api.ConditionTypeGatewaySidecarEnabled, true, "", "")

		e := modes(api.DeploymentSpec{}, status)
		require.Equal(t, "Native", e["INTEGRATION_AUTHENTICATION_MODE"])
		require.Equal(t, "RBAC", e["INTEGRATION_AUTHORIZATION_MODE"])
		require.Equal(t, "Native", e["INTEGRATION_AUTHORIZATION_MODE_COREDB"])
	})

	t.Run("Platform RBAC, CoreDB RBAC (rbac-coredb enabled)", func(t *testing.T) {
		enableFeatureWithDependencies(t, features.RBACCoreDB())

		var status api.DeploymentStatus
		status.Conditions.Update(api.ConditionTypeGatewaySidecarEnabled, true, "", "")

		e := modes(api.DeploymentSpec{}, status)
		require.Equal(t, "Native", e["INTEGRATION_AUTHENTICATION_MODE"])
		require.Equal(t, "RBAC", e["INTEGRATION_AUTHORIZATION_MODE"])
		// The rbac-coredb feature wires --server.external-rbac-service, so the ArangoDB core enforces RBAC.
		require.Equal(t, "RBAC", e["INTEGRATION_AUTHORIZATION_MODE_COREDB"])
	})
}
