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

// DeploymentStatus contains the status part of a Cluster resource.
type DeploymentStatus struct {
	// Phase holds the current lifetime phase of the deployment
	Phase DeploymentPhase `json:"phase"`
	// Reason contains a human readable reason for reaching the current state (can be empty)
	Reason string `json:"reason,omitempty"` // Reason for current state

	// ServiceName holds the name of the Service a client can use (inside the k8s cluster)
	// to access ArangoDB.
	ServiceName string `json:"serviceName,omitempty"`
	// SyncServiceName holds the name of the Service a client can use (inside the k8s cluster)
	// to access syncmasters (only set when dc2dc synchronization is enabled).
	SyncServiceName string `json:"syncServiceName,omitempty"`

	// Images holds a list of ArangoDB images with their ID and ArangoDB version.
	Images ImageInfoList `json:"arangodb-images,omitempty"`

	// Members holds the status for all members in all server groups
	Members DeploymentStatusMembers `json:"members"`

	// Conditions specific to the entire deployment
	Conditions ConditionList `json:"conditions,omitempty"`

	// Plan to update this deployment
	Plan Plan `json:"plan,omitempty"`

	// AcceptedSpec contains the last specification that was accepted by the operator.
	AcceptedSpec *DeploymentSpec `json:"accepted-spec,omitempty"`

	// SecretHashes keeps a sha256 hash of secret values, so we can
	// detect changes in secret values.
	SecretHashes *SecretHashes `json:"secret-hashes,omitempty"`
}
