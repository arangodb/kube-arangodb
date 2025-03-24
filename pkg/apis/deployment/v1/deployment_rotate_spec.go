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

package v1

import "github.com/arangodb/kube-arangodb/pkg/apis/shared"

type DeploymentRotateSpec struct {
	// Order defines the Rotation order
	// +doc/enum: coordinatorFirst|Runs restart of coordinators before DBServers.
	// +doc/enum: standard|Default restart order.
	Order *DeploymentSpecOrder `json:"order,omitempty"`
}

func (d *DeploymentRotateSpec) Get() DeploymentRotateSpec {
	if d == nil {
		return DeploymentRotateSpec{}
	}

	return *d
}

func (d *DeploymentRotateSpec) GetOrder(def *DeploymentSpecOrder) DeploymentSpecOrder {
	if d == nil || d.Order == nil {
		if def == nil {
			return DeploymentSpecOrderCoordinatorFirst
		}

		return *def
	}

	return *d.Order
}

func (d *DeploymentRotateSpec) Validate() error {
	if d == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceError("order", shared.ValidateOptionalInterface(d.Order)),
	)
}
