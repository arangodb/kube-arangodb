//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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

package v2alpha1

import (
	"github.com/arangodb/kube-arangodb/pkg/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConditionType is a strongly typed condition name
type ConditionType string

func (c ConditionType) String() string {
	return string(c)
}

const (
	// ConditionTypeReady indicates that the member or entire deployment is ready and running normally.
	ConditionTypeReady ConditionType = "Ready"
	// ConditionTypeTerminated indicates that the member has terminated and will not restart.
	ConditionTypeTerminated ConditionType = "Terminated"
	// ConditionTypeAutoUpgrade indicates that the member has to be started with `--database.auto-upgrade` once.
	ConditionTypeAutoUpgrade ConditionType = "AutoUpgrade"
	// ConditionTypeCleanedOut indicates that the member (dbserver) has been cleaned out.
	// Always check in combination with ConditionTypeTerminated.
	ConditionTypeCleanedOut ConditionType = "CleanedOut"
	// ConditionTypeAgentRecoveryNeeded indicates that the member (agent) will no
	// longer recover from its current volume and there has to be rebuild
	// using the recovery procedure.
	ConditionTypeAgentRecoveryNeeded ConditionType = "AgentRecoveryNeeded"
	// ConditionTypePodSchedulingFailure indicates that one or more pods belonging to the deployment cannot be schedule.
	ConditionTypePodSchedulingFailure ConditionType = "PodSchedulingFailure"
	// ConditionTypeSecretsChanged indicates that the value of one of more secrets used by
	// the deployment have changed. Once that is the case, the operator will no longer
	// touch the deployment, until the original secrets have been restored.
	ConditionTypeSecretsChanged ConditionType = "SecretsChanged"
	// ConditionTypeMemberOfCluster indicates that the member is a known member of the ArangoDB cluster.
	ConditionTypeMemberOfCluster ConditionType = "MemberOfCluster"
	// ConditionTypeBootstrapCompleted indicates that the initial cluster bootstrap has been completed.
	ConditionTypeBootstrapCompleted ConditionType = "BootstrapCompleted"
	// ConditionTypeBootstrapSucceded indicates that the initial cluster bootstrap completed successfully.
	ConditionTypeBootstrapSucceded ConditionType = "BootstrapSucceded"
	// ConditionTypeTerminating indicates that the member is terminating but not yet terminated.
	ConditionTypeTerminating ConditionType = "Terminating"
	// ConditionTypeTerminating indicates that the deployment is up to date.
	ConditionTypeUpToDate ConditionType = "UpToDate"
	// ConditionTypeMarkedToRemove indicates that the member is marked to be removed.
	ConditionTypeMarkedToRemove ConditionType = "MarkedToRemove"
	// ConditionTypeUpgradeFailed indicates that upgrade failed
	ConditionTypeUpgradeFailed ConditionType = "UpgradeFailed"
	// ConditionTypeMaintenanceMode indicates that Maintenance is enabled
	ConditionTypeMaintenanceMode ConditionType = "MaintenanceMode"
	// ConditionTypePendingRestart indicates that restart is required
	ConditionTypePendingRestart ConditionType = "PendingRestart"
	// ConditionTypeRestart indicates that restart will be started
	ConditionTypeRestart ConditionType = "Restart"
	// ConditionTypePendingTLSRotation indicates that TLS rotation is pending
	ConditionTypePendingTLSRotation ConditionType = "PendingTLSRotation"
	// ConditionTypePendingUpdate indicates that runtime update is pending
	ConditionTypePendingUpdate ConditionType = "PendingUpdate"
	// ConditionTypeUpdating indicates that runtime update is in progress
	ConditionTypeUpdating ConditionType = "Updating"
	// ConditionTypeUpdateFailed indicates that runtime update failed
	ConditionTypeUpdateFailed ConditionType = "UpdateFailed"
)

// Condition represents one current condition of a deployment or deployment member.
// A condition might not show up if it is not happening.
// For example, if a cluster is not upgrading, the Upgrading condition would not show up.
type Condition struct {
	// Type of  condition.
	Type ConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

func (c Condition) IsTrue() bool {
	return c.Status == v1.ConditionTrue
}

// ConditionList is a list of conditions.
// Each type is allowed only once.
type ConditionList []Condition

// Equal checks for equality
func (list ConditionList) Equal(other ConditionList) bool {
	if len(list) != len(other) {
		return false
	}

	for i := 0; i < len(list); i++ {
		c, found := other.Get(list[i].Type)
		if !found {
			return false
		}

		if !list[i].Equal(c) {
			return false
		}
	}

	return true
}

// Equal checks for equality
func (c Condition) Equal(other Condition) bool {
	return c.Type == other.Type &&
		c.Status == other.Status &&
		util.TimeCompareEqual(c.LastUpdateTime, other.LastUpdateTime) &&
		util.TimeCompareEqual(c.LastTransitionTime, other.LastTransitionTime) &&
		c.Reason == other.Reason &&
		c.Message == other.Message
}

// IsTrue return true when a condition with given type exists and its status is `True`.
func (list ConditionList) IsTrue(conditionType ConditionType) bool {
	c, found := list.Get(conditionType)
	return found && c.IsTrue()
}

// Get a condition by type.
// Returns true if found, false if not found.
func (list ConditionList) Get(conditionType ConditionType) (Condition, bool) {
	// Covers nil and empty lists
	if len(list) == 0 {
		return Condition{}, false
	}

	for _, x := range list {
		if x.Type == conditionType {
			return x, true
		}
	}
	// Not found
	return Condition{}, false
}

// Touch update condition LastUpdateTime if condition exists
func (list *ConditionList) Touch(conditionType ConditionType) bool {
	src := *list
	for i, x := range src {
		if x.Type == conditionType {
			src[i].LastUpdateTime = metav1.Now()
			return true
		}
	}

	return false
}

// Update the condition, replacing an old condition with same type (if any)
// Returns true when changes were made, false otherwise.
func (list *ConditionList) Update(conditionType ConditionType, status bool, reason, message string) bool {
	src := *list
	statusX := v1.ConditionFalse
	if status {
		statusX = v1.ConditionTrue
	}
	for i, x := range src {
		if x.Type == conditionType {
			if x.Status != statusX {
				// Transition to another status
				src[i].Status = statusX
				now := metav1.Now()
				src[i].LastTransitionTime = now
				src[i].LastUpdateTime = now
				src[i].Reason = reason
				src[i].Message = message
			} else if x.Reason != reason || x.Message != message {
				src[i].LastUpdateTime = metav1.Now()
				src[i].Reason = reason
				src[i].Message = message
			} else {
				return false
			}
			return true
		}
	}
	// Not found
	now := metav1.Now()
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
