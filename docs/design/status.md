# Status

The status field of the `CustomResource` must contain all persistent state needed to
create & maintain the cluster.

## `status.state: string`

This field contains the current status of the cluster.
Possible values are:

- `Creating` when the cluster is first be created.
- `Ready` when all pods if the cluster are in running state.
- `Scaling` when pods are being added to an existing cluster or removed from an existing cluster.
- `Upgrading` when cluster is in the process of being upgraded to another version.

## `status.members.<group>.[x].state: string`

This field contains the pod state of server x of this group.
Possible values are:

- `Creating` when the pod is about to be created.
- `Ready` when the pod has been created.
- `Draining` when a dbserver pod is being drained.
- `ShuttingDown` when a server is in the process of shutting down.

## `status.members.<group>.[x].podName: string`

This field contains the name of the current pod that runs server x of this group.

## `status.members.<group>.[x].clusterID: string`

This field contains the unique cluster ID of server x of this group.
The field is only valid for groups `single`, `agents`, `dbservers` & `coordinators`.
