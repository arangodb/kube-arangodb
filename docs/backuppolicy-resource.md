# ArangoBackupPolicy Custom Resource

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

## ArangoBackup Custom Resource Spec

```yaml
apiVersion: "backup.arangodb.com/v1"
kind: "ArangoBackupPolicy"
metadata:
  name: "example-arangodb-backup-policy"
spec:
  schedule: "*/15 * * * * "
  selector:
    matchLabels:
      labelName: "labelValue"
    matchExpressions: []
  template:
    options:
      timeout: 3
      force: true
    upload:
      repositoryURL: "s3:/..."
      credentialsSecretName: "secret-name"
status:
  scheduled: "time"
  message: "message"
```

## `spec: Object`

Spec of the ArangoBackupPolicy Custom Resource

Required: true

Default: {}

### `spec.schedule: String`

Schedule definition. Parser from https://godoc.org/github.com/robfig/cron

Required: true

Default: ""

### `spec.allowConcurrent: String`

If false, ArangoBackup will not be created when previous backups are not finished.
`ScheduleSkipped` event will be published in that case.

Required: false

Default: True

### `spec.maxBackups: Integer`

If > 0, then old healthy backups of that policy will be removed to ensure that only `maxBackups` are present at same time.
`CleanedUpOldBackups` event will be published on automatic removal of old backups.

Required: false

Default: 0

### `spec.selector: Object`

Selector definition for selecting matching ArangoBackup Custom Resources. Parser from https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#LabelSelector

Required: false

Default: {}

### `spec.template: ArangoBackupTemplate`

Template for the ArangoBackup Custom Resource

Required: false

Default: {}

### `spec.template.options: ArangoBackup - spec.options`

ArangoBackup options

Required: false

Default: {}

### `spec.template.upload: ArangoBackup - spec.upload`

ArangoBackup upload configuration

Required: false

Default: {}

## `status: Object`

Status of the ArangoBackupPolicy Custom Resource managed by operator

Required: true

Default: {}

### `status.scheduled: TimeStamp`

Next scheduled time in UTC

Required: true

Default: ""

### `status.message: String`

Message from the operator in case of failure - schedule not valid, ArangoBackupPolicy not valid

Required: false

Default: ""
