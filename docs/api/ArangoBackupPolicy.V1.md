# API Reference for ArangoBackupPolicy V1

## Spec

### .spec.allowConcurrent

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_policy_spec.go#L35)</sup>

AllowConcurrent if false, ArangoBackup will not be created when previous Backups are not finished

Default Value: `true`

***

### .spec.maxBackups

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_policy_spec.go#L43)</sup>

MaxBackups defines how many backups should be kept in history (per deployment). Oldest healthy Backups will be deleted.
If not specified or 0 then no limit is applied

Default Value: `0`

***

### .spec.schedule

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_policy_spec.go#L32)</sup>

Schedule is cron-compatible specification of backup schedule
Parsed by https://godoc.org/github.com/robfig/cron

***

### .spec.selector

Type: `meta.LabelSelector` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_policy_spec.go#L39)</sup>

DeploymentSelector Selector definition for selecting matching ArangoBackup Custom Resources.

Links:
* [Kubernetes Documentation](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#LabelSelector)

***

### .spec.template.backoff.iterations

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec_backoff.go#L31)</sup>

Iterations defines number of iterations before reaching MaxDelay. Default to 5

***

### .spec.template.backoff.max_delay

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec_backoff.go#L29)</sup>

MaxDelay defines maximum delay in seconds. Default to 600

***

### .spec.template.backoff.max_iterations

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec_backoff.go#L33)</sup>

MaxIterations defines maximum number of iterations after backoff will be disabled. Default to nil (no limit)

***

### .spec.template.backoff.min_delay

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec_backoff.go#L27)</sup>

MinDelay defines minimum delay in seconds. Default to 30

***

### .spec.template.lifetime

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_policy_spec.go#L61)</sup>

Lifetime is the time after which the backup will be deleted. Format: "1.5h" or "2h45m".

***

### .spec.template.options.allowInconsistent

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec.go#L66)</sup>

AllowInconsistent flag for Backup creation request.
If this value is set to true, backup is taken even if we are not able to acquire lock.

Default Value: `false`

This field is **immutable**: can't be changed after backup creation

***

### .spec.template.options.timeout

Type: `number` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec.go#L61)</sup>

Timeout for Backup creation request in seconds. Works only when AsyncBackupCreation feature is set to false.

Default Value: `30`

This field is **immutable**: can't be changed after backup creation

***

### .spec.template.upload.credentialsSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_spec.go#L81)</sup>

CredentialsSecretName is the name of the secret used while accessing repository

Links:
* [Defining a secret for backup upload or download](/docs/backup-resource.md#defining-a-secret-for-backup-upload-or-download)

This field is **immutable**: can't be changed after backup creation

***

### .spec.template.upload.repositoryURL

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

### .status.message

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_policy_status.go#L33)</sup>

Message from the operator in case of failures - schedule not valid, ArangoBackupPolicy not valid

***

### .status.scheduled

Type: `meta.Time` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/backup/v1/backup_policy_status.go#L31)</sup>

Scheduled Next scheduled time in UTC

