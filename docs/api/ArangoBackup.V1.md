# API Reference for ArangoBackup V1

## Spec

### .spec.backoff.iterations: int

Iterations defines number of iterations before reaching MaxDelay. Default to 5

[Code Reference](/pkg/apis/backup/v1/backup_spec_backoff.go#L31)

### .spec.backoff.max_delay: int

MaxDelay defines maximum delay in seconds. Default to 600

[Code Reference](/pkg/apis/backup/v1/backup_spec_backoff.go#L29)

### .spec.backoff.max_iterations: int

MaxIterations defines maximum number of iterations after backoff will be disabled. Default to nil (no limit)

[Code Reference](/pkg/apis/backup/v1/backup_spec_backoff.go#L33)

### .spec.backoff.min_delay: int

MinDelay defines minimum delay in seconds. Default to 30

[Code Reference](/pkg/apis/backup/v1/backup_spec_backoff.go#L27)

### .spec.deployment.name: string

Name of the ArangoDeployment Custom Resource within same namespace as ArangoBackup Custom Resource.

This field is **immutable**: can't be changed after backup creation

[Code Reference](/pkg/apis/backup/v1/backup_spec.go#L54)

### .spec.download.credentialsSecretName: string

CredentialsSecretName is the name of the secret used while accessing repository

Links:
* [Defining a secret for backup upload or download](/docs/backup-resource.md#defining-a-secret-for-backup-upload-or-download)

This field is **immutable**: can't be changed after backup creation

[Code Reference](/pkg/apis/backup/v1/backup_spec.go#L81)

### .spec.download.id: string

ID of the ArangoBackup to be downloaded

This field is **immutable**: can't be changed after backup creation

[Code Reference](/pkg/apis/backup/v1/backup_spec.go#L89)

### .spec.download.repositoryURL: string

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

[Code Reference](/pkg/apis/backup/v1/backup_spec.go#L77)

### .spec.lifetime: int64

Lifetime is the time after which the backup will be deleted. Format: "1.5h" or "2h45m".

[Code Reference](/pkg/apis/backup/v1/backup_spec.go#L47)

### .spec.options.allowInconsistent: bool

AllowInconsistent flag for Backup creation request.
If this value is set to true, backup is taken even if we are not able to acquire lock.

Default Value: false

This field is **immutable**: can't be changed after backup creation

[Code Reference](/pkg/apis/backup/v1/backup_spec.go#L66)

### .spec.options.timeout: float32

Timeout for Backup creation request in seconds.

Default Value: 30

This field is **immutable**: can't be changed after backup creation

[Code Reference](/pkg/apis/backup/v1/backup_spec.go#L61)

### .spec.policyName: string

PolicyName name of the ArangoBackupPolicy which created this Custom Resource

This field is **immutable**: can't be changed after backup creation

[Code Reference](/pkg/apis/backup/v1/backup_spec.go#L42)

### .spec.upload.credentialsSecretName: string

CredentialsSecretName is the name of the secret used while accessing repository

Links:
* [Defining a secret for backup upload or download](/docs/backup-resource.md#defining-a-secret-for-backup-upload-or-download)

This field is **immutable**: can't be changed after backup creation

[Code Reference](/pkg/apis/backup/v1/backup_spec.go#L81)

### .spec.upload.repositoryURL: string

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

[Code Reference](/pkg/apis/backup/v1/backup_spec.go#L77)

## Status

### .status.available: bool

Available Determines if we can restore from ArangoBackup

[Code Reference](/pkg/apis/backup/v1/backup_status.go#L36)

### .status.backoff.iterations: int

[Code Reference](/pkg/apis/backup/v1/backup_status_backoff.go#L30)

### .status.backup.downloaded: bool

Downloaded Determines if ArangoBackup has been downloaded.

[Code Reference](/pkg/apis/backup/v1/backup_status.go#L66)

### .status.backup.id: string

[Code Reference](/pkg/apis/backup/v1/backup_status.go#L56)

### .status.backup.imported: bool

[Code Reference](/pkg/apis/backup/v1/backup_status.go#L67)

### .status.backup.keys: []string

[Code Reference](/pkg/apis/backup/v1/backup_status.go#L70)

### .status.backup.numberOfDBServers: uint

NumberOfDBServers Cluster size of the Backup in ArangoDB

[Code Reference](/pkg/apis/backup/v1/backup_status.go#L62)

### .status.backup.potentiallyInconsistent: bool

[Code Reference](/pkg/apis/backup/v1/backup_status.go#L58)

### .status.backup.sizeInBytes: uint64

SizeInBytes Size of the Backup in ArangoDB.

[Code Reference](/pkg/apis/backup/v1/backup_status.go#L60)

### .status.backup.uploaded: bool

Uploaded Determines if ArangoBackup has been uploaded

[Code Reference](/pkg/apis/backup/v1/backup_status.go#L64)

### .status.backup.version: string

[Code Reference](/pkg/apis/backup/v1/backup_status.go#L57)

### .status.message: string

Message for the state this object is in.

[Code Reference](/pkg/apis/backup/v1/backup_state.go#L86)

### .status.progress.jobID: string

JobID ArangoDB job ID for uploading or downloading

[Code Reference](/pkg/apis/backup/v1/backup_state.go#L109)

### .status.progress.progress: string

Progress ArangoDB job progress in percents

Example:
```yaml
90%
```

[Code Reference](/pkg/apis/backup/v1/backup_state.go#L112)

### .status.state: string

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

[Code Reference](/pkg/apis/backup/v1/backup_state.go#L80)

