---
layout: page
title: List of Plan Actions
nav_order: 11
---

# ArangoDB Operator Actions

## List

[START_INJECT]: # (actionsTable)

| Action | Timeout | Description | Internal | Optional | Edition |
|:---:|:---:|:---:|:---:|:---:|:---:|
| AddMember | 10m0s | Adds new member to the Member list | no | no | Community & Enterprise |
| AppendTLSCACertificate | 30m0s | Append Certificate into CA TrustStore | no | no | Enterprise Only |
| ArangoMemberUpdatePodSpec | 10m0s | Propagate Member Pod spec (requested) | no | no | Community & Enterprise |
| ArangoMemberUpdatePodStatus | 10m0s | Propagate Member Pod status (current) | no | no | Community & Enterprise |
| BackupRestore | 15m0s | Restore selected Backup | no | no | Enterprise Only |
| BackupRestoreClean | 15m0s | Clean restore status in case of restore spec change | no | no | Enterprise Only |
| BootstrapSetPassword | 10m0s | Change password during bootstrap procedure | no | no | Community & Enterprise |
| BootstrapUpdate | 10m0s | Update bootstrap status | no | no | Community & Enterprise |
| CleanMemberService | 30m0s | Removes Server Service | no | no | Community & Enterprise |
| CleanOutMember | 48h0m0s | Run the CleanOut job on member | no | no | Community & Enterprise |
| CleanTLSCACertificate | 30m0s | Remove Certificate from CA TrustStore | no | no | Enterprise Only |
| CleanTLSKeyfileCertificate | 30m0s | Remove old TLS certificate from server | no | no | Enterprise Only |
| ClusterMemberCleanup | 10m0s | Remove member from Cluster if it is gone already (Coordinators) | no | no | Community & Enterprise |
| ~~DisableClusterScaling~~ | 10m0s | Disable Cluster Scaling integration | no | no | Community & Enterprise |
| DisableMaintenance | 10m0s | Disable ArangoDB maintenance mode | no | no | Community & Enterprise |
| DisableMemberMaintenance | 10m0s | Disable ArangoDB DBServer maintenance mode | no | no | Enterprise Only |
| ~~EnableClusterScaling~~ | 10m0s | Enable Cluster Scaling integration | no | no | Community & Enterprise |
| EnableMaintenance | 10m0s | Enable ArangoDB maintenance mode | no | no | Community & Enterprise |
| EnableMemberMaintenance | 10m0s | Enable ArangoDB DBServer maintenance mode | no | no | Enterprise Only |
| EncryptionKeyAdd | 10m0s | Add the encryption key to the pool | no | no | Enterprise Only |
| EncryptionKeyPropagated | 10m0s | Update condition of encryption propagation | no | no | Enterprise Only |
| EncryptionKeyRefresh | 10m0s | Refresh the encryption keys on member | no | no | Enterprise Only |
| EncryptionKeyRemove | 10m0s | Remove the encryption key to the pool | no | no | Enterprise Only |
| EncryptionKeyStatusUpdate | 10m0s | Update status of encryption propagation | no | no | Enterprise Only |
| EnforceResignLeadership | 45m0s | Run the ResignLeadership job on DBServer and checks data compatibility after | no | yes | Community & Enterprise |
| Idle | 10m0s | Define idle operation in case if preconditions are not meet | no | no | Community & Enterprise |
| JWTAdd | 10m0s | Adds new JWT to the pool | no | no | Enterprise Only |
| JWTClean | 10m0s | Remove JWT key from the pool | no | no | Enterprise Only |
| JWTPropagated | 10m0s | Update condition of JWT propagation | no | no | Enterprise Only |
| JWTRefresh | 10m0s | Refresh current JWT secrets on the member | no | no | Enterprise Only |
| JWTSetActive | 10m0s | Change active JWT key on the cluster | no | no | Enterprise Only |
| JWTStatusUpdate | 10m0s | Update status of JWT propagation | no | no | Enterprise Only |
| KillMemberPod | 10m0s | Execute Delete on Pod (put pod in Terminating state) | no | no | Community & Enterprise |
| LicenseSet | 10m0s | Update Cluster license (3.9+) | no | no | Community & Enterprise |
| MarkToRemoveMember | 10m0s | Marks member to be removed. Used when member Pod is annotated with replace annotation | no | no | Community & Enterprise |
| MemberPhaseUpdate | 10m0s | Change member phase | no | no | Community & Enterprise |
| ~~MemberRIDUpdate~~ | 10m0s | Update Run ID of member | no | no | Community & Enterprise |
| MemberStatusSync | 10m0s | Sync ArangoMember Status with ArangoDeployment Status, to keep Member information up to date | no | no | Community & Enterprise |
| PVCResize | 30m0s | Start the resize procedure. Updates PVC Requests field | no | no | Community & Enterprise |
| PVCResized | 15m0s | Waits for PVC resize to be completed | no | no | Community & Enterprise |
| PlaceHolder | 10m0s | Empty placeholder action | no | no | Community & Enterprise |
| RebalancerCheck | 10m0s | Check Rebalancer job progress | no | no | Enterprise Only |
| RebalancerCheckV2 | 10m0s | Check Rebalancer job progress | no | no | Community & Enterprise |
| RebalancerClean | 10m0s | Cleans Rebalancer jobs | no | no | Enterprise Only |
| RebalancerCleanV2 | 10m0s | Cleans Rebalancer jobs | no | no | Community & Enterprise |
| RebalancerGenerate | 10m0s | Generates the Rebalancer plan | yes | no | Enterprise Only |
| RebalancerGenerateV2 | 10m0s | Generates the Rebalancer plan | yes | no | Community & Enterprise |
| RebuildOutSyncedShards | 24h0m0s | Run Rebuild Out Synced Shards procedure for DBServers | no | no | Community & Enterprise |
| RecreateMember | 15m0s | Recreate member with same ID and Data | no | no | Community & Enterprise |
| RefreshTLSCA | 30m0s | Refresh internal CA | no | no | Enterprise Only |
| RefreshTLSKeyfileCertificate | 30m0s | Recreate Server TLS Certificate secret | no | no | Enterprise Only |
| RemoveMember | 15m0s | Removes member from the Cluster and Status | no | no | Community & Enterprise |
| RemoveMemberPVC | 15m0s | Removes member PVC and enforce recreate procedure | no | no | Community & Enterprise |
| RenewTLSCACertificate | 30m0s | Recreate Managed CA secret | no | no | Enterprise Only |
| RenewTLSCertificate | 30m0s | Recreate Server TLS Certificate secret | no | no | Enterprise Only |
| ResignLeadership | 30m0s | Run the ResignLeadership job on DBServer | no | yes | Community & Enterprise |
| ResourceSync | 10m0s | Runs the Resource sync | no | no | Community & Enterprise |
| RotateMember | 15m0s | Waits for Pod restart and recreation | no | no | Community & Enterprise |
| RotateStartMember | 15m0s | Start member rotation. After this action member is down | no | no | Community & Enterprise |
| RotateStopMember | 15m0s | Finalize member rotation. After this action member is started back | no | no | Community & Enterprise |
| RuntimeContainerArgsLogLevelUpdate | 10m0s | Change ArangoDB Member log levels in runtime | no | no | Community & Enterprise |
| RuntimeContainerImageUpdate | 10m0s | Update Container Image in runtime | no | no | Community & Enterprise |
| RuntimeContainerSyncTolerations | 10m0s | Update Pod Tolerations in runtime | no | no | Community & Enterprise |
| ~~SetCondition~~ | 10m0s | Set deployment condition | no | no | Community & Enterprise |
| SetConditionV2 | 10m0s | Set deployment condition | no | no | Community & Enterprise |
| SetCurrentImage | 6h0m0s | Update deployment current image after image discovery | no | no | Community & Enterprise |
| SetCurrentMemberArch | 10m0s | Set current member architecture | no | no | Community & Enterprise |
| SetMaintenanceCondition | 10m0s | Update ArangoDB maintenance condition | no | no | Community & Enterprise |
| ~~SetMemberCondition~~ | 10m0s | Set member condition | no | no | Community & Enterprise |
| SetMemberConditionV2 | 10m0s | Set member condition | no | no | Community & Enterprise |
| SetMemberCurrentImage | 10m0s | Update Member current image | no | no | Community & Enterprise |
| ShutdownMember | 30m0s | Sends Shutdown requests and waits for container to be stopped | no | no | Community & Enterprise |
| TLSKeyStatusUpdate | 10m0s | Update Status of TLS propagation process | no | no | Enterprise Only |
| TLSPropagated | 10m0s | Update TLS propagation condition | no | no | Enterprise Only |
| TimezoneSecretSet | 30m0s | Set timezone details in cluster | no | no | Community & Enterprise |
| TopologyDisable | 10m0s | Disable TopologyAwareness | no | no | Enterprise Only |
| TopologyEnable | 10m0s | Enable TopologyAwareness | no | no | Enterprise Only |
| TopologyMemberAssignment | 10m0s | Update TopologyAwareness Members assignments | no | no | Enterprise Only |
| TopologyZonesUpdate | 10m0s | Update TopologyAwareness Zones info | no | no | Enterprise Only |
| UpToDateUpdate | 10m0s | Update UpToDate condition | no | no | Community & Enterprise |
| UpdateTLSSNI | 10m0s | Update certificate in SNI | no | no | Enterprise Only |
| UpgradeMember | 6h0m0s | Run the Upgrade procedure on member | no | no | Community & Enterprise |
| WaitForMemberInSync | 30m0s | Wait for member to be in sync. In case of DBServer waits for shards. In case of Agents to catch-up on Agency index | no | no | Community & Enterprise |
| WaitForMemberReady | 30m0s | Wait for member Ready condition | no | no | Community & Enterprise |
| WaitForMemberUp | 30m0s | Wait for member to be responsive | no | no | Community & Enterprise |
[END_INJECT]: # (actionsTable)

