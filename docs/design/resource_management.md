# Resource Management

## overrideDetectedTotalMemory

The `spec.<group>.overrideDetectedTotalMemory` flag is an option that allows users to override the total memory available to the ArangoDB member 
by automatically injecting `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` ENV variable into the container with the value of `spec.<group>.resources.limits.memory`.

Sample:

```yaml
apiVersion: database.arangodb.com/v1
kind: ArangoDeployment
metadata:
  name: cluster
spec:
  mode: Cluster
  dbservers:
    overrideDetectedTotalMemory: true
    resources:
      limits:
        memory: 1Gi
```
