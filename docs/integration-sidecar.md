# Integration 

## Profile

## Sidecar

### Resource Types

Integration Sidecar is supported in a few resources managed by Operator:

- ArangoSchedulerDeployment (scheduler.arangodb.com/v1beta1)
- ArangoSchedulerBatchJob (scheduler.arangodb.com/v1beta1)
- ArangoSchedulerCronJob (scheduler.arangodb.com/v1beta1)
- ArangoSchedulerPod (scheduler.arangodb.com/v1beta1)

To enable integration sidecar for specific deployment label needs to be defined:

```yaml
metadata:
  labels:
    profiles.arangodb.com/deployment: << deployment name >>
```

### Integrations

To enable integration in specific version, labels needs to be added:

```yaml
metadata:
  labels:
    integration.profiles.arangodb.com/<< integration name >>: << integration version >>
```

#### [Authentication V1](/integrations/authentication/v1/definition/README.md)

Authentication Integration Sidecar

To enable:

```yaml
metadata:
  labels:
    integration.profiles.arangodb.com/authn: v1
```

#### [Authorization V0](/integrations/authorization/v0/definition/README.md)

Authorization Integration Sidecar

To enable:

```yaml
metadata:
  labels:
    integration.profiles.arangodb.com/authz: v0
```

#### [Scheduler V2](/integrations/scheduler/v2/definition/README.md)

Scheduler Integration Sidecar

To enable:

```yaml
metadata:
  labels:
    integration.profiles.arangodb.com/sched: v2
```

#### [Storage V2](/integrations/storage/v2/definition/README.md)

Storage Integration Sidecar

To enable:

```yaml
metadata:
  labels:
    integration.profiles.arangodb.com/storage: v2
```

### Envs

#### INTEGRATION_API_ADDRESS

Integration Service API Address

Example: `localhost:1234`

#### INTEGRATION_SERVICE_ADDRESS

Integration Service API Address

Example: `localhost:1234`

#### ARANGO_DEPLOYMENT_NAME

ArangoDeployment name.

Example: `deployment`

#### ARANGO_DEPLOYMENT_ENDPOINT

HTTP/S Endpoint of the ArangoDeployment Internal Service.

Example: `https://deployment.default.svc:8529`