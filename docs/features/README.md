---
layout: page
nav_order: 10
title: List of all features
---

## List of Community Edition features

| Feature | Operator Version | Introduced | ArangoDB Version | ArangoDB Edition | State | Enabled | Flag | Remarks |
|:--- |:--- |:--- |:--- |:--- |:--- |:--- |:--- |:--- |
| ArangoPlatform OpenID SSO | 1.2.49 | 1.2.49 | >= 3.8.0 | Community, Enterprise | Beta | True | N/A | Support for ArangoPlatform SSO with OpenID |
| ArangoPlatform OpenID SSO Refresh | 1.2.49 | 1.2.49 | >= 3.8.0 | Community, Enterprise | Alpha | True | N/A | Support for ArangoPlatform SSO with OpenID Refresh |
| ArangoPlatform | 1.2.49 | 1.2.43 | >= 3.8.0 | Community, Enterprise | Beta | True | N/A | ArangoPlatform Solution with support for ArangoDeployment Gateway Group |
| Cleanup Imported Backups | 1.2.41 | 1.2.41 | >= 3.8.0 | Community, Enterprise | Production | False | --deployment.feature.backup-cleanup | Cleanup backups created outside of the Operator and imported into Kubernetes ArangoBackup |
| Upscale resources spec in init containers | 1.2.36 | 1.2.36 | >= 3.8.0 | Community, Enterprise | Production | True | --deployment.feature.init-containers-upscale-resources | Upscale resources spec to built-in init containers if they are not specified or lower |
| Create backups asynchronously | 1.2.35 | 1.2.41 | >= 3.8.0 | Community, Enterprise | Production | True | --deployment.feature.async-backup-creation | Create backups asynchronously to avoid blocking the operator and reaching the timeout |
| Enforced ResignLeadership | 1.2.34 | 1.2.34 | >= 3.8.0 | Community, Enterprise | Production | True | --deployment.feature.enforced-resign-leadership | Enforce ResignLeadership and ensure that Leaders are moved from restarted DBServer |
| Copy resources spec to init containers | 1.2.33 | 1.2.33 | >= 3.8.0 | Community, Enterprise | Production | True | --deployment.feature.init-containers-copy-resources | Copy resources spec to built-in init containers if they are not specified |
| [Rebalancer V2](rebalancer_v2.md) | 1.2.31 | 1.2.31 | >= 3.10.0 | Community, Enterprise | Alpha | False | --deployment.feature.rebalancer-v2 | N/A |
| [Secured containers](secured_containers.md) | 1.2.31 | 1.2.31 | >= 3.8.0 | Community, Enterprise | Alpha | False | --deployment.feature.secured-containers | If set to True Operator will run containers in secure mode |
| Version Check V2 | 1.2.31 | 1.2.31 | >= 3.8.0 | Community, Enterprise | Alpha | False | --deployment.feature.upgrade-version-check-V2 | N/A |
| [Operator Ephemeral Volumes](ephemeral_volumes.md) | 1.2.31 | 1.2.2 | >= 3.8.0 | Community, Enterprise | Beta | False | --deployment.feature.ephemeral-volumes | N/A |
| [Force Rebuild Out Synced Shards](rebuild_out_synced_shards.md) | 1.2.27 | 1.2.27 | >= 3.8.0 | Community, Enterprise | Production | False | --deployment.feature.force-rebuild-out-synced-shards | It should be used only if user is aware of the risks. |
| [Spec Default Restore](deployment_spec_defaults.md) | 1.2.25 | 1.2.21 | >= 3.8.0 | Community, Enterprise | Beta | True | --deployment.feature.deployment-spec-defaults-restore | If set to False Operator will not change ArangoDeployment Spec |
| Version Check | 1.2.23 | 1.1.4 | >= 3.8.0 | Community, Enterprise | Production | True | --deployment.feature.upgrade-version-check | N/A |
| [Failover Leader service](failover_leader_service.md) | 1.2.13 | 1.2.13 | < 3.12.0 | Community, Enterprise | Production | False | --deployment.feature.failover-leadership | N/A |
| Graceful Restart | 1.2.5 | 1.0.7 | >= 3.8.0 | Community, Enterprise | Production | True | ---deployment.feature.graceful-shutdown | N/A |
| Optional Graceful Restart | 1.2.0 | 1.2.5 | >= 3.8.0 | Community, Enterprise | Production | False | --deployment.feature.optional-graceful-shutdown | N/A |
| Operator Internal Metrics Exporter | 1.2.0 | 1.2.0 | >= 3.8.0 | Community, Enterprise | Production | True | --deployment.feature.metrics-exporter | N/A |
| Operator Maintenance Management Support | 1.2.0 | 1.0.7 | >= 3.8.0 | Community, Enterprise | Production | True | --deployment.feature.maintenance | N/A |
| Encryption Key Rotation Support | 1.2.0 | 1.0.3 | >= 3.8.0 | Enterprise | NotSupported | False | --deployment.feature.encryption-rotation | N/A |
| TLS Runtime Rotation Support | 1.1.0 | 1.0.4 | >= 3.8.0 | Enterprise | Production | True | --deployment.feature.tls-rotation | N/A |
| JWT Rotation Support | 1.1.0 | 1.0.3 | >= 3.8.0 | Enterprise | Production | True | --deployment.feature.jwt-rotation | N/A |
| Operator Single Mode | 1.0.4 | 1.0.4 | >= 3.8.0 | Community, Enterprise | Production | False | --mode.single | Only 1 instance of Operator allowed in namespace when feature is enabled |
| TLS SNI Support | 1.0.3 | 1.0.3 | >= 3.8.0 | Enterprise | Production | True | --deployment.feature.tls-sni | N/A |
| ActiveFailover Support | 1.0.0 | 1.0.0 | < 3.12.0 | Community, Enterprise | Production | True | --deployment.feature.active-failover | N/A |
| Disabling of liveness probes | 0.3.11 | 0.3.10 | >= 3.8.0 | Community, Enterprise | Production | True | N/A | N/A |
| Pod Disruption Budgets | 0.3.11 | 0.3.10 | >= 3.8.0 | Community, Enterprise | Production | True | N/A | N/A |
| Prometheus Metrics Exporter | 0.3.11 | 0.3.10 | >= 3.8.0 | Community, Enterprise | Production | True | N/A | Prometheus required |
| Sidecar Containers | 0.3.11 | 0.3.10 | >= 3.8.0 | Community, Enterprise | Production | True | N/A | N/A |
| Volume Claim Templates | 0.3.11 | 0.3.10 | >= 3.8.0 | Community, Enterprise | Production | True | N/A | N/A |
| Volume Resizing | 0.3.11 | 0.3.10 | >= 3.8.0 | Community, Enterprise | Production | True | N/A | N/A |


## List of Enterprise Edition features

| Feature | Operator Version | Introduced | ArangoDB Version | ArangoDB Edition | State | Enabled | Flag | Remarks |
|:--- |:--- |:--- |:--- |:--- |:--- |:--- |:--- |:--- |
| ArangoML integration | 1.2.36 | 1.2.36 | >= 3.8.0 | Enterprise | Alpha | True | N/A | Support for ArangoML CRDs |
| AgencyCache | 1.2.30 | 1.2.30 | >= 3.8.0 | Enterprise | Production | True | N/A | Enable Agency Cache mechanism in the Operator (Increase limit of the nodes) |
| Member Maintenance Support | 1.2.25 | 1.2.16 | >= 3.8.0 | Enterprise | Production | True | N/A | Enable Member Maintenance during planned restarts |
| [Rebalancer](rebalancer.md) | 1.2.15 | 1.2.5 | >= 3.8.0 | Enterprise | Production | True | N/A | N/A |
| [TopologyAwareness](../design/topology_awareness.md) | 1.2.4 | 1.2.4 | >= 3.8.0 | Enterprise | Production | True | N/A | N/A |


