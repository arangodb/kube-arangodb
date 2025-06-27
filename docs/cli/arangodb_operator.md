---
layout: page
parent: Binaries
title: arangodb_operator
---

# ArangoDB Operator Command

[START_INJECT]: # (arangodb_operator_cmd)
```
Usage:
  arangodb_operator [flags]
  arangodb_operator [command]

Available Commands:
  admin           Administration operations
  completion      Generate the autocompletion script for the specified shell
  crd           CRD operations
  debug-package Generate debug package for debugging
  exporter        
  features        Describe all operator features
  help            Help about any command
  integration     
  storage         
  task          
  version         
  webhook         

Flags:
      --action.PVCResize.concurrency int                       Define limit of concurrent PVC Resizes on the cluster (default 32)
      --agency.refresh-delay duration                          The Agency refresh delay (0 = no delay) (default 500ms)
      --agency.refresh-interval duration                       The Agency refresh interval (0 = do not refresh)
      --agency.retries int                                     The Agency retries (0 = no retries) (default 1)
      --api.enabled                                            Enable operator HTTP and gRPC API (default true)
      --api.grpc-port int                                      gRPC API port to listen on (default 8728)
      --api.http-port int                                      HTTP API port to listen on (default 8628)
      --api.jwt-key-secret-name string                         Name of secret containing key used to sign JWT. If there is no such secret present, value will be saved here (default "arangodb-operator-api-jwt-key")
      --api.jwt-secret-name string                             Name of secret which will contain JWT to authenticate API requests. (default "arangodb-operator-api-jwt")
      --api.tls-secret-name string                             Name of secret containing tls.crt & tls.key for HTTPS API (if empty, self-signed certificate is used)
      --backup-concurrent-uploads int                          Number of concurrent uploads per deployment (default 4)
      --chaos.allowed                                          Set to allow chaos in deployments. Only activated when allowed and enabled in deployment
      --crd.install                                            Install missing CRD if access is possible (default true)
      --crd.preserve-unknown-fields stringArray                Controls which CRD should have enabled preserve unknown fields in validation schema <crd-name>=<true/false>. To apply for all, use crd-name 'all'.
      --crd.validation-schema stringArray                      Overrides default set of CRDs which should have validation schema enabled <crd-name>=<true/false>. To apply for all, use crd-name 'all'.
      --deployment.feature.active-failover                     Support for ActiveFailover mode - Required ArangoDB >= 3.8.0, < 3.12 (default true)
      --deployment.feature.agency-poll                         Enable Agency Poll for Enterprise deployments - Required ArangoDB >= 3.8.0 (default true)
      --deployment.feature.all                                 Enable ALL Features
      --deployment.feature.async-backup-creation               Create backups asynchronously to avoid blocking the operator and reaching the timeout - Required ArangoDB >= 3.8.0 (default true)
      --deployment.feature.backup-cleanup                      Cleanup imported backups if required - Required ArangoDB >= 3.8.0
      --deployment.feature.backup-policy-until-propagation     Sets Until field in the Backup based on next schedule time - Required ArangoDB >= 3.8.0 (default true)
      --deployment.feature.deployment-spec-defaults-restore    Restore defaults from last accepted state of deployment - Required ArangoDB >= 3.8.0 (default true)
      --deployment.feature.enforced-resign-leadership          Enforce ResignLeadership and ensure that Leaders are moved from restarted DBServer - Required ArangoDB >= 3.8.0 (default true)
      --deployment.feature.ephemeral-volumes                   Enables ephemeral volumes for apps and tmp directory - Required ArangoDB >= 3.8.0
      --deployment.feature.failover-leadership                 Support for leadership in fail-over mode - Required ArangoDB >= 3.8.0, < 3.12
      --deployment.feature.gateway                             Defines if gateway extension is enabled - Required ArangoDB >= 3.8.0 (default true)
      --deployment.feature.init-containers-copy-resources      Copy resources spec to built-in init containers if they are not specified - Required ArangoDB >= 3.8.0 (default true)
      --deployment.feature.init-containers-upscale-resources   Copy resources spec to built-in init containers if they are not specified or lower - Required ArangoDB >= 3.8.0 (default true)
      --deployment.feature.local-storage.pass-reclaim-policy   [LocalStorage] Pass ReclaimPolicy from StorageClass instead of using hardcoded Retain - Required ArangoDB >= 3.8.0
      --deployment.feature.local-volume-replacement-check      Replace volume for local-storage if volume is unschedulable (ex. node is gone) - Required ArangoDB >= 3.8.0
      --deployment.feature.random-pod-names                    Enables generating random pod names - Required ArangoDB >= 3.8.0
      --deployment.feature.rebalancer-v2                       Rebalancer V2 feature - Required ArangoDB >= 3.10.0
      --deployment.feature.replace-migration                   During member replacement shards are migrated directly to the new server - Required ArangoDB >= 3.8.0 (default true)
      --deployment.feature.restart-policy-always               Allow to restart containers with always restart policy - Required ArangoDB >= 3.8.0
      --deployment.feature.secured-containers                  Create server's containers with non root privileges. It enables 'ephemeral-volumes' feature implicitly - Required ArangoDB >= 3.8.0
      --deployment.feature.sensitive-information-protection    Hide sensitive information from metrics and logs - Required ArangoDB >= 3.8.0
      --deployment.feature.short-pod-names                     Enable Short Pod Names - Required ArangoDB >= 3.8.0
      --deployment.feature.timezone-management                 Enable timezone management for pods - Required ArangoDB >= 3.8.0
      --deployment.feature.tls-sni                             TLS SNI Support - Required ArangoDB EE >= 3.8.0 (default true)
      --deployment.feature.upgrade-version-check               Enable initContainer with pre version check - Required ArangoDB >= 3.8.0 (default true)
      --deployment.feature.upgrade-version-check-v2            Enable initContainer with pre version check based by Operator - Required ArangoDB >= 3.8.0
      --features-config-map-name string                        Name of the Feature Map ConfigMap (default "arangodb-operator-feature-config-map")
  -h, --help                                                   help for arangodb_operator
      --http1.keep-alive                                       If false, disables HTTP keep-alives and will only use the connection to the server for a single HTTP request (default true)
      --http1.transport.dial-timeout duration                  Maximum amount of time a dial will wait for a connect to complete (default 30s)
      --http1.transport.idle-conn-timeout duration             Maximum amount of time an idle (keep-alive) connection will remain idle before closing itself. Zero means no limit (default 1m30s)
      --http1.transport.idle-conn-timeout-short duration       Maximum amount of time an idle (keep-alive) connection will remain idle before closing itself. Zero means no limit (default 100ms)
      --http1.transport.keep-alive-timeout duration            Interval between keep-alive probes for an active network connection (default 1m30s)
      --http1.transport.keep-alive-timeout-short duration      Interval between keep-alive probes for an active network connection (default 100ms)
      --http1.transport.max-idle-conns int                     Maximum number of idle (keep-alive) connections across all hosts. Zero means no limit (default 100)
      --http1.transport.tls-handshake-timeout duration         Maximum amount of time to wait for a TLS handshake. Zero means no timeout (default 10s)
      --image.discovery.status                                 Discover Operator Image from Pod Status by default. When disabled Pod Spec is used. (default true)
      --image.discovery.timeout duration                       Timeout for image discovery process (default 1m0s)
      --internal.scaling-integration                           Enable Scaling Integration
      --kubernetes.burst int                                   Burst for the k8s API (default 256)
      --kubernetes.max-batch-size int                          Size of batch during objects read (default 256)
      --kubernetes.qps float32                                 Number of queries per second for k8s API (default 32)
      --leader.label.skip                                      Skips Leader Label for the Pod
      --log.format string                                      Set log format. Allowed values: 'pretty', 'JSON'. If empty, default format is used (default "pretty")
      --log.level stringArray                                  Set log levels in format <level> or <logger>=<level>. Possible loggers: action, agency, api-server, assertion, backup-operator, chaos-monkey, crd, deployment, deployment-ci, deployment-reconcile, deployment-replication, deployment-resilience, deployment-resources, deployment-storage, deployment-storage-pc, deployment-storage-service, generic-parent-operator, helm, http, inspector, installer, integration-authn-v1, integration-config-v1, integration-envoy-auth-v3, integration-envoy-auth-v3-impl-auth-bearer, integration-envoy-auth-v3-impl-auth-cookie, integration-envoy-auth-v3-impl-custom-openid, integration-envoy-auth-v3-impl-pass-mode, integration-meta-v1, integration-scheduler-v2, integration-storage-v1-s3, integration-storage-v2, integrations, k8s-client, kubernetes, kubernetes-access, kubernetes-client, kubernetes-informer, monitor, networking-route-operator, operator, operator-arangojob-handler, operator-v2, operator-v2-event, operator-v2-worker, panics, platform-chart-operator, platform-pod-shutdown, platform-service-operator, platform-storage-operator, pod_compare, root, root-event-recorder, scheduler-batchjob-operator, scheduler-cronjob-operator, scheduler-deployment-operator, scheduler-pod-operator, scheduler-profile-operator, server, server-authentication, webhook (default [info])
      --log.sampling                                           If true, operator will try to minimize duplication of logging events (default true)
      --log.stdout                                             If true, operator will log to the stdout (default true)
      --memory-limit uint                                      Define memory limit for hard shutdown and the dump of goroutines. Used for testing
      --metrics.excluded-prefixes stringArray                  List of the excluded metrics prefixes
      --mode.single                                            Enable single mode in Operator. WARNING: There should be only one replica of Operator, otherwise Operator can take unexpected actions
      --operator.analytics                                     Enable to run the Analytics operator
      --operator.apps                                          Enable to run the ArangoApps operator
      --operator.backup                                        Enable to run the ArangoBackup operator
      --operator.deployment                                    Enable to run the ArangoDeployment operator
      --operator.deployment-replication                        Enable to run the ArangoDeploymentReplication operator
      --operator.ml                                            Enable to run the ArangoML operator
      --operator.networking                                    Enable to run the Networking operator
      --operator.platform                                      Enable to run the Platform operator
      --operator.reconciliation.retry.count int                Count of retries during Object Update operations in the Reconciliation loop (default 25)
      --operator.reconciliation.retry.delay duration           Delay between Object Update operations in the Reconciliation loop (default 1s)
      --operator.scheduler                                     Enable to run the Scheduler operator
      --operator.storage                                       Enable to run the ArangoLocalStorage operator
      --operator.version                                       Enable only version endpoint in Operator
      --reconciliation.delay duration                          Delay between reconciliation loops (<= 0 -> Disabled)
      --server.admin-secret-name string                        Name of secret containing username + password for login to the dashboard (default "arangodb-operator-dashboard")
      --server.allow-anonymous-access                          Allow anonymous access to the dashboard
      --server.host string                                     Host to listen on (default "0.0.0.0")
      --server.port int                                        Port to listen on (default 8528)
      --server.tls-secret-name string                          Name of secret containing tls.crt & tls.key for HTTPS server (if empty, self-signed certificate is used)
      --shutdown.delay duration                                The delay before running shutdown handlers (default 2s)
      --shutdown.timeout duration                              Timeout for shutdown handlers (default 30s)
      --timeout.agency duration                                The Agency read timeout (default 10s)
      --timeout.arangod duration                               The request timeout to the ArangoDB (default 5s)
      --timeout.arangod-check duration                         The version check request timeout to the ArangoDB (default 2s)
      --timeout.backup-arangod duration                        The request timeout to the ArangoDB during backup calls (default 30s)
      --timeout.backup-upload duration                         The request timeout to the ArangoDB during uploading files (default 5m0s)
      --timeout.force-delete-pod-grace-period duration         Default period when ArangoDB Pod should be forcefully removed after all containers were stopped - set to 0 to disable forceful removals (default 15m0s)
      --timeout.k8s duration                                   The request timeout to the kubernetes (default 2s)
      --timeout.pod-scheduling-grace-period duration           Default period when ArangoDB Pod should be deleted in case of scheduling info change - set to 0 to disable (default 15s)
      --timeout.reconciliation duration                        The reconciliation timeout to the ArangoDB CR (default 1m0s)
      --timeout.shard-rebuild duration                         Timeout after which particular out-synced shard is considered as failed and rebuild is triggered (default 1h0m0s)
      --timeout.shard-rebuild-retry duration                   Timeout after which rebuild shards retry flow is triggered (default 4h0m0s)

Use "arangodb_operator [command] --help" for more information about a command.
```
[END_INJECT]: # (arangodb_operator_cmd)
