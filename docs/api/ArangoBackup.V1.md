---
layout: page
parent: CRD reference
title: ArangoBackup V1
---

# API Reference for ArangoBackup V1

## Spec

### .spec.backoff.iterations

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/backup/v1/backup_spec_backoff.go#L38)</sup>

Iterations defines number of iterations before reaching MaxDelay. Default to 5

Default Value: `5`

***

### .spec.backoff.max_delay

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/backup/v1/backup_spec_backoff.go#L35)</sup>

MaxDelay defines maximum delay in seconds. Default to 600

Default Value: `600`

***

### .spec.backoff.max_iterations

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/backup/v1/backup_spec_backoff.go#L40)</sup>

MaxIterations defines maximum number of iterations after backoff will be disabled. Default to nil (no limit)

***

### .spec.backoff.min_delay

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/backup/v1/backup_spec_backoff.go#L32)</sup>

MinDelay defines minimum delay in seconds. Default to 30

Default Value: `30`

***

### .spec.deployment.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/backup/v1/backup_spec.go#L55)</sup>

Name of the ArangoDeployment Custom Resource within same namespace as ArangoBackup Custom Resource.

This field is **immutable**: can't be changed after backup creation

***

### .spec.download.autoDelete

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/backup/v1/backup_spec.go#L85)</sup>

AutoDelete removes the ArangoBackup resource (which removes the backup from the cluster) after successful upload

Default Value: `false`

***

### .spec.download.credentialsSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/backup/v1/backup_spec.go#L82)</sup>

CredentialsSecretName is the name of the secret used while accessing repository

Links:
* [Defining a secret for backup upload or download](../backup-resource.md#defining-a-secret-for-backup-upload-or-download)

This field is **immutable**: can't be changed after backup creation

***

### .spec.download.id

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/backup/v1/backup_spec.go#L93)</sup>

ID of the ArangoBackup to be downloaded

This field is **immutable**: can't be changed after backup creation

***

### .spec.download.repositoryURL

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/backup/v1/backup_spec.go#L78)</sup>

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

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/backup/v1/backup_spec.go#L48)</sup>

Lifetime is the time after which the backup will be deleted. Format: "1.5h" or "2h45m".

***

### .spec.options.allowInconsistent

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/backup/v1/backup_spec.go#L67)</sup>

AllowInconsistent flag for Backup creation request.
If this value is set to true, backup is taken even if we are not able to acquire lock.

Default Value: `false`

This field is **immutable**: can't be changed after backup creation

***

### .spec.options.timeout

Type: `number` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/backup/v1/backup_spec.go#L62)</sup>

Timeout for Backup creation request in seconds. Works only when AsyncBackupCreation feature is set to false.

Default Value: `30`

This field is **immutable**: can't be changed after backup creation

***

### .spec.policyName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/backup/v1/backup_spec.go#L42)</sup>

PolicyName name of the ArangoBackupPolicy which created this Custom Resource

This field is **immutable**: can't be changed after backup creation

***

### .spec.upload.autoDelete

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/backup/v1/backup_spec.go#L85)</sup>

AutoDelete removes the ArangoBackup resource (which removes the backup from the cluster) after successful upload

Default Value: `false`

***

### .spec.upload.credentialsSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/backup/v1/backup_spec.go#L82)</sup>

CredentialsSecretName is the name of the secret used while accessing repository

Links:
* [Defining a secret for backup upload or download](../backup-resource.md#defining-a-secret-for-backup-upload-or-download)

This field is **immutable**: can't be changed after backup creation

***

### .spec.upload.repositoryURL

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/backup/v1/backup_spec.go#L78)</sup>

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

