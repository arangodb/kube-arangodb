# ArangoDB Kubernetes Operator

- [Intro](#intro)
- [Using the ArangoDB Kubernetes Operator](using-the-operator.md)
- [Architecture overview](design/README.md)
- [Features description and usage](features/README.md)
- [Custom Resources API Reference](api/README.md)
- [Operator Metrics & Alerts](generated/metrics/README.md)
- [Operator Actions](generated/actions.md)
- [Authentication](authentication.md)
- [Custom resources overview](crds.md):
  - [ArangoDeployment](deployment-resource-reference.md)
  - [ArangoDeploymentReplication](deployment-replication-resource-reference.md)
  - [ArangoLocalStorage](storage-resource.md)
  - [Backup](backup-resource.md)
  - [BackupPolicy](backuppolicy-resource.md)
- [Configuration and secrets](configuration-and-secrets.md)
- [Configuring your driver for ArangoDB access](driver-configuration.md)
- [Using Helm](helm.md)
- [Collecting metrics](metrics.md)
- [Services & Load balancer](services-and-load-balancer.md)
- [Storage configuration](storage.md)
- [Secure connections (TLS)](tls.md)
- [Upgrading ArangoDB version](upgrading.md)
- [Scaling your ArangoDB deployment](scaling.md)
- [Draining the Kubernetes nodes](draining-nodes.md)
- Known issues (TBD)
- [Troubleshooting](troubleshooting.md)
- [How-to ...](how-to/README.md)

## Intro

The ArangoDB Kubernetes Operator (`kube-arangodb`) is a set of operators
that you deploy in your Kubernetes cluster to:

- Manage deployments of the ArangoDB database
- Manage backups
- Provide `PersistentVolumes` on local storage of your nodes for optimal storage performance.
- Configure ArangoDB Datacenter-to-Datacenter Replication

Each of these uses involves a different custom resource.

- Use an [ArangoDeployment resource](deployment-resource-reference.md) to create an ArangoDB database deployment.
- Use an [ArangoMember resource](api/ArangoMember.V1.md) to observe and adjust individual deployment members.
- Use an [ArangoBackup](backup-resource.md) and [ArangoBackupPolicy](backuppolicy-resource.md) resources to create ArangoDB backups.
- Use an [ArangoLocalStorage resource](storage-resource.md) to provide local `PersistentVolumes` for optimal I/O performance.
- Use an [ArangoDeploymentReplication resource](deployment-replication-resource-reference.md) to configure ArangoDB Datacenter-to-Datacenter Replication.

Continue with [Using the ArangoDB Kubernetes Operator](using-the-operator.md)
to learn how to install the ArangoDB Kubernetes operator and create
your first deployment.

For more information about the production readiness state, please refer to the
[main README file](https://github.com/arangodb/kube-arangodb#production-readiness-state).
