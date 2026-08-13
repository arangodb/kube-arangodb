---
layout: page
nav_order: 10
title: List of all features
---

## List of Community Edition features

| Feature | Operator Version | Introduced | ArangoDB Version | ArangoDB Edition | State | Enabled | Flag | Remarks |
|:--- |:--- |:--- |:--- |:--- |:--- |:--- |:--- |:--- |
| ArangoDeployment Status Subresource | 1.5.0 | 1.5.0 | >= 3.8.0 | Community, Enterprise | Production | False | --deployment.feature.enable-arango-deployment-status | Ensures the status subresource on the ArangoDeployment v1 CRD when enabled; when disabled the operator leaves it as the chart ships it (neither adds nor removes it) |
| Backup Policy Until Propagation | 1.4.4 | 1.4.4 | >= 3.8.0 | Community, Enterprise | Alpha | True | --deployment.feature.backup-policy-until-propagation | Sets Until field in the Backup based on next schedule time |
| Central Services | 1.4.4 | 1.4.4 | >= 3.8.0 | Enterprise | Alpha | False | --deployment.feature.central-services | Enables Central Services |
| [Harden ArangoDB Server Containers](harden.md) | 1.4.4 | 1.4.4 | >= 3.8.0 | Community, Enterprise | Alpha | False | --deployment.feature.harden | Adds hardening arguments to the ArangoDB server containers |
| JWT Asymmetric Key | 1.4.4 | 1.4.4 | >= 3.8.0 | Community, Enterprise | Alpha | False | --deployment.feature.jwt-asymmetric-key | Uses Asymmetric Key as a default in ArangoDB |
| Random Pod Names | 1.4.4 | 1.4.4 | >= 3.8.0 | Community, Enterprise | Alpha | False | --deployment.feature.random-pod-names | Enables generating random pod names |
| Replace Migration | 1.4.4 | 1.4.4 | >= 3.8.0 | Community, Enterprise | Alpha | True | --deployment.feature.replace-migration | During member replacement shards are migrated directly to the new server |
| Sensitive Information Protection | 1.4.4 | 1.4.4 | >= 3.8.0 | Community, Enterprise | Alpha | False | --deployment.feature.sensitive-information-protection | Hide sensitive information from metrics and logs |
| Gateway Sidecar | 1.4.2 | 1.4.2 | >= 3.8.0 | Enterprise | Production | False | --deployment.feature.gateway-sidecar | Enables Gateway Integration |
| ArangoPlatform OpenID SSO | 1.2.49 | 1.2.49 | >= 3.8.0 | Community, Enterprise | Beta | True | N/A | Support for ArangoPlatform SSO with OpenID |
| ArangoPlatform OpenID SSO Refresh | 1.2.49 | 1.2.49 | >= 3.8.0 | Community, Enterprise | Alpha | True | N/A | Support for ArangoPlatform SSO with OpenID Refresh |
| ArangoPlatform | 1.2.49 | 1.2.43 | >= 3.8.0 | Community, Enterprise | Beta | True | N/A | ArangoPlatform Solution with support for ArangoDeployment Gateway Group |
| Gateway | 1.2.43 | 1.2.43 | >= 3.8.0 | Community, Enterprise | Production | True | --deployment.feature.gateway | Defines if gateway extension is enabled |
| Cleanup Imported Backups | 1.2.41 | 1.2.41 | >= 3.8.0 | Community, Enterprise | Production | False | --deployment.feature.backup-cleanup | Cleanup backups created outside of the Operator and imported into Kubernetes ArangoBackup |
| Upscale resources spec in init containers | 1.2.36 | 1.2.36 | >= 3.8.0 | Community, Enterprise | Production | True | --deployment.feature.init-containers-upscale-resources | Upscale resources spec to built-in init containers if they are not specified or lower |
| Create backups asynchronously | 1.2.35 | 1.2.41 | >= 3.8.0 | Community, Enterprise | Production | True | --deployment.feature.async-backup-creation | Create backups asynchronously to avoid blocking the operator and reaching the timeout |
| Enforced ResignLeadership | 1.2.34 | 1.2.34 | >= 3.8.0 | Community, Enterprise | Production | True | --deployment.feature.enforced-resign-leadership | Enforce ResignLeadership and ensure that Leaders are moved from restarted DBServer |
| Copy resources spec to init containers | 1.2.33 | 1.2.33 | >= 3.8.0 | Community, Enterprise | Production | True | --deployment.feature.init-containers-copy-resources | Copy resources spec to built-in init containers if they are not specified |
| [Rebalancer V2](rebalancer_v2.md) | 1.2.31 | 1.2.31 | >= 3.10.0 | Community, Enterprise | Alpha | False | --deployment.feature.rebalancer-v2 | N/A |
| [Secured containers](secured_containers.md) | 1.2.31 | 1.2.31 | >= 3.8.0 | Community, Enterprise | Alpha | False | --deployment.feature.secured-containers | If set to True Operator will run containers in secure mode |
| Version Check V2 | 1.2.31 | 1.2.31 | >= 3.8.0 | Community, Enterprise | Alpha | False | --deployment.feature.upgrade-version-check-v2 | Enable initContainer with pre version check based by Operator |
| [Operator Ephemeral Volumes](ephemeral_volumes.md) | 1.2.31 | 1.2.2 | >= 3.8.0 | Community, Enterprise | Beta | False | --deployment.feature.ephemeral-volumes | N/A |
| Agency Poll | 1.2.30 | 1.2.30 | >= 3.8.0 | Community, Enterprise | Production | True | --deployment.feature.agency-poll | Enable Agency Poll for Enterprise deployments |
| Local Volume Replacement Check | 1.2.28 | 1.2.28 | >= 3.8.0 | Community, Enterprise | Production | False | --deployment.feature.local-volume-replacement-check | Replace volume for local-storage if volume is unschedulable (ex. node is gone) |
| [Force Rebuild Out Synced Shards](rebuild_out_synced_shards.md) | 1.2.27 | 1.2.27 | >= 3.8.0 | Community, Enterprise | Production | False | --deployment.feature.force-rebuild-out-synced-shards | It should be used only if user is aware of the risks. |
| [Spec Default Restore](deployment_spec_defaults.md) | 1.2.25 | 1.2.21 | >= 3.8.0 | Community, Enterprise | Beta | True | --deployment.feature.deployment-spec-defaults-restore | If set to False Operator will not change ArangoDeployment Spec |
| Version Check | 1.2.23 | 1.1.4 | >= 3.8.0 | Community, Enterprise | Production | True | --deployment.feature.upgrade-version-check | N/A |
| Timezone Management | 1.2.16 | 1.2.16 | >= 3.8.0 | Community, Enterprise | Production | False | --deployment.feature.timezone-management | Enable timezone management for pods |
| Restart Policy Always | 1.2.14 | 1.2.14 | >= 3.8.0 | Community, Enterprise | Production | False | --deployment.feature.restart-policy-always | Allow to restart containers with always restart policy |
| [Failover Leader service](failover_leader_service.md) | 1.2.13 | 1.2.13 | < 3.12.0 | Community, Enterprise | Production | False | --deployment.feature.failover-leadership | N/A |
| Graceful Restart | 1.2.5 | 1.0.7 | >= 3.8.0 | Community, Enterprise | Production | True | ---deployment.feature.graceful-shutdown | N/A |
| Short Pod Names | 1.2.4 | 1.2.4 | >= 3.8.0 | Community, Enterprise | Production | False | --deployment.feature.short-pod-names | Enable Short Pod Names |
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
| AgencyCache | 1.2.30 | 1.2.30 | >= 3.8.0 | Enterprise | Production | True | N/A | Enable Agency Cache mechanism in the Operator (Increase limit of the nodes) |
| Member Maintenance Support | 1.2.25 | 1.2.16 | >= 3.8.0 | Enterprise | Production | True | N/A | Enable Member Maintenance during planned restarts |
| [Rebalancer](rebalancer.md) | 1.2.15 | 1.2.5 | >= 3.8.0 | Enterprise | Production | True | N/A | N/A |
| [TopologyAwareness](../design/topology_awareness.md) | 1.2.4 | 1.2.4 | >= 3.8.0 | Enterprise | Production | True | N/A | N/A |


