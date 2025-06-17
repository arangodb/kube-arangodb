---
layout: page
parent: Binaries
title: arangodb_operator_platform
---

# ArangoDB Operator Platform Command

[START_INJECT]: # (arangodb_operator_platform_cmd)
```
Usage:
  arangodb_operator_platform [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  package     Release Package related operations
  registry    Registry related operations
  service     Service related operations

Flags:
  -h, --help               help for arangodb_operator_platform
  -n, --namespace string   Kubernetes Namespace (default "default")

Use "arangodb_operator_platform [command] --help" for more information about a command.
```
[END_INJECT]: # (arangodb_operator_platform_cmd)

# ArangoDB Operator Platform Registry Command

[START_INJECT]: # (arangodb_operator_platform_registry_cmd)
```
Registry related operations

Usage:
  arangodb_operator_platform registry [command]

Available Commands:
  install     Manages the Chart Installation
  status      Describes Charts Status

Flags:
  -h, --help                       help for registry
      --platform.endpoint string   Platform Repository URL (default "https://arangodb-platform-prd-chart-registry.s3.amazonaws.com")
      --platform.name string       Kubernetes Platform Name (name of the ArangoDeployment)

Global Flags:
  -n, --namespace string   Kubernetes Namespace (default "default")

Use "arangodb_operator_platform registry [command] --help" for more information about a command.
```
[END_INJECT]: # (arangodb_operator_platform_registry_cmd)

# ArangoDB Operator Platform Registry Install Command

[START_INJECT]: # (arangodb_operator_platform_registry_install_cmd)
```
Manages the Chart Installation

Usage:
  arangodb_operator_platform registry install [flags] [...charts]

Flags:
  -a, --all                        Runs on all items
  -h, --help                       help for install
  -o, --output string              Output format. Allowed table, json, yaml (default "table")
      --platform.endpoint string   Platform Repository URL (default "https://arangodb-platform-prd-chart-registry.s3.amazonaws.com")
      --platform.name string       Kubernetes Platform Name (name of the ArangoDeployment)
  -u, --upgrade                    Enable upgrade procedure

Global Flags:
  -n, --namespace string   Kubernetes Namespace (default "default")
```
[END_INJECT]: # (arangodb_operator_platform_registry_install_cmd)

# ArangoDB Operator Platform Registry Status Command

[START_INJECT]: # (arangodb_operator_platform_registry_status_cmd)
```
Describes Charts Status

Usage:
  arangodb_operator_platform registry status [flags]

Flags:
  -h, --help                       help for status
  -o, --output string              Output format. Allowed table, json, yaml (default "table")
      --platform.endpoint string   Platform Repository URL (default "https://arangodb-platform-prd-chart-registry.s3.amazonaws.com")
      --platform.name string       Kubernetes Platform Name (name of the ArangoDeployment)

Global Flags:
  -n, --namespace string   Kubernetes Namespace (default "default")
```
[END_INJECT]: # (arangodb_operator_platform_registry_status_cmd)

# ArangoDB Operator Platform Service Command

[START_INJECT]: # (arangodb_operator_platform_service_cmd)
```
Service related operations

Usage:
  arangodb_operator_platform service [command]

Available Commands:
  enable         Manages Service Installation/Management
  enable-service Manages Service Installation/Management
  status         Shows Service Status

Flags:
  -h, --help                       help for service
      --platform.endpoint string   Platform Repository URL (default "https://arangodb-platform-prd-chart-registry.s3.amazonaws.com")
      --platform.name string       Kubernetes Platform Name (name of the ArangoDeployment)

Global Flags:
  -n, --namespace string   Kubernetes Namespace (default "default")

Use "arangodb_operator_platform service [command] --help" for more information about a command.
```
[END_INJECT]: # (arangodb_operator_platform_service_cmd)

# ArangoDB Operator Platform Service Enable Command

