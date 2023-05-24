//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

import (
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

// ConditionType is a strongly typed condition name
type ConditionType string

func (c ConditionType) String() string {
	return string(c)
}

const (
	// ConditionTypeReady indicates that the member or entire deployment is ready and running normally.
	ConditionTypeReady ConditionType = "Ready"
	// ConditionTypeStarted indicates that the member was ready at least once.
	ConditionTypeStarted ConditionType = "Started"
	// ConditionTypeReachable indicates that the member is reachable.
	ConditionTypeReachable ConditionType = "Reachable"
	// ConditionTypeScheduled indicates that the member primary pod is scheduled.
	ConditionTypeScheduled ConditionType = "Scheduled"
	// ConditionTypeServing indicates that the member core services are running.
	ConditionTypeServing ConditionType = "Serving"
	// ConditionTypeActive indicates that the member server container started.
	ConditionTypeActive ConditionType = "Active"
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
	// ConditionTypeMemberOfCluster indicates that the member is a known member of the ArangoDB cluster.
	ConditionTypeMemberOfCluster ConditionType = "MemberOfCluster"

	// ConditionTypeTerminating indicates that the member is terminating but not yet terminated.
	ConditionTypeTerminating ConditionType = "Terminating"
	// ConditionTypeUpToDate indicates that the deployment is up to date.
	ConditionTypeUpToDate ConditionType = "UpToDate"
	// ConditionTypeSpecAccepted indicates that the deployment spec has been accepted.
	ConditionTypeSpecAccepted ConditionType = "SpecAccepted"
	// ConditionTypeSpecPropagated indicates that the deployment has been at least once UpToDate after spec acceptance.
	ConditionTypeSpecPropagated ConditionType = "SpecPropagated"
	// ConditionTypeMemberVolumeUnschedulable indicates that the member cannot schedued due to volume issue.
	ConditionTypeMemberVolumeUnschedulable ConditionType = "MemberVolumeUnschedulable"
	// ConditionTypeMarkedToRemove indicates that the member is marked to be removed.
	ConditionTypeMarkedToRemove ConditionType = "MarkedToRemove"
	// ConditionTypeScaleDownCandidate indicates that the member will be picked in ScaleDown operaion.
	ConditionTypeScaleDownCandidate ConditionType = "ScaleDownCandidate"
	// ConditionTypeUpgradeFailed indicates that upgrade failed
	ConditionTypeUpgradeFailed ConditionType = "UpgradeFailed"
	// ConditionTypeArchitectureMismatch indicates that the member has a different architecture than the deployment.
	ConditionTypeArchitectureMismatch ConditionType = "ArchitectureMismatch"
	// ConditionTypeArchitectureChangeCannotBeApplied indicates that the member has a different architecture than the requested one.
	ConditionTypeArchitectureChangeCannotBeApplied ConditionType = "ArchitectureChangeCannotBeApplied"

	// ConditionTypeMemberMaintenanceMode indicates that Maintenance is enabled on particular member
	ConditionTypeMemberMaintenanceMode ConditionType = "MemberMaintenanceMode"
	// ConditionTypeMaintenanceMode indicates that Maintenance is enabled
	ConditionTypeMaintenanceMode ConditionType = "MaintenanceMode"

	// ConditionTypePendingRestart indicates that restart is required
	ConditionTypePendingRestart ConditionType = "PendingRestart"
	// ConditionTypeRestart indicates that restart will be started
	ConditionTypeRestart ConditionType = "Restart"
	// MemberReplacementRequired indicates that the member requires a replacement to proceed with next actions.
	MemberReplacementRequired ConditionType = "MemberReplacementRequired"

	// ConditionTypePendingTLSRotation indicates that TLS rotation is pending
	ConditionTypePendingTLSRotation ConditionType = "PendingTLSRotation"

	// ConditionTypePendingUpdate indicates that runtime update is pending
	ConditionTypePendingUpdate ConditionType = "PendingUpdate"
	// ConditionTypeUpdating indicates that runtime update is in progress
	ConditionTypeUpdating ConditionType = "Updating"
	// ConditionTypeUpdateFailed indicates that runtime update failed
	ConditionTypeUpdateFailed ConditionType = "UpdateFailed"

	// ConditionTypeTopologyAware indicates that the member is deployed with TopologyAwareness.
	ConditionTypeTopologyAware ConditionType = "TopologyAware"

	// ConditionTypePVCResizePending indicates that the member has to be restarted due to PVC Resized pending action
	ConditionTypePVCResizePending ConditionType = "PVCResizePending"

	// ConditionTypeLicenseSet indicates that license V2 is set on cluster.
	ConditionTypeLicenseSet ConditionType = "LicenseSet"

	// ConditionTypeBackupInProgress indicates that there is Backup in progress on cluster
	ConditionTypeBackupInProgress ConditionType = "BackupInProgress"
	// ConditionTypeUpgradeInProgress indicates that there is upgrade in progress on cluster
	ConditionTypeUpgradeInProgress ConditionType = "UpgradeInProgress"
	// ConditionTypeUpdateInProgress indicates that there is update in progress on cluster
	ConditionTypeUpdateInProgress ConditionType = "UpdateInProgress"

	// ConditionTypeMaintenance indicates that maintenance is enabled on cluster
	ConditionTypeMaintenance ConditionType = "Maintenance"

	// ConditionTypeSyncEnabled Define if sync is enabled
	ConditionTypeSyncEnabled ConditionType = "SyncEnabled"
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
	// Hash keep propagation hash id, for example checksum of secret
	Hash string `json:"hash,omitempty"`
	// Params keeps additional params for the condition
	Params ConditionParams `json:"params,omitempty"`
}

func (c Condition) IsTrue() bool {
	return c.Status == core.ConditionTrue
}

// Equal checks for equality
func (c Condition) Equal(other Condition) bool {
	return c.Type == other.Type &&
		c.Status == other.Status &&
		c.Params.Equal(other.Params) &&
		util.TimeCompareEqual(c.LastUpdateTime, other.LastUpdateTime) &&
		util.TimeCompareEqual(c.LastTransitionTime, other.LastTransitionTime) &&
		c.Reason == other.Reason &&
		c.Message == other.Message &&
		c.Hash == other.Hash
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

// IsTrue return true when a condition with given type exists and its status is `True`.
func (list ConditionList) IsTrue(conditionType ConditionType) bool {
	c, found := list.Get(conditionType)
	return found && c.IsTrue()
}

// Check create a condition checker.
func (list ConditionList) Check(conditionType ConditionType) ConditionCheck {
	c, ok := list.Get(conditionType)

	return conditionCheck{
		condition: c,
		exists:    ok,
	}
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
			src[i].LastUpdateTime = meta.Now()
			return true
		}
	}

	return false
}

