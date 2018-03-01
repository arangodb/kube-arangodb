//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package v1alpha

// DeploymentState is a strongly typed state of a deployment
type DeploymentState string

const (
	// DeploymentStateNone indicates that the state is not set yet
	DeploymentStateNone DeploymentState = ""
	// DeploymentStateCreating indicates that the deployment is being created
	DeploymentStateCreating DeploymentState = "Creating"
	// DeploymentStateRunning indicates that all servers are running and reachable
	DeploymentStateRunning DeploymentState = "Running"
	// DeploymentStateScaling indicates that servers are being added or removed
	DeploymentStateScaling DeploymentState = "Scaling"
	// DeploymentStateUpgrading indicates that a version upgrade is in progress
	DeploymentStateUpgrading DeploymentState = "Upgrading"
	// DeploymentStateFailed indicates that a deployment is in a failed state
	// from which automatic recovery is impossible. Inspect `Reason` for more info.
	DeploymentStateFailed DeploymentState = "Failed"
)

// IsFailed returns true if given state is DeploymentStateFailed
func (cs DeploymentState) IsFailed() bool {
	return cs == DeploymentStateFailed
}
