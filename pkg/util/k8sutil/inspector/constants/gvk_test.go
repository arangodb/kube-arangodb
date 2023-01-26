//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	"reflect"
	"testing"

	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func Test_GVK(t *testing.T) {
	testGVK(t, ArangoClusterSynchronizationGKv1(), &api.ArangoClusterSynchronization{}, api.ArangoClusterSynchronization{})
	testGVK(t, ArangoMemberGKv1(), &api.ArangoMember{}, api.ArangoMember{})
	testGVK(t, ArangoTaskGKv1(), &api.ArangoTask{}, api.ArangoTask{})
	testGVK(t, EndpointsGKv1(), &core.Endpoints{}, core.Endpoints{})
	testGVK(t, NodeGKv1(), &core.Node{}, core.Node{})
	testGVK(t, PodDisruptionBudgetGKv1(), &policy.PodDisruptionBudget{}, policy.PodDisruptionBudget{})
	testGVK(t, PodGKv1(), &core.Pod{}, core.Pod{})
	testGVK(t, ServiceAccountGKv1(), &core.ServiceAccount{}, core.ServiceAccount{})
	testGVK(t, ServiceGKv1(), &core.Service{}, core.Service{})
	testGVK(t, PersistentVolumeClaimGKv1(), &core.PersistentVolumeClaim{}, core.PersistentVolumeClaim{})
	testGVK(t, SecretGKv1(), &core.Secret{}, core.Secret{})
	testGVK(t, ServiceMonitorGKv1(), &monitoring.ServiceMonitor{}, monitoring.ServiceMonitor{})
	testGVK(t, ArangoDeploymentGKv1(), &api.ArangoDeployment{}, api.ArangoDeployment{})
}

func testGVK(t *testing.T, gvk schema.GroupVersionKind, in ...interface{}) {
	t.Run(gvk.String(), func(t *testing.T) {
		for _, z := range in {
			zt := reflect.TypeOf(z)
			t.Run(zt.String(), func(t *testing.T) {
				g, ok := ExtractGVKFromObject(z)
				require.True(t, ok)
				require.Equal(t, gvk, g)
			})
		}
	})
}
