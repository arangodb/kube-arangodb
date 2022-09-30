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

import (
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConditionType is a strongly typed condition name
type ConditionType string

const (
	// ConditionTypeConfigured indicates that the replication has been configured.
	ConditionTypeConfigured ConditionType = "Configured"
	// ConditionTypeEnsuredInSync indicates that the replication consistency was checked.
	ConditionTypeEnsuredInSync ConditionType = "EnsuredInSync"
	// ConditionTypeAborted indicates that the replication was canceled with abort option.
	ConditionTypeAborted ConditionType = "Aborted"

	// ConditionConfiguredReasonActive describes synchronization as active.
	ConditionConfiguredReasonActive = "Active"
	// ConditionConfiguredReasonInactive describes synchronization as inactive.
	ConditionConfiguredReasonInactive = "Inactive"
	// ConditionConfiguredReasonInvalid describes synchronization as active.
	ConditionConfiguredReasonInvalid = "Invalid"
)

// Condition represents one current condition of a deployment or deployment member.
// A condition might not show up if it is not happening.
// For example, if a cluster is not upgrading, the Upgrading condition would not show up.
type Condition struct {
	// Type of  condition.
	Type ConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status core.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime meta.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime meta.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// ConditionList is a list of conditions.
// Each type is allowed only once.
type ConditionList []Condition

// IsTrue return true when a condition with given type exists and its status is `True`.
func (list ConditionList) IsTrue(conditionType ConditionType) bool {
	c, found := list.Get(conditionType)
	return found && c.Status == core.ConditionTrue
}

// Get a condition by type.
// Returns true if found, false if not found.
func (list ConditionList) Get(conditionType ConditionType) (Condition, bool) {
	for _, x := range list {
		if x.Type == conditionType {
			return x, true
		}
	}
	// Not found
	return Condition{}, false
}

// Update the condition, replacing an old condition with same type (if any)
// Returns true when changes were made, false otherwise.
func (list *ConditionList) Update(conditionType ConditionType, status bool, reason, message string) bool {
	src := *list
	statusX := core.ConditionFalse
	if status {
		statusX = core.ConditionTrue
	}
	for i, x := range src {
		if x.Type == conditionType {
			if x.Status != statusX {
				// Transition to another status
				src[i].Status = statusX
				now := meta.Now()
				src[i].LastTransitionTime = now
				src[i].LastUpdateTime = now
				src[i].Reason = reason
				src[i].Message = message
			} else if x.Reason != reason || x.Message != message {
				src[i].LastUpdateTime = meta.Now()
				src[i].Reason = reason
				src[i].Message = message
			} else {
				return false
			}
			return true
		}
	}
	// Not found
	now := meta.Now()
	*list = append(src, Condition{
		Type:               conditionType,
		LastUpdateTime:     now,
		LastTransitionTime: now,
		Status:             statusX,
		Reason:             reason,
		Message:            message,
	})
	return true
}

// Remove the condition with given type.
// Returns true if removed, or false if not found.
func (list *ConditionList) Remove(conditionType ConditionType) bool {
	src := *list
	for i, x := range src {
		if x.Type == conditionType {
			*list = append(src[:i], src[i+1:]...)
			return true
		}
	}
	// Not found
	return false
}
