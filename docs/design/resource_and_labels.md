# Resources and labels

The ArangoDB operator will create the following Kubernetes resources for specified
cluster deployment models.

## Single server

For a single server deployment, the following k8s resources are created:

- `Pod` running ArangoDB single server named `<cluster-name>`.
  - Labels:
    - `app=arangodb`
    - `arangodb_cluster_name: <cluster-name>`
    - `role: single`
- `PersistentVolumeClaim` for, data stored in the single server, named `<cluster-name>_pvc`.
  - Labels:
    - `app=arangodb`
    - `arangodb_cluster_name: <cluster-name>`
    - `role: single`
- `Service` for accessing the single server, named `<cluster-name>`.
  The service will provide access to the single server from within the Kubernetes cluster.
  - Labels:
    - `app=arangodb`
    - `arangodb_cluster_name: <cluster-name>`
    - `role: single`

## Full cluster

For a full cluster deployment, the following Kubernetes resources are created:

- `Pods` running ArangoDB agent named `<cluster-name>_agent_<x>`.
  - Labels:
    - `app=arangodb`
    - `arangodb_cluster_name: <cluster-name>`
    - `role: agent`

- `PersistentVolumeClaims` for, data stored in the agents, named `<cluster-name>_agent_pvc_<x>`.
  - Labels:
    - `app=arangodb`
    - `arangodb_cluster_name: <cluster-name>`
    - `role: agent`

- `Pods` running ArangoDB coordinators named `<cluster-name>_coordinator_<x>`.
  - Labels:
    - `app=arangodb`
    - `arangodb_cluster_name: <cluster-name>`
    - `role: coordinator`
  - Note: Coordinators are configured to use an `emptyDir` volume since
     they do not need persistent storage.

- `Pods` running ArangoDB dbservers named `<cluster-name>_dbserver_<x>`.
  - Labels:
    - `app=arangodb`
    - `arangodb_cluster_name: <cluster-name>`
    - `role: dbserver`

- `PersistentVolumeClaims` for, data stored in the dbservers, named `<cluster-name>_dbserver_pvc_<x>`.
  - Labels:
    - `app=arangodb`
    - `arangodb_cluster_name: <cluster-name>`
    - `role: dbserver`

- Headless `Service` for accessing the all server, named `<cluster-name>_servers`.
  The service will provide access all server server from within the k8s cluster.
  - Labels:
    - `app=arangodb`
    - `arangodb_cluster_name: <cluster-name>`
  - Selector:
    - `app=arangodb`
    - `arangodb_cluster_name: <cluster-name>`

- `Service` for accessing the all coordinators, named `<cluster-name>`.
  The service will provide access all coordinators from within the k8s cluster.
  - Labels:
    - `app=arangodb`
    - `arangodb_cluster_name: <cluster-name>`
    - `role: coordinator`
  - Selector:
    - `app=arangodb`
    - `arangodb_cluster_name: <cluster-name>`
    - `role: coordinator`

## Full cluster with DC2DC

For a full cluster with datacenter replication deployment,
the same resources are created as for a Full cluster, with the following
additions:

- `Pods` running ArangoSync workers named `<cluster-name>_syncworker_<x>`.
  - Labels:
    - `app=arangodb`
    - `arangodb_cluster_name: <cluster-name>`
    - `role: syncworker`

- `Pods` running ArangoSync master named `<cluster-name>_syncmaster_<x>`.
  - Labels:
    - `app=arangodb`
    - `arangodb_cluster_name: <cluster-name>`
    - `role: syncmaster`

- `Service` for accessing the sync masters, named `<cluster-name>_sync`.
  The service will provide access to all syncmaster from within the Kubernetes cluster.
