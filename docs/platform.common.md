---
layout: page
title: Common Use Cases
parent: ArangoDBPlatform
nav_order: 1
---

# Custom ImagePullSecrets

To inject ImagePullSecrets to all pods across the Platform [ArangoProfile](./arango-profile-resource.md) can be used.

Example:
```yaml
apiVersion: scheduler.arangodb.com/v1beta1
kind: ArangoProfile
metadata:
  name: image-secrets
spec:
  selectors:
    label:
      matchLabels: {}
  template:
    pod:
      imagePullSecrets:
        - <Secret Name>
    priority: 129
```

# Scheduling Platform Pods on Specific Nodes

To pin all Platform pods to a dedicated node pool (for example accredited or isolated nodes), an
[ArangoProfile](./arango-profile-resource.md) can inject a `nodeSelector`, `tolerations` and
`affinity` into every matching pod.

Example:
```yaml
apiVersion: scheduler.arangodb.com/v1beta1
kind: ArangoProfile
metadata:
  name: node-placement
spec:
  selectors:
    label:
      matchLabels: {}
  template:
    priority: 200
    pod:
      nodeSelector:
        node-pool: arango-platform
      tolerations:
        - key: dedicated
          operator: Equal
          value: arango-platform
          effect: NoSchedule
```

Notes:

* A pod is injected only if it carries the `profiles.arangodb.com/apply` or
  `profiles.arangodb.com/deployment` label, and only on CREATE, so existing pods must be restarted to
  pick up a new or changed profile.
* Profiles are namespaced: create one per namespace.
* The `nodeSelector` is **merged into** the pod's existing selector (keys set by the chart or workload
  are kept), rather than replacing it. Across several matching profiles the keys are merged and the
  higher `priority` wins on a conflict. A value of `"-"` removes a key set by a lower-priority profile
  or by the pod. A profile that sets no `nodeSelector` leaves the pod's selector untouched.
* `tolerations` are added only if not already present. `affinity`, when set, replaces the pod's
  affinity.
* Database members (ArangoDeployment) are scheduled through `spec.<group>.nodeSelector`,
  `spec.<group>.tolerations` and `spec.<group>.affinity` on the ArangoDeployment resource itself; set
  those there in addition to the profile above.