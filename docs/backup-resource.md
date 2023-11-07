# ArangoBackup Custom Resource

[Full CustomResourceDefinition reference ->](./api/ArangoBackup.V1.md)

The ArangoBackup Operator creates and maintains ArangoBackups
in a Kubernetes cluster, given a Backup specification.
This deployment specification is a `CustomResource` following
a `CustomResourceDefinition` created by the operator.


## Defining a secret for backup upload or download

`credentialsSecretName` in `spec.download` and `spec.upload` expects the next structure for secret:

```yaml
apiVersion: v1
data:
  token: <json token>
kind: Secret
metadata:
  name: <name>
type: Opaque
```

`JSON Token` options are described on the [rclone](https://rclone.org/) page.
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

AWS S3 example - based on [rclone S3](https://rclone.org/s3/) documentation and interactive process:

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

##### Use IAM with Amazon EKS

Instead of creating and distributing your AWS credentials to the containers or
using the Amazon EC2 instance's role, you can associate an IAM role with a
Kubernetes service account and configure pods to use the service account.

1. Create a Policy to access the S3 bucket.

   ```bash
   aws iam create-policy \
   --policy-name S3-ACCESS_ROLE \
   --policy-document \
   '{
   "Version": "2012-10-17",
   "Statement": [
       {
           "Effect": "Allow",
           "Action": "s3:ListAllMyBuckets",
           "Resource": "*"
       },
       {
           "Effect": "Allow",
           "Action": "*",
           "Resource": "arn:aws:s3:::MY_BUCKET"
       },
       {
           "Effect": "Allow",
           "Action": "*",
           "Resource": "arn:aws:s3:::MY_BUCKET/*"
       }
   ]
   }'
   ```

2. Create an IAM role for the service account (SA).

   ```bash
   eksctl create iamserviceaccount \
     --name SA_NAME \
     --namespace NAMESPACE \
     --cluster CLUSTER_NAME \
     --attach-policy-arn arn:aws:iam::ACCOUNT_ID:policy/S3-ACCESS_ROLE \
     --approve
   ```

3. Ensure that you use that SA in your ArangoDeployment for `dbservers` and
   `coordinators`.

   ```yaml
   apiVersion: database.arangodb.com/v1
   kind: ArangoDeployment
   metadata:
     name: cluster
   spec:
     image: arangodb/enterprise
     mode: Cluster

     dbservers:
       serviceAccountName: SA_NAME
     coordinators:
       serviceAccountName: SA_NAME
   ```

4. Create a `Secret` Kubernetes object with a configuration for S3.

   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: arangodb-cluster-backup-credentials
   type: Opaque
   stringData:
     token: |
       {
         "s3": {
           "type": "s3",
           "provider": "AWS",
           "env_auth": "true",
           "location_constraint": "eu-central-1",
           "region": "eu-central-1",
           "acl": "private",
           "no_check_bucket": "true"
         }
       }
   ```

5. Create an `ArangoBackup` Kubernetes object with upload to S3.

   ```yaml
   apiVersion: "backup.arangodb.com/v1alpha"
   kind: "ArangoBackup"
   metadata:
     name: backup
   spec:
     deployment:
       name: MY_DEPLOYMENT
     upload:
       repositoryURL: "s3:MY_BUCKET"
       credentialsSecretName: arangodb-cluster-backup-credentials
   ```

## Examples:

### Create simple Backup

```yaml
apiVersion: "backup.arangodb.com/v1"
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
apiVersion: "backup.arangodb.com/v1"
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
apiVersion: "backup.arangodb.com/v1"
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

## Restore

To restore a data for deployment for specific backup, use `spec.restoreFrom` field of [ArangoDeployment](api/ArangoDeployment.V1.md#specrestorefrom-string).

