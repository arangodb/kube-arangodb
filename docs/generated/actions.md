---
layout: page
title: List of Plan Actions
nav_order: 11
---

# ArangoDB Operator Actions

## List

[START_INJECT]: # (actionsTable)

| Action | Internal | Timeout | Optional | Edition | Description |
|:---:|:---:|:---:|:---:|:---:|:---:|
| AddMember | no | 10m0s | no | Community & Enterprise | Adds new member to the Member list |
| AppendTLSCACertificate | no | 30m0s | no | Enterprise Only | Append Certificate into CA TrustStore |
| ArangoMemberUpdatePodSpec | no | 10m0s | no | Community & Enterprise | Propagate Member Pod spec (requested) |
| ArangoMemberUpdatePodStatus | no | 10m0s | no | Community & Enterprise | Propagate Member Pod status (current) |
| BackupRestore | no | 15m0s | no | Enterprise Only | Restore selected Backup |
| BackupRestoreClean | no | 15m0s | no | Enterprise Only | Clean restore status in case of restore spec change |
| BootstrapSetPassword | no | 10m0s | no | Community & Enterprise | Change password during bootstrap procedure |
| BootstrapUpdate | no | 10m0s | no | Community & Enterprise | Update bootstrap status |
| CleanMemberService | no | 30m0s | no | Community & Enterprise | Removes Server Service |
| CleanOutMember | no | 48h0m0s | no | Community & Enterprise | Run the CleanOut job on member |
| CleanTLSCACertificate | no | 30m0s | no | Enterprise Only | Remove Certificate from CA TrustStore |
| CleanTLSKeyfileCertificate | no | 30m0s | no | Enterprise Only | Remove old TLS certificate from server |
| ClusterMemberCleanup | no | 10m0s | no | Community & Enterprise | Remove member from Cluster if it is gone already (Coordinators) |
| Delay | no | 10m0s | yes | Community & Enterprise | Define delay operation |
| ~~DisableClusterScaling~~ | no | 10m0s | no | Community & Enterprise | Disable Cluster Scaling integration |
| DisableMaintenance | no | 10m0s | no | Community & Enterprise | Disable ArangoDB maintenance mode |
| DisableMemberMaintenance | no | 10m0s | no | Enterprise Only | Disable ArangoDB DBServer maintenance mode |
| ~~EnableClusterScaling~~ | no | 10m0s | no | Community & Enterprise | Enable Cluster Scaling integration |
| EnableMaintenance | no | 10m0s | no | Community & Enterprise | Enable ArangoDB maintenance mode |
| EnableMemberMaintenance | no | 10m0s | no | Enterprise Only | Enable ArangoDB DBServer maintenance mode |
| EncryptionKeyAdd | no | 10m0s | no | Enterprise Only | Add the encryption key to the pool |
| EncryptionKeyPropagated | no | 10m0s | no | Enterprise Only | Update condition of encryption propagation |
| EncryptionKeyRefresh | no | 10m0s | no | Enterprise Only | Refresh the encryption keys on member |
| EncryptionKeyRemove | no | 10m0s | no | Enterprise Only | Remove the encryption key to the pool |
| EncryptionKeyStatusUpdate | no | 10m0s | no | Enterprise Only | Update status of encryption propagation |
| EnforceResignLeadership | no | 45m0s | yes | Community & Enterprise | Run the ResignLeadership job on DBServer and checks data compatibility after |
| Idle | no | 10m0s | no | Community & Enterprise | Define idle operation in case if preconditions are not meet |
| JWTAdd | no | 10m0s | no | Enterprise Only | Adds new JWT to the pool |
| JWTClean | no | 10m0s | no | Enterprise Only | Remove JWT key from the pool |
| JWTPropagated | no | 10m0s | no | Enterprise Only | Update condition of JWT propagation |
| JWTRefresh | no | 10m0s | no | Enterprise Only | Refresh current JWT secrets on the member |
| JWTSetActive | no | 10m0s | no | Enterprise Only | Change active JWT key on the cluster |
| JWTStatusUpdate | no | 10m0s | no | Enterprise Only | Update status of JWT propagation |
| KillMemberPod | no | 10m0s | no | Community & Enterprise | Execute Delete on Pod (put pod in Terminating state) |
| LicenseSet | no | 10m0s | no | Community & Enterprise | Update Cluster license (3.9+) |
| MarkToRemoveMember | no | 10m0s | no | Community & Enterprise | Marks member to be removed. Used when member Pod is annotated with replace annotation |
| MemberPhaseUpdate | no | 10m0s | no | Community & Enterprise | Change member phase |
| ~~MemberRIDUpdate~~ | no | 10m0s | no | Community & Enterprise | Update Run ID of member |
| MemberStatusSync | no | 10m0s | no | Community & Enterprise | Sync ArangoMember Status with ArangoDeployment Status, to keep Member information up to date |
| MigrateMember | no | 48h0m0s | no | Community & Enterprise | Run the data movement actions on the member (migration) |
| PVCResize | no | 30m0s | no | Community & Enterprise | Start the resize procedure. Updates PVC Requests field |
| PVCResized | no | 15m0s | no | Community & Enterprise | Waits for PVC resize to be completed |
| PlaceHolder | no | 10m0s | no | Community & Enterprise | Empty placeholder action |
| RebalancerCheck | no | 10m0s | no | Enterprise Only | Check Rebalancer job progress |
| RebalancerCheckV2 | no | 10m0s | no | Community & Enterprise | Check Rebalancer job progress |
| RebalancerClean | no | 10m0s | no | Enterprise Only | Cleans Rebalancer jobs |
| RebalancerCleanV2 | no | 10m0s | no | Community & Enterprise | Cleans Rebalancer jobs |
| RebalancerGenerate | yes | 10m0s | no | Enterprise Only | Generates the Rebalancer plan |
| RebalancerGenerateV2 | yes | 10m0s | no | Community & Enterprise | Generates the Rebalancer plan |
| RebuildOutSyncedShards | no | 24h0m0s | no | Community & Enterprise | Run Rebuild Out Synced Shards procedure for DBServers |
| RecreateMember | no | 15m0s | no | Community & Enterprise | Recreate member with same ID and Data |
| RefreshTLSCA | no | 30m0s | no | Enterprise Only | Refresh internal CA |
| RefreshTLSKeyfileCertificate | no | 30m0s | no | Enterprise Only | Recreate Server TLS Certificate secret |
| RemoveMember | no | 15m0s | no | Community & Enterprise | Removes member from the Cluster and Status |
| RemoveMemberPVC | no | 15m0s | no | Community & Enterprise | Removes member PVC and enforce recreate procedure |
| RenewTLSCACertificate | no | 30m0s | no | Enterprise Only | Recreate Managed CA secret |
| RenewTLSCertificate | no | 30m0s | no | Enterprise Only | Recreate Server TLS Certificate secret |
| ResignLeadership | no | 30m0s | yes | Community & Enterprise | Run the ResignLeadership job on DBServer |
| ResourceSync | no | 10m0s | no | Community & Enterprise | Runs the Resource sync |
| RotateMember | no | 15m0s | no | Community & Enterprise | Waits for Pod restart and recreation |
| RotateStartMember | no | 15m0s | no | Community & Enterprise | Start member rotation. After this action member is down |
| RotateStopMember | no | 15m0s | no | Community & Enterprise | Finalize member rotation. After this action member is started back |
| RuntimeContainerArgsLogLevelUpdate | no | 10m0s | no | Community & Enterprise | Change ArangoDB Member log levels in runtime |
| RuntimeContainerImageUpdate | no | 10m0s | no | Community & Enterprise | Update Container Image in runtime |
| RuntimeContainerSyncTolerations | no | 10m0s | no | Community & Enterprise | Update Pod Tolerations in runtime |
| ~~SetCondition~~ | no | 10m0s | no | Community & Enterprise | Set deployment condition |
| SetConditionV2 | no | 10m0s | no | Community & Enterprise | Set deployment condition |
| SetCurrentImage | no | 6h0m0s | no | Community & Enterprise | Update deployment current image after image discovery |
| SetCurrentMemberArch | no | 10m0s | no | Community & Enterprise | Set current member architecture |
| SetMaintenanceCondition | no | 10m0s | no | Community & Enterprise | Update ArangoDB maintenance condition |
| ~~SetMemberCondition~~ | no | 10m0s | no | Community & Enterprise | Set member condition |
| SetMemberConditionV2 | no | 10m0s | no | Community & Enterprise | Set member condition |
| SetMemberCurrentImage | no | 10m0s | no | Community & Enterprise | Update Member current image |
| ShutdownMember | no | 30m0s | no | Community & Enterprise | Sends Shutdown requests and waits for container to be stopped |
| TLSKeyStatusUpdate | no | 10m0s | no | Enterprise Only | Update Status of TLS propagation process |
| TLSPropagated | no | 10m0s | no | Enterprise Only | Update TLS propagation condition |
| TimezoneSecretSet | no | 30m0s | no | Community & Enterprise | Set timezone details in cluster |
| TopologyDisable | no | 10m0s | no | Enterprise Only | Disable TopologyAwareness |
| TopologyEnable | no | 10m0s | no | Enterprise Only | Enable TopologyAwareness |
| TopologyMemberAssignment | no | 10m0s | no | Enterprise Only | Update TopologyAwareness Members assignments |
| TopologyZonesUpdate | no | 10m0s | no | Enterprise Only | Update TopologyAwareness Zones info |
| UpToDateUpdate | no | 10m0s | no | Community & Enterprise | Update UpToDate condition |
| UpdateTLSSNI | no | 10m0s | no | Enterprise Only | Update certificate in SNI |
| UpgradeMember | no | 6h0m0s | no | Community & Enterprise | Run the Upgrade procedure on member |
| WaitForMemberInSync | no | 30m0s | no | Community & Enterprise | Wait for member to be in sync. In case of DBServer waits for shards. In case of Agents to catch-up on Agency index |
| WaitForMemberReady | no | 30m0s | no | Community & Enterprise | Wait for member Ready condition |
| WaitForMemberUp | no | 30m0s | no | Community & Enterprise | Wait for member to be responsive |

