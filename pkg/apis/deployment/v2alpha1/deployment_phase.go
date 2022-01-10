//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package v2alpha1

// DeploymentPhase is a strongly typed lifetime phase of a deployment
type DeploymentPhase string

const (
	// DeploymentPhaseNone indicates that the phase is not set yet
	DeploymentPhaseNone DeploymentPhase = ""
	// DeploymentPhaseRunning indicates that the deployment is under control of the
	// ArangoDeployment operator.
	DeploymentPhaseRunning DeploymentPhase = "Running"
	// DeploymentPhaseFailed indicates that a deployment is in a failed state
	// from which automatic recovery is impossible. Inspect `Reason` for more info.
	DeploymentPhaseFailed DeploymentPhase = "Failed"
)

// IsFailed returns true if given state is DeploymentStateFailed
func (cs DeploymentPhase) IsFailed() bool {
	return cs == DeploymentPhaseFailed
}
