# API Reference for ArangoBackup V1

## Spec

### .spec.backoff.iterations

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec_backoff.go#L31)</sup>

Iterations defines number of iterations before reaching MaxDelay. Default to 5

***

### .spec.backoff.max_delay

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec_backoff.go#L29)</sup>

MaxDelay defines maximum delay in seconds. Default to 600

***

### .spec.backoff.max_iterations

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec_backoff.go#L33)</sup>

MaxIterations defines maximum number of iterations after backoff will be disabled. Default to nil (no limit)

***

### .spec.backoff.min_delay

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec_backoff.go#L27)</sup>

MinDelay defines minimum delay in seconds. Default to 30

***

### .spec.deployment.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec.go#L54)</sup>

Name of the ArangoDeployment Custom Resource within same namespace as ArangoBackup Custom Resource.

This field is **immutable**: can't be changed after backup creation

***

### .spec.download.credentialsSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec.go#L81)</sup>

CredentialsSecretName is the name of the secret used while accessing repository

Links:
* [Defining a secret for backup upload or download](/docs/backup-resource.md#defining-a-secret-for-backup-upload-or-download)

This field is **immutable**: can't be changed after backup creation

***

### .spec.download.id

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec.go#L89)</sup>

ID of the ArangoBackup to be downloaded

This field is **immutable**: can't be changed after backup creation

***

### .spec.download.repositoryURL

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec.go#L77)</sup>

RepositoryURL is the URL path for file storage
Same repositoryURL needs to be defined in `credentialsSecretName` if protocol is other than local.
Format: `<protocol>:/<path>`

Links:
* [rclone.org](https://rclone.org/docs/#syntax-of-remote-paths)

Example:
```yaml
s3://my-bucket/test
azure://test
```

This field is **immutable**: can't be changed after backup creation

***

### .spec.lifetime

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec.go#L47)</sup>

Lifetime is the time after which the backup will be deleted. Format: "1.5h" or "2h45m".

***

### .spec.options.allowInconsistent

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec.go#L66)</sup>

AllowInconsistent flag for Backup creation request.
If this value is set to true, backup is taken even if we are not able to acquire lock.

Default Value: `false`

This field is **immutable**: can't be changed after backup creation

***

### .spec.options.timeout

Type: `number` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec.go#L61)</sup>

Timeout for Backup creation request in seconds. Works only when AsyncBackupCreation feature is set to false.

Default Value: `30`

This field is **immutable**: can't be changed after backup creation

***

### .spec.policyName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec.go#L42)</sup>

PolicyName name of the ArangoBackupPolicy which created this Custom Resource

This field is **immutable**: can't be changed after backup creation

***

### .spec.upload.credentialsSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec.go#L81)</sup>

CredentialsSecretName is the name of the secret used while accessing repository

Links:
* [Defining a secret for backup upload or download](/docs/backup-resource.md#defining-a-secret-for-backup-upload-or-download)

This field is **immutable**: can't be changed after backup creation

***

### .spec.upload.repositoryURL

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec.go#L77)</sup>

RepositoryURL is the URL path for file storage
Same repositoryURL needs to be defined in `credentialsSecretName` if protocol is other than local.
Format: `<protocol>:/<path>`

Links:
* [rclone.org](https://rclone.org/docs/#syntax-of-remote-paths)

Example:
```yaml
s3://my-bucket/test
azure://test
```

This field is **immutable**: can't be changed after backup creation

## Status

### .status.available

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_status.go#L36)</sup>

Available Determines if we can restore from ArangoBackup

***

### .status.backoff.iterations

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_status_backoff.go#L30)</sup>

***

### .status.backup.downloaded

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_status.go#L66)</sup>

Downloaded Determines if ArangoBackup has been downloaded.

***

### .status.backup.id

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_status.go#L56)</sup>

***

### .status.backup.imported

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_status.go#L67)</sup>

***

### .status.backup.keys

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_status.go#L70)</sup>

***

### .status.backup.numberOfDBServers

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_status.go#L62)</sup>

NumberOfDBServers Cluster size of the Backup in ArangoDB

***

### .status.backup.potentiallyInconsistent

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_status.go#L58)</sup>

***

### .status.backup.sizeInBytes

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_status.go#L60)</sup>

SizeInBytes Size of the Backup in ArangoDB.

***

### .status.backup.uploaded

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_status.go#L64)</sup>

Uploaded Determines if ArangoBackup has been uploaded

***

### .status.backup.version

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_status.go#L57)</sup>

***

### .status.message

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_state.go#L88)</sup>

Message for the state this object is in.

***

### .status.progress.jobID

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_state.go#L111)</sup>

JobID ArangoDB job ID for uploading or downloading

***

### .status.progress.progress

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_state.go#L114)</sup>

Progress ArangoDB job progress in percents

Example:
```yaml
90%
```

***

### .status.state

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_state.go#L82)</sup>

State holds the current high level state of the backup

Possible Values: 
* Pending (default) - state in which Custom Resource is queued. If Backup is possible changed to "Scheduled"
* Scheduled - state which will start create/download process
* Download - state in which download request will be created on ArangoDB
* DownloadError - state when download failed
* Downloading - state for downloading progress
* Create - state for creation, field available set to true
* Upload - state in which upload request will be created on ArangoDB
* Uploading - state for uploading progress
* UploadError - state when uploading failed
* Ready - state when Backup is finished
* Deleted - state when Backup was once in ready, but has been deleted
* Failed - state for failure
* Unavailable - state when Backup is not available on the ArangoDB. It can happen in case of upgrades, node restarts etc.

