# Upgrading ArangoDB version

The ArangoDB Kubernetes Operator supports upgrading an ArangoDB from
one version to the next.

**Warning!**
It is highly recommended to take a backup of your data before upgrading ArangoDB
using [arangodump](https://docs.arangodb.com/stable/components/tools/arangodump/) or [ArangoBackup CR](backup-resource.md).

## Upgrade an ArangoDB deployment version

To upgrade a cluster, change the version by changing
the `spec.image` setting and then apply the updated
custom resource using:

```bash
kubectl apply -f yourCustomResourceFile.yaml
```

The ArangoDB operator will perform an sequential upgrade
of all servers in your deployment. Only one server is upgraded
at a time.

For patch level upgrades (e.g. 3.9.2 to 3.9.3) each server
is stopped and restarted with the new version.

For minor level upgrades (e.g. 3.9.2 to 3.10.0) each server
is stopped, then the new version is started with `--database.auto-upgrade`
and once that is finish the new version is started with the normal arguments.

The process for major level upgrades depends on the specific version.

## Upgrade an ArangoDB deployment to Enterprise edition

In order to upgrade a cluster from community to enterprise, we have to
go through to the following adjustements to an existing deployment:

* [Add a license key](https://arangodb.github.io/kube-arangodb/docs/how-to/set_license.html)
to the cluster
* Adjust `spec.image` to a valid enterprise image string
* Add `spec.license.secretName` to the introduced license key

```bash
kubectl apply -f yourCustomResourceFile.yaml
```

The actual upgrade procedure follows exactly the same steps as
described above for a simple version upgrade in a sequential manner.

Regardless of if you are not only changing the images of community and
enterprise of the same major, minor and patch levels, or upgrade both
to enterprise and a different version, the procedure is only performed
once in a combined step of upgrading version and edition.
