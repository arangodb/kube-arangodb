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

package arangod

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func newDatabaseClientTestDeployment(mode api.DeploymentMode, group api.ServerGroup, ready map[string]bool) *api.ArangoDeployment {
	depl := &api.ArangoDeployment{
		ObjectMeta: meta.ObjectMeta{
			Name:      "deployment",
			Namespace: "namespace",
		},
		Spec: api.DeploymentSpec{
			Mode: mode.New(),
		},
	}

	members := api.MemberStatusList{}
	for _, id := range []string{"member-a", "member-b", "member-c"} {
		isReady, ok := ready[id]
		if !ok {
			continue
		}

		member := api.MemberStatus{ID: id}
		member.Conditions.Update(api.ConditionTypeReady, isReady, "", "")
		members = append(members, member)
	}

	switch group {
	case api.ServerGroupCoordinators:
		depl.Status.Members.Coordinators = members
	case api.ServerGroupSingle:
		depl.Status.Members.Single = members
	}

	return depl
}

func newDatabaseClientTestServices(t *testing.T, depl *api.ArangoDeployment, group api.ServerGroup) *fake.Clientset {
	client := fake.NewSimpleClientset()

	for _, member := range depl.Status.Members.MembersOfGroup(group) {
		_, err := client.CoreV1().Services(depl.GetNamespace()).Create(context.Background(), &core.Service{
			ObjectMeta: meta.ObjectMeta{
				Name:      member.ArangoMemberName(depl.GetName(), group),
				Namespace: depl.GetNamespace(),
			},
			Spec: core.ServiceSpec{
				ClusterIP: "10.0.0.1",
			},
		}, meta.CreateOptions{})
		require.NoError(t, err)
	}

	return client
}

func Test_DatabaseClientEndpoints_Cluster_DialsReadyCoordinatorsDirectly(t *testing.T) {
	depl := newDatabaseClientTestDeployment(api.DeploymentModeCluster, api.ServerGroupCoordinators, map[string]bool{
		"member-a": true,
		"member-b": false,
		"member-c": true,
	})
	client := newDatabaseClientTestServices(t, depl, api.ServerGroupCoordinators)

	endpoints, err := databaseClientEndpoints(context.Background(), client.CoreV1(), depl)
	require.NoError(t, err)

	require.Equal(t, []string{
		k8sutil.CreatePodDNSName(depl, api.ServerGroupCoordinators.AsRole(), "member-a"),
		k8sutil.CreatePodDNSName(depl, api.ServerGroupCoordinators.AsRole(), "member-c"),
	}, endpoints)
	require.NotContains(t, endpoints, k8sutil.CreateDatabaseClientServiceDNSName(depl))
}

func Test_DatabaseClientEndpoints_Cluster_FollowsCommunicationMethod(t *testing.T) {
	depl := newDatabaseClientTestDeployment(api.DeploymentModeCluster, api.ServerGroupCoordinators, map[string]bool{
		"member-a": true,
	})
	depl.Spec.CommunicationMethod = api.DeploymentCommunicationMethodDNS.New()
	client := newDatabaseClientTestServices(t, depl, api.ServerGroupCoordinators)

	endpoints, err := databaseClientEndpoints(context.Background(), client.CoreV1(), depl)
	require.NoError(t, err)

	memberServiceName := depl.Status.Members.Coordinators[0].ArangoMemberName(depl.GetName(), api.ServerGroupCoordinators)
	require.Equal(t, []string{memberServiceName + ".namespace.svc"}, endpoints)
}

func Test_DatabaseClientEndpoints_Single_DialsMemberDirectly(t *testing.T) {
	depl := newDatabaseClientTestDeployment(api.DeploymentModeSingle, api.ServerGroupSingle, map[string]bool{
		"member-a": true,
	})
	client := newDatabaseClientTestServices(t, depl, api.ServerGroupSingle)

	endpoints, err := databaseClientEndpoints(context.Background(), client.CoreV1(), depl)
	require.NoError(t, err)

	require.Equal(t, []string{k8sutil.CreatePodDNSName(depl, api.ServerGroupSingle.AsRole(), "member-a")}, endpoints)
}

func Test_DatabaseClientEndpoints_ActiveFailover_KeepsLeaderService(t *testing.T) {
	depl := newDatabaseClientTestDeployment(api.DeploymentModeActiveFailover, api.ServerGroupSingle, map[string]bool{
		"member-a": true,
		"member-b": true,
	})

	endpoints, err := databaseClientEndpoints(context.Background(), fake.NewSimpleClientset().CoreV1(), depl)
	require.NoError(t, err)

	require.Equal(t, []string{k8sutil.CreateDatabaseClientServiceDNSName(depl)}, endpoints)
}

func Test_DatabaseClientEndpoints_NoReadyMember(t *testing.T) {
	depl := newDatabaseClientTestDeployment(api.DeploymentModeCluster, api.ServerGroupCoordinators, map[string]bool{
		"member-a": false,
	})
	client := newDatabaseClientTestServices(t, depl, api.ServerGroupCoordinators)

	_, err := databaseClientEndpoints(context.Background(), client.CoreV1(), depl)
	require.EqualError(t, err, "No ready coordinator member to connect to")
}

func Test_SharedHTTPTransports_AreReused(t *testing.T) {
	require.Same(t, sharedHTTPTransport(), sharedHTTPTransport())
	require.Same(t, sharedHTTPSTransport(), sharedHTTPSTransport())
	require.Same(t, sharedHTTPTransportShortTimeout(), sharedHTTPTransportShortTimeout())
	require.Same(t, sharedHTTPSTransportShortTimeout(), sharedHTTPSTransportShortTimeout())
}
