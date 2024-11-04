# Integration 

## Profile

## Sidecar

### Resource Types

Integration Sidecar is supported in a few resources managed by Operator:

- ArangoSchedulerDeployment (scheduler.arangodb.com/v1beta1)
- ArangoSchedulerBatchJob (scheduler.arangodb.com/v1beta1)
- ArangoSchedulerCronJob (scheduler.arangodb.com/v1beta1)
- ArangoSchedulerPod (scheduler.arangodb.com/v1beta1)

### Envs

#### ARANGO_DEPLOYMENT_NAME

ArangoDeployment name.

Example: `deployment`

#### ARANGO_DEPLOYMENT_ENDPOINT

HTTP/S Endpoint of the ArangoDeployment Internal Service.

Example: `https://deployment.default.svc:8529`
