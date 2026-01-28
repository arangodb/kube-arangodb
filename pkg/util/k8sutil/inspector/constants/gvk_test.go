//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package constants

import (
	"testing"

	monitoringApi "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1beta1"
	permissionApi "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
)

func Test_GVK(t *testing.T) {
	testGVK(t, ArangoClusterSynchronizationGKv1(), &api.ArangoClusterSynchronization{})
	testGVK(t, ArangoMemberGKv1(), &api.ArangoMember{})
	testGVK(t, ArangoTaskGKv1(), &api.ArangoTask{})
	testGVK(t, ArangoRouteGKv1Beta1(), &networkingApi.ArangoRoute{})
	testGVK(t, ArangoProfileGKv1Beta1(), &schedulerApi.ArangoProfile{})
	testGVK(t, ArangoTaskGKv1(), &api.ArangoTask{})
	testGVK(t, EndpointsGKv1(), &core.Endpoints{})
	testGVK(t, NodeGKv1(), &core.Node{})
	testGVK(t, PodDisruptionBudgetGKv1(), &policy.PodDisruptionBudget{})
	testGVK(t, PodGKv1(), &core.Pod{})
	testGVK(t, ServiceAccountGKv1(), &core.ServiceAccount{})
	testGVK(t, ServiceGKv1(), &core.Service{})
	testGVK(t, PersistentVolumeClaimGKv1(), &core.PersistentVolumeClaim{})
	testGVK(t, SecretGKv1(), &core.Secret{})
	testGVK(t, ServiceMonitorGKv1(), &monitoringApi.ServiceMonitor{})
	testGVK(t, ArangoDeploymentGKv1(), &api.ArangoDeployment{})
	testGVK(t, ArangoPlatformStorageGKv1Beta1(), &platformApi.ArangoPlatformStorage{})
	testGVK(t, ArangoPermissionTokenGKv1Alpha1(), &permissionApi.ArangoPermissionToken{})
}

func testGVK[T meta.Object](t *testing.T, gvk schema.GroupVersionKind, in T) {
	t.Run(gvk.String(), func(t *testing.T) {
		g, ok := ExtractGVKFromObject(in)
		require.True(t, ok)
		require.Equal(t, gvk, g)
	})
}
