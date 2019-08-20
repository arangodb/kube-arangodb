# ArangoBackup Custom Resource

The ArangoDB Backup Operator creates and maintains ArangoDB backups
in a Kubernetes cluster, given a backup specification.
This deployment specification is a `CustomResource` following
a `CustomResourceDefinition` created by the operator.

## Examples:

### Create simple backup

```yaml
apiVersion: "database.arangodb.com/v1alpha"
kind: "ArangoBackup"
metadata:
  name: "example-arangodb-backup"
  namespace: "arangodb"
spec:
  deployment:
    name: "my-deployment"
```

Action:

Create backup on ArangoDeployment named `my-deployment`

### Create and upload backup


```yaml
apiVersion: "database.arangodb.com/v1alpha"
kind: "ArangoBackup"
metadata:
  name: "example-arangodb-backup"
  namespace: "arangodb"
spec:
  deployment:
    name: "my-deployment"
  upload:
    repositoryPath: "S3://test/kube-test"
```

Action:

Create backup on ArangoDeployment named `my-deployment` and upload it to `S3://test/kube-test`

### Download backup


```yaml
apiVersion: "database.arangodb.com/v1alpha"
kind: "ArangoBackup"
metadata:
  name: "example-arangodb-backup"
  namespace: "arangodb"
spec:
  deployment:
    name: "my-deployment"
  download:
    repositoryPath: "S3://test/kube-test"
    id: "backup-id"
```

Download backup with id `backup-id` from `S3://test/kube-test`  on ArangoDeployment named `my-deployment`

## ArangoBackup Custom Resource Spec:

```yaml
apiVersion: "database.arangodb.com/v1alpha"
kind: "ArangoBackup"
metadata:
  name: "example-arangodb-backup"
  namespace: "arangodb"
spec:
  policyName: "my-policy"
  deployment:
    name: "my-deployment"
  options:
    timeout: 3
    force: true
  download:
    repositoryPath: "s3:/..."
    credentialsSecretName: "secret-name"
    id: "backup-id"
  upload:
    repositoryPath: "s3:/..."
    credentialsSecretName: "secret-name"
status:
  state: "Ready"
  message: "Message details" -
  progress:
    jobID: "id"
    progress: "10%"
  backup:
    id: "id"
    version: "3.6.0-dev"
    forced: true
    uploaded: true
    downloaded: true
    createdAt: "time"
  available: true
```

## `spec: Object`

Spec of the ArangoBackup object.

Required: true

Default: {}

### `spec.deployment: Object`

ArangoDeployment specification. This field is immutable

Required: true

Default: {}

#### `spec.policyName: String`

Name of the ArangoBackupPolicy which created this object

Required: false

Default: ""

#### `spec.deployment.name: String`

Name of the ArangoDeployment Custom Resource within same namespace as ArangoBackup object

Required: true

Default: ""

### `spec.options: Object`

Backup options. This field is immutable

Required: false

Default: {}

#### `spec.options.timeout: float`

Timeout for backup creation request in seconds

Required: false

Default: 30

#### `spec.options.force: bool`

Force flag for backup creation request

TODO: Point to ArangoDB documentation

Required: false

Default: false

### `spec.download: Object`

Backup download settings

Explicit with: `spec.upload`

Required: false

Default: {}

#### `spec.download.repositoryPath: string`

# TODO: Point to the backup API definition

Required: true

Default: ""

#### `spec.download.credentialsSecretName: string`

Name of the secrets used while accessing repository

# TODO: Point to the credential structure

Required: false

Default: ""

#### `spec.download.id: string`

ID of the ArangoDB backup to be downloaded

Required: true

Default: ""

### `spec.upload: Object`

Backup upload settings

Explicit with: `spec.download`

Required: false

Default: {}

#### `spec.upload.repositoryPath: string`

# TODO: Point to the backup API definition

Required: true

Default: ""

#### `spec.upload.credentialsSecretName: string`

Name of the secrets used while accessing repository

# TODO: Point to the credential structure

Required: false

Default: ""

## `status: Object`

Status of the arangoBackup object. This field is managed by subresource and only by operator

Advertised fields:
- `.status.state` - current object state
- `.status.message` - additional message for current state

Required: true

Default: {}

### `status.state: enum`

State of the ArangoBackup object

Required: true

Default: ""

Possible states:
- "" - default state, changed to "Pending"
- "Pending" - state in which object is queued. If backup is possible changed to "Scheduled"
- "Scheduled" - state which will start create/download process
- "Download" - state in which download request will be created on ArangoDB
- "Downloading" - state for downloading progress
- "Create" - state for creation, field available set to true
- "Upload" - state in which upload request will be created on ArangoDB
- "Uploading" - state for uploading progress
- "Ready" - state when backup object is finished
- "Deleted" - state when backup was once in ready, but was deleted
- "Failed" - state for failure

### `status.message: string`

State message of the ArangoBackup object

Required: false

Default: ""

### `status.progress: string`

Progress info of the uploading and downloading process

Required: false

Default: {}

#### `status.progress.jobID: string`

ArangoDeployment job ID

Required: true

Default: ""

#### `status.progress.progress: string`

ArangoDeployment job progress

Required: true

Default: "0%"


### `status.backup: string`

Progress info of the uploading and downloading process

Required: false

Default: {}

#### `status.backup.id: string`

ArangoDB backup object ID

Required: true

Default: ""

#### `status.backup.version: string`

ArangoDB backup object version

Required: true

Default: ""

#### `status.backup.forced: bool`

ArangoDB backup forced flag

Required: false

Default: false

#### `status.backup.uploaded: bool`

Determines if ArangoDB backup was uploaded

Required: false

Default: false

#### `status.backup.downloaded: bool`

Determines if ArangoDB backup was downloaded

Required: false

Default: false

#### `status.backup.createdAt: bool`

ArangoDB backup object creation time

Required: true

Default: now()

### `status.available: bool`

Determines if we can recover from ArangoDB backup

Required: true

Default: false