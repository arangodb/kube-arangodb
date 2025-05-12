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
      --crd.skip stringArray                      Controls which CRD should be skipped.
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
      --crd.skip stringArray                      Controls which CRD should be skipped.
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
      --crd.skip stringArray                      Controls which CRD should be skipped.
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
      --generator.arango-analytics-gae          Define if generator arango-analytics-gae is enabled (default true)
      --generator.arango-backup-backup          Define if generator arango-backup-backup is enabled (default true)
      --generator.arango-backup-backuppolicy    Define if generator arango-backup-backuppolicy is enabled (default true)
      --generator.arango-database-acs           Define if generator arango-database-acs is enabled (default true)
      --generator.arango-database-deployment    Define if generator arango-database-deployment is enabled (default true)
      --generator.arango-database-member        Define if generator arango-database-member is enabled (default true)
      --generator.arango-database-task          Define if generator arango-database-task is enabled (default true)
      --generator.arango-ml-batchjob            Define if generator arango-ml-batchjob is enabled (default true)
      --generator.arango-ml-cronjob             Define if generator arango-ml-cronjob is enabled (default true)
      --generator.arango-ml-extension           Define if generator arango-ml-extension is enabled (default true)
      --generator.arango-ml-storage             Define if generator arango-ml-storage is enabled (default true)
      --generator.arango-networking-route       Define if generator arango-networking-route is enabled (default true)
      --generator.arango-platform-chart         Define if generator arango-platform-chart is enabled (default true)
      --generator.arango-platform-storage       Define if generator arango-platform-storage is enabled (default true)
      --generator.arango-scheduler-batchjob     Define if generator arango-scheduler-batchjob is enabled (default true)
      --generator.arango-scheduler-cronjob      Define if generator arango-scheduler-cronjob is enabled (default true)
      --generator.arango-scheduler-deployment   Define if generator arango-scheduler-deployment is enabled (default true)
      --generator.arango-scheduler-pod          Define if generator arango-scheduler-pod is enabled (default true)
      --generator.arango-scheduler-profile      Define if generator arango-scheduler-profile is enabled (default true)
      --generator.helm-releases                 Define if generator helm-releases is enabled (default true)
      --generator.kubernetes-apps-deployment    Define if generator kubernetes-apps-deployment is enabled (default true)
      --generator.kubernetes-apps-replicaset    Define if generator kubernetes-apps-replicaset is enabled (default true)
      --generator.kubernetes-apps-statefulset   Define if generator kubernetes-apps-statefulset is enabled (default true)
      --generator.kubernetes-batch-cronjob      Define if generator kubernetes-batch-cronjob is enabled (default true)
      --generator.kubernetes-batch-job          Define if generator kubernetes-batch-job is enabled (default true)
      --generator.kubernetes-core-configmap     Define if generator kubernetes-core-configmap is enabled (default true)
      --generator.kubernetes-core-event         Define if generator kubernetes-core-event is enabled (default true)
      --generator.kubernetes-core-pod           Define if generator kubernetes-core-pod is enabled (default true)
      --generator.kubernetes-core-secret        Define if generator kubernetes-core-secret is enabled (default true)
      --generator.kubernetes-core-service       Define if generator kubernetes-core-service is enabled (default true)
      --generator.prometheus-monitoring         Define if generator prometheus-monitoring is enabled (default true)
  -h, --help                                    help for debug-package
      --hide-sensitive-data                     Hide sensitive data (default true)
  -n, --namespace string                        Kubernetes namespace (default "default")
  -o, --output -                                Output of the result gz file. If set to - then stdout is used (default "out.tar.gz")
      --pod-logs                                Collect pod logs (default true)
```
[END_INJECT]: # (arangodb_operator_ops_cmd_debug_package)
