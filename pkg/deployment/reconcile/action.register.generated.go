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

	_ Action        = &actionDisableMaintenance{}
	_ actionFactory = newDisableMaintenanceAction

	_ Action        = &actionDisableMemberMaintenance{}
	_ actionFactory = newDisableMemberMaintenanceAction

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

	_ Action        = &actionEnforceResignLeadership{}
	_ actionFactory = newEnforceResignLeadershipAction

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

	_ Action        = &actionMemberStatusSync{}
	_ actionFactory = newMemberStatusSyncAction

	_ Action        = &actionPVCResize{}
	_ actionFactory = newPVCResizeAction

	_ Action        = &actionPVCResized{}
	_ actionFactory = newPVCResizedAction

	_ Action        = &actionPlaceHolder{}
	_ actionFactory = newPlaceHolderAction

	_ Action        = &actionRebalancerCheck{}
	_ actionFactory = newRebalancerCheckAction

	_ Action        = &actionRebalancerCheckV2{}
	_ actionFactory = newRebalancerCheckV2Action

	_ Action        = &actionRebalancerClean{}
	_ actionFactory = newRebalancerCleanAction

	_ Action        = &actionRebalancerCleanV2{}
	_ actionFactory = newRebalancerCleanV2Action

	_ Action        = &actionRebalancerGenerate{}
	_ actionFactory = newRebalancerGenerateAction

	_ Action        = &actionRebalancerGenerateV2{}
	_ actionFactory = newRebalancerGenerateV2Action

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

	_ Action        = &actionSetConditionV2{}
	_ actionFactory = newSetConditionV2Action

	_ Action        = &actionSetCurrentImage{}
	_ actionFactory = newSetCurrentImageAction

	_ Action        = &actionSetCurrentMemberArch{}
	_ actionFactory = newSetCurrentMemberArchAction

	_ Action        = &actionSetMaintenanceCondition{}
	_ actionFactory = newSetMaintenanceConditionAction

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
		// Get Action type
		action := api.ActionTypeAddMember

		// Get Action defition
		function := newAddMemberAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// AppendTLSCACertificate
	{
		// Get Action type
		action := api.ActionTypeAppendTLSCACertificate

		// Get Action defition
		function := newAppendTLSCACertificateAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// ArangoMemberUpdatePodSpec
	{
		// Get Action type
		action := api.ActionTypeArangoMemberUpdatePodSpec

		// Get Action defition
		function := newArangoMemberUpdatePodSpecAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// ArangoMemberUpdatePodStatus
	{
		// Get Action type
		action := api.ActionTypeArangoMemberUpdatePodStatus

		// Get Action defition
		function := newArangoMemberUpdatePodStatusAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// BackupRestore
	{
		// Get Action type
		action := api.ActionTypeBackupRestore

		// Get Action defition
		function := newBackupRestoreAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// BackupRestoreClean
	{
		// Get Action type
		action := api.ActionTypeBackupRestoreClean

		// Get Action defition
		function := newBackupRestoreCleanAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// BootstrapSetPassword
	{
		// Get Action type
		action := api.ActionTypeBootstrapSetPassword

		// Get Action defition
		function := newBootstrapSetPasswordAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// BootstrapUpdate
	{
		// Get Action type
		action := api.ActionTypeBootstrapUpdate

		// Get Action defition
		function := newBootstrapUpdateAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// CleanMemberService
	{
		// Get Action type
		action := api.ActionTypeCleanMemberService

		// Get Action defition
		function := newCleanMemberServiceAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// CleanOutMember
	{
		// Get Action type
		action := api.ActionTypeCleanOutMember

		// Get Action defition
		function := newCleanOutMemberAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// CleanTLSCACertificate
	{
		// Get Action type
		action := api.ActionTypeCleanTLSCACertificate

		// Get Action defition
		function := newCleanTLSCACertificateAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// CleanTLSKeyfileCertificate
	{
		// Get Action type
		action := api.ActionTypeCleanTLSKeyfileCertificate

		// Get Action defition
		function := newCleanTLSKeyfileCertificateAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// ClusterMemberCleanup
	{
		// Get Action type
		action := api.ActionTypeClusterMemberCleanup

		// Get Action defition
		function := newClusterMemberCleanupAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// DisableClusterScaling
	{
		// Get Action type
		// nolint:staticcheck
		action := api.ActionTypeDisableClusterScaling

		// Get Empty (Deprecated) Action Definition
		function := newDeprecatedAction

		// Register action
		registerAction(action, function)
	}

	// DisableMaintenance
	{
		// Get Action type
		action := api.ActionTypeDisableMaintenance

		// Get Action defition
		function := newDisableMaintenanceAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// DisableMemberMaintenance
	{
		// Get Action type
		action := api.ActionTypeDisableMemberMaintenance

		// Get Action defition
		function := newDisableMemberMaintenanceAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// EnableClusterScaling
	{
		// Get Action type
		// nolint:staticcheck
		action := api.ActionTypeEnableClusterScaling

		// Get Empty (Deprecated) Action Definition
		function := newDeprecatedAction

		// Register action
		registerAction(action, function)
	}

	// EnableMaintenance
	{
		// Get Action type
		action := api.ActionTypeEnableMaintenance

		// Get Action defition
		function := newEnableMaintenanceAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// EnableMemberMaintenance
	{
		// Get Action type
		action := api.ActionTypeEnableMemberMaintenance

		// Get Action defition
		function := newEnableMemberMaintenanceAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// EncryptionKeyAdd
	{
		// Get Action type
		action := api.ActionTypeEncryptionKeyAdd

		// Get Action defition
		function := newEncryptionKeyAddAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// EncryptionKeyPropagated
	{
		// Get Action type
		action := api.ActionTypeEncryptionKeyPropagated

		// Get Action defition
		function := newEncryptionKeyPropagatedAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// EncryptionKeyRefresh
	{
		// Get Action type
		action := api.ActionTypeEncryptionKeyRefresh

		// Get Action defition
		function := newEncryptionKeyRefreshAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// EncryptionKeyRemove
	{
		// Get Action type
		action := api.ActionTypeEncryptionKeyRemove

		// Get Action defition
		function := newEncryptionKeyRemoveAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// EncryptionKeyStatusUpdate
	{
		// Get Action type
		action := api.ActionTypeEncryptionKeyStatusUpdate

		// Get Action defition
		function := newEncryptionKeyStatusUpdateAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// EnforceResignLeadership
	{
		// Get Action type
		action := api.ActionTypeEnforceResignLeadership

		// Get Action defition
		function := newEnforceResignLeadershipAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// Idle
	{
		// Get Action type
		action := api.ActionTypeIdle

		// Get Action defition
		function := newIdleAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// JWTAdd
	{
		// Get Action type
		action := api.ActionTypeJWTAdd

		// Get Action defition
		function := newJWTAddAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// JWTClean
	{
		// Get Action type
		action := api.ActionTypeJWTClean

		// Get Action defition
		function := newJWTCleanAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// JWTPropagated
	{
		// Get Action type
		action := api.ActionTypeJWTPropagated

		// Get Action defition
		function := newJWTPropagatedAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// JWTRefresh
	{
		// Get Action type
		action := api.ActionTypeJWTRefresh

		// Get Action defition
		function := newJWTRefreshAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// JWTSetActive
	{
		// Get Action type
		action := api.ActionTypeJWTSetActive

		// Get Action defition
		function := newJWTSetActiveAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// JWTStatusUpdate
	{
		// Get Action type
		action := api.ActionTypeJWTStatusUpdate

		// Get Action defition
		function := newJWTStatusUpdateAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// KillMemberPod
	{
		// Get Action type
		action := api.ActionTypeKillMemberPod

		// Get Action defition
		function := newKillMemberPodAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// LicenseSet
	{
		// Get Action type
		action := api.ActionTypeLicenseSet

		// Get Action defition
		function := newLicenseSetAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// MarkToRemoveMember
	{
		// Get Action type
		action := api.ActionTypeMarkToRemoveMember

		// Get Action defition
		function := newMarkToRemoveMemberAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// MemberPhaseUpdate
	{
		// Get Action type
		action := api.ActionTypeMemberPhaseUpdate

		// Get Action defition
		function := newMemberPhaseUpdateAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// MemberRIDUpdate
	{
		// Get Action type
		// nolint:staticcheck
		action := api.ActionTypeMemberRIDUpdate

		// Get Empty (Deprecated) Action Definition
		function := newDeprecatedAction

		// Register action
		registerAction(action, function)
	}

	// MemberStatusSync
	{
		// Get Action type
		action := api.ActionTypeMemberStatusSync

		// Get Action defition
		function := newMemberStatusSyncAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// PVCResize
	{
		// Get Action type
		action := api.ActionTypePVCResize

		// Get Action defition
		function := newPVCResizeAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// PVCResized
	{
		// Get Action type
		action := api.ActionTypePVCResized

		// Get Action defition
		function := newPVCResizedAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// PlaceHolder
	{
		// Get Action type
		action := api.ActionTypePlaceHolder

		// Get Action defition
		function := newPlaceHolderAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RebalancerCheck
	{
		// Get Action type
		action := api.ActionTypeRebalancerCheck

		// Get Action defition
		function := newRebalancerCheckAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RebalancerCheckV2
	{
		// Get Action type
		action := api.ActionTypeRebalancerCheckV2

		// Get Action defition
		function := newRebalancerCheckV2Action

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RebalancerClean
	{
		// Get Action type
		action := api.ActionTypeRebalancerClean

		// Get Action defition
		function := newRebalancerCleanAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RebalancerCleanV2
	{
		// Get Action type
		action := api.ActionTypeRebalancerCleanV2

		// Get Action defition
		function := newRebalancerCleanV2Action

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RebalancerGenerate
	{
		// Get Action type
		action := api.ActionTypeRebalancerGenerate

		// Get Action defition
		function := newRebalancerGenerateAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RebalancerGenerateV2
	{
		// Get Action type
		action := api.ActionTypeRebalancerGenerateV2

		// Get Action defition
		function := newRebalancerGenerateV2Action

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RebuildOutSyncedShards
	{
		// Get Action type
		action := api.ActionTypeRebuildOutSyncedShards

		// Get Action defition
		function := newRebuildOutSyncedShardsAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RecreateMember
	{
		// Get Action type
		action := api.ActionTypeRecreateMember

		// Get Action defition
		function := newRecreateMemberAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RefreshTLSKeyfileCertificate
	{
		// Get Action type
		action := api.ActionTypeRefreshTLSKeyfileCertificate

		// Get Action defition
		function := newRefreshTLSKeyfileCertificateAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RemoveMember
	{
		// Get Action type
		action := api.ActionTypeRemoveMember

		// Get Action defition
		function := newRemoveMemberAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RemoveMemberPVC
	{
		// Get Action type
		action := api.ActionTypeRemoveMemberPVC

		// Get Action defition
		function := newRemoveMemberPVCAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RenewTLSCACertificate
	{
		// Get Action type
		action := api.ActionTypeRenewTLSCACertificate

		// Get Action defition
		function := newRenewTLSCACertificateAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RenewTLSCertificate
	{
		// Get Action type
		action := api.ActionTypeRenewTLSCertificate

		// Get Action defition
		function := newRenewTLSCertificateAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// ResignLeadership
	{
		// Get Action type
		action := api.ActionTypeResignLeadership

		// Get Action defition
		function := newResignLeadershipAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// ResourceSync
	{
		// Get Action type
		action := api.ActionTypeResourceSync

		// Get Action defition
		function := newResourceSyncAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RotateMember
	{
		// Get Action type
		action := api.ActionTypeRotateMember

		// Get Action defition
		function := newRotateMemberAction

		// Wrap action main function

		// With StartupFailureGracePeriod
		function = withActionStartFailureGracePeriod(function, 60*time.Second)

		// Register action
		registerAction(action, function)
	}

	// RotateStartMember
	{
		// Get Action type
		action := api.ActionTypeRotateStartMember

		// Get Action defition
		function := newRotateStartMemberAction

		// Wrap action main function

		// With StartupFailureGracePeriod
		function = withActionStartFailureGracePeriod(function, 60*time.Second)

		// Register action
		registerAction(action, function)
	}

	// RotateStopMember
	{
		// Get Action type
		action := api.ActionTypeRotateStopMember

		// Get Action defition
		function := newRotateStopMemberAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RuntimeContainerArgsLogLevelUpdate
	{
		// Get Action type
		action := api.ActionTypeRuntimeContainerArgsLogLevelUpdate

		// Get Action defition
		function := newRuntimeContainerArgsLogLevelUpdateAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RuntimeContainerImageUpdate
	{
		// Get Action type
		action := api.ActionTypeRuntimeContainerImageUpdate

		// Get Action defition
		function := newRuntimeContainerImageUpdateAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// RuntimeContainerSyncTolerations
	{
		// Get Action type
		action := api.ActionTypeRuntimeContainerSyncTolerations

		// Get Action defition
		function := newRuntimeContainerSyncTolerationsAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// SetCondition
	{
		// Get Action type
		// nolint:staticcheck
		action := api.ActionTypeSetCondition

		// Get Empty (Deprecated) Action Definition
		function := newDeprecatedAction

		// Register action
		registerAction(action, function)
	}

	// SetConditionV2
	{
		// Get Action type
		action := api.ActionTypeSetConditionV2

		// Get Action defition
		function := newSetConditionV2Action

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// SetCurrentImage
	{
		// Get Action type
		action := api.ActionTypeSetCurrentImage

		// Get Action defition
		function := newSetCurrentImageAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// SetCurrentMemberArch
	{
		// Get Action type
		action := api.ActionTypeSetCurrentMemberArch

		// Get Action defition
		function := newSetCurrentMemberArchAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// SetMaintenanceCondition
	{
		// Get Action type
		action := api.ActionTypeSetMaintenanceCondition

		// Get Action defition
		function := newSetMaintenanceConditionAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// SetMemberCondition
	{
		// Get Action type
		// nolint:staticcheck
		action := api.ActionTypeSetMemberCondition

		// Get Empty (Deprecated) Action Definition
		function := newDeprecatedAction

		// Register action
		registerAction(action, function)
	}

	// SetMemberConditionV2
	{
		// Get Action type
		action := api.ActionTypeSetMemberConditionV2

		// Get Action defition
		function := newSetMemberConditionV2Action

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// SetMemberCurrentImage
	{
		// Get Action type
		action := api.ActionTypeSetMemberCurrentImage

		// Get Action defition
		function := newSetMemberCurrentImageAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// ShutdownMember
	{
		// Get Action type
		action := api.ActionTypeShutdownMember

		// Get Action defition
		function := newShutdownMemberAction

		// Wrap action main function

		// With StartupFailureGracePeriod
		function = withActionStartFailureGracePeriod(function, 60*time.Second)

		// Register action
		registerAction(action, function)
	}

	// TLSKeyStatusUpdate
	{
		// Get Action type
		action := api.ActionTypeTLSKeyStatusUpdate

		// Get Action defition
		function := newTLSKeyStatusUpdateAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// TLSPropagated
	{
		// Get Action type
		action := api.ActionTypeTLSPropagated

		// Get Action defition
		function := newTLSPropagatedAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// TimezoneSecretSet
	{
		// Get Action type
		action := api.ActionTypeTimezoneSecretSet

		// Get Action defition
		function := newTimezoneSecretSetAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// TopologyDisable
	{
		// Get Action type
		action := api.ActionTypeTopologyDisable

		// Get Action defition
		function := newTopologyDisableAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// TopologyEnable
	{
		// Get Action type
		action := api.ActionTypeTopologyEnable

		// Get Action defition
		function := newTopologyEnableAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// TopologyMemberAssignment
	{
		// Get Action type
		action := api.ActionTypeTopologyMemberAssignment

		// Get Action defition
		function := newTopologyMemberAssignmentAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// TopologyZonesUpdate
	{
		// Get Action type
		action := api.ActionTypeTopologyZonesUpdate

		// Get Action defition
		function := newTopologyZonesUpdateAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// UpToDateUpdate
	{
		// Get Action type
		action := api.ActionTypeUpToDateUpdate

		// Get Action defition
		function := newUpToDateUpdateAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// UpdateTLSSNI
	{
		// Get Action type
		action := api.ActionTypeUpdateTLSSNI

		// Get Action defition
		function := newUpdateTLSSNIAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// UpgradeMember
	{
		// Get Action type
		action := api.ActionTypeUpgradeMember

		// Get Action defition
		function := newUpgradeMemberAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// WaitForMemberInSync
	{
		// Get Action type
		action := api.ActionTypeWaitForMemberInSync

		// Get Action defition
		function := newWaitForMemberInSyncAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// WaitForMemberReady
	{
		// Get Action type
		action := api.ActionTypeWaitForMemberReady

		// Get Action defition
		function := newWaitForMemberReadyAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

	// WaitForMemberUp
	{
		// Get Action type
		action := api.ActionTypeWaitForMemberUp

		// Get Action defition
		function := newWaitForMemberUpAction

		// Wrap action main function

		// Register action
		registerAction(action, function)
	}

}
