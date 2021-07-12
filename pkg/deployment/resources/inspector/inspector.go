//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Tomasz Mielech
//

package inspector

import (
	"context"
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"

	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringClient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"

	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	"k8s.io/client-go/kubernetes"
)

// SecretReadInterface has methods to work with Secret resources with ReadOnly mode.
type SecretReadInterface interface {
	Get(ctx context.Context, name string, opts meta.GetOptions) (*core.Secret, error)
}

func NewInspector(ctx context.Context, k kubernetes.Interface, m monitoringClient.MonitoringV1Interface, c versioned.Interface, namespace string) (inspectorInterface.Inspector, error) {
	i := &inspector{
		namespace: namespace,
		k:         k,
		m:         m,
		c:         c,
	}

	if err := i.Refresh(ctx); err != nil {
		return nil, err
	}

	return i, nil
}

func NewEmptyInspector() inspectorInterface.Inspector {
	return NewInspectorFromData(nil, nil, nil, nil, nil, nil, nil, nil)
}

func NewInspectorFromData(pods map[string]*core.Pod,
	secrets map[string]*core.Secret,
	pvcs map[string]*core.PersistentVolumeClaim,
	services map[string]*core.Service,
	serviceAccounts map[string]*core.ServiceAccount,
	podDisruptionBudgets map[string]*policy.PodDisruptionBudget,
	serviceMonitors map[string]*monitoring.ServiceMonitor,
	arangoMembers map[string]*api.ArangoMember) inspectorInterface.Inspector {
	return &inspector{
		pods:                 pods,
		secrets:              secrets,
		pvcs:                 pvcs,
		services:             services,
		serviceAccounts:      serviceAccounts,
		podDisruptionBudgets: podDisruptionBudgets,
		serviceMonitors:      serviceMonitors,
		arangoMembers:        arangoMembers,
	}
}

type inspector struct {
	lock sync.Mutex

	namespace string

	k kubernetes.Interface
	m monitoringClient.MonitoringV1Interface
	c versioned.Interface

	pods                 map[string]*core.Pod
	secrets              map[string]*core.Secret
	pvcs                 map[string]*core.PersistentVolumeClaim
	services             map[string]*core.Service
	serviceAccounts      map[string]*core.ServiceAccount
	podDisruptionBudgets map[string]*policy.PodDisruptionBudget
	serviceMonitors      map[string]*monitoring.ServiceMonitor
	arangoMembers        map[string]*api.ArangoMember
}

func (i *inspector) Refresh(ctx context.Context) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.namespace == "" {
		return errors.New("Inspector created fro mstatic data")
	}

	pods, err := podsToMap(ctx, i.k, i.namespace)
	if err != nil {
		return err
	}

	secrets, err := secretsToMap(ctx, i.k, i.namespace)
	if err != nil {
		return err
	}

	pvcs, err := pvcsToMap(ctx, i.k, i.namespace)
	if err != nil {
		return err
	}

	services, err := servicesToMap(ctx, i.k, i.namespace)
	if err != nil {
		return err
	}

	serviceAccounts, err := serviceAccountsToMap(ctx, i.k, i.namespace)
	if err != nil {
		return err
	}

	podDisruptionBudgets, err := podDisruptionBudgetsToMap(ctx, i.k, i.namespace)
	if err != nil {
		return err
	}

	serviceMonitors, err := serviceMonitorsToMap(ctx, i.m, i.namespace)
	if err != nil {
		return err
	}

	arangoMembers, err := arangoMembersToMap(ctx, i.c, i.namespace)
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
	i.arangoMembers = arangoMembers

	return nil
}