## ArangoDeployment spec

[START_INJECT]: # (actionsModYaml)

```yaml
spec:
  timeouts:
    actions:
      AddMember: 10m0s
      AppendTLSCACertificate: 30m0s
      ArangoMemberUpdatePodSpec: 10m0s
      ArangoMemberUpdatePodStatus: 10m0s
      BackupRestore: 15m0s
      BackupRestoreClean: 15m0s
      BootstrapSetPassword: 10m0s
      BootstrapUpdate: 10m0s
      CleanMemberService: 30m0s
      CleanOutMember: 48h0m0s
      CleanTLSCACertificate: 30m0s
      CleanTLSKeyfileCertificate: 30m0s
      ClusterMemberCleanup: 10m0s
      DisableClusterScaling: 10m0s
      DisableMaintenance: 10m0s
      DisableMemberMaintenance: 10m0s
      EnableClusterScaling: 10m0s
      EnableMaintenance: 10m0s
      EnableMemberMaintenance: 10m0s
      EncryptionKeyAdd: 10m0s
      EncryptionKeyPropagated: 10m0s
      EncryptionKeyRefresh: 10m0s
      EncryptionKeyRemove: 10m0s
      EncryptionKeyStatusUpdate: 10m0s
      EnforceResignLeadership: 45m0s
      Idle: 10m0s
      JWTAdd: 10m0s
      JWTClean: 10m0s
      JWTPropagated: 10m0s
      JWTRefresh: 10m0s
      JWTSetActive: 10m0s
      JWTStatusUpdate: 10m0s
      KillMemberPod: 10m0s
      LicenseSet: 10m0s
      MarkToRemoveMember: 10m0s
      MemberPhaseUpdate: 10m0s
      MemberRIDUpdate: 10m0s
      MemberStatusSync: 10m0s
      PVCResize: 30m0s
      PVCResized: 15m0s
      PlaceHolder: 10m0s
      RebalancerCheck: 10m0s
      RebalancerCheckV2: 10m0s
      RebalancerClean: 10m0s
      RebalancerCleanV2: 10m0s
      RebalancerGenerate: 10m0s
      RebalancerGenerateV2: 10m0s
      RebuildOutSyncedShards: 24h0m0s
      RecreateMember: 15m0s
      RefreshTLSCA: 30m0s
      RefreshTLSKeyfileCertificate: 30m0s
      RemoveMember: 15m0s
      RemoveMemberPVC: 15m0s
      RenewTLSCACertificate: 30m0s
      RenewTLSCertificate: 30m0s
      ResignLeadership: 30m0s
      ResourceSync: 10m0s
      RotateMember: 15m0s
      RotateStartMember: 15m0s
      RotateStopMember: 15m0s
      RuntimeContainerArgsLogLevelUpdate: 10m0s
      RuntimeContainerImageUpdate: 10m0s
      RuntimeContainerSyncTolerations: 10m0s
      SetCondition: 10m0s
      SetConditionV2: 10m0s
      SetCurrentImage: 6h0m0s
      SetCurrentMemberArch: 10m0s
      SetMaintenanceCondition: 10m0s
      SetMemberCondition: 10m0s
      SetMemberConditionV2: 10m0s
      SetMemberCurrentImage: 10m0s
      ShutdownMember: 30m0s
      TLSKeyStatusUpdate: 10m0s
      TLSPropagated: 10m0s
      TimezoneSecretSet: 30m0s
      TopologyDisable: 10m0s
      TopologyEnable: 10m0s
      TopologyMemberAssignment: 10m0s
      TopologyZonesUpdate: 10m0s
      UpToDateUpdate: 10m0s
      UpdateTLSSNI: 10m0s
      UpgradeMember: 6h0m0s
      WaitForMemberInSync: 30m0s
      WaitForMemberReady: 30m0s
      WaitForMemberUp: 30m0s

```
[END_INJECT]: # (actionsModYaml)
