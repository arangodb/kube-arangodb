---
layout: page
title: Backups
parent: Design overview
---

# ArangoBackup

## Lifetime

The Lifetime of an ArangoBackup let us define the time an ArangoBackup is available in the system. 
E.g.: if we want to keep the ArangoBackup for 1 day, we can set the Lifetime to 1 day. After 1 day the ArangoBackup will be deleted automatically.

```yaml
apiVersion: "backup.arangodb.com/v1alpha"
kind: "ArangoBackup"
metadata:
  name: backup-with-one-day-lifetime
spec:
  deployment:
    name: deployment
  lifetime: 1d
```

## Upload

You can upload the backup to a remote storage.
Here is an example for uploading the backup to AWS S3.

```yaml
apiVersion: "backup.arangodb.com/v1alpha"
kind: "ArangoBackup"
metadata:
  name: backup-and-upload
spec:
  deployment:
    name: deployment
  upload:
    repositoryURL: "s3:BUCKET_NAME"
    credentialsSecretName: upload-credentials
```

To make this work, you need to create a `upload-credentials` Secret with the credentials for the remote storage:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: upload-credentials
type: Opaque
stringData:
  token: |
    {
      "s3": {
        "type": "s3",
        "provider": "AWS",
        "env_auth": "false",
        "region": "eu-central-1",
        "location_constraint": "eu-central-1",
        "access_key_id": "ACCESS_KEY_ID",
        "secret_access_key": "SECRECT_ACCESS_KEY",
        "no_check_bucket": "true"
      }
    }
```

The configuration options are passed to [rclone](https://rclone.org/s3/), which
is used to transfer Hot Backups to and from object storage. Note the following:

- **`acl`**: AWS buckets created since April 2023 default to _Bucket owner
  enforced_ Object Ownership, which rejects requests with an ACL header. Omit
  the `acl` key (or set it to `""`) for such buckets. It may still be required
  for some S3-compatible providers and for older AWS buckets with ACLs
  re-enabled.
- **Region**: For AWS S3 with a region other than `us-east-1`, set the
  `location_constraint` to the region, `"no_check_bucket": "true"`, or both.
  Otherwise rclone (v1.68.0 and later) sends an unspecified location constraint
  that AWS rejects with an `IllegalLocationConstraintException`.
- **Checksums**: For S3-compatible providers (e.g. GCS, Ceph, MinIO, Wasabi),
  uploads may fail unless you set `"use_data_integrity_protections": "false"`,
  because rclone (v1.68.0 and later) defaults to CRC32/CRC64 checksums while
  these providers may expect MD5.
- **Provider quirks**: rclone auto-handles quirks for known providers (e.g.
  `use_x_id`, `sign_accept_encoding`, `use_multipart_uploads`). You may need to
  set these manually if your provider is not recognized.
