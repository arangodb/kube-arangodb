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

package inspector

import (
	"context"
	"time"

	monitoringApi "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/servicemonitor"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/version"
)

func init() {
	requireRegisterInspectorLoader(serviceMonitorsInspectorLoaderObj)
}

var serviceMonitorsInspectorLoaderObj = serviceMonitorsInspectorLoader{}

type serviceMonitorsInspectorLoader struct {
}

func (p serviceMonitorsInspectorLoader) Component() definitions.Component {
	return definitions.ServiceMonitor
}

func (p serviceMonitorsInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q serviceMonitorsInspector

	q.v1 = newInspectorVersion[*monitoringApi.ServiceMonitorList, *monitoringApi.ServiceMonitor](ctx,
		constants.ServiceMonitorGRv1(),
		constants.ServiceMonitorGKv1(),
		i.client.Monitoring().MonitoringV1().ServiceMonitors(i.namespace),
		servicemonitor.List())

	i.serviceMonitors = &q
	q.state = i
	q.last = time.Now()
}

func (p serviceMonitorsInspectorLoader) Verify(i *inspectorState) error {
	return nil
}

func (p serviceMonitorsInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.serviceMonitors != nil {
		if !override {
			return
		}
	}

	to.serviceMonitors = from.serviceMonitors
	to.serviceMonitors.state = to
}

func (p serviceMonitorsInspectorLoader) Name() string {
	return "serviceMonitors"
}

type serviceMonitorsInspector struct {
	state *inspectorState

	last time.Time

	v1 *inspectorVersion[*monitoringApi.ServiceMonitor]
}

func (p *serviceMonitorsInspector) LastRefresh() time.Time {
	return p.last
}

func (p *serviceMonitorsInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, serviceMonitorsInspectorLoaderObj)
}

func (p *serviceMonitorsInspector) Version() version.Version {
	return version.V1
}

func (p *serviceMonitorsInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.ServiceMonitor()
}

func (p *serviceMonitorsInspector) validate() error {
	if p == nil {
		return errors.Errorf("ServiceMonitorInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1.validate()
}

func (p *serviceMonitorsInspector) V1() (generic.Inspector[*monitoringApi.ServiceMonitor], error) {
	if p.v1.err != nil {
		return nil, p.v1.err
	}

	return p.v1, nil
}
