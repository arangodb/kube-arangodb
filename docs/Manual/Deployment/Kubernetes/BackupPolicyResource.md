# ArangoBackupPolicy Custom Resource

The ArangoBackupPolicy represents schedule definition for creating ArangoBackup objects by operator.
This deployment specification is a `CustomResource` following
a `CustomResourceDefinition` created by the operator.

## Examples:

### Create schedule for all deployments


```yaml
apiVersion: "database.arangodb.com/v1alpha"
kind: "ArangoBackupPolicy"
metadata:
  name: "example-arangodb-backup-policy"
spec:
  schedule: "*/15 * * * * *"
```

Action:

Create an ArangoBackup object for each ArangoBackup every 15 seconds

### Create schedule for selected deployments


```yaml
apiVersion: "database.arangodb.com/v1alpha"
kind: "ArangoBackupPolicy"
metadata:
  name: "example-arangodb-backup-policy"
spec:
  schedule: "*/15 * * * * *"
  selector:
    matchLabels:
      labelName: "labelValue"
```

Action:

Create an ArangoBackup object for selected ArangoBackup every 15 seconds

### Create schedule for all deployments and upload


```yaml
apiVersion: "database.arangodb.com/v1alpha"
kind: "ArangoBackupPolicy"
metadata:
  name: "example-arangodb-backup-policy"
spec:
  schedule: "*/15 * * * * *"
  template:
      upload:
        repositoryURL: "s3:/..."
        credentialsSecretName: "secret-name"
```

Create an ArangoBackup object for each ArangoBackup every 15 seconds and upload to repositoryURL

## ArangoBackup Custom Resource Spec:

```yaml
apiVersion: "database.arangodb.com/v1alpha"
kind: "ArangoBackupPolicy"
metadata:
  name: "example-arangodb-backup-policy"
spec:
  schedule: "*/15 * * * * *"
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

Spec of the ArangoBackupPolicy object

Required: true

Default: {}

### `spec.schedule: String`

Schedule definition. Parser from https://godoc.org/github.com/robfig/cron

Required: true

Default: ""

### `spec.selector: Object`

Selector definition. Parser from https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#LabelSelector

Required: false

Default: {}

### `spec.template: ArangoBackupTemplate`

Template for the ArangoBackup object

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

Status of the ArangoBackupPolicy object managed by operator

Required: true

Default: {}

### `spec.scheduled: String`

Next scheduled time

Required: true

Default: ""

### `spec.message: String`

Message from the operator in case of failure - schedule not valid, ArangoBackupPolicy not valid

Required: false

Default: ""