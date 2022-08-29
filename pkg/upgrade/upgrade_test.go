//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package upgrade

import (
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func genNewVersionAppender(v *[]api.Version, version api.Version) Upgrade {
	return newUpgrade(version, func(obj api.ArangoDeployment, status *api.DeploymentStatus, cache interfaces.Inspector) error {
		*v = append(*v, version)
		return nil
	})
}

func Test_Verify(t *testing.T) {
	require.NoError(t, upgrades.Verify())
}

func Test_Verify_WrongOrder(t *testing.T) {
	t.Run("Invalid version - starts from 0", func(t *testing.T) {
		var u Upgrades

		var v []api.Version

		u = append(u, genNewVersionAppender(&v, api.Version{
			Major: 1,
			Minor: 1,
			Patch: 1,
			ID:    0,
		}))

		_, err := u.Execute(api.ArangoDeployment{}, nil, nil)
		require.EqualError(t, err, "Invalid version in 1.1.1 - got 0, expected 1")
	})
	t.Run("Invalid version - missing middle", func(t *testing.T) {
		var u Upgrades

		var v []api.Version

		u = append(u,
			genNewVersionAppender(&v, api.Version{
				Major: 1,
				Minor: 1,
				Patch: 1,
				ID:    1,
			}),
			genNewVersionAppender(&v, api.Version{
				Major: 1,
				Minor: 1,
				Patch: 1,
				ID:    3,
			}),
		)

		_, err := u.Execute(api.ArangoDeployment{}, nil, nil)
		require.EqualError(t, err, "Invalid version in 1.1.1 - got 3, expected 2")
	})
	t.Run("Valid multi version", func(t *testing.T) {
		var u Upgrades

		var v []api.Version

		u = append(u,
			genNewVersionAppender(&v, api.Version{
				Major: 1,
				Minor: 1,
				Patch: 1,
				ID:    1,
			}),
			genNewVersionAppender(&v, api.Version{
				Major: 1,
				Minor: 1,
				Patch: 2,
				ID:    1,
			}),
		)

		_, err := u.Execute(api.ArangoDeployment{}, nil, nil)
		require.NoError(t, err)
		require.Len(t, v, 2)
		require.Equal(t, api.Version{
			Major: 1,
			Minor: 1,
			Patch: 1,
			ID:    1,
		}, v[0])
		require.Equal(t, api.Version{
			Major: 1,
			Minor: 1,
			Patch: 2,
			ID:    1,
		}, v[1])
	})
	t.Run("Valid multi version - rev order", func(t *testing.T) {
		var u Upgrades

		var v []api.Version

		u = append(u,
			genNewVersionAppender(&v, api.Version{
				Major: 1,
				Minor: 1,
				Patch: 2,
				ID:    1,
			}),
			genNewVersionAppender(&v, api.Version{
				Major: 1,
				Minor: 1,
				Patch: 1,
				ID:    1,
			}),
		)

		_, err := u.Execute(api.ArangoDeployment{}, nil, nil)
		require.NoError(t, err)
		require.Len(t, v, 2)
		require.Equal(t, api.Version{
			Major: 1,
			Minor: 1,
			Patch: 1,
			ID:    1,
		}, v[0])
		require.Equal(t, api.Version{
			Major: 1,
			Minor: 1,
			Patch: 2,
			ID:    1,
		}, v[1])
	})
	t.Run("Valid multi version - only upgrade", func(t *testing.T) {
		var u Upgrades

		var v []api.Version

		obj := api.ArangoDeployment{
			Status: api.DeploymentStatus{
				Version: &api.Version{
					Major: 1,
					Minor: 1,
					Patch: 2,
					ID:    1,
				},
			},
		}

		u = append(u,
			genNewVersionAppender(&v, api.Version{
				Major: 1,
				Minor: 1,
				Patch: 2,
				ID:    1,
			}),
			genNewVersionAppender(&v, api.Version{
				Major: 1,
				Minor: 1,
				Patch: 1,
				ID:    1,
			}),
		)

		status := obj.Status.DeepCopy()

		_, err := u.Execute(obj, status, nil)
		require.NoError(t, err)
		require.Len(t, v, 1)
		require.Equal(t, api.Version{
			Major: 1,
			Minor: 1,
			Patch: 2,
			ID:    1,
		}, v[0])
	})
}

func Test_RunUpgrade(t *testing.T) {
	obj := api.ArangoDeployment{}

	t.Run("Prepare", func(t *testing.T) {
		testMemberCIDAppendPrepare(t, &obj)
	})

	status := obj.Status.DeepCopy()

	c := kclient.NewFakeClient()

	i := tests.NewInspector(t, c)

	t.Run("Upgrade", func(t *testing.T) {
		changed, err := RunUpgrade(obj, status, i)
		require.NoError(t, err)
		require.True(t, changed)
	})

	obj.Status = *status

	t.Run("Check", func(t *testing.T) {
		testMemberCIDAppendCheck(t, obj)
	})

	require.NotNil(t, obj.Status.Version)
	require.Equal(t, upgrades[len(upgrades)-1].Version(), *obj.Status.Version)
}
