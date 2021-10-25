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

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/servicemonitor"
	monitoringGroup "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringClient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (i *inspector) IterateServiceMonitors(action servicemonitor.Action, filters ...servicemonitor.Filter) error {
	for _, serviceMonitor := range i.ServiceMonitors() {
		if err := i.iterateServiceMonitor(serviceMonitor, action, filters...); err != nil {
			return err
		}
	}
	return nil
}

func (i *inspector) iterateServiceMonitor(serviceMonitor *monitoring.ServiceMonitor, action servicemonitor.Action, filters ...servicemonitor.Filter) error {
	for _, filter := range filters {
		if !filter(serviceMonitor) {
			return nil
		}
	}

	return action(serviceMonitor)
}

func (i *inspector) ServiceMonitors() []*monitoring.ServiceMonitor {
	i.lock.Lock()
	defer i.lock.Unlock()

	var r []*monitoring.ServiceMonitor
	for _, sms := range i.serviceMonitors {
		r = append(r, sms)
	}

	return r
}

func (i *inspector) ServiceMonitor(name string) (*monitoring.ServiceMonitor, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	serviceMonitor, ok := i.serviceMonitors[name]
	if !ok {
		return nil, false
	}

	return serviceMonitor, true
}

func (i *inspector) ServiceMonitorReadInterface() servicemonitor.ReadInterface {
	return &serviceMonitorReadInterface{i: i}
}

type serviceMonitorReadInterface struct {
	i *inspector
}

func (s serviceMonitorReadInterface) Get(ctx context.Context, name string, opts meta.GetOptions) (*monitoring.ServiceMonitor, error) {
	if s, ok := s.i.ServiceMonitor(name); !ok {
		return nil, apiErrors.NewNotFound(schema.GroupResource{
			Group:    monitoringGroup.GroupName,
			Resource: "servicemonitors",
		}, name)
	} else {
		return s, nil
	}
}

func serviceMonitorsToMap(ctx context.Context, inspector *inspector, m monitoringClient.MonitoringV1Interface, namespace string) func() error {
	return func() error {
		serviceMonitors := getServiceMonitors(ctx, m, namespace, "")

		serviceMonitorMap := map[string]*monitoring.ServiceMonitor{}

		for _, serviceMonitor := range serviceMonitors {
			_, exists := serviceMonitorMap[serviceMonitor.GetName()]
			if exists {
				return errors.Newf("ServiceMonitor %s already exists in map, error received", serviceMonitor.GetName())
			}

			serviceMonitorMap[serviceMonitor.GetName()] = serviceMonitor
		}

		inspector.serviceMonitors = serviceMonitorMap

		return nil
	}
}

func getServiceMonitors(ctx context.Context, m monitoringClient.MonitoringV1Interface, namespace, cont string) []*monitoring.ServiceMonitor {
	ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
	defer cancel()
	serviceMonitors, err := m.ServiceMonitors(namespace).List(ctxChild, meta.ListOptions{
		Limit:    128,
		Continue: cont,
	})

	if err != nil {
		return []*monitoring.ServiceMonitor{}
	}

	return serviceMonitors.Items
}

func FilterServiceMonitorsByLabels(labels map[string]string) servicemonitor.Filter {
	return func(serviceMonitor *monitoring.ServiceMonitor) bool {
		for key, value := range labels {
			v, ok := serviceMonitor.Labels[key]
			if !ok {
				return false
			}

			if v != value {
				return false
			}
		}

		return true
	}
}
