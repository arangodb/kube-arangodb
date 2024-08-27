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
  completion  Generate the autocompletion script for the specified shell
  crd         CRD operations
  help        Help about any command
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
