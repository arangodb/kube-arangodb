//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package prometheus

import (
	"context"

	monitoringApi "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/list"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func PrometheusMonitoring(f shared.FactoryGen) {
	f.Register("monitoring", true, shared.WithKubernetesItems[*monitoringApi.ServiceMonitor](kubernetesBackupV1ServiceMonitorList, shared.WithDefinitions[*monitoringApi.ServiceMonitor]))
}

func kubernetesBackupV1ServiceMonitorList(ctx context.Context, client kclient.Client, namespace string) ([]*monitoringApi.ServiceMonitor, error) {
	return list.ListObjects[*monitoringApi.ServiceMonitorList, *monitoringApi.ServiceMonitor](ctx, client.Monitoring().MonitoringV1().ServiceMonitors(namespace), func(result *monitoringApi.ServiceMonitorList) []*monitoringApi.ServiceMonitor {
		q := make([]*monitoringApi.ServiceMonitor, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}
