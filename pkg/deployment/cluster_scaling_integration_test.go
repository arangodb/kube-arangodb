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

package deployment

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	memberState "github.com/arangodb/kube-arangodb/pkg/deployment/member"
)

// Without a serving member in the member state the database calls must fail
// instead of falling back to the database client Service.
func newClusterScalingTestDeployment(t *testing.T) *Deployment {
	d, _ := createTestDeployment(t, Config{}, &api.ArangoDeployment{
		Spec: api.DeploymentSpec{
			Mode: api.DeploymentModeCluster.New(),
		},
	})
	d.memberState = memberState.NewStateInspector(d)
	return d
}

func Test_ClusterScaling_InspectCluster_UsesMemberStateClient(t *testing.T) {
	ci := newClusterScalingIntegration(newClusterScalingTestDeployment(t))

	require.ErrorContains(t, ci.inspectCluster(context.Background(), true), "ArangoDB is not reachable")
}

func Test_ClusterScaling_CleanClusterServers_UsesMemberStateClient(t *testing.T) {
	ci := newClusterScalingIntegration(newClusterScalingTestDeployment(t))

	require.ErrorContains(t, ci.cleanClusterServers(context.Background()), "ArangoDB is not reachable")
}

func Test_SetNumberOfServers_UsesMemberStateClient(t *testing.T) {
	d := newClusterScalingTestDeployment(t)

	require.ErrorContains(t, d.SetNumberOfServers(context.Background(), nil, nil), "ArangoDB is not reachable")
}
