//
// DISCLAIMER
//
// Copyright 2023-2026 ArangoDB GmbH, Cologne, Germany
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
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
)

// ConditionType is a strongly typed condition name
type ConditionType = sharedApi.ConditionType

const (
	// ConditionTypeReady indicates that the member or entire deployment is ready and running normally.
	ConditionTypeReady ConditionType = "Ready"
	// ConditionTypeStarted indicates that the member was ready at least once.
	ConditionTypeStarted ConditionType = "Started"
	// ConditionTypeReachable indicates that the member is reachable.
	ConditionTypeReachable ConditionType = "Reachable"
	// ConditionTypeScheduled indicates that the member primary pod is scheduled.
	ConditionTypeScheduled ConditionType = "Scheduled"
	// ConditionTypeScheduleSpecChanged indicates that the member schedule spec was changed.
	ConditionTypeScheduleSpecChanged ConditionType = "ScheduleSpecChanged"
	// ConditionTypeServing indicates that the member core services are running.
	ConditionTypeServing ConditionType = "Serving"
	// ConditionTypeActive indicates that the member server container started.
	ConditionTypeActive ConditionType = "Active"
	// ConditionTypeTerminated indicates that the member has terminated and will not restart.
	ConditionTypeTerminated ConditionType = "Terminated"
	// ConditionTypeAutoUpgrade indicates that the member has to be started with `--database.auto-upgrade` once.
	ConditionTypeAutoUpgrade ConditionType = "AutoUpgrade"
	// ConditionTypeUpgradeAllowed indicates that the member upgrade is allowed in the manual procedure.
	ConditionTypeUpgradeAllowed ConditionType = "UpgradeAllowed"

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

	// ConditionTypeSyncEnabled Define if DBServer contains any data
	ConditionTypeDBServerWithData ConditionType = "DBServerWithData"
	// ConditionTypeSyncEnabled Define if DBServer contains any active data leaders
	ConditionTypeDBServerWithDataLeader ConditionType = "DBServerWithDataLeader"

	// ConditionTypeGatewayConfig contains current config checksum of the Gateway
	ConditionTypeGatewayConfig ConditionType = "GatewayConfig"

	// ConditionTypeGatewaySidecarEnabled indicates that the sidecar gateway is enabled.
	ConditionTypeGatewaySidecarEnabled ConditionType = "GatewaySidecarEnabled"
)

// Condition represents one current condition of a deployment or deployment member.
// A condition might not show up if it is not happening.
// For example, if a cluster is not upgrading, the Upgrading condition would not show up.
type Condition = sharedApi.Condition

// ConditionList is a list of conditions.
// Each type is allowed only once.
type ConditionList = sharedApi.ConditionList
