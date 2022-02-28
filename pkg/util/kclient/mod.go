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

package kclient

import (
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangomember"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/poddisruptionbudget"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/serviceaccount"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/servicemonitor"
)

func NewModInterface(client Client, namespace string) ModInterface {
	return modInterface{
		client:    client,
		namespace: namespace,
	}
}

type ModInterface interface {
	Secrets() secret.ModInterface
	Pods() pod.ModInterface
	Services() service.ModInterface
	ServiceAccounts() serviceaccount.ModInterface
	PersistentVolumeClaims() persistentvolumeclaim.ModInterface
	PodDisruptionBudgets() poddisruptionbudget.ModInterface
	ServiceMonitors() servicemonitor.ModInterface
	ArangoMembers() arangomember.ModInterface
}

type modInterface struct {
	client    Client
	namespace string
}

func (m modInterface) PersistentVolumeClaims() persistentvolumeclaim.ModInterface {
	return m.client.Kubernetes().CoreV1().PersistentVolumeClaims(m.namespace)
}

func (m modInterface) PodDisruptionBudgets() poddisruptionbudget.ModInterface {
	return m.client.Kubernetes().PolicyV1beta1().PodDisruptionBudgets(m.namespace)
}

func (m modInterface) ServiceMonitors() servicemonitor.ModInterface {
	return m.client.Monitoring().MonitoringV1().ServiceMonitors(m.namespace)
}

func (m modInterface) ArangoMembers() arangomember.ModInterface {
	return m.client.Arango().DatabaseV1().ArangoMembers(m.namespace)
}

func (m modInterface) Services() service.ModInterface {
	return m.client.Kubernetes().CoreV1().Services(m.namespace)
}

func (m modInterface) ServiceAccounts() serviceaccount.ModInterface {
	return m.client.Kubernetes().CoreV1().ServiceAccounts(m.namespace)
}

func (m modInterface) Pods() pod.ModInterface {
	return m.client.Kubernetes().CoreV1().Pods(m.namespace)
}

func (m modInterface) Secrets() secret.ModInterface {
	return m.client.Kubernetes().CoreV1().Secrets(m.namespace)
}
