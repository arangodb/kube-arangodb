//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package kclient

import (
	monitoringApi "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
)

func NewModInterface(client Client, namespace string) ModInterface {
	return modInterface{
		client:    client,
		namespace: namespace,
	}
}

type ModInterface interface {
	Secrets() generic.ModClient[*core.Secret]
	Pods() generic.ModClient[*core.Pod]
	Services() generic.ModClient[*core.Service]
	ServiceAccounts() generic.ModClient[*core.ServiceAccount]
	PersistentVolumeClaims() generic.ModClient[*core.PersistentVolumeClaim]
	PodDisruptionBudgets() generic.ModClient[*policy.PodDisruptionBudget]
	ServiceMonitors() generic.ModClient[*monitoringApi.ServiceMonitor]
	ArangoMembers() generic.ModStatusClient[*api.ArangoMember]
}

type modInterface struct {
	client    Client
	namespace string
}

func (m modInterface) PersistentVolumeClaims() generic.ModClient[*core.PersistentVolumeClaim] {
	return m.client.Kubernetes().CoreV1().PersistentVolumeClaims(m.namespace)
}

func (m modInterface) PodDisruptionBudgets() generic.ModClient[*policy.PodDisruptionBudget] {
	return m.client.Kubernetes().PolicyV1().PodDisruptionBudgets(m.namespace)
}

func (m modInterface) ServiceMonitors() generic.ModClient[*monitoringApi.ServiceMonitor] {
	return m.client.Monitoring().MonitoringV1().ServiceMonitors(m.namespace)
}

func (m modInterface) ArangoMembers() generic.ModStatusClient[*api.ArangoMember] {
	return m.client.Arango().DatabaseV1().ArangoMembers(m.namespace)
}

func (m modInterface) Services() generic.ModClient[*core.Service] {
	return m.client.Kubernetes().CoreV1().Services(m.namespace)
}

func (m modInterface) ServiceAccounts() generic.ModClient[*core.ServiceAccount] {
	return m.client.Kubernetes().CoreV1().ServiceAccounts(m.namespace)
}

func (m modInterface) Pods() generic.ModClient[*core.Pod] {
	return m.client.Kubernetes().CoreV1().Pods(m.namespace)
}

func (m modInterface) Secrets() generic.ModClient[*core.Secret] {
	return m.client.Kubernetes().CoreV1().Secrets(m.namespace)
}
