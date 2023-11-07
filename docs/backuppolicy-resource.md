# ArangoBackupPolicy Custom Resource

[Full CustomResourceDefinition reference ->](./api/ArangoBackupPolicy.V1.md)

The ArangoBackupPolicy represents schedule definition for creating ArangoBackup Custom Resources by operator.
This deployment specification is a `CustomResource` following a `CustomResourceDefinition` created by the operator.

## Examples

### Create schedule for all deployments

You can create an ArangoBackup Custom Resource for each ArangoBackup every 15 minutes.

```yaml
apiVersion: "backup.arangodb.com/v1"
kind: "ArangoBackupPolicy"
metadata:
  name: "example-arangodb-backup-policy"
spec:
  schedule: "*/15 * * * *"
```

### Create schedule for selected deployments

You can create an ArangoBackup Custom Resource for selected ArangoBackups every 15 minutes.

```yaml
apiVersion: "backup.arangodb.com/v1"
kind: "ArangoBackupPolicy"
metadata:
  name: "example-arangodb-backup-policy"
spec:
  schedule: "*/15 * * * *"
  selector:
    matchLabels:
      labelName: "labelValue"
```

### Create schedule for all deployments and upload

You can create an ArangoBackup Custom Resource for each ArangoBackup every 15
minutes and upload it to the specified repositoryURL.

```yaml
apiVersion: "backup.arangodb.com/v1"
kind: "ArangoBackupPolicy"
metadata:
  name: "example-arangodb-backup-policy"
spec:
  schedule: "*/15 * * * * "
  template:
    upload:
      repositoryURL: "s3:/..."
      credentialsSecretName: "secret-name"
```

### Create schedule for all deployments, don't allow parallel backup runs, keep limited number of backups

You can create an ArangoBackup Custom Resource for each ArangoBackup every 15
minutes. You can keep 10 backups per deployment at the same time, and delete the
oldest ones. Don't allow to run backup if previous backup is not finished.

```yaml
apiVersion: "backup.arangodb.com/v1"
kind: "ArangoBackupPolicy"
metadata:
  name: "example-arangodb-backup-policy"
spec:
  schedule: "*/15 * * * *"
  maxBackups: 10
  allowConcurrent: False
```
