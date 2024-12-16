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
| [arangodb_operator_agency_errors](./arangodb_operator_agency_errors.md) | arangodb_operator | agency | Current count of agency cache fetch errors | Counter |
| [arangodb_operator_agency_fetches](./arangodb_operator_agency_fetches.md) | arangodb_operator | agency | Current count of agency cache fetches | Counter |
| [arangodb_operator_agency_index](./arangodb_operator_agency_index.md) | arangodb_operator | agency | Current index of the agency cache | Gauge |
| [arangodb_operator_agency_cache_health_present](./arangodb_operator_agency_cache_health_present.md) | arangodb_operator | agency_cache | Determines if local agency cache health is present | Gauge |
| [arangodb_operator_agency_cache_healthy](./arangodb_operator_agency_cache_healthy.md) | arangodb_operator | agency_cache | Determines if agency is healthy | Gauge |
| [arangodb_operator_agency_cache_leaders](./arangodb_operator_agency_cache_leaders.md) | arangodb_operator | agency_cache | Determines agency leader vote count | Gauge |
| [arangodb_operator_agency_cache_member_commit_offset](./arangodb_operator_agency_cache_member_commit_offset.md) | arangodb_operator | agency_cache | Determines agency member commit offset | Gauge |
| [arangodb_operator_agency_cache_member_serving](./arangodb_operator_agency_cache_member_serving.md) | arangodb_operator | agency_cache | Determines if agency member is reachable | Gauge |
| [arangodb_operator_agency_cache_present](./arangodb_operator_agency_cache_present.md) | arangodb_operator | agency_cache | Determines if local agency cache is present | Gauge |
| [arangodb_operator_agency_cache_serving](./arangodb_operator_agency_cache_serving.md) | arangodb_operator | agency_cache | Determines if agency is serving | Gauge |
| [arangodb_operator_deployment_conditions](./arangodb_operator_deployment_conditions.md) | arangodb_operator | deployment | Representation of the ArangoDeployment condition state (true/false) | Gauge |
| [arangodb_operator_engine_assertions](./arangodb_operator_engine_assertions.md) | arangodb_operator | engine | Number of assertions invoked during Operator runtime | Counter |
| [arangodb_operator_engine_ops_alerts](./arangodb_operator_engine_ops_alerts.md) | arangodb_operator | engine | Counter for actions which requires ops attention | Counter |
| [arangodb_operator_engine_panics_recovered](./arangodb_operator_engine_panics_recovered.md) | arangodb_operator | engine | Number of Panics recovered inside Operator reconciliation loop | Counter |
| [arangodb_operator_kubernetes_client_request_errors](./arangodb_operator_kubernetes_client_request_errors.md) | arangodb_operator | kubernetes_client | Number of Kubernetes Client request errors | Counter |
| [arangodb_operator_kubernetes_client_requests](./arangodb_operator_kubernetes_client_requests.md) | arangodb_operator | kubernetes_client | Number of Kubernetes Client requests | Counter |
| [arangodb_operator_members_conditions](./arangodb_operator_members_conditions.md) | arangodb_operator | members | Representation of the ArangoMember condition state (true/false) | Gauge |
| [arangodb_operator_members_unexpected_container_exit_codes](./arangodb_operator_members_unexpected_container_exit_codes.md) | arangodb_operator | members | Counter of unexpected restarts in pod (Containers/InitContainers/EphemeralContainers) | Counter |
| [arangodb_operator_objects_processed](./arangodb_operator_objects_processed.md) | arangodb_operator | objects | Number of the processed objects | Counter |
| [arangodb_operator_rebalancer_enabled](./arangodb_operator_rebalancer_enabled.md) | arangodb_operator | rebalancer | Determines if rebalancer is enabled | Gauge |
| [arangodb_operator_rebalancer_moves_current](./arangodb_operator_rebalancer_moves_current.md) | arangodb_operator | rebalancer | Define how many moves are currently in progress | Gauge |
| [arangodb_operator_rebalancer_moves_failed](./arangodb_operator_rebalancer_moves_failed.md) | arangodb_operator | rebalancer | Define how many moves failed | Counter |
| [arangodb_operator_rebalancer_moves_generated](./arangodb_operator_rebalancer_moves_generated.md) | arangodb_operator | rebalancer | Define how many moves were generated | Counter |
| [arangodb_operator_rebalancer_moves_succeeded](./arangodb_operator_rebalancer_moves_succeeded.md) | arangodb_operator | rebalancer | Define how many moves succeeded | Counter |
| [arangodb_operator_resources_arangodeployment_accepted](./arangodb_operator_resources_arangodeployment_accepted.md) | arangodb_operator | resources | Defines if ArangoDeployment has been accepted | Gauge |
| [arangodb_operator_resources_arangodeployment_immutable_errors](./arangodb_operator_resources_arangodeployment_immutable_errors.md) | arangodb_operator | resources | Counter for deployment immutable errors | Counter |
| [arangodb_operator_resources_arangodeployment_propagated](./arangodb_operator_resources_arangodeployment_propagated.md) | arangodb_operator | resources | Defines if ArangoDeployment Spec is propagated | Gauge |
| [arangodb_operator_resources_arangodeployment_status_restores](./arangodb_operator_resources_arangodeployment_status_restores.md) | arangodb_operator | resources | Counter for deployment status restored | Counter |
| [arangodb_operator_resources_arangodeployment_uptodate](./arangodb_operator_resources_arangodeployment_uptodate.md) | arangodb_operator | resources | Defines if ArangoDeployment is uptodate | Gauge |
| [arangodb_operator_resources_arangodeployment_validation_errors](./arangodb_operator_resources_arangodeployment_validation_errors.md) | arangodb_operator | resources | Counter for deployment validation errors | Counter |
| [arangodb_operator_resources_arangodeploymentreplication_active](./arangodb_operator_resources_arangodeploymentreplication_active.md) | arangodb_operator | resources | Defines if ArangoDeploymentReplication is configured and running | Gauge |
| [arangodb_operator_resources_arangodeploymentreplication_failed](./arangodb_operator_resources_arangodeploymentreplication_failed.md) | arangodb_operator | resources | Defines if ArangoDeploymentReplication is in Failed phase | Gauge |
| [arangodb_resources_deployment_config_map_duration](./arangodb_resources_deployment_config_map_duration.md) | arangodb_resources | deployment_config_map | Duration of inspected ConfigMaps by Deployment in seconds | Gauge |
| [arangodb_resources_deployment_config_map_inspected](./arangodb_resources_deployment_config_map_inspected.md) | arangodb_resources | deployment_config_map | Number of inspected ConfigMaps by Deployment | Counter |
[END_INJECT]: # (metricsTable)
