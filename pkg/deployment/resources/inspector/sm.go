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

package inspector

import (
	"context"
	"time"

	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
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
	p.loadV1(ctx, i, &q)
	i.serviceMonitors = &q
	q.state = i
	q.last = time.Now()
}

func (p serviceMonitorsInspectorLoader) loadV1(ctx context.Context, i *inspectorState, q *serviceMonitorsInspector) {
	var z serviceMonitorsInspectorV1

	z.serviceMonitorInspector = q

	z.serviceMonitors, z.err = p.getV1ServiceMonitors(ctx, i)

	q.v1 = &z
}

func (p serviceMonitorsInspectorLoader) getV1ServiceMonitors(ctx context.Context, i *inspectorState) (map[string]*monitoring.ServiceMonitor, error) {
	objs, err := p.getV1ServiceMonitorsList(ctx, i)
	if err != nil {
		return nil, err
	}

	r := make(map[string]*monitoring.ServiceMonitor, len(objs))

	for id := range objs {
		r[objs[id].GetName()] = objs[id]
	}

	return r, nil
}

func (p serviceMonitorsInspectorLoader) getV1ServiceMonitorsList(ctx context.Context, i *inspectorState) ([]*monitoring.ServiceMonitor, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Monitoring().MonitoringV1().ServiceMonitors(i.namespace).List(ctxChild, meta.ListOptions{
		Limit: globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
	})

	if err != nil {
		return nil, err
	}

	items := obj.Items
	cont := obj.Continue
	var s = int64(len(items))

	if z := obj.RemainingItemCount; z != nil {
		s += *z
	}

	ptrs := make([]*monitoring.ServiceMonitor, 0, s)

	for {
		ptrs = append(ptrs, items...)
		if cont == "" {
			break
		}

		items, cont, err = p.getV1ServiceMonitorsListRequest(ctx, i, cont)
		if err != nil {
			return nil, err
		}
	}

	return ptrs, nil
}

func (p serviceMonitorsInspectorLoader) getV1ServiceMonitorsListRequest(ctx context.Context, i *inspectorState, cont string) ([]*monitoring.ServiceMonitor, string, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Monitoring().MonitoringV1().ServiceMonitors(i.namespace).List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, "", err
	}

	return obj.Items, obj.Continue, err
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

	v1 *serviceMonitorsInspectorV1
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
		return errors.Newf("ServiceMonitorInspector is nil")
	}

	if p.state == nil {
		return errors.Newf("Parent is nil")
	}

	return p.v1.validate()
}
