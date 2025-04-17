---
layout: page
title: Integration Sidecar Shutdown V1
grand_parent: ArangoDBPlatform
parent: Integration Sidecars
---

# Shutdown V1

Definitions:

- [Service](https://github.com/arangodb/kube-arangodb/blob/1.2.44/integrations/shutdown/v1/definition/shutdown.proto)

Operator will send shutdown request once all containers marked with annotation are stopped.

Example:

```yaml
metadata:
  annotations:
    core.shutdown.arangodb.com/app: "wait"
    core.shutdown.arangodb.com/app2: "wait"
    container.shutdown.arangodb.com/app3: port1
spec:
  containers:
    - name: app
    - name: app2
    - name: app3
      ports:
        name: port1
```

Pod will receive shutdown request on port `port1` if containers `app` and `app2` will be in non running state.
