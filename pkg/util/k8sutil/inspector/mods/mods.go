//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

package mods

import (
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1alpha1"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
)

type PodsMods interface {
	V1() generic.ModClient[*core.Pod]
}

type ServiceAccountsMods interface {
	V1() generic.ModClient[*core.ServiceAccount]
}

type SecretsMods interface {
	V1() generic.ModClient[*core.Secret]
}

type ConfigMapsMods interface {
	V1() generic.ModClient[*core.ConfigMap]
}

type PersistentVolumeClaimsMods interface {
	V1() generic.ModClient[*core.PersistentVolumeClaim]
}

type ServicesMods interface {
	V1() generic.ModClient[*core.Service]
}

type EndpointsMods interface {
	V1() generic.ModClient[*core.Endpoints]
}

type ServiceMonitorsMods interface {
	V1() generic.ModClient[*monitoring.ServiceMonitor]
}

type PodDisruptionBudgetsMods interface {
	V1() generic.ModClient[*policy.PodDisruptionBudget]
}

type ArangoMemberMods interface {
	V1() generic.ModStatusClient[*api.ArangoMember]
}

type ArangoTaskMods interface {
	V1() generic.ModStatusClient[*api.ArangoTask]
}

type ArangoClusterSynchronizationMods interface {
	V1() generic.ModStatusClient[*api.ArangoClusterSynchronization]
}

type ArangoRouteMods interface {
	V1Alpha1() generic.ModStatusClient[*networkingApi.ArangoRoute]
}

type ArangoProfileMods interface {
	V1Beta1() generic.ModStatusClient[*schedulerApi.ArangoProfile]
}

type Mods interface {
	PodsModInterface() PodsMods
	ServiceAccountsModInterface() ServiceAccountsMods
	SecretsModInterface() SecretsMods
	ConfigMapsModInterface() ConfigMapsMods
	PersistentVolumeClaimsModInterface() PersistentVolumeClaimsMods
	ServicesModInterface() ServicesMods
	EndpointsModInterface() EndpointsMods
	ServiceMonitorsModInterface() ServiceMonitorsMods
	PodDisruptionBudgetsModInterface() PodDisruptionBudgetsMods

	ArangoMemberModInterface() ArangoMemberMods
	ArangoTaskModInterface() ArangoTaskMods
	ArangoClusterSynchronizationModInterface() ArangoClusterSynchronizationMods
	ArangoRouteModInterface() ArangoRouteMods
	ArangoProfileModInterface() ArangoProfileMods
}
