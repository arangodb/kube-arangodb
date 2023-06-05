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

package reconcile

import (
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

var (
	// Ensure implementation

	_ Action        = &actionAddMember{}
	_ actionFactory = newAddMemberAction

	_ Action        = &actionAppendTLSCACertificate{}
	_ actionFactory = newAppendTLSCACertificateAction

	_ Action        = &actionArangoMemberUpdatePodSpec{}
	_ actionFactory = newArangoMemberUpdatePodSpecAction

	_ Action        = &actionArangoMemberUpdatePodStatus{}
	_ actionFactory = newArangoMemberUpdatePodStatusAction

	_ Action        = &actionBackupRestore{}
	_ actionFactory = newBackupRestoreAction

	_ Action        = &actionBackupRestoreClean{}
	_ actionFactory = newBackupRestoreCleanAction

	_ Action        = &actionBootstrapSetPassword{}
	_ actionFactory = newBootstrapSetPasswordAction

	_ Action        = &actionBootstrapUpdate{}
	_ actionFactory = newBootstrapUpdateAction

	_ Action        = &actionCleanMemberService{}
	_ actionFactory = newCleanMemberServiceAction

	_ Action        = &actionCleanOutMember{}
	_ actionFactory = newCleanOutMemberAction

	_ Action        = &actionCleanTLSCACertificate{}
	_ actionFactory = newCleanTLSCACertificateAction

	_ Action        = &actionCleanTLSKeyfileCertificate{}
	_ actionFactory = newCleanTLSKeyfileCertificateAction

	_ Action        = &actionClusterMemberCleanup{}
	_ actionFactory = newClusterMemberCleanupAction

	_ Action        = &actionDisableClusterScaling{}
	_ actionFactory = newDisableClusterScalingAction

	_ Action        = &actionDisableMaintenance{}
	_ actionFactory = newDisableMaintenanceAction

	_ Action        = &actionDisableMemberMaintenance{}
	_ actionFactory = newDisableMemberMaintenanceAction

	_ Action        = &actionEnableClusterScaling{}
	_ actionFactory = newEnableClusterScalingAction

	_ Action        = &actionEnableMaintenance{}
	_ actionFactory = newEnableMaintenanceAction

	_ Action        = &actionEnableMemberMaintenance{}
	_ actionFactory = newEnableMemberMaintenanceAction

	_ Action        = &actionEncryptionKeyAdd{}
	_ actionFactory = newEncryptionKeyAddAction

	_ Action        = &actionEncryptionKeyPropagated{}
	_ actionFactory = newEncryptionKeyPropagatedAction

	_ Action        = &actionEncryptionKeyRefresh{}
	_ actionFactory = newEncryptionKeyRefreshAction

	_ Action        = &actionEncryptionKeyRemove{}
	_ actionFactory = newEncryptionKeyRemoveAction

	_ Action        = &actionEncryptionKeyStatusUpdate{}
	_ actionFactory = newEncryptionKeyStatusUpdateAction

	_ Action        = &actionIdle{}
	_ actionFactory = newIdleAction

	_ Action        = &actionJWTAdd{}
	_ actionFactory = newJWTAddAction

	_ Action        = &actionJWTClean{}
	_ actionFactory = newJWTCleanAction

	_ Action        = &actionJWTPropagated{}
	_ actionFactory = newJWTPropagatedAction

	_ Action        = &actionJWTRefresh{}
	_ actionFactory = newJWTRefreshAction

	_ Action        = &actionJWTSetActive{}
	_ actionFactory = newJWTSetActiveAction

	_ Action        = &actionJWTStatusUpdate{}
	_ actionFactory = newJWTStatusUpdateAction

	_ Action        = &actionKillMemberPod{}
	_ actionFactory = newKillMemberPodAction

	_ Action        = &actionLicenseSet{}
	_ actionFactory = newLicenseSetAction

	_ Action        = &actionMarkToRemoveMember{}
	_ actionFactory = newMarkToRemoveMemberAction

	_ Action        = &actionMemberPhaseUpdate{}
	_ actionFactory = newMemberPhaseUpdateAction

	_ Action        = &actionMemberRIDUpdate{}
	_ actionFactory = newMemberRIDUpdateAction

	_ Action        = &actionPVCResize{}
	_ actionFactory = newPVCResizeAction

	_ Action        = &actionPVCResized{}
	_ actionFactory = newPVCResizedAction

	_ Action        = &actionPlaceHolder{}
	_ actionFactory = newPlaceHolderAction

	_ Action        = &actionRebalancerCheck{}
	_ actionFactory = newRebalancerCheckAction

	_ Action        = &actionRebalancerClean{}
	_ actionFactory = newRebalancerCleanAction

	_ Action        = &actionRebalancerGenerate{}
	_ actionFactory = newRebalancerGenerateAction

	_ Action        = &actionRebuildOutSyncedShards{}
	_ actionFactory = newRebuildOutSyncedShardsAction

	_ Action        = &actionRecreateMember{}
	_ actionFactory = newRecreateMemberAction

	_ Action        = &actionRefreshTLSKeyfileCertificate{}
	_ actionFactory = newRefreshTLSKeyfileCertificateAction

	_ Action        = &actionRemoveMember{}
	_ actionFactory = newRemoveMemberAction

	_ Action        = &actionRemoveMemberPVC{}
	_ actionFactory = newRemoveMemberPVCAction

	_ Action        = &actionRenewTLSCACertificate{}
	_ actionFactory = newRenewTLSCACertificateAction

	_ Action        = &actionRenewTLSCertificate{}
	_ actionFactory = newRenewTLSCertificateAction

	_ Action        = &actionResignLeadership{}
	_ actionFactory = newResignLeadershipAction

	_ Action        = &actionResourceSync{}
	_ actionFactory = newResourceSyncAction

	_ Action        = &actionRotateMember{}
	_ actionFactory = newRotateMemberAction

	_ Action        = &actionRotateStartMember{}
	_ actionFactory = newRotateStartMemberAction

	_ Action        = &actionRotateStopMember{}
	_ actionFactory = newRotateStopMemberAction

	_ Action        = &actionRuntimeContainerArgsLogLevelUpdate{}
	_ actionFactory = newRuntimeContainerArgsLogLevelUpdateAction

	_ Action        = &actionRuntimeContainerImageUpdate{}
	_ actionFactory = newRuntimeContainerImageUpdateAction

	_ Action        = &actionRuntimeContainerSyncTolerations{}
	_ actionFactory = newRuntimeContainerSyncTolerationsAction

	_ Action        = &actionSetCondition{}
	_ actionFactory = newSetConditionAction

	_ Action        = &actionSetConditionV2{}
	_ actionFactory = newSetConditionV2Action

	_ Action        = &actionSetCurrentImage{}
	_ actionFactory = newSetCurrentImageAction

	_ Action        = &actionSetCurrentMemberArch{}
	_ actionFactory = newSetCurrentMemberArchAction

	_ Action        = &actionSetMaintenanceCondition{}
	_ actionFactory = newSetMaintenanceConditionAction

	_ Action        = &actionSetMemberCondition{}
	_ actionFactory = newSetMemberConditionAction

	_ Action        = &actionSetMemberConditionV2{}
	_ actionFactory = newSetMemberConditionV2Action

	_ Action        = &actionSetMemberCurrentImage{}
	_ actionFactory = newSetMemberCurrentImageAction

	_ Action        = &actionShutdownMember{}
	_ actionFactory = newShutdownMemberAction

	_ Action        = &actionTLSKeyStatusUpdate{}
	_ actionFactory = newTLSKeyStatusUpdateAction

	_ Action        = &actionTLSPropagated{}
	_ actionFactory = newTLSPropagatedAction

	_ Action        = &actionTimezoneSecretSet{}
	_ actionFactory = newTimezoneSecretSetAction

	_ Action        = &actionTopologyDisable{}
	_ actionFactory = newTopologyDisableAction

	_ Action        = &actionTopologyEnable{}
	_ actionFactory = newTopologyEnableAction

	_ Action        = &actionTopologyMemberAssignment{}
	_ actionFactory = newTopologyMemberAssignmentAction

	_ Action        = &actionTopologyZonesUpdate{}
	_ actionFactory = newTopologyZonesUpdateAction

	_ Action        = &actionUpToDateUpdate{}
	_ actionFactory = newUpToDateUpdateAction

	_ Action        = &actionUpdateTLSSNI{}
	_ actionFactory = newUpdateTLSSNIAction

	_ Action        = &actionUpgradeMember{}
	_ actionFactory = newUpgradeMemberAction

	_ Action        = &actionWaitForMemberInSync{}
	_ actionFactory = newWaitForMemberInSyncAction

	_ Action        = &actionWaitForMemberReady{}
	_ actionFactory = newWaitForMemberReadyAction

	_ Action        = &actionWaitForMemberUp{}
	_ actionFactory = newWaitForMemberUpAction
)

func init() {
	// Register all actions

	// AddMember
	{
		// Get Action defition
		function := newAddMemberAction
		action := api.ActionTypeAddMember

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// AppendTLSCACertificate
	{
		// Get Action defition
		function := newAppendTLSCACertificateAction
		action := api.ActionTypeAppendTLSCACertificate

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// ArangoMemberUpdatePodSpec
	{
		// Get Action defition
		function := newArangoMemberUpdatePodSpecAction
		action := api.ActionTypeArangoMemberUpdatePodSpec

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// ArangoMemberUpdatePodStatus
	{
		// Get Action defition
		function := newArangoMemberUpdatePodStatusAction
		action := api.ActionTypeArangoMemberUpdatePodStatus

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// BackupRestore
	{
		// Get Action defition
		function := newBackupRestoreAction
		action := api.ActionTypeBackupRestore

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// BackupRestoreClean
	{
		// Get Action defition
		function := newBackupRestoreCleanAction
		action := api.ActionTypeBackupRestoreClean

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// BootstrapSetPassword
	{
		// Get Action defition
		function := newBootstrapSetPasswordAction
		action := api.ActionTypeBootstrapSetPassword

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// BootstrapUpdate
	{
		// Get Action defition
		function := newBootstrapUpdateAction
		action := api.ActionTypeBootstrapUpdate

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// CleanMemberService
	{
		// Get Action defition
		function := newCleanMemberServiceAction
		action := api.ActionTypeCleanMemberService

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// CleanOutMember
	{
		// Get Action defition
		function := newCleanOutMemberAction
		action := api.ActionTypeCleanOutMember

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// CleanTLSCACertificate
	{
		// Get Action defition
		function := newCleanTLSCACertificateAction
		action := api.ActionTypeCleanTLSCACertificate

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// CleanTLSKeyfileCertificate
	{
		// Get Action defition
		function := newCleanTLSKeyfileCertificateAction
		action := api.ActionTypeCleanTLSKeyfileCertificate

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// ClusterMemberCleanup
	{
		// Get Action defition
		function := newClusterMemberCleanupAction
		action := api.ActionTypeClusterMemberCleanup

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// DisableClusterScaling
	{
		// Get Action defition
		function := newDisableClusterScalingAction
		action := api.ActionTypeDisableClusterScaling

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// DisableMaintenance
	{
		// Get Action defition
		function := newDisableMaintenanceAction
		action := api.ActionTypeDisableMaintenance

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// DisableMemberMaintenance
	{
		// Get Action defition
		function := newDisableMemberMaintenanceAction
		action := api.ActionTypeDisableMemberMaintenance

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// EnableClusterScaling
	{
		// Get Action defition
		function := newEnableClusterScalingAction
		action := api.ActionTypeEnableClusterScaling

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// EnableMaintenance
	{
		// Get Action defition
		function := newEnableMaintenanceAction
		action := api.ActionTypeEnableMaintenance

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// EnableMemberMaintenance
	{
		// Get Action defition
		function := newEnableMemberMaintenanceAction
		action := api.ActionTypeEnableMemberMaintenance

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// EncryptionKeyAdd
	{
		// Get Action defition
		function := newEncryptionKeyAddAction
		action := api.ActionTypeEncryptionKeyAdd

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// EncryptionKeyPropagated
	{
		// Get Action defition
		function := newEncryptionKeyPropagatedAction
		action := api.ActionTypeEncryptionKeyPropagated

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// EncryptionKeyRefresh
	{
		// Get Action defition
		function := newEncryptionKeyRefreshAction
		action := api.ActionTypeEncryptionKeyRefresh

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// EncryptionKeyRemove
	{
		// Get Action defition
		function := newEncryptionKeyRemoveAction
		action := api.ActionTypeEncryptionKeyRemove

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// EncryptionKeyStatusUpdate
	{
		// Get Action defition
		function := newEncryptionKeyStatusUpdateAction
		action := api.ActionTypeEncryptionKeyStatusUpdate

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// Idle
	{
		// Get Action defition
		function := newIdleAction
		action := api.ActionTypeIdle

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// JWTAdd
	{
		// Get Action defition
		function := newJWTAddAction
		action := api.ActionTypeJWTAdd

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// JWTClean
	{
		// Get Action defition
		function := newJWTCleanAction
		action := api.ActionTypeJWTClean

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// JWTPropagated
	{
		// Get Action defition
		function := newJWTPropagatedAction
		action := api.ActionTypeJWTPropagated

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// JWTRefresh
	{
		// Get Action defition
		function := newJWTRefreshAction
		action := api.ActionTypeJWTRefresh

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// JWTSetActive
	{
		// Get Action defition
		function := newJWTSetActiveAction
		action := api.ActionTypeJWTSetActive

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// JWTStatusUpdate
	{
		// Get Action defition
		function := newJWTStatusUpdateAction
		action := api.ActionTypeJWTStatusUpdate

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// KillMemberPod
	{
		// Get Action defition
		function := newKillMemberPodAction
		action := api.ActionTypeKillMemberPod

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// LicenseSet
	{
		// Get Action defition
		function := newLicenseSetAction
		action := api.ActionTypeLicenseSet

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// MarkToRemoveMember
	{
		// Get Action defition
		function := newMarkToRemoveMemberAction
		action := api.ActionTypeMarkToRemoveMember

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// MemberPhaseUpdate
	{
		// Get Action defition
		function := newMemberPhaseUpdateAction
		action := api.ActionTypeMemberPhaseUpdate

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// MemberRIDUpdate
	{
		// Get Action defition
		function := newMemberRIDUpdateAction
		action := api.ActionTypeMemberRIDUpdate

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// PVCResize
	{
		// Get Action defition
		function := newPVCResizeAction
		action := api.ActionTypePVCResize

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// PVCResized
	{
		// Get Action defition
		function := newPVCResizedAction
		action := api.ActionTypePVCResized

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// PlaceHolder
	{
		// Get Action defition
		function := newPlaceHolderAction
		action := api.ActionTypePlaceHolder

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RebalancerCheck
	{
		// Get Action defition
		function := newRebalancerCheckAction
		action := api.ActionTypeRebalancerCheck

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RebalancerClean
	{
		// Get Action defition
		function := newRebalancerCleanAction
		action := api.ActionTypeRebalancerClean

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RebalancerGenerate
	{
		// Get Action defition
		function := newRebalancerGenerateAction
		action := api.ActionTypeRebalancerGenerate

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RebuildOutSyncedShards
	{
		// Get Action defition
		function := newRebuildOutSyncedShardsAction
		action := api.ActionTypeRebuildOutSyncedShards

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RecreateMember
	{
		// Get Action defition
		function := newRecreateMemberAction
		action := api.ActionTypeRecreateMember

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RefreshTLSKeyfileCertificate
	{
		// Get Action defition
		function := newRefreshTLSKeyfileCertificateAction
		action := api.ActionTypeRefreshTLSKeyfileCertificate

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RemoveMember
	{
		// Get Action defition
		function := newRemoveMemberAction
		action := api.ActionTypeRemoveMember

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RemoveMemberPVC
	{
		// Get Action defition
		function := newRemoveMemberPVCAction
		action := api.ActionTypeRemoveMemberPVC

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RenewTLSCACertificate
	{
		// Get Action defition
		function := newRenewTLSCACertificateAction
		action := api.ActionTypeRenewTLSCACertificate

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RenewTLSCertificate
	{
		// Get Action defition
		function := newRenewTLSCertificateAction
		action := api.ActionTypeRenewTLSCertificate

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// ResignLeadership
	{
		// Get Action defition
		function := newResignLeadershipAction
		action := api.ActionTypeResignLeadership

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// ResourceSync
	{
		// Get Action defition
		function := newResourceSyncAction
		action := api.ActionTypeResourceSync

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RotateMember
	{
		// Get Action defition
		function := newRotateMemberAction
		action := api.ActionTypeRotateMember

		// Wrap action main function
		function = withActionStartFailureGracePeriod(function, 60*time.Second)

		// Register action
		registerAction(action, function)
	}

	// RotateStartMember
	{
		// Get Action defition
		function := newRotateStartMemberAction
		action := api.ActionTypeRotateStartMember

		// Wrap action main function
		function = withActionStartFailureGracePeriod(function, 60*time.Second)

		// Register action
		registerAction(action, function)
	}

	// RotateStopMember
	{
		// Get Action defition
		function := newRotateStopMemberAction
		action := api.ActionTypeRotateStopMember

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RuntimeContainerArgsLogLevelUpdate
	{
		// Get Action defition
		function := newRuntimeContainerArgsLogLevelUpdateAction
		action := api.ActionTypeRuntimeContainerArgsLogLevelUpdate

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RuntimeContainerImageUpdate
	{
		// Get Action defition
		function := newRuntimeContainerImageUpdateAction
		action := api.ActionTypeRuntimeContainerImageUpdate

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RuntimeContainerSyncTolerations
	{
		// Get Action defition
		function := newRuntimeContainerSyncTolerationsAction
		action := api.ActionTypeRuntimeContainerSyncTolerations

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// SetCondition
	{
		// Get Action defition
		function := newSetConditionAction
		action := api.ActionTypeSetCondition

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// SetConditionV2
	{
		// Get Action defition
		function := newSetConditionV2Action
		action := api.ActionTypeSetConditionV2

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// SetCurrentImage
	{
		// Get Action defition
		function := newSetCurrentImageAction
		action := api.ActionTypeSetCurrentImage

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// SetCurrentMemberArch
	{
		// Get Action defition
		function := newSetCurrentMemberArchAction
		action := api.ActionTypeSetCurrentMemberArch

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// SetMaintenanceCondition
	{
		// Get Action defition
		function := newSetMaintenanceConditionAction
		action := api.ActionTypeSetMaintenanceCondition

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// SetMemberCondition
	{
		// Get Action defition
		function := newSetMemberConditionAction
		action := api.ActionTypeSetMemberCondition

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// SetMemberConditionV2
	{
		// Get Action defition
		function := newSetMemberConditionV2Action
		action := api.ActionTypeSetMemberConditionV2

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// SetMemberCurrentImage
	{
		// Get Action defition
		function := newSetMemberCurrentImageAction
		action := api.ActionTypeSetMemberCurrentImage

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// ShutdownMember
	{
		// Get Action defition
		function := newShutdownMemberAction
		action := api.ActionTypeShutdownMember

		// Wrap action main function
		function = withActionStartFailureGracePeriod(function, 60*time.Second)

		// Register action
		registerAction(action, function)
	}

	// TLSKeyStatusUpdate
	{
		// Get Action defition
		function := newTLSKeyStatusUpdateAction
		action := api.ActionTypeTLSKeyStatusUpdate

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// TLSPropagated
	{
		// Get Action defition
		function := newTLSPropagatedAction
		action := api.ActionTypeTLSPropagated

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// TimezoneSecretSet
	{
		// Get Action defition
		function := newTimezoneSecretSetAction
		action := api.ActionTypeTimezoneSecretSet

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// TopologyDisable
	{
		// Get Action defition
		function := newTopologyDisableAction
		action := api.ActionTypeTopologyDisable

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// TopologyEnable
	{
		// Get Action defition
		function := newTopologyEnableAction
		action := api.ActionTypeTopologyEnable

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// TopologyMemberAssignment
	{
		// Get Action defition
		function := newTopologyMemberAssignmentAction
		action := api.ActionTypeTopologyMemberAssignment

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// TopologyZonesUpdate
	{
		// Get Action defition
		function := newTopologyZonesUpdateAction
		action := api.ActionTypeTopologyZonesUpdate

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// UpToDateUpdate
	{
		// Get Action defition
		function := newUpToDateUpdateAction
		action := api.ActionTypeUpToDateUpdate

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// UpdateTLSSNI
	{
		// Get Action defition
		function := newUpdateTLSSNIAction
		action := api.ActionTypeUpdateTLSSNI

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// UpgradeMember
	{
		// Get Action defition
		function := newUpgradeMemberAction
		action := api.ActionTypeUpgradeMember

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// WaitForMemberInSync
	{
		// Get Action defition
		function := newWaitForMemberInSyncAction
		action := api.ActionTypeWaitForMemberInSync

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// WaitForMemberReady
	{
		// Get Action defition
		function := newWaitForMemberReadyAction
		action := api.ActionTypeWaitForMemberReady

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// WaitForMemberUp
	{
		// Get Action defition
		function := newWaitForMemberUpAction
		action := api.ActionTypeWaitForMemberUp

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}
}
