# Introduction

Kubernetes ArangoDB Ingress for custom certificates.

ArangoDB supports more than only HTTP protocol, so simple Ingress is not enough.

## Before

Before Ingress proxy will be installed certificate secret needs to be created:

```
kubectl -n <deployment namespace> create secret tls <secret name> --cert <path to cert> --key <path to key>
```

## Installation

To install Ingress:
```
helm install --name <my ingress name> --namespace <deployment namespace> <path to kube-arangodb repository>/chart/arangodb-ingress-proxy --set replicas=2 --set tls=TLS Secret name> --set deployment=<ArangoDeployment name>
```