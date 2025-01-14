---
layout: page
parent: Binaries
title: arangodb_operator_ops
---

# ArangoDB Operator Ops Command

[START_INJECT]: # (arangodb_operator_ops_cmd)
```
Usage:
  arangodb_operator_ops [flags]
  arangodb_operator_ops [command]

Available Commands:
  completion    Generate the autocompletion script for the specified shell
  crd           CRD operations
  debug-package Generate debug package for debugging
  help          Help about any command
  task          

Flags:
  -h, --help   help for arangodb_operator_ops

Use "arangodb_operator_ops [command] --help" for more information about a command.
```
[END_INJECT]: # (arangodb_operator_ops_cmd)

# ArangoDB Operator Ops CRD Subcommand

[START_INJECT]: # (arangodb_operator_ops_cmd_crd)
```
CRD operations

Usage:
  arangodb_operator_ops crd [flags]
  arangodb_operator_ops crd [command]

Available Commands:
  generate    Generates YAML of all required CRDs
  install     Install and update all required CRDs

Flags:
      --crd.force-update                          Enforce CRD Schema update
      --crd.preserve-unknown-fields stringArray   Controls which CRD should have enabled preserve unknown fields in validation schema <crd-name>=<true/false>.
      --crd.validation-schema stringArray         Controls which CRD should have validation schema <crd-name>=<true/false>.
  -h, --help                                      help for crd

Use "arangodb_operator_ops crd [command] --help" for more information about a command.
```
[END_INJECT]: # (arangodb_operator_ops_cmd_crd)

# ArangoDB Operator Ops CRD Install Subcommand

[START_INJECT]: # (arangodb_operator_ops_cmd_crd_install)
```
Install and update all required CRDs

Usage:
  arangodb_operator_ops crd install [flags]

Flags:
  -h, --help   help for install

Global Flags:
      --crd.force-update                          Enforce CRD Schema update
      --crd.preserve-unknown-fields stringArray   Controls which CRD should have enabled preserve unknown fields in validation schema <crd-name>=<true/false>.
      --crd.validation-schema stringArray         Controls which CRD should have validation schema <crd-name>=<true/false>.
```
[END_INJECT]: # (arangodb_operator_ops_cmd_crd_install)

# ArangoDB Operator Ops CRD Generate Subcommand

[START_INJECT]: # (arangodb_operator_ops_cmd_crd_generate)
```
Generates YAML of all required CRDs

Usage:
  arangodb_operator_ops crd generate [flags]

Flags:
  -h, --help   help for generate

Global Flags:
      --crd.force-update                          Enforce CRD Schema update
      --crd.preserve-unknown-fields stringArray   Controls which CRD should have enabled preserve unknown fields in validation schema <crd-name>=<true/false>.
      --crd.validation-schema stringArray         Controls which CRD should have validation schema <crd-name>=<true/false>.
```
[END_INJECT]: # (arangodb_operator_ops_cmd_crd_generate)

# ArangoDB Operator Ops CRD Install Subcommand

[START_INJECT]: # (arangodb_operator_ops_cmd_debug_package)
```
Generate debug package for debugging

Usage:
  arangodb_operator_ops debug-package [flags]

Flags:
      --generator.agency-dump             Define if generator agency-dump is enabled (default true)
      --generator.analytics               Define if generator analytics is enabled (default true)
      --generator.backupBackup            Define if generator backupBackup is enabled (default true)
      --generator.deployments             Define if generator deployments is enabled (default true)
      --generator.kubernetes-configmaps   Define if generator kubernetes-configmaps is enabled (default true)
      --generator.kubernetes-events       Define if generator kubernetes-events is enabled (default true)
      --generator.kubernetes-pods         Define if generator kubernetes-pods is enabled (default true)
      --generator.kubernetes-secrets      Define if generator kubernetes-secrets is enabled (default true)
      --generator.kubernetes-services     Define if generator kubernetes-services is enabled (default true)
      --generator.ml                      Define if generator ml is enabled (default true)
      --generator.networking              Define if generator networking is enabled (default true)
      --generator.platform                Define if generator platform is enabled (default true)
      --generator.scheduler               Define if generator scheduler is enabled (default true)
  -h, --help                              help for debug-package
      --hide-sensitive-data               Hide sensitive data (default true)
  -n, --namespace string                  Kubernetes namespace (default "default")
  -o, --output -                          Output of the result gz file. If set to - then stdout is used (default "out.tar.gz")
      --pod-logs                          Collect pod logs (default true)
```
[END_INJECT]: # (arangodb_operator_ops_cmd_debug_package)
