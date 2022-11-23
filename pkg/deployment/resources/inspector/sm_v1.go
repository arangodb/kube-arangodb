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

package inspector

import (
	"context"

	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	ins "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/servicemonitor/v1"
)

func (p *serviceMonitorsInspector) V1() (ins.Inspector, error) {
	if p.v1.err != nil {
		return nil, p.v1.err
	}

	return p.v1, nil
}

type serviceMonitorsInspectorV1 struct {
	serviceMonitorInspector *serviceMonitorsInspector

	serviceMonitors map[string]*monitoring.ServiceMonitor
	err             error
}

func (p *serviceMonitorsInspectorV1) validate() error {
	if p == nil {
		return errors.Newf("ServiceMonitorsV1Inspector is nil")
	}

	if p.serviceMonitorInspector == nil {
		return errors.Newf("Parent is nil")
	}

	if p.serviceMonitors == nil && p.err == nil {
		return errors.Newf("ServiceMonitors or err should be not nil")
	}

	if p.serviceMonitors != nil && p.err != nil {
		return errors.Newf("ServiceMonitors or err cannot be not nil together")
	}

	return nil
}

func (p *serviceMonitorsInspectorV1) ServiceMonitors() []*monitoring.ServiceMonitor {
	var r []*monitoring.ServiceMonitor
	for _, serviceMonitor := range p.serviceMonitors {
		r = append(r, serviceMonitor)
	}

	return r
}

func (p *serviceMonitorsInspectorV1) GetSimple(name string) (*monitoring.ServiceMonitor, bool) {
	serviceMonitor, ok := p.serviceMonitors[name]
	if !ok {
		return nil, false
	}

	return serviceMonitor, true
}

func (p *serviceMonitorsInspectorV1) Iterate(action ins.Action, filters ...ins.Filter) error {
	for _, serviceMonitor := range p.serviceMonitors {
		if err := p.iterateServiceMonitor(serviceMonitor, action, filters...); err != nil {
			return err
		}
	}

	return nil
}

func (p *serviceMonitorsInspectorV1) iterateServiceMonitor(serviceMonitor *monitoring.ServiceMonitor, action ins.Action, filters ...ins.Filter) error {
	for _, f := range filters {
		if f == nil {
			continue
		}

		if !f(serviceMonitor) {
			return nil
		}
	}

	return action(serviceMonitor)
}

func (p *serviceMonitorsInspectorV1) Read() ins.ReadInterface {
	return p
}

func (p *serviceMonitorsInspectorV1) Get(ctx context.Context, name string, opts meta.GetOptions) (*monitoring.ServiceMonitor, error) {
	if s, ok := p.GetSimple(name); !ok {
		return nil, apiErrors.NewNotFound(constants.ServiceMonitorGR(), name)
	} else {
		return s, nil
	}
}
