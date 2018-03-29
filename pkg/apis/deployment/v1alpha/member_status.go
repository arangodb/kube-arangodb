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

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	// RecentTerminatons holds the times when this member was recently terminated.
	// First entry is the oldest. (do not add omitempty, since we want to be able to switch from a list to an empty list)
	RecentTerminations []metav1.Time `json:"recent-terminations"`
}

// RemoveTerminationsBefore removes all recent terminations before the given timestamp.
// It returns the number of terminations that have been removed.
func (s *MemberStatus) RemoveTerminationsBefore(timestamp time.Time) int {
	removed := 0
	for {
		if len(s.RecentTerminations) == 0 {
			// Nothing left
			return removed
		}
		if s.RecentTerminations[0].Time.Before(timestamp) {
			// Let's remove it
			s.RecentTerminations = s.RecentTerminations[1:]
			removed++
		} else {
			// First (oldest) is not before given timestamp, we're done
			return removed
		}
	}
}

// RecentTerminationsSince returns the number of terminations since the given timestamp.
func (s MemberStatus) RecentTerminationsSince(timestamp time.Time) int {
	count := 0
	for idx := len(s.RecentTerminations) - 1; idx >= 0; idx-- {
		if s.RecentTerminations[idx].Time.Before(timestamp) {
			// This termination is before the timestamp, so we're done
			return count
		}
		count++
	}
	return count
}
