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
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func Test_Actions(t *testing.T) {
	// Iterate over all actions

	t.Run("AddMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeAddMember)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeAddMember.Internal())
		})
	})

	t.Run("AppendTLSCACertificate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeAppendTLSCACertificate)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeAppendTLSCACertificate.Internal())
		})
	})

	t.Run("ArangoMemberUpdatePodSpec", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeArangoMemberUpdatePodSpec)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeArangoMemberUpdatePodSpec.Internal())
		})
	})

	t.Run("ArangoMemberUpdatePodStatus", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeArangoMemberUpdatePodStatus)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeArangoMemberUpdatePodStatus.Internal())
		})
	})

	t.Run("BackupRestore", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeBackupRestore)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeBackupRestore.Internal())
		})
	})

	t.Run("BackupRestoreClean", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeBackupRestoreClean)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeBackupRestoreClean.Internal())
		})
	})

	t.Run("BootstrapSetPassword", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeBootstrapSetPassword)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeBootstrapSetPassword.Internal())
		})
	})

	t.Run("BootstrapUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeBootstrapUpdate)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeBootstrapUpdate.Internal())
		})
	})

	t.Run("CleanOutMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeCleanOutMember)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeCleanOutMember.Internal())
		})
	})

	t.Run("CleanTLSCACertificate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeCleanTLSCACertificate)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeCleanTLSCACertificate.Internal())
		})
	})

	t.Run("CleanTLSKeyfileCertificate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeCleanTLSKeyfileCertificate)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeCleanTLSKeyfileCertificate.Internal())
		})
	})

	t.Run("ClusterMemberCleanup", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeClusterMemberCleanup)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeClusterMemberCleanup.Internal())
		})
	})

	t.Run("DisableClusterScaling", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeDisableClusterScaling)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeDisableClusterScaling.Internal())
		})
	})

	t.Run("DisableMaintenance", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeDisableMaintenance)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeDisableMaintenance.Internal())
		})
	})

	t.Run("DisableMemberMaintenance", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeDisableMemberMaintenance)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeDisableMemberMaintenance.Internal())
		})
	})

	t.Run("EnableClusterScaling", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeEnableClusterScaling)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeEnableClusterScaling.Internal())
		})
	})

	t.Run("EnableMaintenance", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeEnableMaintenance)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeEnableMaintenance.Internal())
		})
	})

	t.Run("EnableMemberMaintenance", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeEnableMemberMaintenance)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeEnableMemberMaintenance.Internal())
		})
	})

	t.Run("EncryptionKeyAdd", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeEncryptionKeyAdd)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeEncryptionKeyAdd.Internal())
		})
	})

	t.Run("EncryptionKeyPropagated", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeEncryptionKeyPropagated)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeEncryptionKeyPropagated.Internal())
		})
	})

	t.Run("EncryptionKeyRefresh", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeEncryptionKeyRefresh)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeEncryptionKeyRefresh.Internal())
		})
	})

	t.Run("EncryptionKeyRemove", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeEncryptionKeyRemove)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeEncryptionKeyRemove.Internal())
		})
	})

	t.Run("EncryptionKeyStatusUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeEncryptionKeyStatusUpdate)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeEncryptionKeyStatusUpdate.Internal())
		})
	})

	t.Run("Idle", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeIdle)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeIdle.Internal())
		})
	})

	t.Run("JWTAdd", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeJWTAdd)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeJWTAdd.Internal())
		})
	})

	t.Run("JWTClean", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeJWTClean)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeJWTClean.Internal())
		})
	})

	t.Run("JWTPropagated", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeJWTPropagated)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeJWTPropagated.Internal())
		})
	})

	t.Run("JWTRefresh", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeJWTRefresh)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeJWTRefresh.Internal())
		})
	})

	t.Run("JWTSetActive", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeJWTSetActive)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeJWTSetActive.Internal())
		})
	})

	t.Run("JWTStatusUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeJWTStatusUpdate)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeJWTStatusUpdate.Internal())
		})
	})

	t.Run("KillMemberPod", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeKillMemberPod)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeKillMemberPod.Internal())
		})
	})

	t.Run("LicenseSet", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeLicenseSet)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeLicenseSet.Internal())
		})
	})

	t.Run("MarkToRemoveMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeMarkToRemoveMember)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeMarkToRemoveMember.Internal())
		})
	})

	t.Run("MemberPhaseUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeMemberPhaseUpdate)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeMemberPhaseUpdate.Internal())
		})
	})

	t.Run("MemberRIDUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeMemberRIDUpdate)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeMemberRIDUpdate.Internal())
		})
	})

	t.Run("PVCResize", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypePVCResize)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypePVCResize.Internal())
		})
	})

	t.Run("PVCResized", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypePVCResized)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypePVCResized.Internal())
		})
	})

	t.Run("PlaceHolder", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypePlaceHolder)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypePlaceHolder.Internal())
		})
	})

	t.Run("RebalancerCheck", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRebalancerCheck)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeRebalancerCheck.Internal())
		})
	})

	t.Run("RebalancerClean", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRebalancerClean)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeRebalancerClean.Internal())
		})
	})

	t.Run("RebalancerGenerate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRebalancerGenerate)
		t.Run("Internal", func(t *testing.T) {
			require.True(t, api.ActionTypeRebalancerGenerate.Internal())
		})
	})

	t.Run("RecreateMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRecreateMember)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeRecreateMember.Internal())
		})
	})

	t.Run("RefreshTLSKeyfileCertificate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRefreshTLSKeyfileCertificate)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeRefreshTLSKeyfileCertificate.Internal())
		})
	})

	t.Run("RemoveMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRemoveMember)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeRemoveMember.Internal())
		})
	})

	t.Run("RenewTLSCACertificate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRenewTLSCACertificate)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeRenewTLSCACertificate.Internal())
		})
	})

	t.Run("RenewTLSCertificate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRenewTLSCertificate)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeRenewTLSCertificate.Internal())
		})
	})

	t.Run("ResignLeadership", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeResignLeadership)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeResignLeadership.Internal())
		})
	})

	t.Run("ResourceSync", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeResourceSync)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeResourceSync.Internal())
		})
	})

	t.Run("RotateMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRotateMember)
		ActionsWrapWithActionStartFailureGracePeriod(t, api.ActionTypeRotateMember, 60*time.Second)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeRotateMember.Internal())
		})
	})

	t.Run("RotateStartMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRotateStartMember)
		ActionsWrapWithActionStartFailureGracePeriod(t, api.ActionTypeRotateStartMember, 60*time.Second)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeRotateStartMember.Internal())
		})
	})

	t.Run("RotateStopMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRotateStopMember)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeRotateStopMember.Internal())
		})
	})

	t.Run("RuntimeContainerArgsLogLevelUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRuntimeContainerArgsLogLevelUpdate)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeRuntimeContainerArgsLogLevelUpdate.Internal())
		})
	})

	t.Run("RuntimeContainerImageUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRuntimeContainerImageUpdate)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeRuntimeContainerImageUpdate.Internal())
		})
	})

	t.Run("RuntimeContainerSyncTolerations", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRuntimeContainerSyncTolerations)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeRuntimeContainerSyncTolerations.Internal())
		})
	})

	t.Run("SetCondition", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeSetCondition)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeSetCondition.Internal())
		})
	})

	t.Run("SetConditionV2", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeSetConditionV2)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeSetConditionV2.Internal())
		})
	})

	t.Run("SetCurrentImage", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeSetCurrentImage)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeSetCurrentImage.Internal())
		})
	})

	t.Run("SetCurrentMemberArch", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeSetCurrentMemberArch)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeSetCurrentMemberArch.Internal())
		})
	})

	t.Run("SetMaintenanceCondition", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeSetMaintenanceCondition)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeSetMaintenanceCondition.Internal())
		})
	})

	t.Run("SetMemberCondition", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeSetMemberCondition)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeSetMemberCondition.Internal())
		})
	})

	t.Run("SetMemberConditionV2", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeSetMemberConditionV2)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeSetMemberConditionV2.Internal())
		})
	})

	t.Run("SetMemberCurrentImage", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeSetMemberCurrentImage)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeSetMemberCurrentImage.Internal())
		})
	})

	t.Run("ShutdownMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeShutdownMember)
		ActionsWrapWithActionStartFailureGracePeriod(t, api.ActionTypeShutdownMember, 60*time.Second)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeShutdownMember.Internal())
		})
	})

	t.Run("TLSKeyStatusUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeTLSKeyStatusUpdate)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeTLSKeyStatusUpdate.Internal())
		})
	})

	t.Run("TLSPropagated", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeTLSPropagated)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeTLSPropagated.Internal())
		})
	})

	t.Run("TimezoneSecretSet", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeTimezoneSecretSet)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeTimezoneSecretSet.Internal())
		})
	})

	t.Run("TopologyDisable", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeTopologyDisable)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeTopologyDisable.Internal())
		})
	})

	t.Run("TopologyEnable", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeTopologyEnable)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeTopologyEnable.Internal())
		})
	})

	t.Run("TopologyMemberAssignment", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeTopologyMemberAssignment)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeTopologyMemberAssignment.Internal())
		})
	})

	t.Run("TopologyZonesUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeTopologyZonesUpdate)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeTopologyZonesUpdate.Internal())
		})
	})

	t.Run("UpToDateUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeUpToDateUpdate)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeUpToDateUpdate.Internal())
		})
	})

	t.Run("UpdateTLSSNI", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeUpdateTLSSNI)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeUpdateTLSSNI.Internal())
		})
	})

	t.Run("UpgradeMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeUpgradeMember)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeUpgradeMember.Internal())
		})
	})

	t.Run("WaitForMemberInSync", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeWaitForMemberInSync)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeWaitForMemberInSync.Internal())
		})
	})

	t.Run("WaitForMemberReady", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeWaitForMemberReady)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeWaitForMemberReady.Internal())
		})
	})

	t.Run("WaitForMemberUp", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeWaitForMemberUp)
		t.Run("Internal", func(t *testing.T) {
			require.False(t, api.ActionTypeWaitForMemberUp.Internal())
		})
	})
}

func ActionsExistence(t *testing.T, action api.ActionType) {
	t.Run("Existence", func(t *testing.T) {
		_, ok := getActionFactory(action)
		require.True(t, ok)
	})
}

func ActionsWrapWithActionStartFailureGracePeriod(t *testing.T, action api.ActionType, timeout time.Duration) {
	t.Run("WrapWithActionStartFailureGracePeriod", func(t *testing.T) {
		f, ok := getActionFactory(action)
		require.True(t, ok)

		a := extractAction(f)
		require.NotNil(t, a)

		z, ok := a.(*actionStartFailureGracePeriod)
		require.True(t, ok)

		require.Equal(t, z.failureGracePeriod, timeout)
	})
}

func extractAction(f actionFactory) Action {
	return f(api.Action{}, nil)
}
