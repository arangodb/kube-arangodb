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

	serviceMonitorv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/servicemonitor/v1"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (p serviceMonitorsMod) V1() serviceMonitorv1.ModInterface {
	return serviceMonitorsModV1(p)
}

type serviceMonitorsModV1 struct {
	i *inspectorState
}

func (p serviceMonitorsModV1) client() monitoringv1.ServiceMonitorInterface {
	return p.i.Client().Monitoring().MonitoringV1().ServiceMonitors(p.i.Namespace())
}

func (p serviceMonitorsModV1) Create(ctx context.Context, serviceMonitor *monitoring.ServiceMonitor, opts meta.CreateOptions) (*monitoring.ServiceMonitor, error) {
	if serviceMonitor, err := p.client().Create(ctx, serviceMonitor, opts); err != nil {
		return serviceMonitor, err
	} else {
		p.i.GetThrottles().ServiceMonitor().Invalidate()
		return serviceMonitor, err
	}
}

func (p serviceMonitorsModV1) Update(ctx context.Context, serviceMonitor *monitoring.ServiceMonitor, opts meta.UpdateOptions) (*monitoring.ServiceMonitor, error) {
	if serviceMonitor, err := p.client().Update(ctx, serviceMonitor, opts); err != nil {
		return serviceMonitor, err
	} else {
		p.i.GetThrottles().ServiceMonitor().Invalidate()
		return serviceMonitor, err
	}
}

func (p serviceMonitorsModV1) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result *monitoring.ServiceMonitor, err error) {
	if serviceMonitor, err := p.client().Patch(ctx, name, pt, data, opts, subresources...); err != nil {
		return serviceMonitor, err
	} else {
		p.i.GetThrottles().ServiceMonitor().Invalidate()
		return serviceMonitor, err
	}
}

func (p serviceMonitorsModV1) Delete(ctx context.Context, name string, opts meta.DeleteOptions) error {
	if err := p.client().Delete(ctx, name, opts); err != nil {
		return err
	} else {
		p.i.GetThrottles().ServiceMonitor().Invalidate()
		return err
	}
}
