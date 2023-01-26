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
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/mods"
	servicemonitorv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/servicemonitor/v1"
)

func (i *inspectorState) ServiceMonitorsModInterface() mods.ServiceMonitorsMods {
	return serviceMonitorsMod{
		i: i,
	}
}

type serviceMonitorsMod struct {
	i *inspectorState
}

func (p serviceMonitorsMod) V1() servicemonitorv1.ModInterface {
	return wrapMod[*monitoring.ServiceMonitor](definitions.ServiceMonitor, p.i.GetThrottles, generic.WithModStatusGetter[*monitoring.ServiceMonitor](constants.ServiceMonitorGKv1(), p.clientv1))
}

func (p serviceMonitorsMod) clientv1() generic.ModClient[*monitoring.ServiceMonitor] {
	return p.i.Client().Monitoring().MonitoringV1().ServiceMonitors(p.i.Namespace())
}
