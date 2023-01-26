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
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func ExtractGVKFromObject(in interface{}) (schema.GroupVersionKind, bool) {
	if in != nil {
		switch in.(type) {
		case *api.ArangoClusterSynchronization, api.ArangoClusterSynchronization:
			return ArangoClusterSynchronizationGKv1(), true
		case *api.ArangoMember, api.ArangoMember:
			return ArangoMemberGKv1(), true
		case *api.ArangoTask, api.ArangoTask:
			return ArangoTaskGKv1(), true
		case *core.Endpoints, core.Endpoints:
			return EndpointsGKv1(), true
		case *core.Node, core.Node:
			return NodeGKv1(), true
		case *policy.PodDisruptionBudget, policy.PodDisruptionBudget:
			return PodDisruptionBudgetGKv1(), true
		case *core.Pod, core.Pod:
			return PodGKv1(), true
		case *core.ServiceAccount, core.ServiceAccount:
			return ServiceAccountGKv1(), true
		case *core.PersistentVolumeClaim, core.PersistentVolumeClaim:
			return PersistentVolumeClaimGKv1(), true
		case *core.Secret, core.Secret:
			return SecretGKv1(), true
		case *core.Service, core.Service:
			return ServiceGKv1(), true
		case *monitoring.ServiceMonitor, monitoring.ServiceMonitor:
			return ServiceMonitorGKv1(), true
		case *api.ArangoDeployment, api.ArangoDeployment:
			return ArangoDeploymentGKv1(), true
		}
	}

	return schema.GroupVersionKind{}, false
}
