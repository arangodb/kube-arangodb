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
	})

	t.Run("AppendTLSCACertificate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeAppendTLSCACertificate)
	})

	t.Run("ArangoMemberUpdatePodSpec", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeArangoMemberUpdatePodSpec)
	})

	t.Run("ArangoMemberUpdatePodStatus", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeArangoMemberUpdatePodStatus)
	})

	t.Run("BackupRestore", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeBackupRestore)
	})

	t.Run("BackupRestoreClean", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeBackupRestoreClean)
	})

	t.Run("BootstrapSetPassword", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeBootstrapSetPassword)
	})

	t.Run("BootstrapUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeBootstrapUpdate)
	})

	t.Run("CleanOutMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeCleanOutMember)
	})

	t.Run("CleanTLSCACertificate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeCleanTLSCACertificate)
	})

	t.Run("CleanTLSKeyfileCertificate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeCleanTLSKeyfileCertificate)
	})

	t.Run("ClusterMemberCleanup", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeClusterMemberCleanup)
	})

	t.Run("DisableClusterScaling", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeDisableClusterScaling)
	})

	t.Run("DisableMaintenance", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeDisableMaintenance)
	})

	t.Run("DisableMemberMaintenance", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeDisableMemberMaintenance)
	})

	t.Run("EnableClusterScaling", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeEnableClusterScaling)
	})

	t.Run("EnableMaintenance", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeEnableMaintenance)
	})

	t.Run("EnableMemberMaintenance", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeEnableMemberMaintenance)
	})

	t.Run("EncryptionKeyAdd", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeEncryptionKeyAdd)
	})

	t.Run("EncryptionKeyPropagated", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeEncryptionKeyPropagated)
	})

	t.Run("EncryptionKeyRefresh", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeEncryptionKeyRefresh)
	})

	t.Run("EncryptionKeyRemove", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeEncryptionKeyRemove)
	})

	t.Run("EncryptionKeyStatusUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeEncryptionKeyStatusUpdate)
	})

	t.Run("Idle", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeIdle)
	})

	t.Run("JWTAdd", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeJWTAdd)
	})

	t.Run("JWTClean", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeJWTClean)
	})

	t.Run("JWTPropagated", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeJWTPropagated)
	})

	t.Run("JWTRefresh", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeJWTRefresh)
	})

	t.Run("JWTSetActive", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeJWTSetActive)
	})

	t.Run("JWTStatusUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeJWTStatusUpdate)
	})

	t.Run("KillMemberPod", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeKillMemberPod)
	})

	t.Run("LicenseSet", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeLicenseSet)
	})

	t.Run("MarkToRemoveMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeMarkToRemoveMember)
	})

	t.Run("MemberPhaseUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeMemberPhaseUpdate)
	})

	t.Run("MemberRIDUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeMemberRIDUpdate)
	})

	t.Run("PVCResize", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypePVCResize)
	})

	t.Run("PVCResized", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypePVCResized)
	})

	t.Run("PlaceHolder", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypePlaceHolder)
	})

	t.Run("RebalancerCheck", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRebalancerCheck)
	})

	t.Run("RebalancerClean", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRebalancerClean)
	})

	t.Run("RebalancerGenerate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRebalancerGenerate)
	})

	t.Run("RecreateMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRecreateMember)
	})

	t.Run("RefreshTLSKeyfileCertificate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRefreshTLSKeyfileCertificate)
	})

	t.Run("RemoveMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRemoveMember)
	})

	t.Run("RenewTLSCACertificate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRenewTLSCACertificate)
	})

	t.Run("RenewTLSCertificate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRenewTLSCertificate)
	})

	t.Run("ResignLeadership", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeResignLeadership)
	})

	t.Run("ResourceSync", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeResourceSync)
	})

	t.Run("RotateMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRotateMember)
		ActionsWrapWithActionStartFailureGracePeriod(t, api.ActionTypeRotateMember, 60*time.Second)
	})

	t.Run("RotateStartMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRotateStartMember)
		ActionsWrapWithActionStartFailureGracePeriod(t, api.ActionTypeRotateStartMember, 60*time.Second)
	})

	t.Run("RotateStopMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRotateStopMember)
	})

	t.Run("RuntimeContainerArgsLogLevelUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRuntimeContainerArgsLogLevelUpdate)
	})

	t.Run("RuntimeContainerImageUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeRuntimeContainerImageUpdate)
	})

	t.Run("SetCondition", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeSetCondition)
	})

	t.Run("SetConditionV2", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeSetConditionV2)
	})

	t.Run("SetCurrentImage", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeSetCurrentImage)
	})

	t.Run("SetMaintenanceCondition", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeSetMaintenanceCondition)
	})

	t.Run("SetMemberCondition", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeSetMemberCondition)
	})

	t.Run("SetMemberConditionV2", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeSetMemberConditionV2)
	})

	t.Run("SetMemberCurrentImage", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeSetMemberCurrentImage)
	})

	t.Run("ShutdownMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeShutdownMember)
		ActionsWrapWithActionStartFailureGracePeriod(t, api.ActionTypeShutdownMember, 60*time.Second)
	})

	t.Run("TLSKeyStatusUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeTLSKeyStatusUpdate)
	})

	t.Run("TLSPropagated", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeTLSPropagated)
	})

	t.Run("TimezoneSecretSet", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeTimezoneSecretSet)
	})

	t.Run("TopologyDisable", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeTopologyDisable)
	})

	t.Run("TopologyEnable", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeTopologyEnable)
	})

	t.Run("TopologyMemberAssignment", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeTopologyMemberAssignment)
	})

	t.Run("TopologyZonesUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeTopologyZonesUpdate)
	})

	t.Run("UpToDateUpdate", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeUpToDateUpdate)
	})

	t.Run("UpdateTLSSNI", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeUpdateTLSSNI)
	})

	t.Run("UpgradeMember", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeUpgradeMember)
	})

	t.Run("WaitForMemberInSync", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeWaitForMemberInSync)
	})

	t.Run("WaitForMemberUp", func(t *testing.T) {
		ActionsExistence(t, api.ActionTypeWaitForMemberUp)
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
