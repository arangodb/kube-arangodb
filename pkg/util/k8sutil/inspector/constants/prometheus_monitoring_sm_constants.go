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

package constants

import (
	monitoringApi "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ServiceMonitor
const (
	ServiceMonitorGroup     = "monitoring.coreos.com"
	ServiceMonitorResource  = "servicemonitors"
	ServiceMonitorKind      = "ServiceMonitor"
	ServiceMonitorVersionV1 = "v1"
)

func init() {
	register[*monitoringApi.ServiceMonitor](ServiceMonitorGKv1(), ServiceMonitorGRv1())
}

func ServiceMonitorGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ServiceMonitorGroup,
		Kind:  ServiceMonitorKind,
	}
}

func ServiceMonitorGKv1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ServiceMonitorGroup,
		Kind:    ServiceMonitorKind,
		Version: ServiceMonitorVersionV1,
	}
}

func ServiceMonitorGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ServiceMonitorGroup,
		Resource: ServiceMonitorResource,
	}
}

func ServiceMonitorGRv1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ServiceMonitorGroup,
		Resource: ServiceMonitorResource,
		Version:  ServiceMonitorVersionV1,
	}
}
