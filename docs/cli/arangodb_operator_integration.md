---
layout: page
parent: Binaries
title: arangodb_operator_integration
---

# ArangoDB Operator Integration Command

[START_INJECT]: # (arangodb_operator_integration_cmd)
```
Usage:
  arangodb_operator_integration [flags]
  arangodb_operator_integration [command]

Available Commands:
  client      
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command

Flags:
      --health.address string                                       Address to expose health service (Env: HEALTH_ADDRESS) (default "0.0.0.0:9091")
      --health.auth.token string                                    Token for health service (when auth service is token) (Env: HEALTH_AUTH_TOKEN)
      --health.auth.type string                                     Auth type for health service (Env: HEALTH_AUTH_TYPE) (default "None")
      --health.shutdown.enabled                                     Determines if shutdown service should be enabled and exposed (Env: HEALTH_SHUTDOWN_ENABLED) (default true)
      --health.tls.keyfile string                                   Path to the keyfile (Env: HEALTH_TLS_KEYFILE)
  -h, --help                                                        help for arangodb_operator_integration
      --integration.authentication.v1                               Enable AuthenticationV1 Integration Service (Env: INTEGRATION_AUTHENTICATION_V1)
      --integration.authentication.v1.enabled                       Defines if Authentication is enabled (Env: INTEGRATION_AUTHENTICATION_V1_ENABLED) (default true)
      --integration.authentication.v1.external                      Defones if External access to service authentication.v1 is enabled (Env: INTEGRATION_AUTHENTICATION_V1_EXTERNAL)
      --integration.authentication.v1.internal                      Defones if Internal access to service authentication.v1 is enabled (Env: INTEGRATION_AUTHENTICATION_V1_INTERNAL) (default true)
      --integration.authentication.v1.path string                   Path to the JWT Folder (Env: INTEGRATION_AUTHENTICATION_V1_PATH)
      --integration.authentication.v1.token.allowed strings         Allowed users for the Token (Env: INTEGRATION_AUTHENTICATION_V1_TOKEN_ALLOWED)
      --integration.authentication.v1.token.max-size uint16         Max Token max size in bytes (Env: INTEGRATION_AUTHENTICATION_V1_TOKEN_MAX_SIZE) (default 64)
      --integration.authentication.v1.token.ttl.default duration    Default Token TTL (Env: INTEGRATION_AUTHENTICATION_V1_TOKEN_TTL_DEFAULT) (default 1h0m0s)
      --integration.authentication.v1.token.ttl.max duration        Max Token TTL (Env: INTEGRATION_AUTHENTICATION_V1_TOKEN_TTL_MAX) (default 1h0m0s)
      --integration.authentication.v1.token.ttl.min duration        Min Token TTL (Env: INTEGRATION_AUTHENTICATION_V1_TOKEN_TTL_MIN) (default 1m0s)
      --integration.authentication.v1.token.user string             Default user of the Token (Env: INTEGRATION_AUTHENTICATION_V1_TOKEN_USER) (default "root")
      --integration.authentication.v1.ttl duration                  TTL of the JWT cache (Env: INTEGRATION_AUTHENTICATION_V1_TTL) (default 15s)
      --integration.authorization.v0                                Enable AuthorizationV0 Integration Service (Env: INTEGRATION_AUTHORIZATION_V0)
      --integration.authorization.v0.external                       Defones if External access to service authorization.v0 is enabled (Env: INTEGRATION_AUTHORIZATION_V0_EXTERNAL)
      --integration.authorization.v0.internal                       Defones if Internal access to service authorization.v0 is enabled (Env: INTEGRATION_AUTHORIZATION_V0_INTERNAL) (default true)
      --integration.config.v1                                       Enable ConfigV1 Integration Service (Env: INTEGRATION_CONFIG_V1)
      --integration.config.v1.external                              Defones if External access to service config.v1 is enabled (Env: INTEGRATION_CONFIG_V1_EXTERNAL)
      --integration.config.v1.internal                              Defones if Internal access to service config.v1 is enabled (Env: INTEGRATION_CONFIG_V1_INTERNAL) (default true)
      --integration.config.v1.module strings                        Module in the reference <name>=<abs path> (Env: INTEGRATION_CONFIG_V1_MODULE)
      --integration.envoy.auth.v3                                   Enable EnvoyAuthV3 Integration Service (Env: INTEGRATION_ENVOY_AUTH_V3)
      --integration.envoy.auth.v3.external                          Defones if External access to service envoy.auth.v3 is enabled (Env: INTEGRATION_ENVOY_AUTH_V3_EXTERNAL)
      --integration.envoy.auth.v3.internal                          Defones if Internal access to service envoy.auth.v3 is enabled (Env: INTEGRATION_ENVOY_AUTH_V3_INTERNAL) (default true)
      --integration.scheduler.v1                                    SchedulerV1 Integration (Env: INTEGRATION_SCHEDULER_V1)
      --integration.scheduler.v1.external                           Defones if External access to service scheduler.v1 is enabled (Env: INTEGRATION_SCHEDULER_V1_EXTERNAL)
      --integration.scheduler.v1.internal                           Defones if Internal access to service scheduler.v1 is enabled (Env: INTEGRATION_SCHEDULER_V1_INTERNAL) (default true)
      --integration.scheduler.v1.namespace string                   Kubernetes Namespace (Env: INTEGRATION_SCHEDULER_V1_NAMESPACE) (default "default")
      --integration.scheduler.v1.verify-access                      Verify the CRD Access (Env: INTEGRATION_SCHEDULER_V1_VERIFY_ACCESS) (default true)
      --integration.scheduler.v2                                    SchedulerV2 Integration (Env: INTEGRATION_SCHEDULER_V2)
      --integration.scheduler.v2.deployment string                  ArangoDeployment Name (Env: INTEGRATION_SCHEDULER_V2_DEPLOYMENT)
      --integration.scheduler.v2.external                           Defones if External access to service scheduler.v2 is enabled (Env: INTEGRATION_SCHEDULER_V2_EXTERNAL)
      --integration.scheduler.v2.internal                           Defones if Internal access to service scheduler.v2 is enabled (Env: INTEGRATION_SCHEDULER_V2_INTERNAL) (default true)
      --integration.scheduler.v2.namespace string                   Kubernetes Namespace (Env: INTEGRATION_SCHEDULER_V2_NAMESPACE) (default "default")
      --integration.shutdown.v1                                     ShutdownV1 Handler (Env: INTEGRATION_SHUTDOWN_V1)
      --integration.shutdown.v1.external                            Defones if External access to service shutdown.v1 is enabled (Env: INTEGRATION_SHUTDOWN_V1_EXTERNAL)
      --integration.shutdown.v1.internal                            Defones if Internal access to service shutdown.v1 is enabled (Env: INTEGRATION_SHUTDOWN_V1_INTERNAL) (default true)
      --integration.storage.v2                                      StorageBucket V2 Integration (Env: INTEGRATION_STORAGE_V2)
      --integration.storage.v2.external                             Defones if External access to service storage.v2 is enabled (Env: INTEGRATION_STORAGE_V2_EXTERNAL)
      --integration.storage.v2.internal                             Defones if Internal access to service storage.v2 is enabled (Env: INTEGRATION_STORAGE_V2_INTERNAL) (default true)
      --integration.storage.v2.s3.allow-insecure                    If set to true, the Endpoint certificates won't be checked (Env: INTEGRATION_STORAGE_V2_S3_ALLOW_INSECURE)
      --integration.storage.v2.s3.bucket.name string                Bucket name (Env: INTEGRATION_STORAGE_V2_S3_BUCKET_NAME)
      --integration.storage.v2.s3.bucket.prefix string              Bucket Prefix (Env: INTEGRATION_STORAGE_V2_S3_BUCKET_PREFIX)
      --integration.storage.v2.s3.ca strings                        Path to file containing CA certificate to validate endpoint connection (Env: INTEGRATION_STORAGE_V2_S3_CA)
      --integration.storage.v2.s3.disable-ssl                       If set to true, the SSL won't be used when connecting to Endpoint (Env: INTEGRATION_STORAGE_V2_S3_DISABLE_SSL)
      --integration.storage.v2.s3.endpoint string                   Endpoint of S3 API implementation (Env: INTEGRATION_STORAGE_V2_S3_ENDPOINT)
      --integration.storage.v2.s3.provider.file.access-key string   Path to file containing S3 AccessKey (Env: INTEGRATION_STORAGE_V2_S3_PROVIDER_FILE_ACCESS_KEY)
      --integration.storage.v2.s3.provider.file.secret-key string   Path to file containing S3 SecretKey (Env: INTEGRATION_STORAGE_V2_S3_PROVIDER_FILE_SECRET_KEY)
      --integration.storage.v2.s3.provider.type string              S3 Credentials Provider type (Env: INTEGRATION_STORAGE_V2_S3_PROVIDER_TYPE) (default "file")
      --integration.storage.v2.s3.region string                     Region (Env: INTEGRATION_STORAGE_V2_S3_REGION)
      --integration.storage.v2.type string                          Type of the Storage Integration (Env: INTEGRATION_STORAGE_V2_TYPE) (default "s3")
      --services.address string                                     Address to expose internal services (Env: SERVICES_ADDRESS) (default "127.0.0.1:9092")
      --services.auth.token string                                  Token for internal service (when auth service is token) (Env: SERVICES_AUTH_TOKEN)
      --services.auth.type string                                   Auth type for internal service (Env: SERVICES_AUTH_TYPE) (default "None")
      --services.enabled                                            Defines if internal access is enabled (Env: SERVICES_ENABLED) (default true)
      --services.external.address string                            Address to expose external services (Env: SERVICES_EXTERNAL_ADDRESS) (default "0.0.0.0:9093")
      --services.external.auth.token string                         Token for external service (when auth service is token) (Env: SERVICES_EXTERNAL_AUTH_TOKEN)
      --services.external.auth.type string                          Auth type for external service (Env: SERVICES_EXTERNAL_AUTH_TYPE) (default "None")
      --services.external.enabled                                   Defines if external access is enabled (Env: SERVICES_EXTERNAL_ENABLED)
      --services.external.tls.keyfile string                        Path to the keyfile (Env: SERVICES_EXTERNAL_TLS_KEYFILE)
      --services.tls.keyfile string                                 Path to the keyfile (Env: SERVICES_TLS_KEYFILE)

Use "arangodb_operator_integration [command] --help" for more information about a command.
```
[END_INJECT]: # (arangodb_operator_integration_cmd)

# ArangoDB Operator Integration Client Subcommand

[START_INJECT]: # (arangodb_operator_integration_cmd_client)
```
Usage:
  arangodb_operator_integration client [command]

Available Commands:
  health      
  pong        
  shutdown    

Flags:
      --address string   GRPC Service Address (default "127.0.0.1:8080")
  -h, --help             help for client
      --tls.ca string    Path to the custom CA
      --tls.enabled      Defines if GRPC is protected with TLS
      --tls.fallback     Enables TLS Fallback
      --tls.insecure     Enables Insecure TLS Connection
      --token string     GRPC Token

Use "arangodb_operator_integration client [command] --help" for more information about a command.
```
[END_INJECT]: # (arangodb_operator_integration_cmd_client)
