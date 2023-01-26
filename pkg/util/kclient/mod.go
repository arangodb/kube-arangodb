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

package kclient

import (
	arangomemberv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangomember/v1"
	persistentvolumeclaimv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim/v1"
	podv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod/v1"
	poddisruptionbudgetv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/poddisruptionbudget/v1"
	secretv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret/v1"
	servicev1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service/v1"
	serviceaccountv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/serviceaccount/v1"
	servicemonitorv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/servicemonitor/v1"
)

func NewModInterface(client Client, namespace string) ModInterface {
	return modInterface{
		client:    client,
		namespace: namespace,
	}
}

type ModInterface interface {
	Secrets() secretv1.ModInterface
	Pods() podv1.ModInterface
	Services() servicev1.ModInterface
	ServiceAccounts() serviceaccountv1.ModInterface
	PersistentVolumeClaims() persistentvolumeclaimv1.ModInterface
	PodDisruptionBudgets() poddisruptionbudgetv1.ModInterface
	ServiceMonitors() servicemonitorv1.ModInterface
	ArangoMembers() arangomemberv1.ModInterface
}

type modInterface struct {
	client    Client
	namespace string
}

func (m modInterface) PersistentVolumeClaims() persistentvolumeclaimv1.ModInterface {
	return m.client.Kubernetes().CoreV1().PersistentVolumeClaims(m.namespace)
}

func (m modInterface) PodDisruptionBudgets() poddisruptionbudgetv1.ModInterface {
	return m.client.Kubernetes().PolicyV1().PodDisruptionBudgets(m.namespace)
}

func (m modInterface) ServiceMonitors() servicemonitorv1.ModInterface {
	return m.client.Monitoring().MonitoringV1().ServiceMonitors(m.namespace)
}

func (m modInterface) ArangoMembers() arangomemberv1.ModInterface {
	return m.client.Arango().DatabaseV1().ArangoMembers(m.namespace)
}

func (m modInterface) Services() servicev1.ModInterface {
	return m.client.Kubernetes().CoreV1().Services(m.namespace)
}

func (m modInterface) ServiceAccounts() serviceaccountv1.ModInterface {
	return m.client.Kubernetes().CoreV1().ServiceAccounts(m.namespace)
}

func (m modInterface) Pods() podv1.ModInterface {
	return m.client.Kubernetes().CoreV1().Pods(m.namespace)
}

func (m modInterface) Secrets() secretv1.ModInterface {
	return m.client.Kubernetes().CoreV1().Secrets(m.namespace)
}
