# Lifecycle hooks & Finalizers

The ArangoDB operator expects full control of the `Pods` and `PersistentVolumeClaims` it creates.
Therefore, it takes measures to prevent the removal of those resources
until it is safe to do so.

To achieve this, the server containers in the `Pods` have
a `preStop` hook configured and finalizers are added to the `Pods`
and `PersistentVolumeClaims`.

The `preStop` hook executes a binary that waits until all finalizers of
the current pod have been removed.
Until this `preStop` hook terminates, Kubernetes will not send a `TERM` signal
to the processes inside the container, which ensures that the server remains running
until it is safe to stop them.

The operator performs all actions needed when a delete of a `Pod` or
`PersistentVolumeClaims` has been triggered.
E.g. for a dbserver it cleans out the server if the `Pod` and `PersistentVolumeClaim` are being deleted.

## Lifecycle init-container

Because the binary that is called in the `preStop` hook is not part of a standard
ArangoDB docker image, it has to be brought into the filesystem of a `Pod`.
This is done by an initial container that copies the binary to an `emptyDir` volume that
is shared between the init-container and the server container.

## Finalizers

The ArangoDB operators adds the following finalizers to `Pods`:
- `dbserver.database.arangodb.com/drain`: Added to DBServers, removed only when the dbserver can be restarted or is completely drained
- `agent.database.arangodb.com/agency-serving`: Added to Agents, removed only when enough agents are left to keep the agency serving
- `pod.database.arangodb.com/delay`: Delays pod termination
- `database.arangodb.com/graceful-shutdown`: Added to All members, indicating the need for graceful shutdown

The ArangoDB operators adds the following finalizers to `PersistentVolumeClaims`:
- `pvc.database.arangodb.com/member-exists`: Removed only when its member no longer exists or can be safely rebuild

The ArangoDB operators adds the following finalizers to `ArangoDeployment`:
- `database.arangodb.com/remove-child-finalizers`: Clean-ups finalizers from all children resources

The ArangoDB operators adds the following finalizers to `ArangoDeploymentReplication`:
- `replication.database.arangodb.com/stop-sync`: Stops deployment-to-deployment replication
