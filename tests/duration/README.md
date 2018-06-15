# Kube-ArangoDB duration test

This test is a simple application that keeps accessing the database with various requests.

## Building

In root of kube-arangodb repository, run:

```bash
make docker-duration-test
```

## Running

Start an ArangoDB `Cluster` deployment.

Run:

```bash
kubectl run \
    --image=${DOCKERNAMESPACE}/kube-arangodb-durationtest:dev \
    --image-pull-policy=Always duration-test \
    -- \
    --cluster=https://<deployment-name>.<namespace>.svc:8529 \
    --username=root
```

To remove the test, run:

```bash
kubectl delete -n <namespace> deployment/duration-test
```
