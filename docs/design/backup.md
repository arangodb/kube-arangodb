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
        "access_key_id": "ACCESS_KEY_ID",
        "secret_access_key": "SECRECT_ACCESS_KEY",
        "acl": "private",
        "no_check_bucket": "true"
      }
    }
```