[END_INJECT]: # (actionsTable)

## ArangoDeployment spec

[START_INJECT]: # (actionsModYaml)

```yaml
spec:
  timeouts:
    actions:
      AddMember:
        Duration: 600000000000
      AppendTLSCACertificate:
        Duration: 1800000000000
      ArangoMemberUpdatePodSpec:
        Duration: 600000000000
      ArangoMemberUpdatePodStatus:
        Duration: 600000000000
      BackupRestore:
        Duration: 900000000000
      BackupRestoreClean:
        Duration: 900000000000
      BootstrapSetPassword:
        Duration: 600000000000
      BootstrapUpdate:
        Duration: 600000000000
      CleanMemberService:
        Duration: 1800000000000
      CleanOutMember:
        Duration: 172800000000000
      CleanTLSCACertificate:
        Duration: 1800000000000
      CleanTLSKeyfileCertificate:
        Duration: 1800000000000
      ClusterMemberCleanup:
        Duration: 600000000000
      Delay:
        Duration: 600000000000
      DisableClusterScaling:
        Duration: 600000000000
      DisableMaintenance:
        Duration: 600000000000
      DisableMemberMaintenance:
        Duration: 600000000000
      EnableClusterScaling:
        Duration: 600000000000
      EnableMaintenance:
        Duration: 600000000000
      EnableMemberMaintenance:
        Duration: 600000000000
      EncryptionKeyAdd:
        Duration: 600000000000
      EncryptionKeyPropagated:
        Duration: 600000000000
      EncryptionKeyRefresh:
        Duration: 600000000000
      EncryptionKeyRemove:
        Duration: 600000000000
      EncryptionKeyStatusUpdate:
        Duration: 600000000000
      EnforceResignLeadership:
        Duration: 2700000000000
      Idle:
        Duration: 600000000000
      JWTAdd:
        Duration: 600000000000
      JWTClean:
        Duration: 600000000000
      JWTPropagated:
        Duration: 600000000000
      JWTRefresh:
        Duration: 600000000000
      JWTSetActive:
        Duration: 600000000000
      JWTStatusUpdate:
        Duration: 600000000000
      KillMemberPod:
        Duration: 600000000000
      LicenseSet:
        Duration: 600000000000
      MarkToRemoveMember:
        Duration: 600000000000
      MemberPhaseUpdate:
        Duration: 600000000000
      MemberRIDUpdate:
        Duration: 600000000000
      MemberStatusSync:
        Duration: 600000000000
      MigrateMember:
        Duration: 172800000000000
      PVCResize:
        Duration: 1800000000000
      PVCResized:
        Duration: 900000000000
      PlaceHolder:
        Duration: 600000000000
      RebalancerCheck:
        Duration: 600000000000
      RebalancerCheckV2:
        Duration: 600000000000
      RebalancerClean:
        Duration: 600000000000
      RebalancerCleanV2:
        Duration: 600000000000
      RebalancerGenerate:
        Duration: 600000000000
      RebalancerGenerateV2:
        Duration: 600000000000
      RebuildOutSyncedShards:
        Duration: 86400000000000
      RecreateMember:
        Duration: 900000000000
      RefreshTLSCA:
        Duration: 1800000000000
      RefreshTLSKeyfileCertificate:
        Duration: 1800000000000
      RemoveMember:
        Duration: 900000000000
      RemoveMemberPVC:
        Duration: 900000000000
      RenewTLSCACertificate:
        Duration: 1800000000000
      RenewTLSCertificate:
        Duration: 1800000000000
      ResignLeadership:
        Duration: 1800000000000
      ResourceSync:
        Duration: 600000000000
      RotateMember:
        Duration: 900000000000
      RotateStartMember:
        Duration: 900000000000
      RotateStopMember:
        Duration: 900000000000
      RuntimeContainerArgsLogLevelUpdate:
        Duration: 600000000000
      RuntimeContainerImageUpdate:
        Duration: 600000000000
      RuntimeContainerSyncTolerations:
        Duration: 600000000000
      SetCondition:
        Duration: 600000000000
      SetConditionV2:
        Duration: 600000000000
      SetCurrentImage:
        Duration: 21600000000000
      SetCurrentMemberArch:
        Duration: 600000000000
      SetMaintenanceCondition:
        Duration: 600000000000
      SetMemberCondition:
        Duration: 600000000000
      SetMemberConditionV2:
        Duration: 600000000000
      SetMemberCurrentImage:
        Duration: 600000000000
      ShutdownMember:
        Duration: 1800000000000
      TLSKeyStatusUpdate:
        Duration: 600000000000
      TLSPropagated:
        Duration: 600000000000
      TimezoneSecretSet:
        Duration: 1800000000000
      TopologyDisable:
        Duration: 600000000000
      TopologyEnable:
        Duration: 600000000000
      TopologyMemberAssignment:
        Duration: 600000000000
      TopologyZonesUpdate:
        Duration: 600000000000
      UpToDateUpdate:
        Duration: 600000000000
      UpdateTLSSNI:
        Duration: 600000000000
      UpgradeMember:
        Duration: 21600000000000
      WaitForMemberInSync:
        Duration: 1800000000000
      WaitForMemberReady:
        Duration: 1800000000000
      WaitForMemberUp:
        Duration: 1800000000000

```
[END_INJECT]: # (actionsModYaml)
