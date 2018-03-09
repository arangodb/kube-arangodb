# Test clusters

The ArangoDB operator is tested on various types of kubernetes clusters.

To prepare a cluster for running the ArangoDB operator tests,
do the following:

- Create a `kubectl` config file for accessing the cluster.
- Use that config file.
- Run `./scripts/kube_configure_test_cluster.sh`. This creates a `ConfigMap`
  named `arango-operator-test` in the `kube-system` namespace containing the
  following environment variables.

```bash
REQUIRE_LOCAL_STORAGE=1
```
