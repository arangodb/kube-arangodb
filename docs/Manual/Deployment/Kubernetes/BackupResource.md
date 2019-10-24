# ArangoBackup Custom Resource

The ArangoBackup Operator creates and maintains ArangoBackups
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
    credentialsSecretName: "my-s3-rclone-credentials"
```

Action:

Create Backup on ArangoDeployment named `my-deployment` and upload it to `S3://test/kube-test`.


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
    credentialsSecretName: "my-s3-rclone-credentials"
    id: "backup-id"
```

Download Backup with id `backup-id` from `S3://test/kube-test`  on ArangoDeployment named `my-deployment`

### Restore

Information about restoring can be found in [ArangoDeployment](./DeploymentResource.md).

## Advertised fields

List of custom columns in CRD specification for Kubectl:
- `.spec.policyName` - optional name of the policy
- `.spec.deployment.name` - name of the deployment
- `.status.state` - current ArangoBackup Custom Resource state
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
  time: "time"
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
    sizeInBytes: 1
    numberOfDBServers: 3
  available: true
```

## `spec: Object`

Spec of the ArangoBackup Custom Resource.

Required: true

Default: {}

### `spec.deployment: Object`

ArangoDeployment specification.

Field is immutable.

Required: true

Default: {}

#### `spec.deployment.name: String`

Name of the ArangoDeployment Custom Resource within same namespace as ArangoBackup Custom Resource.

Field is immutable.

Required: true

Default: ""

#### `spec.policyName: String`

Name of the ArangoBackupPolicy which created this Custom Resource

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

#### `spec.options.allowInconsistent: bool`

AllowInconsistent flag for Backup creation request.
If this value is set to true, backup is taken even if we are not able to acquire lock.

Field is immutable.

Required: false

Default: false

### `spec.download: Object`

Backup download settings.

Field is immutable.

Required: false

Default: {}

#### `spec.download.repositoryURL: string`

Field is immutable. Protocol needs to be defined in `spec.download.credentialsSecretName` if protocol is other than local.

Mode protocols can be found at [rclone.org](https://rclone.org/).

Format: `<protocol>:/<path>`

Examples:
- `s3://my-bucket/test`
- `azure://test`

Required: true

Default: ""

#### `spec.download.credentialsSecretName: string`

Field is immutable. Name of the secret used while accessing repository

Secret structure:

```yaml
apiVersion: v1
data:
  token: <json token>
kind: Secret
metadata:
  name: <name>
type: Opaque
```

`JSON Token` options are described on the [Rclone](https://rclone.org/) page.
We can define more than one protocols at same time in one secret.

This field is defined in json format:

```json
{
  "<protocol>": {
    "type":"<type>",
    ...parameters
    }
}
```

AWS S3 example - based on [Rclone S3](https://rclone.org/s3/) documentation and interactive process:

```json
{
  "S3": {
    "type": "s3", # Choose s3 type
    "provider": "AWS", # Choose one of the providers
    "env_auth": "false", # Define credentials in next step instead of using ENV
    "access_key_id": "xxx",
    "secret_access_key": "xxx",
    "region": "eu-west-2", # Choose region
    "acl": "private", # Set permissions on newly created remote object
  }
}
```

and you can from now use `S3://bucket/path`.

Required: false

Default: ""

#### `spec.download.id: string`

ID of the ArangoBackup to be downloaded.

Field is immutable.

Required: true

Default: ""

### `spec.upload: Object`

Backup upload settings.

This field can be removed and created again with different values. This operation will trigger upload again.
Fields in Custom Resource Spec Upload are immutable.

Required: false

Default: {}

#### `spec.upload.repositoryURL: string`

Same structure as `spec.download.repositoryURL`.

Required: true

Default: ""

#### `spec.upload.credentialsSecretName: string`

Same structure as `spec.download.credentialsSecretName`.

Required: false

Default: ""

## `status: Object`

Status of the ArangoBackup Custom Resource. This field is managed by subresource and only by operator

Required: true

Default: {}

### `status.state: enum`

State of the ArangoBackup object.

Required: true

Default: ""

Possible states:
- "" - default state, changed to "Pending"
- "Pending" - state in which Custom Resource is queued. If Backup is possible changed to "Scheduled"
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
- "Unavailable" - state when Backup is not available on the ArangoDB. It can happen in case of upgrades, node restarts etc.

### `status.time: timestamp`

Time in UTC when state of the ArangoBackup Custom Resource changed.

Required: true

Default: ""

### `status.message: string`

State message of the ArangoBackup Custom Resource.

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

ArangoBackup details.

Required: true

Default: {}

#### `status.backup.id: string`

ArangoBackup ID.

Required: true

Default: ""

#### `status.backup.version: string`

ArangoBackup version.

Required: true

Default: ""

#### `status.backup.potentiallyInconsistent: bool`

ArangoBackup potentially inconsistent flag.

Required: false

Default: false

#### `status.backup.uploaded: bool`

Determines if ArangoBackup has been uploaded.

Required: false

Default: false

#### `status.backup.downloaded: bool`

Determines if ArangoBackup has been downloaded.

Required: false

Default: false

#### `status.backup.createdAt: TimeStamp`

ArangoBackup Custom Resource creation time in UTC.

Required: true

Default: now()

#### `status.backup.sizeInBytes: uint64`

Size of the Backup in ArangoDB.

Required: true

Default: 0

#### `status.backup.numberOfDBServers: uint`

Cluster size of the Backup in ArangoDB.

Required: true

Default: 0

### `status.available: bool`

Determines if we can restore from ArangoBackup.

Required: true

Default: false