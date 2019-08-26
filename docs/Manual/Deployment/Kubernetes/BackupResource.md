# ArangoBackup Custom Resource

The ArangoDB Backup Operator creates and maintains ArangoDB Backups
in a Kubernetes cluster, given a Backup specification.
This deployment specification is a `CustomResource` following
a `CustomResourceDefinition` created by the operator.

## Examples:

### Create simple Backup

```yaml
apiVersion: "backup.arangodb.com/v1alpha"
kind: "ArangoBackup"
metadata:
  name: "example-arangodb-backup"
  namespace: "arangodb"
spec:
  deployment:
    name: "my-deployment"
```

Action:

Create Backup on ArangoDeployment named `my-deployment`

### Create and upload Backup


```yaml
apiVersion: "backup.arangodb.com/v1alpha"
kind: "ArangoBackup"
metadata:
  name: "example-arangodb-backup"
  namespace: "arangodb"
spec:
  deployment:
    name: "my-deployment"
  upload:
    repositoryURL: "S3://test/kube-test"
```

Action:

Create Backup on ArangoDeployment named `my-deployment` and upload it to `S3://test/kube-test`

### Download Backup


```yaml
apiVersion: "backup.arangodb.com/v1alpha"
kind: "ArangoBackup"
metadata:
  name: "example-arangodb-backup"
  namespace: "arangodb"
spec:
  deployment:
    name: "my-deployment"
  download:
    repositoryURL: "S3://test/kube-test"
    id: "backup-id"
```

Download Backup with id `backup-id` from `S3://test/kube-test`  on ArangoDeployment named `my-deployment`

## Advertised fields

List of custom columns in CRD specification for Kubectl:
- `.spec.policyName` - optional name of the policy
- `.spec.deployment.name` - name of the deployment
- `.status.state` - current object state
- `.status.message` - additional message for current state

## ArangoBackup Custom Resource Spec:

```yaml
apiVersion: "backup.arangodb.com/v1alpha"
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
    repositoryURL: "s3:/..."
    credentialsSecretName: "secret-name"
    id: "backup-id"
  upload:
    repositoryURL: "s3:/..."
    credentialsSecretName: "secret-name"
status:
  state: "Ready"
  message: "Message details" -
  progress:
    jobID: "id"
    progress: "10%"
  Backup:
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

ArangoDeployment specification.

Field is immutable.

Required: true

Default: {}

#### `spec.deployment.name: String`

Name of the ArangoDeployment Custom Resource within same namespace as ArangoBackup object.

Field is immutable.

Required: true

Default: ""

#### `spec.policyName: String`

Name of the ArangoBackupPolicy which created this object

Field is immutable.

Required: false

Default: ""

### `spec.options: Object`

Backup options.

Field is immutable.

Required: false

Default: {}

#### `spec.options.timeout: float`

Timeout for Backup creation request in seconds.

Field is immutable.

Required: false

Default: 30

#### `spec.options.force: bool`

Force flag for Backup creation request.

Field is immutable.

TODO: Point to ArangoDB documentation

Required: false

Default: false

### `spec.download: Object`

Backup download settings.

Field is immutable.

Required: false

Default: {}

#### `spec.download.repositoryURL: string`

Field is immutable.

TODO: Point to the Backup API definition

Required: true

Default: ""

#### `spec.download.credentialsSecretName: string`

Name of the secrets used while accessing repository

Field is immutable.

TODO: Point to the credential structure

Required: false

Default: ""

#### `spec.download.id: string`

ID of the ArangoDB Backup to be downloaded.

Field is immutable.

Required: true

Default: ""

### `spec.upload: Object`

Backup upload settings.

This field can be removed and created again with different values. This operation will trigger upload again.
Fields in object are immutable.

Required: false

Default: {}

#### `spec.upload.repositoryURL: string`

Field is immutable.

TODO: Point to the Backup API definition

Required: true

Default: ""

#### `spec.upload.credentialsSecretName: string`

Name of the secrets used while accessing repository

Field is immutable.

TODO: Point to the credential structure

Required: false

Default: ""

## `status: Object`

Status of the arangoBackup object. This field is managed by subresource and only by operator

Required: true

Default: {}

### `status.state: enum`

State of the ArangoBackup object.

Required: true

Default: ""

Possible states:
- "" - default state, changed to "Pending"
- "Pending" - state in which object is queued. If Backup is possible changed to "Scheduled"
- "Scheduled" - state which will start create/download process
- "Download" - state in which download request will be created on ArangoDB
- "DownloadError" - state when download failed
- "Downloading" - state for downloading progress
- "Create" - state for creation, field available set to true
- "Upload" - state in which upload request will be created on ArangoDB
- "Uploading" - state for uploading progress
- "UploadError" - state when uploading failed
- "Ready" - state when Backup is finished
- "Deleted" - state when Backup was once in ready, but has been deleted
- "Failed" - state for failure

### `status.message: string`

State message of the ArangoBackup object.

Required: false

Default: ""

### `status.progress: object`

Progress info of the uploading and downloading process.

Required: false

Default: {}

#### `status.progress.jobID: string`

ArangoDB job ID for uploading or downloading.

Required: true

Default: ""

#### `status.progress.progress: string`

ArangoDeployment job progress.

Required: true

Default: "0%"


### `status.backup: object`

ArangoDB Backup details.

Required: true

Default: {}

#### `status.backup.id: string`

ArangoDB Backup ID.

Required: true

Default: ""

#### `status.backup.version: string`

ArangoDB Backup version.

Required: true

Default: ""

#### `status.backup.forced: bool`

ArangoDB Backup forced flag.

Required: false

Default: false

#### `status.backup.uploaded: bool`

Determines if ArangoDB Backup has been uploaded.

Required: false

Default: false

#### `status.backup.downloaded: bool`

Determines if ArangoDB Backup has been downloaded.

Required: false

Default: false

#### `status.backup.createdAt: TimeStamp`

ArangoDB Backup object creation time.

Required: true

Default: now()

### `status.available: bool`

Determines if we can restore from ArangoDB Backup.

Required: true

Default: false