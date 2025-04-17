---
layout: page
parent: Custom resources overview
title: ArangoProfile
---

# ArangoProfile Custom Resource

[Full CustomResourceDefinition reference ->](./api/ArangoProfile.V1Beta1.md)

# Integration

## Enablement

In order to enable Injection one of the two Labels needs to be present on Pod:

- `profiles.arangodb.com/apply` with any value
- `profiles.arangodb.com/deployment` with value set to the existing Deployment name

## Injection

### Selector

Using [Selector](./api/ArangoProfile.V1Beta1.md) `.spec.selectors.label` you can select which profiles are going to be applied on the Pod.

To not match any pod:
```yaml
apiVersion: scheduler.arangodb.com/v1beta1
kind: ArangoProfile
metadata:
  name: example
spec:
  selectors: {}
  template: ...
```

To match all pods:
```yaml
apiVersion: scheduler.arangodb.com/v1beta1
kind: ArangoProfile
metadata:
  name: example
spec:
  selectors:
    label:
      matchLabels: {}
  template: ...
```

To match specific pods (with label key=value):
```yaml
apiVersion: scheduler.arangodb.com/v1beta1
kind: ArangoProfile
metadata:
  name: example
spec:
  selectors:
    label:
      matchLabels: 
        key: value
  template: ...
```

### Selection

Profiles can be injected using name (not only selectors).

In order to inject specific profiles to the pod use label (split by `,`):

```yaml
metadata:
  annotations:
    profiles.arangodb.com/profiles: "gpu"
```

or

```yaml
metadata:
  annotations:
    profiles.arangodb.com/profiles: "gpu,internal"
```
