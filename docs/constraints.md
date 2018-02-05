# Constraints

The ArangoDB operator tries to honor various constraints to support high availability of
the ArangoDB cluster.

## Run agents on separate machines

It is essential for HA that agents are running on separate nodes.
To ensure this, the agent Pods are configured with pod-anti-affinity.

```yaml
kind: Pod
spec:
  affinity:
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchExpressions:
            - key: app
              operator: In
              values:
              - arangodb
            - key: arangodb_cluster_name
              operator: In
              values:
              - <cluster-name>
            - key: role
              operator: In
              values:
              - agent
```

## Run dbservers on separate machines

It is needed for HA that dbservers are running on separate nodes.
To ensure this, the dbserver Pods are configured with pod-anti-affinity.

```yaml
kind: Pod
spec:
  affinity:
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchExpressions:
            - key: app
              operator: In
              values:
              - arangodb
            - key: arangodb_cluster_name
              operator: In
              values:
              - <cluster-name>
            - key: role
              operator: In
              values:
              - dbserver
```

Q: Do we want to allow multiple dbservers on a single node?
   If so, we should use `preferredDuringSchedulingIgnoredDuringExecution`
   antiy-affinity.
