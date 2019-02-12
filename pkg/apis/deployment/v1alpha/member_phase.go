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

// MemberPhase is a strongly typed lifetime phase of a deployment member
type MemberPhase string

const (
	// MemberPhaseNone indicates that the state is not set yet
	MemberPhaseNone MemberPhase = ""
	// MemberPhaseCreated indicates that all resources needed for the member have been created
	MemberPhaseCreated MemberPhase = "Created"
	// MemberPhaseFailed indicates that the member is gone beyond hope of recovery. It must be replaced with a new member.
	MemberPhaseFailed MemberPhase = "Failed"
	// MemberPhaseCleanOut indicates that a dbserver is in the process of being cleaned out
	MemberPhaseCleanOut MemberPhase = "CleanOut"
	// MemberPhaseDrain indicates that a dbserver is in the process of being cleaned out as result of draining a node
	MemberPhaseDrain MemberPhase = "Drain"
	// MemberPhaseShuttingDown indicates that a member is shutting down
	MemberPhaseShuttingDown MemberPhase = "ShuttingDown"
	// MemberPhaseRotating indicates that a member is being rotated
	MemberPhaseRotating MemberPhase = "Rotating"
	// MemberPhaseUpgrading indicates that a member is in the process of upgrading its database data format
	MemberPhaseUpgrading MemberPhase = "Upgrading"
)

// IsFailed returns true when given phase == "Failed"
func (p MemberPhase) IsFailed() bool {
	return p == MemberPhaseFailed
}

// IsCreatedOrDrain returns true when given phase is MemberPhaseCreated or MemberPhaseDrain
func (p MemberPhase) IsCreatedOrDrain() bool {
	return p == MemberPhaseCreated || p == MemberPhaseDrain
}
