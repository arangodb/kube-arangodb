# API Reference for ArangoBackupPolicy V1

## Spec

### .spec.allowConcurrent: bool

AllowConcurrent if false, ArangoBackup will not be created when previous Backups are not finished

Default Value: true

[Code Reference](/pkg/apis/backup/v1/backup_policy_spec.go#L35)

### .spec.maxBackups: int

MaxBackups defines how many backups should be kept in history (per deployment). Oldest healthy Backups will be deleted.
If not specified or 0 then no limit is applied

Default Value: 0

[Code Reference](/pkg/apis/backup/v1/backup_policy_spec.go#L43)

### .spec.schedule: string

Schedule is cron-compatible specification of backup schedule
Parsed by https://godoc.org/github.com/robfig/cron

[Code Reference](/pkg/apis/backup/v1/backup_policy_spec.go#L32)

### .spec.selector: meta.LabelSelector

DeploymentSelector Selector definition for selecting matching ArangoBackup Custom Resources.

Links:
* [Kubernetes Documentation](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#LabelSelector)

[Code Reference](/pkg/apis/backup/v1/backup_policy_spec.go#L39)

### .spec.template.backoff.iterations: int

Iterations defines number of iterations before reaching MaxDelay. Default to 5

[Code Reference](/pkg/apis/backup/v1/backup_spec_backoff.go#L31)

### .spec.template.backoff.max_delay: int

MaxDelay defines maximum delay in seconds. Default to 600

[Code Reference](/pkg/apis/backup/v1/backup_spec_backoff.go#L29)

### .spec.template.backoff.max_iterations: int

MaxIterations defines maximum number of iterations after backoff will be disabled. Default to nil (no limit)

[Code Reference](/pkg/apis/backup/v1/backup_spec_backoff.go#L33)

### .spec.template.backoff.min_delay: int

MinDelay defines minimum delay in seconds. Default to 30

[Code Reference](/pkg/apis/backup/v1/backup_spec_backoff.go#L27)

### .spec.template.lifetime: int64

Lifetime is the time after which the backup will be deleted. Format: "1.5h" or "2h45m".

[Code Reference](/pkg/apis/backup/v1/backup_policy_spec.go#L61)

### .spec.template.options.allowInconsistent: bool

AllowInconsistent flag for Backup creation request.
If this value is set to true, backup is taken even if we are not able to acquire lock.

Default Value: false

This field is **immutable**: can't be changed after backup creation

[Code Reference](/pkg/apis/backup/v1/backup_spec.go#L66)

### .spec.template.options.timeout: float32

Timeout for Backup creation request in seconds.

Default Value: 30

This field is **immutable**: can't be changed after backup creation

[Code Reference](/pkg/apis/backup/v1/backup_spec.go#L61)

### .spec.template.upload.credentialsSecretName: string

CredentialsSecretName is the name of the secret used while accessing repository

Links:
* [Defining a secret for backup upload or download](/docs/backup-resource.md#defining-a-secret-for-backup-upload-or-download)

This field is **immutable**: can't be changed after backup creation

[Code Reference](/pkg/apis/backup/v1/backup_spec.go#L81)

### .spec.template.upload.repositoryURL: string

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

### .status.message: string

Message from the operator in case of failures - schedule not valid, ArangoBackupPolicy not valid

[Code Reference](/pkg/apis/backup/v1/backup_policy_status.go#L33)

### .status.scheduled: meta.Time

Scheduled Next scheduled time in UTC

[Code Reference](/pkg/apis/backup/v1/backup_policy_status.go#L31)

