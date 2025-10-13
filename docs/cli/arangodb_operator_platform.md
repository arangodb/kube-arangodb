---
layout: page
parent: Binaries
title: arangodb_operator_platform
---

# ArangoDB Operator Platform License Command

[START_INJECT]: # (arangodb_operator_platform_license_cmd)
```
License related Operations

Usage:
  arangodb_operator_platform license [command]

Available Commands:
  activate    Activates the License on ArangoDB Endpoint
  generate    Generate the License
  secret      Creates Platform Secret with Registry credentials

Flags:
  -h, --help   help for license

Global Flags:
      --kubeconfig string   Kubernetes Config File
  -n, --namespace string    Kubernetes Namespace (default "default")

Use "arangodb_operator_platform license [command] --help" for more information about a command.
```
[END_INJECT]: # (arangodb_operator_platform_license_cmd)

# ArangoDB Operator Platform License Activate Command

[START_INJECT]: # (arangodb_operator_platform_license_activate_cmd)
```
Activates the License on ArangoDB Endpoint

Usage:
  arangodb_operator_platform license activate [flags]

Flags:
      --arangodb.endpoint string       ArangoDB Endpoint
  -h, --help                           help for activate
      --license.client.id string       LicenseManager Client ID
      --license.client.secret string   LicenseManager Client Secret
      --license.endpoint string        LicenseManager Endpoint (default "license.arangodb.com")
      --license.interval duration      Interval of the license synchronization

Global Flags:
      --kubeconfig string   Kubernetes Config File
  -n, --namespace string    Kubernetes Namespace (default "default")
```
[END_INJECT]: # (arangodb_operator_platform_license_activate_cmd)

# ArangoDB Operator Platform License Generate Command

[START_INJECT]: # (arangodb_operator_platform_license_generate_cmd)
```
Generate the License

Usage:
  arangodb_operator_platform license generate [flags]

Flags:
      --deployment.id string           Deployment ID
  -h, --help                           help for generate
      --license.client.id string       LicenseManager Client ID
      --license.client.secret string   LicenseManager Client Secret
      --license.endpoint string        LicenseManager Endpoint (default "license.arangodb.com")

Global Flags:
      --kubeconfig string   Kubernetes Config File
  -n, --namespace string    Kubernetes Namespace (default "default")
```
[END_INJECT]: # (arangodb_operator_platform_license_generate_cmd)

# ArangoDB Operator Platform License Secret Command

[START_INJECT]: # (arangodb_operator_platform_license_secret_cmd)
```
Creates Platform Secret with Registry credentials

Usage:
  arangodb_operator_platform license secret [flags]

Flags:
  -h, --help                           help for secret
      --license.client.id string       LicenseManager Client ID
      --license.client.secret string   LicenseManager Client Secret
      --license.client.stage strings   LicenseManager Stages (default [prd])
      --license.endpoint string        LicenseManager Endpoint (default "license.arangodb.com")
      --secret string                  Kubernetes Secret Name

Global Flags:
      --kubeconfig string   Kubernetes Config File
  -n, --namespace string    Kubernetes Namespace (default "default")
```
[END_INJECT]: # (arangodb_operator_platform_license_secret_cmd)

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
  registry    Points all images to the new registry

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

