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
	monitoring "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringClient "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"github.com/pkg/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceMonitorFilter func(serviceMonitor *monitoring.ServiceMonitor) bool
type ServiceMonitorAction func(serviceMonitor *monitoring.ServiceMonitor) error

func (i *inspector) IterateServiceMonitors(action ServiceMonitorAction, filters ...ServiceMonitorFilter) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	for _, serviceMonitor := range i.serviceMonitors {
		if err := i.iterateServiceMonitor(serviceMonitor, action, filters...); err != nil {
			return err
		}
	}
	return nil
}

func (i *inspector) iterateServiceMonitor(serviceMonitor *monitoring.ServiceMonitor, action ServiceMonitorAction, filters ...ServiceMonitorFilter) error {
	for _, filter := range filters {
		if !filter(serviceMonitor) {
			return nil
		}
	}

	return action(serviceMonitor)
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

func serviceMonitorsToMap(m monitoringClient.MonitoringV1Interface, namespace string) (map[string]*monitoring.ServiceMonitor, error) {
	serviceMonitors, err := getServiceMonitors(m, namespace, "")
	if err != nil {
		return nil, err
	}

	serviceMonitorMap := map[string]*monitoring.ServiceMonitor{}

	for _, serviceMonitor := range serviceMonitors {
		_, exists := serviceMonitorMap[serviceMonitor.GetName()]
		if exists {
			return nil, errors.Errorf("ServiceMonitor %s already exists in map, error received", serviceMonitor.GetName())
		}

		serviceMonitorMap[serviceMonitor.GetName()] = serviceMonitor
	}

	return serviceMonitorMap, nil
}

func getServiceMonitors(m monitoringClient.MonitoringV1Interface, namespace, cont string) ([]*monitoring.ServiceMonitor, error) {
	serviceMonitors, err := m.ServiceMonitors(namespace).List(meta.ListOptions{
		Limit:    128,
		Continue: cont,
	})

	if err != nil {
		return []*monitoring.ServiceMonitor{}, nil
	}

	return serviceMonitors.Items, nil
}

func FilterServiceMonitorsByLabels(labels map[string]string) ServiceMonitorFilter {
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
