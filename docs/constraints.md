# Constraints

The ArangoDB operator tries to honor various constraints to support high availability of
the ArangoDB cluster.

## Run agents and dbservers on separate machines

It is essential for resilience and high availability that no two agents
are running on the same node and no two dbservers are running running
on the same node.

To ensure this, the agent and dbserver Pods are configured with pod-anti-affinity.

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
              - agent (or dbserver)
```

The settings used for pod affinity are based on the `spec.environment` setting.

For a `development` environment we use `preferredDuringSchedulingIgnoredDuringExecution`
so deployments are still possible on tiny clusters like `minikube`.

For `production` environments we enforce (anti-)affinity using
`requiredDuringSchedulingIgnoredDuringExecution`.

## Run coordinators on separate machines

It is preferred to run coordinators of separate machines.

To achieve this, the coordinator Pods are configured with pod-anti-affinity
using the `preferredDuringSchedulingIgnoredDuringExecution` setting.

## Run syncworkers on same machine as dbserver

It is preferred to run syncworkers on the same machine as
dbservers.

To achieve this, the syncworker Pods are configured with pod-affinity
using the `preferredDuringSchedulingIgnoredDuringExecution` setting.
