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

package v1

import shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"

type DeploymentUpgradeSpec struct {
	// AutoUpgrade flag specifies if upgrade should be auto-injected, even if is not required (in case of stuck)
	// +doc/default: false
	AutoUpgrade bool `json:"autoUpgrade"`
	// DebugLog flag specifies if containers running upgrade process should print more debugging information.
	// This applies only to init containers.
	// +doc/default: false
	DebugLog bool `json:"debugLog"`
	// Order defines the Upgrade order
	// +doc/enum: standard|Default restart order.
	// +doc/enum: coordinatorFirst|Runs restart of coordinators before DBServers.
	Order *DeploymentSpecOrder `json:"order,omitempty"`
}

func (d *DeploymentUpgradeSpec) Get() DeploymentUpgradeSpec {
	if d == nil {
		return DeploymentUpgradeSpec{}
	}

	return *d
}

func (d *DeploymentUpgradeSpec) GetOrder(def *DeploymentSpecOrder) DeploymentSpecOrder {
	if d == nil || d.Order == nil {
		if def == nil {
			return DeploymentSpecOrderStandard
		}

		return *def
	}

	return *d.Order
}

func (d *DeploymentUpgradeSpec) Validate() error {
	if d == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceError("order", shared.ValidateOptionalInterface(d.Order)),
	)
}
