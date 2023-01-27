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

package v2alpha1

// MemberPhase is a strongly typed lifetime phase of a deployment member
type MemberPhase string

const (
	// MemberPhaseNone indicates that the state is not set yet
	MemberPhaseNone MemberPhase = ""
	// MemberPhasePending indicates that member propagation has been started
	MemberPhasePending MemberPhase = "Pending"
	// MemberPhaseCreated indicates that all resources needed for the member have been created
	MemberPhaseCreated MemberPhase = "Created"
	// MemberPhaseCreationFailed indicates that creation of member resources was not possible, fallback to MemberPhaseCreated state
	MemberPhaseCreationFailed MemberPhase = "CreationFailed"
	// MemberPhaseFailed indicates that the member is gone beyond hope of recovery. It must be replaced with a new member.
	MemberPhaseFailed MemberPhase = "Failed"
	// MemberPhaseCleanOut indicates that a dbserver is in the process of being cleaned out
	MemberPhaseCleanOut MemberPhase = "CleanOut"
	// MemberPhaseDrain indicates that a dbserver is in the process of being cleaned out as result of draining a node
	MemberPhaseDrain MemberPhase = "Drain"
	// MemberPhaseResign indicates that a dbserver is in the process of resigning for a shutdown
	MemberPhaseResign MemberPhase = "Resign"
	// MemberPhaseShuttingDown indicates that a member is shutting down
	MemberPhaseShuttingDown MemberPhase = "ShuttingDown"
	// MemberPhaseRotating indicates that a member is being rotated
	MemberPhaseRotating MemberPhase = "Rotating"
	// MemberPhaseRotateStart indicates that a member is being rotated but wont get up outside of plan
	MemberPhaseRotateStart MemberPhase = "RotateStart"
	// MemberPhaseUpgrading indicates that a member is in the process of upgrading its database data format
	MemberPhaseUpgrading MemberPhase = "Upgrading"
)

// IsPending returns true when given phase == "" OR "Pending"
func (p MemberPhase) IsPending() bool {
	return p == MemberPhaseNone || p == MemberPhasePending || p == MemberPhaseCreationFailed
}

// IsFailed returns true when given phase == "Failed"
func (p MemberPhase) IsFailed() bool {
	return p == MemberPhaseFailed
}

// IsReady returns true when given phase == "Created"
func (p MemberPhase) IsReady() bool {
	return p == MemberPhaseCreated
}

// IsCreatedOrDrain returns true when given phase is MemberPhaseCreated or MemberPhaseDrain
func (p MemberPhase) IsCreatedOrDrain() bool {
	return p == MemberPhaseCreated || p == MemberPhaseDrain
}

// String returns string from MemberPhase
func (p MemberPhase) String() string {
	return string(p)
}

// GetPhase parses string into phase
func GetPhase(phase string) (MemberPhase, bool) {
	switch p := MemberPhase(phase); p {
	case MemberPhaseNone, MemberPhasePending, MemberPhaseCreated, MemberPhaseCreationFailed, MemberPhaseFailed, MemberPhaseCleanOut, MemberPhaseDrain, MemberPhaseResign, MemberPhaseShuttingDown, MemberPhaseRotating, MemberPhaseRotateStart, MemberPhaseUpgrading:
		return p, true
	default:
		return "", false
	}
}
