# ArangoDB Operator Metrics

## List

|               Action               | Timeout |        Edition         |                                                    Description                                                     |
|:----------------------------------:|:-------:|:----------------------:|:------------------------------------------------------------------------------------------------------------------:|
|             AddMember              |  10m0s  | Community & Enterprise |                                         Adds new member to the Member list                                         |
|       AppendTLSCACertificate       |  30m0s  |    Enterprise Only     |                                       Append Certificate into CA TrustStore                                        |
|     ArangoMemberUpdatePodSpec      |  10m0s  | Community & Enterprise |                                       Propagate Member Pod spec (requested)                                        |
|    ArangoMemberUpdatePodStatus     |  10m0s  | Community & Enterprise |                                       Propagate Member Pod status (current)                                        |
|           BackupRestore            |  15m0s  |    Enterprise Only     |                                              Restore selected Backup                                               |
|         BackupRestoreClean         |  15m0s  |    Enterprise Only     |                                Clean restore status in case of restore spec change                                 |
|        BootstrapSetPassword        |  10m0s  | Community & Enterprise |                                     Change password during bootstrap procedure                                     |
|          BootstrapUpdate           |  10m0s  | Community & Enterprise |                                              Update bootstrap status                                               |
|           CleanOutMember           | 48h0m0s | Community & Enterprise |                                           Run the CleanOut job on member                                           |
|       CleanTLSCACertificate        |  30m0s  |    Enterprise Only     |                                       Remove Certificate from CA TrustStore                                        |
|     CleanTLSKeyfileCertificate     |  30m0s  |    Enterprise Only     |                                       Remove old TLS certificate from server                                       |
|        ClusterMemberCleanup        |  10m0s  | Community & Enterprise |                          Remove member from Cluster if it is gone already (Coordinators)                           |
|       DisableClusterScaling        |  10m0s  | Community & Enterprise |                                  (Deprecated) Disable Cluster Scaling integration                                  |
|         DisableMaintenance         |  10m0s  | Community & Enterprise |                                         Disable ArangoDB maintenance mode                                          |
|      DisableMemberMaintenance      |  10m0s  |    Enterprise Only     |                                     Disable ArangoDB DBServer maintenance mode                                     |
|        EnableClusterScaling        |  10m0s  | Community & Enterprise |                                  (Deprecated) Enable Cluster Scaling integration                                   |
|         EnableMaintenance          |  10m0s  | Community & Enterprise |                                          Enable ArangoDB maintenance mode                                          |
|      EnableMemberMaintenance       |  10m0s  |    Enterprise Only     |                                     Enable ArangoDB DBServer maintenance mode                                      |
|          EncryptionKeyAdd          |  10m0s  |    Enterprise Only     |                                         Add the encryption key to the pool                                         |
|      EncryptionKeyPropagated       |  10m0s  |    Enterprise Only     |                                     Update condition of encryption propagation                                     |
|        EncryptionKeyRefresh        |  10m0s  |    Enterprise Only     |                                       Refresh the encryption keys on member                                        |
|        EncryptionKeyRemove         |  10m0s  |    Enterprise Only     |                                       Remove the encryption key to the pool                                        |
|     EncryptionKeyStatusUpdate      |  10m0s  |    Enterprise Only     |                                      Update status of encryption propagation                                       |
|                Idle                |  10m0s  | Community & Enterprise |                            Define idle operation in case if preconditions are not meet                             |
|               JWTAdd               |  10m0s  |    Enterprise Only     |                                              Adds new JWT to the pool                                              |
|              JWTClean              |  10m0s  |    Enterprise Only     |                                            Remove JWT key from the pool                                            |
|           JWTPropagated            |  10m0s  |    Enterprise Only     |                                        Update condition of JWT propagation                                         |
|             JWTRefresh             |  10m0s  |    Enterprise Only     |                                     Refresh current JWT secrets on the member                                      |
|            JWTSetActive            |  10m0s  |    Enterprise Only     |                                        Change active JWT key on the cluster                                        |
|          JWTStatusUpdate           |  10m0s  |    Enterprise Only     |                                          Update status of JWT propagation                                          |
|           KillMemberPod            |  10m0s  | Community & Enterprise |                                Execute Delete on Pod 9put pod in Terminating state)                                |
|             LicenseSet             |  10m0s  | Community & Enterprise |                                           Update Cluster license (3.9+)                                            |
|         MarkToRemoveMember         |  10m0s  | Community & Enterprise |               Marks member to be removed. Used when member Pod is annotated with replace annotation                |
|         MemberPhaseUpdate          |  10m0s  | Community & Enterprise |                                                Change member phase                                                 |
|          MemberRIDUpdate           |  10m0s  | Community & Enterprise |                                              Update Run ID of member                                               |
|             PVCResize              |  30m0s  | Community & Enterprise |                               Start the resize procedure. Updates PVC Requests field                               |
|             PVCResized             |  15m0s  | Community & Enterprise |                                        Waits for PVC resize to be completed                                        |
|            PlaceHolder             |  10m0s  | Community & Enterprise |                                              Empty placeholder action                                              |
|          RebalancerCheck           |  10m0s  |    Enterprise Only     |                                           Check Rebalancer job progress                                            |
|          RebalancerClean           |  10m0s  |    Enterprise Only     |                                               Cleans Rebalancer jobs                                               |
|         RebalancerGenerate         |  10m0s  |    Enterprise Only     |                                           Generates the Rebalancer plan                                            |
|           RecreateMember           |  15m0s  | Community & Enterprise |                                       Recreate member with same ID and Data                                        |
|    RefreshTLSKeyfileCertificate    |  30m0s  |    Enterprise Only     |                                       Recreate Server TLS Certificate secret                                       |
|            RemoveMember            |  15m0s  | Community & Enterprise |                                     Removes member from the Cluster and Status                                     |
|       RenewTLSCACertificate        |  30m0s  |    Enterprise Only     |                                             Recreate Managed CA secret                                             |
|        RenewTLSCertificate         |  30m0s  |    Enterprise Only     |                                       Recreate Server TLS Certificate secret                                       |
|          ResignLeadership          |  30m0s  | Community & Enterprise |                                      Run the ResignLeadership job on DBServer                                      |
|            ResourceSync            |  10m0s  | Community & Enterprise |                                               Runs the Resource sync                                               |
|            RotateMember            |  15m0s  | Community & Enterprise |                                        Waits for Pod restart and recreation                                        |
|         RotateStartMember          |  15m0s  | Community & Enterprise |                              Start member rotation. After this action member is down                               |
|          RotateStopMember          |  15m0s  | Community & Enterprise |                         Finalize member rotation. After this action member is started back                         |
| RuntimeContainerArgsLogLevelUpdate |  10m0s  | Community & Enterprise |                                    Change ArangoDB Member log levels in runtime                                    |
|    RuntimeContainerImageUpdate     |  10m0s  | Community & Enterprise |                                         Update Container Image in runtime                                          |
|            SetCondition            |  10m0s  | Community & Enterprise |                                       (Deprecated) Set deployment condition                                        |
|           SetConditionV2           |  10m0s  | Community & Enterprise |                                              Set deployment condition                                              |
|          SetCurrentImage           | 6h0m0s  | Community & Enterprise |                               Update deployment current image after image discovery                                |
|      SetMaintenanceCondition       |  10m0s  | Community & Enterprise |                                       Update ArangoDB maintenance condition                                        |
|         SetMemberCondition         |  10m0s  | Community & Enterprise |                                         (Deprecated) Set member condition                                          |
|        SetMemberConditionV2        |  10m0s  | Community & Enterprise |                                                Set member condition                                                |
|       SetMemberCurrentImage        |  10m0s  | Community & Enterprise |                                            Update Member current image                                             |
|           ShutdownMember           |  30m0s  | Community & Enterprise |                           Sends Shutdown requests and waits for container to be stopped                            |
|         TLSKeyStatusUpdate         |  10m0s  |    Enterprise Only     |                                      Update Status of TLS propagation process                                      |
|           TLSPropagated            |  10m0s  |    Enterprise Only     |                                          Update TLS propagation condition                                          |
|         TimezoneSecretSet          |  30m0s  | Community & Enterprise |                                          Set timezone details in cluster                                           |
|          TopologyDisable           |  10m0s  |    Enterprise Only     |                                             Disable TopologyAwareness                                              |
|           TopologyEnable           |  10m0s  |    Enterprise Only     |                                              Enable TopologyAwareness                                              |
|      TopologyMemberAssignment      |  10m0s  |    Enterprise Only     |                                    Update TopologyAwareness Members assignments                                    |
|        TopologyZonesUpdate         |  10m0s  |    Enterprise Only     |                                        Update TopologyAwareness Zones info                                         |
|           UpToDateUpdate           |  10m0s  | Community & Enterprise |                                             Update UpToDate condition                                              |
|            UpdateTLSSNI            |  10m0s  |    Enterprise Only     |                                             Update certificate in SNI                                              |
|           UpgradeMember            | 6h0m0s  | Community & Enterprise |                                        Run the Upgrade procedure on member                                         |
|        WaitForMemberInSync         |  30m0s  | Community & Enterprise | Wait for member to be in sync. In case of DBServer waits for shards. In case of Agents to catch-up on Agency index |
|          WaitForMemberUp           |  30m0s  | Community & Enterprise |                                          Wait for member to be responsive                                          |


## ArangoDeployment spec

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
      RecreateMember: 15m0s
      RefreshTLSKeyfileCertificate: 30m0s
      RemoveMember: 15m0s
      RenewTLSCACertificate: 30m0s
      RenewTLSCertificate: 30m0s
      ResignLeadership: 30m0s
      ResourceSync: 10m0s
      RotateMember: 15m0s
      RotateStartMember: 15m0s
      RotateStopMember: 15m0s
      RuntimeContainerArgsLogLevelUpdate: 10m0s
      RuntimeContainerImageUpdate: 10m0s
      SetCondition: 10m0s
      SetConditionV2: 10m0s
      SetCurrentImage: 6h0m0s
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
      WaitForMemberUp: 30m0s

```