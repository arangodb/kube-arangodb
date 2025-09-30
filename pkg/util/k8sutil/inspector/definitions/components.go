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

package definitions

type ComponentCount map[Component]int

type Component string

const (
	ArangoClusterSynchronization Component = "ArangoClusterSynchronization"
	ArangoMember                 Component = "ArangoMember"
	ArangoTask                   Component = "ArangoTask"
	ArangoRoute                  Component = "ArangoRoute"
	ArangoProfile                Component = "ArangoProfile"
	ArangoPlatformStorage        Component = "ArangoPlatformStorage"
	ArangoPlatformService        Component = "ArangoPlatformService"
	Node                         Component = "Node"
	PersistentVolume             Component = "PersistentVolume"
	PersistentVolumeClaim        Component = "PersistentVolumeClaim"
	Pod                          Component = "Pod"
	PodDisruptionBudget          Component = "PodDisruptionBudget"
	Secret                       Component = "Secret"
	ConfigMap                    Component = "ConfigMap"
	Service                      Component = "Service"
	ServiceAccount               Component = "ServiceAccount"
	ServiceMonitor               Component = "ServiceMonitor"
	Endpoints                    Component = "Endpoints"
)

func AllComponents() []Component {
	return []Component{
		ArangoClusterSynchronization,
		ArangoMember,
		ArangoTask,
		ArangoRoute,
		ArangoProfile,
		ArangoPlatformStorage,
		ArangoPlatformService,
		Node,
		PersistentVolume,
		PersistentVolumeClaim,
		Pod,
		PodDisruptionBudget,
		Secret,
		ConfigMap,
		Service,
		ServiceAccount,
		ServiceMonitor,
		Endpoints,
	}
}
