---
layout: page
title: List of available metrics
nav_order: 9
has_children: true
has_toc: false
---

# ArangoDB Operator Metrics

## List of the Operator metrics

[START_INJECT]: # (metricsTable)

| Name | Namespace | Group | Type | Description |
|:---:|:---:|:---:|:---:|:--- |
| [arangodb_operator_agency_errors](./arangodb_operator_agency_errors.md) | arangodb_operator | agency | Counter | Current count of agency cache fetch errors |
| [arangodb_operator_agency_fetches](./arangodb_operator_agency_fetches.md) | arangodb_operator | agency | Counter | Current count of agency cache fetches |
| [arangodb_operator_agency_index](./arangodb_operator_agency_index.md) | arangodb_operator | agency | Gauge | Current index of the agency cache |
| [arangodb_operator_agency_cache_health_present](./arangodb_operator_agency_cache_health_present.md) | arangodb_operator | agency_cache | Gauge | Determines if local agency cache health is present |
| [arangodb_operator_agency_cache_healthy](./arangodb_operator_agency_cache_healthy.md) | arangodb_operator | agency_cache | Gauge | Determines if agency is healthy |
| [arangodb_operator_agency_cache_leaders](./arangodb_operator_agency_cache_leaders.md) | arangodb_operator | agency_cache | Gauge | Determines agency leader vote count |
| [arangodb_operator_agency_cache_member_commit_offset](./arangodb_operator_agency_cache_member_commit_offset.md) | arangodb_operator | agency_cache | Gauge | Determines agency member commit offset |
| [arangodb_operator_agency_cache_member_serving](./arangodb_operator_agency_cache_member_serving.md) | arangodb_operator | agency_cache | Gauge | Determines if agency member is reachable |
| [arangodb_operator_agency_cache_present](./arangodb_operator_agency_cache_present.md) | arangodb_operator | agency_cache | Gauge | Determines if local agency cache is present |
| [arangodb_operator_agency_cache_serving](./arangodb_operator_agency_cache_serving.md) | arangodb_operator | agency_cache | Gauge | Determines if agency is serving |
| [arangodb_operator_deployment_conditions](./arangodb_operator_deployment_conditions.md) | arangodb_operator | deployment | Gauge | Representation of the ArangoDeployment condition state (true/false) |
| [arangodb_operator_engine_assertions](./arangodb_operator_engine_assertions.md) | arangodb_operator | engine | Counter | Number of assertions invoked during Operator runtime |
| [arangodb_operator_engine_ops_alerts](./arangodb_operator_engine_ops_alerts.md) | arangodb_operator | engine | Counter | Counter for actions which requires ops attention |
| [arangodb_operator_engine_panics_recovered](./arangodb_operator_engine_panics_recovered.md) | arangodb_operator | engine | Counter | Number of Panics recovered inside Operator reconciliation loop |
| [arangodb_operator_kubernetes_client_request_errors](./arangodb_operator_kubernetes_client_request_errors.md) | arangodb_operator | kubernetes_client | Counter | Number of Kubernetes Client request errors |
| [arangodb_operator_kubernetes_client_requests](./arangodb_operator_kubernetes_client_requests.md) | arangodb_operator | kubernetes_client | Counter | Number of Kubernetes Client requests |
| [arangodb_operator_members_conditions](./arangodb_operator_members_conditions.md) | arangodb_operator | members | Gauge | Representation of the ArangoMember condition state (true/false) |
| [arangodb_operator_members_unexpected_container_exit_codes](./arangodb_operator_members_unexpected_container_exit_codes.md) | arangodb_operator | members | Counter | Counter of unexpected restarts in pod (Containers/InitContainers/EphemeralContainers) |
| [arangodb_operator_objects_processed](./arangodb_operator_objects_processed.md) | arangodb_operator | objects | Counter | Number of the processed objects |
| [arangodb_operator_rebalancer_enabled](./arangodb_operator_rebalancer_enabled.md) | arangodb_operator | rebalancer | Gauge | Determines if rebalancer is enabled |
| [arangodb_operator_rebalancer_moves_current](./arangodb_operator_rebalancer_moves_current.md) | arangodb_operator | rebalancer | Gauge | Define how many moves are currently in progress |
| [arangodb_operator_rebalancer_moves_failed](./arangodb_operator_rebalancer_moves_failed.md) | arangodb_operator | rebalancer | Counter | Define how many moves failed |
| [arangodb_operator_rebalancer_moves_generated](./arangodb_operator_rebalancer_moves_generated.md) | arangodb_operator | rebalancer | Counter | Define how many moves were generated |
| [arangodb_operator_rebalancer_moves_succeeded](./arangodb_operator_rebalancer_moves_succeeded.md) | arangodb_operator | rebalancer | Counter | Define how many moves succeeded |
| [arangodb_operator_resources_arangodeployment_accepted](./arangodb_operator_resources_arangodeployment_accepted.md) | arangodb_operator | resources | Gauge | Defines if ArangoDeployment has been accepted |
| [arangodb_operator_resources_arangodeployment_immutable_errors](./arangodb_operator_resources_arangodeployment_immutable_errors.md) | arangodb_operator | resources | Counter | Counter for deployment immutable errors |
| [arangodb_operator_resources_arangodeployment_propagated](./arangodb_operator_resources_arangodeployment_propagated.md) | arangodb_operator | resources | Gauge | Defines if ArangoDeployment Spec is propagated |
| [arangodb_operator_resources_arangodeployment_status_restores](./arangodb_operator_resources_arangodeployment_status_restores.md) | arangodb_operator | resources | Counter | Counter for deployment status restored |
| [arangodb_operator_resources_arangodeployment_uptodate](./arangodb_operator_resources_arangodeployment_uptodate.md) | arangodb_operator | resources | Gauge | Defines if ArangoDeployment is uptodate |
| [arangodb_operator_resources_arangodeployment_validation_errors](./arangodb_operator_resources_arangodeployment_validation_errors.md) | arangodb_operator | resources | Counter | Counter for deployment validation errors |
| [arangodb_operator_resources_arangodeploymentreplication_active](./arangodb_operator_resources_arangodeploymentreplication_active.md) | arangodb_operator | resources | Gauge | Defines if ArangoDeploymentReplication is configured and running |
| [arangodb_operator_resources_arangodeploymentreplication_failed](./arangodb_operator_resources_arangodeploymentreplication_failed.md) | arangodb_operator | resources | Gauge | Defines if ArangoDeploymentReplication is in Failed phase |
| [arangodb_resources_deployment_config_map_duration](./arangodb_resources_deployment_config_map_duration.md) | arangodb_resources | deployment_config_map | Gauge | Duration of inspected ConfigMaps by Deployment in seconds |
| [arangodb_resources_deployment_config_map_inspected](./arangodb_resources_deployment_config_map_inspected.md) | arangodb_resources | deployment_config_map | Counter | Number of inspected ConfigMaps by Deployment |

[END_INJECT]: # (metricsTable)
