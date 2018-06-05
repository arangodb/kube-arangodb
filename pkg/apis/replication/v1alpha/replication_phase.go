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

// DeploymentReplicationPhase is a strongly typed lifetime phase of a deployment replication
type DeploymentReplicationPhase string

const (
	// DeploymentReplicationPhaseNone indicates that the phase is not set yet
	DeploymentReplicationPhaseNone DeploymentReplicationPhase = ""
	// DeploymentReplicationPhaseRunning indicates that the deployment replication is under control of the
	// ArangoDeploymentReplication operator.
	DeploymentReplicationPhaseRunning DeploymentReplicationPhase = "Running"
	// DeploymentReplicationPhaseFailed indicates that a deployment replication is in a failed state
	// from which automatic recovery is impossible. Inspect `Reason` for more info.
	DeploymentReplicationPhaseFailed DeploymentReplicationPhase = "Failed"
)

// IsFailed returns true if given state is DeploymentStateFailed
func (cs DeploymentReplicationPhase) IsFailed() bool {
	return cs == DeploymentReplicationPhaseFailed
}
