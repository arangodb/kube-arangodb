# ArangoDB Operator Metrics

## List

<!-- START(actionsTable) -->
|               Action               | Internal | Timeout | Optional |        Edition         |                                                    Description                                                     |
|:----------------------------------:|:--------:|:-------:|:--------:|:----------------------:|:------------------------------------------------------------------------------------------------------------------:|
|             AddMember              |    no    |  10m0s  |    no    | Community & Enterprise |                                         Adds new member to the Member list                                         |
|       AppendTLSCACertificate       |    no    |  30m0s  |    no    |    Enterprise Only     |                                       Append Certificate into CA TrustStore                                        |
|     ArangoMemberUpdatePodSpec      |    no    |  10m0s  |    no    | Community & Enterprise |                                       Propagate Member Pod spec (requested)                                        |
|    ArangoMemberUpdatePodStatus     |    no    |  10m0s  |    no    | Community & Enterprise |                                       Propagate Member Pod status (current)                                        |
|           BackupRestore            |    no    |  15m0s  |    no    |    Enterprise Only     |                                              Restore selected Backup                                               |
|         BackupRestoreClean         |    no    |  15m0s  |    no    |    Enterprise Only     |                                Clean restore status in case of restore spec change                                 |
|        BootstrapSetPassword        |    no    |  10m0s  |    no    | Community & Enterprise |                                     Change password during bootstrap procedure                                     |
|          BootstrapUpdate           |    no    |  10m0s  |    no    | Community & Enterprise |                                              Update bootstrap status                                               |
|         CleanMemberService         |    no    |  30m0s  |    no    | Community & Enterprise |                                               Removes Server Service                                               |
|           CleanOutMember           |    no    | 48h0m0s |    no    | Community & Enterprise |                                           Run the CleanOut job on member                                           |
|       CleanTLSCACertificate        |    no    |  30m0s  |    no    |    Enterprise Only     |                                       Remove Certificate from CA TrustStore                                        |
|     CleanTLSKeyfileCertificate     |    no    |  30m0s  |    no    |    Enterprise Only     |                                       Remove old TLS certificate from server                                       |
|        ClusterMemberCleanup        |    no    |  10m0s  |    no    | Community & Enterprise |                          Remove member from Cluster if it is gone already (Coordinators)                           |
|       DisableClusterScaling        |    no    |  10m0s  |    no    | Community & Enterprise |                                  (Deprecated) Disable Cluster Scaling integration                                  |
|         DisableMaintenance         |    no    |  10m0s  |    no    | Community & Enterprise |                                         Disable ArangoDB maintenance mode                                          |
|      DisableMemberMaintenance      |    no    |  10m0s  |    no    |    Enterprise Only     |                                     Disable ArangoDB DBServer maintenance mode                                     |
|        EnableClusterScaling        |    no    |  10m0s  |    no    | Community & Enterprise |                                  (Deprecated) Enable Cluster Scaling integration                                   |
|         EnableMaintenance          |    no    |  10m0s  |    no    | Community & Enterprise |                                          Enable ArangoDB maintenance mode                                          |
|      EnableMemberMaintenance       |    no    |  10m0s  |    no    |    Enterprise Only     |                                     Enable ArangoDB DBServer maintenance mode                                      |
|          EncryptionKeyAdd          |    no    |  10m0s  |    no    |    Enterprise Only     |                                         Add the encryption key to the pool                                         |
|      EncryptionKeyPropagated       |    no    |  10m0s  |    no    |    Enterprise Only     |                                     Update condition of encryption propagation                                     |
|        EncryptionKeyRefresh        |    no    |  10m0s  |    no    |    Enterprise Only     |                                       Refresh the encryption keys on member                                        |
|        EncryptionKeyRemove         |    no    |  10m0s  |    no    |    Enterprise Only     |                                       Remove the encryption key to the pool                                        |
|     EncryptionKeyStatusUpdate      |    no    |  10m0s  |    no    |    Enterprise Only     |                                      Update status of encryption propagation                                       |
|                Idle                |    no    |  10m0s  |    no    | Community & Enterprise |                            Define idle operation in case if preconditions are not meet                             |
|               JWTAdd               |    no    |  10m0s  |    no    |    Enterprise Only     |                                              Adds new JWT to the pool                                              |
|              JWTClean              |    no    |  10m0s  |    no    |    Enterprise Only     |                                            Remove JWT key from the pool                                            |
|           JWTPropagated            |    no    |  10m0s  |    no    |    Enterprise Only     |                                        Update condition of JWT propagation                                         |
|             JWTRefresh             |    no    |  10m0s  |    no    |    Enterprise Only     |                                     Refresh current JWT secrets on the member                                      |
|            JWTSetActive            |    no    |  10m0s  |    no    |    Enterprise Only     |                                        Change active JWT key on the cluster                                        |
|          JWTStatusUpdate           |    no    |  10m0s  |    no    |    Enterprise Only     |                                          Update status of JWT propagation                                          |
|           KillMemberPod            |    no    |  10m0s  |    no    | Community & Enterprise |                                Execute Delete on Pod 9put pod in Terminating state)                                |
|             LicenseSet             |    no    |  10m0s  |    no    | Community & Enterprise |                                           Update Cluster license (3.9+)                                            |
|         MarkToRemoveMember         |    no    |  10m0s  |    no    | Community & Enterprise |               Marks member to be removed. Used when member Pod is annotated with replace annotation                |
|         MemberPhaseUpdate          |    no    |  10m0s  |    no    | Community & Enterprise |                                                Change member phase                                                 |
|          MemberRIDUpdate           |    no    |  10m0s  |    no    | Community & Enterprise |                                              Update Run ID of member                                               |
|             PVCResize              |    no    |  30m0s  |    no    | Community & Enterprise |                               Start the resize procedure. Updates PVC Requests field                               |
|             PVCResized             |    no    |  15m0s  |    no    | Community & Enterprise |                                        Waits for PVC resize to be completed                                        |
|            PlaceHolder             |    no    |  10m0s  |    no    | Community & Enterprise |                                              Empty placeholder action                                              |
|          RebalancerCheck           |    no    |  10m0s  |    no    |    Enterprise Only     |                                           Check Rebalancer job progress                                            |
|          RebalancerClean           |    no    |  10m0s  |    no    |    Enterprise Only     |                                               Cleans Rebalancer jobs                                               |
|         RebalancerGenerate         |   yes    |  10m0s  |    no    |    Enterprise Only     |                                           Generates the Rebalancer plan                                            |
|       RebuildOutSyncedShards       |    no    | 24h0m0s |    no    | Community & Enterprise |                               Run Rebuild Out Synced Shards procedure for DBServers                                |
|           RecreateMember           |    no    |  15m0s  |    no    | Community & Enterprise |                                       Recreate member with same ID and Data                                        |
|    RefreshTLSKeyfileCertificate    |    no    |  30m0s  |    no    |    Enterprise Only     |                                       Recreate Server TLS Certificate secret                                       |
|            RemoveMember            |    no    |  15m0s  |    no    | Community & Enterprise |                                     Removes member from the Cluster and Status                                     |
|          RemoveMemberPVC           |    no    |  15m0s  |    no    | Community & Enterprise |                                 Removes member PVC and enforce recreate procedure                                  |
|       RenewTLSCACertificate        |    no    |  30m0s  |    no    |    Enterprise Only     |                                             Recreate Managed CA secret                                             |
|        RenewTLSCertificate         |    no    |  30m0s  |    no    |    Enterprise Only     |                                       Recreate Server TLS Certificate secret                                       |
|          ResignLeadership          |    no    |  30m0s  |   yes    | Community & Enterprise |                                      Run the ResignLeadership job on DBServer                                      |
|            ResourceSync            |    no    |  10m0s  |    no    | Community & Enterprise |                                               Runs the Resource sync                                               |
|            RotateMember            |    no    |  15m0s  |    no    | Community & Enterprise |                                        Waits for Pod restart and recreation                                        |
|         RotateStartMember          |    no    |  15m0s  |    no    | Community & Enterprise |                              Start member rotation. After this action member is down                               |
|          RotateStopMember          |    no    |  15m0s  |    no    | Community & Enterprise |                         Finalize member rotation. After this action member is started back                         |
| RuntimeContainerArgsLogLevelUpdate |    no    |  10m0s  |    no    | Community & Enterprise |                                    Change ArangoDB Member log levels in runtime                                    |
|    RuntimeContainerImageUpdate     |    no    |  10m0s  |    no    | Community & Enterprise |                                         Update Container Image in runtime                                          |
|  RuntimeContainerSyncTolerations   |    no    |  10m0s  |    no    | Community & Enterprise |                                         Update Pod Tolerations in runtime                                          |
|            SetCondition            |    no    |  10m0s  |    no    | Community & Enterprise |                                       (Deprecated) Set deployment condition                                        |
|           SetConditionV2           |    no    |  10m0s  |    no    | Community & Enterprise |                                              Set deployment condition                                              |
|          SetCurrentImage           |    no    | 6h0m0s  |    no    | Community & Enterprise |                               Update deployment current image after image discovery                                |
|        SetCurrentMemberArch        |    no    |  10m0s  |    no    | Community & Enterprise |                                          Set current member architecture                                           |
|      SetMaintenanceCondition       |    no    |  10m0s  |    no    | Community & Enterprise |                                       Update ArangoDB maintenance condition                                        |
|         SetMemberCondition         |    no    |  10m0s  |    no    | Community & Enterprise |                                         (Deprecated) Set member condition                                          |
|        SetMemberConditionV2        |    no    |  10m0s  |    no    | Community & Enterprise |                                                Set member condition                                                |
|       SetMemberCurrentImage        |    no    |  10m0s  |    no    | Community & Enterprise |                                            Update Member current image                                             |
|           ShutdownMember           |    no    |  30m0s  |    no    | Community & Enterprise |                           Sends Shutdown requests and waits for container to be stopped                            |
|         TLSKeyStatusUpdate         |    no    |  10m0s  |    no    |    Enterprise Only     |                                      Update Status of TLS propagation process                                      |
|           TLSPropagated            |    no    |  10m0s  |    no    |    Enterprise Only     |                                          Update TLS propagation condition                                          |
|         TimezoneSecretSet          |    no    |  30m0s  |    no    | Community & Enterprise |                                          Set timezone details in cluster                                           |
|          TopologyDisable           |    no    |  10m0s  |    no    |    Enterprise Only     |                                             Disable TopologyAwareness                                              |
|           TopologyEnable           |    no    |  10m0s  |    no    |    Enterprise Only     |                                              Enable TopologyAwareness                                              |
|      TopologyMemberAssignment      |    no    |  10m0s  |    no    |    Enterprise Only     |                                    Update TopologyAwareness Members assignments                                    |
|        TopologyZonesUpdate         |    no    |  10m0s  |    no    |    Enterprise Only     |                                        Update TopologyAwareness Zones info                                         |
|           UpToDateUpdate           |    no    |  10m0s  |    no    | Community & Enterprise |                                             Update UpToDate condition                                              |
|            UpdateTLSSNI            |    no    |  10m0s  |    no    |    Enterprise Only     |                                             Update certificate in SNI                                              |
|           UpgradeMember            |    no    | 6h0m0s  |    no    | Community & Enterprise |                                        Run the Upgrade procedure on member                                         |
|        WaitForMemberInSync         |    no    |  30m0s  |    no    | Community & Enterprise | Wait for member to be in sync. In case of DBServer waits for shards. In case of Agents to catch-up on Agency index |
|         WaitForMemberReady         |    no    |  30m0s  |    no    | Community & Enterprise |                                          Wait for member Ready condition                                           |
|          WaitForMemberUp           |    no    |  30m0s  |    no    | Community & Enterprise |                                          Wait for member to be responsive                                          |

<!-- END(actionsTable) -->

## ArangoDeployment spec

<!-- START(actionsModYaml) -->
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
      PVCResize: 30m0s
      PVCResized: 15m0s
      PlaceHolder: 10m0s
      RebalancerCheck: 10m0s
      RebalancerClean: 10m0s
      RebalancerGenerate: 10m0s
      RebuildOutSyncedShards: 24h0m0s
      RecreateMember: 15m0s
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
<!-- END(actionsModYaml) -->