func (list ConditionList) Index(conditionType ConditionType) int {
	for i, x := range list {
		if x.Type == conditionType {
			return i
		}
	}

	return -1
}

func (list *ConditionList) update(conditionType ConditionType, status bool, reason, message, hash string) bool {
	src := *list
	statusX := core.ConditionFalse
	if status {
		statusX = core.ConditionTrue
	}

	index := list.Index(conditionType)

	if index == -1 {
		// Not found
		now := meta.Now()
		*list = append(src, Condition{
			Type:               conditionType,
			LastUpdateTime:     now,
			LastTransitionTime: now,
			Status:             statusX,
			Reason:             reason,
			Message:            message,
			Hash:               hash,
		})
		return true
	}

	if src[index].Status != statusX {
		// Transition to another status
		src[index].Status = statusX
		now := meta.Now()
		src[index].LastTransitionTime = now
		src[index].LastUpdateTime = now
		src[index].Reason = reason
		src[index].Message = message
		src[index].Hash = hash
	} else if src[index].Reason != reason || src[index].Message != message || src[index].Hash != hash {
		src[index].LastUpdateTime = meta.Now()
		src[index].Reason = reason
		src[index].Message = message
		src[index].Hash = hash
	} else {
		return false
	}
	return true
}

// Update the condition, replacing an old condition with same type (if any)
// Returns true when changes were made, false otherwise.
func (list *ConditionList) Update(conditionType ConditionType, status bool, reason, message string) bool {
	return list.update(conditionType, status, reason, message, "")
}

// UpdateWithHash updates the condition, replacing an old condition with same type (if any)
// Returns true when changes were made, false otherwise.
func (list *ConditionList) UpdateWithHash(conditionType ConditionType, status bool, reason, message, hash string) bool {
	return list.update(conditionType, status, reason, message, hash)
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
