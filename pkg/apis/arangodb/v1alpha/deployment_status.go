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

type DeploymentState string

const (
	DeploymentStateNone      DeploymentState = ""
	DeploymentStateCreating  DeploymentState = "Creating"
	DeploymentStateRunning   DeploymentState = "Running"
	DeploymentStateScaling   DeploymentState = "Scaling"
	DeploymentStateUpgrading DeploymentState = "Upgrading"
	DeploymentStateFailed    DeploymentState = "Failed"
)

// IsFailed returns true if given state is DeploymentStateFailed
func (cs DeploymentState) IsFailed() bool {
	return cs == DeploymentStateFailed
}

// DeploymentStatus contains the status part of a Cluster resource.
type DeploymentStatus struct {
	State  DeploymentState `json:"state"`
	Reason string          `json:"reason,omitempty"` // Reason for current state

	Members DeploymentStatusMembers `json:"members"`
}

type DeploymentStatusMembers struct {
	Single       []MemberStatus `json:"single,omitempty"`
	Agents       []MemberStatus `json:"agents,omitempty"`
	DBServers    []MemberStatus `json:"dbservers,omitempty"`
	Coordinators []MemberStatus `json:"coordinators,omitempty"`
	SyncMasters  []MemberStatus `json:"syncmasters,omitempty"`
	SyncWorkers  []MemberStatus `json:"syncworkers,omitempty"`
}

type MemberState string

const (
	MemberStateCreating     MemberState = "Creating"
	MemberStateReady        MemberState = "Ready"
	MemberStateCleanout     MemberState = "Cleanout"
	MemberStateShuttingDown MemberState = "ShuttingDown"
)

type MemberStatus struct {
	State     MemberState `json:"state"`
	ClusterID string      `json:"clusterID"`
	PodName   string      `json:"podName,omitempty"`
}
