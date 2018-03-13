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

// MemberStatus holds the current status of a single member (server)
type MemberStatus struct {
	// ID holds the unique ID of the member.
	// This id is also used within the ArangoDB cluster to identify this server.
	ID string `json:"id"`
	// State holds the current state of this member
	State MemberState `json:"state"`
	// PersistentVolumeClaimName holds the name of the persistent volume claim used for this member (if any).
	PersistentVolumeClaimName string `json:"persistentVolumeClaimName,omitempty"`
	// PodName holds the name of the Pod that currently runs this member
	PodName string `json:"podName,omitempty"`
	// Conditions specific to this member
	Conditions ConditionList `json:"conditions,omitempty"`
}
