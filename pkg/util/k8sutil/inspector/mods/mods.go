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

package mods

import (
	arangoclustersynchronizationv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangoclustersynchronization/v1"
	arangomemberv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangomember/v1"
	arangotaskv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangotask/v1"
	endpointsv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/endpoints/v1"
	persistentvolumeclaimv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim/v1"
	podv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod/v1"
	poddisruptionbudgetv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/poddisruptionbudget/v1"
	secretv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret/v1"
	servicev1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service/v1"
	serviceaccountv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/serviceaccount/v1"
	servicemonitorv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/servicemonitor/v1"
)

type PodsMods interface {
	V1() podv1.ModInterface
}

type ServiceAccountsMods interface {
	V1() serviceaccountv1.ModInterface
}

type SecretsMods interface {
	V1() secretv1.ModInterface
}

type PersistentVolumeClaimsMods interface {
	V1() persistentvolumeclaimv1.ModInterface
}

type ServicesMods interface {
	V1() servicev1.ModInterface
}

type EndpointsMods interface {
	V1() endpointsv1.ModInterface
}

type ServiceMonitorsMods interface {
	V1() servicemonitorv1.ModInterface
}

type PodDisruptionBudgetsMods interface {
	V1() poddisruptionbudgetv1.ModInterface
}

type ArangoMemberMods interface {
	V1() arangomemberv1.ModInterface
}

type ArangoTaskMods interface {
	V1() arangotaskv1.ModInterface
}

type ArangoClusterSynchronizationMods interface {
	V1() arangoclustersynchronizationv1.ModInterface
}

type Mods interface {
	PodsModInterface() PodsMods
	ServiceAccountsModInterface() ServiceAccountsMods
	SecretsModInterface() SecretsMods
	PersistentVolumeClaimsModInterface() PersistentVolumeClaimsMods
	ServicesModInterface() ServicesMods
	EndpointsModInterface() EndpointsMods
	ServiceMonitorsModInterface() ServiceMonitorsMods
	PodDisruptionBudgetsModInterface() PodDisruptionBudgetsMods

	ArangoMemberModInterface() ArangoMemberMods
	ArangoTaskModInterface() ArangoTaskMods
	ArangoClusterSynchronizationModInterface() ArangoClusterSynchronizationMods
}
