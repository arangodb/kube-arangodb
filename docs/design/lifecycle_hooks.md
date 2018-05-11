# Lifecycle hooks

The ArangoDB operator expects full control of the `Pods` and `PersistentVolumeClaims` it creates.
Therefore it takes measures to prevent the removal of those resources
until it is safe to do so.

To achieve this, the server containers in the `Pods` have
a `preStop` hook configured and finalizers are added to the `Pods`
ands `PersistentVolumeClaims`.

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
