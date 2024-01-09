# ArangoMLExtension Custom Resource


#### Enterprise Edition only

[Full CustomResourceDefinition reference ->](./api/ArangoMLExtension.V1Alpha1.md)


You can spin up the [ArangoML](https://github.com/arangoml) engine on existing ArangoDeployment.
That will allow you to train ML models and use them for predictions based on data in your database.

This instruction covers only the steps to run ArangoML in Kubernetes cluster with already running ArangoDeployment.
If you don't have one yet, consider checking [kube-arangodb installation guide](./using-the-operator.md) and [ArangoDeployment CR description](./deployment-resource-reference.md).

### To start ArangoML in your cluster, follow next steps:

1) Enable ML operator. e.g. if you are using Helm package, add `--set "operator.features.ml=true"` option to the Helm command.

2) Create `ArangoMLStorage` CR. This resource provides access for ArangoML to object storage. Currently only S3 API-compatible storages are supported.
  In this example we will use [Minio](https://min.io/) object storage. Please install Minio and make sure the endpoint is available from inside the cluster running ArangoML.

  - Create Kubernetes Secret containing Minio credentials to access S3 API. The secret data should contain two fields: `accessKey` and `secretKey`.
  - Create Kubernetes Secret containing CA certificates to validate connection to endpoint if your Minio installation uses encrypted connection. The secret data should contain two fields: `ca.crt` and `ca.key` (both PEM-encoded).
  - Create ArangoMLStorage resource. Example:
  ```yaml
    apiVersion: ml.arangodb.com/v1alpha1
    kind: ArangoMLStorage
    metadata:
      name: myarangoml-storage
    spec:
      backend:
        s3: # defines access to S3 API
          caSecret: # skip this field if you are not using HTTPS connection to minio
            name: ml-storage-s3-ca
          credentialsSecret:
            name: ml-storage-s3-creds
          allowInsecure: false # set to true if you want to skip certificate check 
          endpoint: https://minio.my-minio-tenant.svc.cluster.local
      bucketName: my-arangoml-bucket # bucket will be created if it does not exist
      mode: # defines how storage proxy is deployed to cluster. Currently only 'sidecar' mode is supported. 
        sidecar: {} # you can configure various parameters for sidecar container here. See full CRD reference for details.
  ```

3) Create `ArangoMLExtension` CR. The name of extension **must** be the same as the name of `ArangoDeployment` and it should be created in the same namespace. 
  Assuming you have ArangoDeployment with name `myarangodb`, create CR:
  ```yaml
    apiVersion: ml.arangodb.com/v1alpha1
    kind: ArangoMLExtension
    metadata:
      name: myarangodb
    spec:
      storage:
        name: myarangoml-storage # name of the ArangoMLStorage created on the previous step
      deployment:
        # you can add here: tolerations, nodeSelector, nodeAffinity, scheduler and many other parameters. See full CRD reference for details.
        replicas: 1 # by default only one pod is running which contains containers for each component (prediction, training, project). You can scale it up or down.
        prediction: 
          image: <prediction-image>
          # you can configure various parameters for container running this component here. See full CRD reference for details.
        project:
          image: <projects-image>
          # you can configure various parameters for container running this component here. See full CRD reference for details.
        training:
          image: <training-image>
          # you can configure various parameters for container running this component here. See full CRD reference for details.
      init: # configuration for Kubernetes Job running initial bootstrap of ArangoML for your cluster.
        image: <init-image>
        # you can add here: tolerations, nodeSelector, nodeAffinity, scheduler and many other parameters. See full CRD reference for details.
      jobsTemplates:
        prediction:
          cpu:
            image: <prediction-job-cpu image>
            # you can configure various parameters for pod and container running this component here. See full CRD reference for details.
          gpu:
            image: <prediction-job-gpu image>
            # you can configure various parameters for pod and container running this component here. See full CRD reference for details.
            resources: # this ensures that pod will be scheduled on GPU-enabled node. Adjust for your environment if neccessary.
              limits:
                nvidia.com/gpu: "1"
              requests:
                nvidia.com/gpu: "1"
        training:
          cpu:
            image: <training-cpu-image>
            # you can configure various parameters for pod and container running this component here. See full CRD reference for details.
          gpu:
            image: <training-gpu-image>
            # you can configure various parameters for pod and container running this component here. See full CRD reference for details.
            resources: # this ensures that pod will be scheduled on GPU-enabled node. Adjust for your environment if neccessary.
              limits:
                nvidia.com/gpu: "1"
              requests:
                nvidia.com/gpu: "1"
  ```

4) After creation of CR, please wait a few minutes for ArangoML initialization to complete. You can check the status for ArangoMLExtension to see current state. Wait for condition `Ready` to be `True`:
```shell
kubectl describe arangomlextension myarangodb
```
```
# ...
status:
  conditions:
    name: Ready
    value: True
```

5) ArangoML now is ready to use! Head to [ArangoML documentation](https://github.com/arangoml) for more details on usage.

**Please note** the ArangoML creates a new database in your ArangoDB cluster for storing meta-information about model training and predictions. Editing or removing this database can cause ArangoML to fail or operate in an unpredictable manner.
