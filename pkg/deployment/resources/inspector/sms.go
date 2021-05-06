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

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/servicemonitor"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringClient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (i *inspector) IterateServiceMonitors(action servicemonitor.ServiceMonitorAction, filters ...servicemonitor.ServiceMonitorFilter) error {
	for _, serviceMonitor := range i.ServiceMonitors() {
		if err := i.iterateServiceMonitor(serviceMonitor, action, filters...); err != nil {
			return err
		}
	}
	return nil
}

func (i *inspector) iterateServiceMonitor(serviceMonitor *monitoring.ServiceMonitor, action servicemonitor.ServiceMonitorAction, filters ...servicemonitor.ServiceMonitorFilter) error {
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

func serviceMonitorsToMap(ctx context.Context, m monitoringClient.MonitoringV1Interface, namespace string) (map[string]*monitoring.ServiceMonitor, error) {
	serviceMonitors := getServiceMonitors(ctx, m, namespace, "")

	serviceMonitorMap := map[string]*monitoring.ServiceMonitor{}

	for _, serviceMonitor := range serviceMonitors {
		_, exists := serviceMonitorMap[serviceMonitor.GetName()]
		if exists {
			return nil, errors.Newf("ServiceMonitor %s already exists in map, error received", serviceMonitor.GetName())
		}

		serviceMonitorMap[serviceMonitor.GetName()] = serviceMonitor
	}

	return serviceMonitorMap, nil
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

func FilterServiceMonitorsByLabels(labels map[string]string) servicemonitor.ServiceMonitorFilter {
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
