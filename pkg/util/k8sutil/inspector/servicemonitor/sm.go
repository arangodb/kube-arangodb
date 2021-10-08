//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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

package servicemonitor

import monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

type Inspector interface {
	ServiceMonitor(name string) (*monitoring.ServiceMonitor, bool)
	IterateServiceMonitors(action Action, filters ...Filter) error
	ServiceMonitorReadInterface() ReadInterface
}

type Filter func(serviceMonitor *monitoring.ServiceMonitor) bool
type Action func(serviceMonitor *monitoring.ServiceMonitor) error
