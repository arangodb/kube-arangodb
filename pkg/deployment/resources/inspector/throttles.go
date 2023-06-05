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
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
)

func NewDefaultThrottle() throttle.Components {
	return throttle.NewThrottleComponents(
		30*time.Second, // ArangoDeploymentSynchronization
		30*time.Second, // ArangoMember
		30*time.Second, // ArangoTask
		30*time.Second, // Node
		30*time.Second, // PV
		15*time.Second, // PVC
		time.Second,    // Pod
		30*time.Second, // PDB
		10*time.Second, // Secret
		10*time.Second, // Service
		30*time.Second, // SA
		30*time.Second, // ServiceMonitor
		15*time.Second) // Endpoints
}
