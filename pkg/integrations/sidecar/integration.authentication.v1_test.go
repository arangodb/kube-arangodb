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
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	schedulerContainerResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container/resources"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
)

func envByName(envs []core.EnvVar, name string) (string, bool) {
	for _, e := range envs {
		if e.Name == name {
			return e.Value, true
		}
	}
	return "", false
}

// Test_IntegrationAuthenticationV1_ArangoJWTFolder asserts that INTEGRATION_ARANGO_JWT_FOLDER is set
// only when the integration mounts the cluster JWT volume (i.e. only when authentication is enabled),
// so the env is never present without the folder it points at.
func Test_IntegrationAuthenticationV1_ArangoJWTFolder(t *testing.T) {
	t.Run("Authenticated sets the env and mounts the volume together", func(t *testing.T) {
		i := IntegrationAuthenticationV1{Core: &Core{}, DeploymentName: "test", Spec: api.DeploymentSpec{}}
		require.True(t, i.Spec.IsAuthenticated())

		envs, err := i.Envs()
		require.NoError(t, err)
		v, ok := envByName(envs, utilConstants.INTEGRATION_ARANGO_JWT_FOLDER.String())
		require.True(t, ok, "env must be set when authenticated")
		require.Equal(t, shared.ClusterJWTSecretVolumeMountDir, v)

		vols, mounts, err := i.Volumes()
		require.NoError(t, err)
		require.NotEmpty(t, vols, "cluster JWT volume must be mounted alongside the env")
		require.NotEmpty(t, mounts)
	})

	t.Run("Unauthenticated omits both the env and the volume", func(t *testing.T) {
		i := IntegrationAuthenticationV1{Core: &Core{}, DeploymentName: "test", Spec: api.DeploymentSpec{
			Authentication: api.AuthenticationSpec{JWTSecretName: util.NewType(api.JWTSecretNameDisabled)},
		}}
		require.False(t, i.Spec.IsAuthenticated())

		envs, err := i.Envs()
		require.NoError(t, err)
		_, ok := envByName(envs, utilConstants.INTEGRATION_ARANGO_JWT_FOLDER.String())
		require.False(t, ok, "env must be absent when the folder is not mounted")

		vols, mounts, err := i.Volumes()
		require.NoError(t, err)
		require.Empty(t, vols)
		require.Empty(t, mounts)
	})
}

// Test_NewIntegration_NoArangoJWTFolderOnSharedContainer is the regression for the move: the shared
// integration container must never carry INTEGRATION_ARANGO_JWT_FOLDER, even when the deployment is
// authenticated. The env belongs to the authentication v1 integration profile (co-located with the
// cluster JWT volume it mounts), so that when that integration is dropped (e.g. served by the central
// service) the env is dropped with it rather than pointing at an unmounted folder.
func Test_NewIntegration_NoArangoJWTFolderOnSharedContainer(t *testing.T) {
	assertNoFolderEnv := func(t *testing.T, spec api.DeploymentSpec) {
		pt, err := NewIntegration("test", spec, api.DeploymentStatus{},
			&schedulerContainerResourcesApi.Image{Image: util.NewType("img")},
			spec.Integration.GetSidecar())
		require.NoError(t, err)
		require.NotNil(t, pt)
		require.NotNil(t, pt.Container)

		c, ok := pt.Container.Containers[ContainerName]
		require.True(t, ok, "integration container must exist")
		require.NotNil(t, c.Environments)

		for _, e := range c.Environments.Env {
			require.NotEqual(t, utilConstants.INTEGRATION_ARANGO_JWT_FOLDER.String(), e.Name,
				"the shared integration container must not set INTEGRATION_ARANGO_JWT_FOLDER")
		}
	}

	t.Run("Authenticated", func(t *testing.T) {
		spec := api.DeploymentSpec{}
		require.True(t, spec.IsAuthenticated())
		assertNoFolderEnv(t, spec)
	})

	t.Run("Unauthenticated", func(t *testing.T) {
		spec := api.DeploymentSpec{Authentication: api.AuthenticationSpec{JWTSecretName: util.NewType(api.JWTSecretNameDisabled)}}
		require.False(t, spec.IsAuthenticated())
		assertNoFolderEnv(t, spec)
	})
}
