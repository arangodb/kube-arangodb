//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package inspector

import (
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	monitoring "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringClient "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"

	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	"k8s.io/client-go/kubernetes"
)

func NewInspector(k kubernetes.Interface, m monitoringClient.MonitoringV1Interface, namespace string) (Inspector, error) {
	pods, err := podsToMap(k, namespace)
	if err != nil {
		return nil, err
	}

	secrets, err := secretsToMap(k, namespace)
	if err != nil {
		return nil, err
	}

	pvcs, err := pvcsToMap(k, namespace)
	if err != nil {
		return nil, err
	}

	services, err := servicesToMap(k, namespace)
	if err != nil {
		return nil, err
	}

	serviceAccounts, err := serviceAccountsToMap(k, namespace)
	if err != nil {
		return nil, err
	}

	podDisruptionBudgets, err := podDisruptionBudgetsToMap(k, namespace)
	if err != nil {
		return nil, err
	}

	serviceMonitors, err := serviceMonitorsToMap(m, namespace)
	if err != nil {
		return nil, err
	}

	return NewInspectorFromData(pods, secrets, pvcs, services, serviceAccounts, podDisruptionBudgets, serviceMonitors), nil
}

func NewEmptyInspector() Inspector {
	return NewInspectorFromData(nil, nil, nil, nil, nil, nil, nil)
}

func NewInspectorFromData(pods map[string]*core.Pod,
	secrets map[string]*core.Secret,
	pvcs map[string]*core.PersistentVolumeClaim,
	services map[string]*core.Service,
	serviceAccounts map[string]*core.ServiceAccount,
	podDisruptionBudgets map[string]*policy.PodDisruptionBudget,
	serviceMonitors map[string]*monitoring.ServiceMonitor) Inspector {
	return &inspector{
		pods:                 pods,
		secrets:              secrets,
		pvcs:                 pvcs,
		services:             services,
		serviceAccounts:      serviceAccounts,
		podDisruptionBudgets: podDisruptionBudgets,
		serviceMonitors:      serviceMonitors,
	}
}

type Inspector interface {
	Refresh(k kubernetes.Interface, m monitoringClient.MonitoringV1Interface, namespace string) error

	Pod(name string) (*core.Pod, bool)
	IteratePods(action PodAction, filters ...PodFilter) error

	Secret(name string) (*core.Secret, bool)
	IterateSecrets(action SecretAction, filters ...SecretFilter) error
	SecretReadInterface() k8sutil.SecretReadInterface

	PersistentVolumeClaim(name string) (*core.PersistentVolumeClaim, bool)
	IteratePersistentVolumeClaims(action PersistentVolumeClaimAction, filters ...PersistentVolumeClaimFilter) error

	Service(name string) (*core.Service, bool)
	IterateServices(action ServiceAction, filters ...ServiceFilter) error

	ServiceAccount(name string) (*core.ServiceAccount, bool)
	IterateServiceAccounts(action ServiceAccountAction, filters ...ServiceAccountFilter) error

	PodDisruptionBudget(name string) (*policy.PodDisruptionBudget, bool)
	IteratePodDisruptionBudgets(action PodDisruptionBudgetAction, filters ...PodDisruptionBudgetFilter) error

	ServiceMonitor(name string) (*monitoring.ServiceMonitor, bool)
	IterateServiceMonitors(action ServiceMonitorAction, filters ...ServiceMonitorFilter) error
}

type inspector struct {
	lock sync.Mutex

	pods                 map[string]*core.Pod
	secrets              map[string]*core.Secret
	pvcs                 map[string]*core.PersistentVolumeClaim
	services             map[string]*core.Service
	serviceAccounts      map[string]*core.ServiceAccount
	podDisruptionBudgets map[string]*policy.PodDisruptionBudget
	serviceMonitors      map[string]*monitoring.ServiceMonitor

	ns string
	k  kubernetes.Interface
	m  monitoringClient.MonitoringV1Interface
}

func (i *inspector) Refresh(k kubernetes.Interface, m monitoringClient.MonitoringV1Interface, namespace string) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	pods, err := podsToMap(k, namespace)
	if err != nil {
		return err
	}

	secrets, err := secretsToMap(k, namespace)
	if err != nil {
		return err
	}

	pvcs, err := pvcsToMap(k, namespace)
	if err != nil {
		return err
	}

	services, err := servicesToMap(k, namespace)
	if err != nil {
		return err
	}

	serviceAccounts, err := serviceAccountsToMap(k, namespace)
	if err != nil {
		return err
	}

	podDisruptionBudgets, err := podDisruptionBudgetsToMap(k, namespace)
	if err != nil {
		return err
	}

	serviceMonitors, err := serviceMonitorsToMap(m, namespace)
	if err != nil {
		return err
	}

	i.pods = pods
	i.secrets = secrets
	i.pvcs = pvcs
	i.services = services
	i.serviceAccounts = serviceAccounts
	i.podDisruptionBudgets = podDisruptionBudgets
	i.serviceMonitors = serviceMonitors

	return nil
}
