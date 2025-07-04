---
layout: page
parent: Binaries
title: arangodb_operator_platform
---

# ArangoDB Operator Platform Package Command

[START_INJECT]: # (arangodb_operator_platform_package_cmd)
```
Release Package related operations

Usage:
  arangodb_operator_platform package [command]

Available Commands:
  dump        Dumps the current setup of the platform
  export      Export the package in the ZIP Format
  import      Imports the package from the ZIP format
  install     Installs the specified setup of the platform
  merge       Merges definitions into single file

Flags:
  -h, --help   help for package

Global Flags:
      --kubeconfig string   Kubernetes Config File
  -n, --namespace string    Kubernetes Namespace (default "default")

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
  -h, --help                   help for dump
      --platform.name string   Kubernetes Platform Name (name of the ArangoDeployment)

Global Flags:
      --kubeconfig string   Kubernetes Config File
  -n, --namespace string    Kubernetes Namespace (default "default")
```
[END_INJECT]: # (arangodb_operator_platform_package_dump_cmd)

# ArangoDB Operator Platform Package Install Command

[START_INJECT]: # (arangodb_operator_platform_package_install_cmd)
```
Installs the specified setup of the platform

Usage:
  arangodb_operator_platform package install [flags] ... packages

Flags:
  -h, --help                       help for install
      --platform.endpoint string   Platform Repository URL (default "https://arangodb-platform-prd-chart-registry.s3.amazonaws.com")
      --platform.name string       Kubernetes Platform Name (name of the ArangoDeployment)

Global Flags:
      --kubeconfig string   Kubernetes Config File
  -n, --namespace string    Kubernetes Namespace (default "default")
```
[END_INJECT]: # (arangodb_operator_platform_package_install_cmd)

