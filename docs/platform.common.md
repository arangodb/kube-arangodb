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