[START_INJECT]: # (arangodb_operator_platform_service_enable_cmd)
```
Manages Service Installation/Management

Usage:
  arangodb_operator_platform service enable [flags] deployment name chart

Flags:
  -h, --help                       help for enable
      --platform.endpoint string   Platform Repository URL (default "https://arangodb-platform-prd-chart-registry.s3.amazonaws.com")
      --platform.name string       Kubernetes Platform Name (name of the ArangoDeployment)
  -f, --values strings             Chart values

Global Flags:
  -n, --namespace string   Kubernetes Namespace (default "default")
```
[END_INJECT]: # (arangodb_operator_platform_service_enable_cmd)

# ArangoDB Operator Platform Service EnableService Command

[START_INJECT]: # (arangodb_operator_platform_service_enableservice_cmd)
```
Manages Service Installation/Management

Usage:
  arangodb_operator_platform service enable-service [flags] deployment chart

Flags:
  -h, --help                       help for enable-service
      --platform.endpoint string   Platform Repository URL (default "https://arangodb-platform-prd-chart-registry.s3.amazonaws.com")
      --platform.name string       Kubernetes Platform Name (name of the ArangoDeployment)
  -f, --values strings             Chart values

Global Flags:
  -n, --namespace string   Kubernetes Namespace (default "default")
```
[END_INJECT]: # (arangodb_operator_platform_service_enableservice_cmd)

# ArangoDB Operator Platform Service Status Command

[START_INJECT]: # (arangodb_operator_platform_service_status_cmd)
```
Shows Service Status

Usage:
  arangodb_operator_platform service status [flags] deployment

Flags:
  -h, --help                       help for status
  -o, --output string              Output format. Allowed table, json, yaml (default "table")
      --platform.endpoint string   Platform Repository URL (default "https://arangodb-platform-prd-chart-registry.s3.amazonaws.com")
      --platform.name string       Kubernetes Platform Name (name of the ArangoDeployment)

Global Flags:
  -n, --namespace string   Kubernetes Namespace (default "default")
```
[END_INJECT]: # (arangodb_operator_platform_service_status_cmd)

# ArangoDB Operator Platform Package Command

[START_INJECT]: # (arangodb_operator_platform_package_cmd)
```
Release Package related operations

Usage:
  arangodb_operator_platform package [command]

Available Commands:
  dump        Dumps the current setup of the platform
  install     Installs the specified setup of the platform

Flags:
  -h, --help                   help for package
      --platform.name string   Kubernetes Platform Name (name of the ArangoDeployment)

Global Flags:
  -n, --namespace string   Kubernetes Namespace (default "default")

Use "arangodb_operator_platform package [command] --help" for more information about a command.
```
[END_INJECT]: # (arangodb_operator_platform_package_cmd)

# ArangoDB Operator Platform Package Dump Command

[START_INJECT]: # (arangodb_operator_platform_package_dump_cmd)
```
Dumps the current setup of the platform

Usage:
  arangodb_operator_platform package dump [flags]

Flags:
  -h, --help   help for dump

Global Flags:
  -n, --namespace string       Kubernetes Namespace (default "default")
      --platform.name string   Kubernetes Platform Name (name of the ArangoDeployment)
```
[END_INJECT]: # (arangodb_operator_platform_package_dump_cmd)

# ArangoDB Operator Platform Package Install Command

[START_INJECT]: # (arangodb_operator_platform_package_install_cmd)
```
Installs the specified setup of the platform

Usage:
  arangodb_operator_platform package install [flags] package

Flags:
  -h, --help                       help for install
      --platform.endpoint string   Platform Repository URL (default "https://arangodb-platform-prd-chart-registry.s3.amazonaws.com")

Global Flags:
  -n, --namespace string       Kubernetes Namespace (default "default")
      --platform.name string   Kubernetes Platform Name (name of the ArangoDeployment)
```
[END_INJECT]: # (arangodb_operator_platform_package_install_cmd)

