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
      --health.address string                                      Address to expose health service (default "0.0.0.0:9091")
      --health.auth.token string                                   Token for health service (when auth service is token)
      --health.auth.type string                                    Auth type for health service (default "None")
      --health.shutdown.enabled                                    Determines if shutdown service should be enabled and exposed (default true)
  -h, --help                                                       help for arangodb_operator_integration
      --integration.authentication.v1                              Enable AuthenticationV1 Integration Service
      --integration.authentication.v1.enabled                      Defines if Authentication is enabled (default true)
      --integration.authentication.v1.external                     Defones if External access to service authentication.v1 is enabled
      --integration.authentication.v1.internal                     Defones if Internal access to service authentication.v1 is enabled (default true)
      --integration.authentication.v1.path string                  Path to the JWT Folder
      --integration.authentication.v1.token.allowed strings        Allowed users for the Token
      --integration.authentication.v1.token.max-size uint16        Max Token max size in bytes (default 64)
      --integration.authentication.v1.token.ttl.default duration   Default Token TTL (default 1h0m0s)
      --integration.authentication.v1.token.ttl.max duration       Max Token TTL (default 1h0m0s)
      --integration.authentication.v1.token.ttl.min duration       Min Token TTL (default 1m0s)
      --integration.authentication.v1.token.user string            Default user of the Token (default "root")
      --integration.authentication.v1.ttl duration                 TTL of the JWT cache (default 15s)
      --integration.authorization.v0                               Enable AuthorizationV0 Integration Service
      --integration.authorization.v0.external                      Defones if External access to service authorization.v0 is enabled
      --integration.authorization.v0.internal                      Defones if Internal access to service authorization.v0 is enabled (default true)
      --integration.config.v1                                      Enable ConfigV1 Integration Service
      --integration.config.v1.external                             Defones if External access to service config.v1 is enabled
      --integration.config.v1.internal                             Defones if Internal access to service config.v1 is enabled (default true)
      --integration.config.v1.module strings                       Module in the reference <name>=<abs path>
      --integration.envoy.auth.v3                                  Enable EnvoyAuthV3 Integration Service
      --integration.envoy.auth.v3.external                         Defones if External access to service envoy.auth.v3 is enabled
      --integration.envoy.auth.v3.internal                         Defones if Internal access to service envoy.auth.v3 is enabled (default true)
      --integration.scheduler.v1                                   SchedulerV1 Integration
      --integration.scheduler.v1.external                          Defones if External access to service scheduler.v1 is enabled
      --integration.scheduler.v1.internal                          Defones if Internal access to service scheduler.v1 is enabled (default true)
      --integration.scheduler.v1.namespace string                  Kubernetes Namespace (default "default")
      --integration.scheduler.v1.verify-access                     Verify the CRD Access (default true)
      --integration.shutdown.v1                                    ShutdownV1 Handler
      --integration.shutdown.v1.external                           Defones if External access to service shutdown.v1 is enabled
      --integration.shutdown.v1.internal                           Defones if Internal access to service shutdown.v1 is enabled (default true)
      --integration.storage.v1                                     StorageBucket Integration
      --integration.storage.v1.external                            Defones if External access to service storage.v1 is enabled
      --integration.storage.v1.internal                            Defones if Internal access to service storage.v1 is enabled (default true)
      --integration.storage.v1.s3.access-key string                Path to file containing S3 AccessKey
      --integration.storage.v1.s3.allow-insecure                   If set to true, the Endpoint certificates won't be checked
      --integration.storage.v1.s3.bucket string                    Bucket name
      --integration.storage.v1.s3.ca-crt string                    Path to file containing CA certificate to validate endpoint connection
      --integration.storage.v1.s3.ca-key string                    Path to file containing keyfile to validate endpoint connection
      --integration.storage.v1.s3.disable-ssl                      If set to true, the SSL won't be used when connecting to Endpoint
      --integration.storage.v1.s3.endpoint string                  Endpoint of S3 API implementation
      --integration.storage.v1.s3.region string                    Region
      --integration.storage.v1.s3.secret-key string                Path to file containing S3 SecretKey
      --integration.storage.v1.type string                         Type of the Storage Integration (default "s3")
      --services.address string                                    Address to expose internal services (default "127.0.0.1:9092")
      --services.auth.token string                                 Token for internal service (when auth service is token)
      --services.auth.type string                                  Auth type for internal service (default "None")
      --services.enabled                                           Defines if internal access is enabled (default true)
      --services.external.address string                           Address to expose external services (default "0.0.0.0:9093")
      --services.external.auth.token string                        Token for external service (when auth service is token)
      --services.external.auth.type string                         Auth type for external service (default "None")
      --services.external.enabled                                  Defines if external access is enabled

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
      --token string     GRPC Token

Use "arangodb_operator_integration client [command] --help" for more information about a command.
```
[END_INJECT]: # (arangodb_operator_integration_cmd_client)
