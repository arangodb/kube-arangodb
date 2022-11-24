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

package v1

// DeploymentReplicationStatus contains the status part of
// an ArangoDeploymentReplication.
type DeploymentReplicationStatus struct {
	// Phase holds the current lifetime phase of the deployment replication
	Phase DeploymentReplicationPhase `json:"phase,omitempty"`
	// Reason contains a human-readable reason for reaching the current phase (can be empty)
	Reason string `json:"reason,omitempty"` // Reason for current phase

	// Conditions specific to the entire deployment replication
	Conditions ConditionList `json:"conditions,omitempty"`

	// Deprecated: this field will be removed in future versions
	// Source contains the detailed status of the source endpoint
	Source EndpointStatus `json:"source"`

	// Deprecated: this field will be removed in future versions
	// Destination contains the detailed status of the destination endpoint
	Destination EndpointStatus `json:"destination"`

	// Deprecated: this field will not be updated anymore
	// CancelFailures records the number of times that the configuration was canceled
	// which resulted in an error.
	CancelFailures int `json:"cancel-failures,omitempty"`

	// IncomingSynchronization contains the incoming synchronization status for all databases
	IncomingSynchronization SynchronizationStatus `json:"incoming-synchronization,omitempty"`
}
