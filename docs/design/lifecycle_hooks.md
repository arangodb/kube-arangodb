# Lifecycle hooks

The ArangoDB operator expects full control of the `Pods` is creates.
Therefore it takes measures to prevent the removal of those `Pods`
until it is say to do so.

To achieve this, the server containers in the `Pods` have
a `preStop` hook configured and finalizers are added to the `Pods`.

The `preStop` hook executes a binary that waits until all finalizers of
the current pod have been removed.
Until this `preStop` hook terminates, Kubernetes will not send a `TERM` signal
to the processes inside the container, which ensures that the server remains running
until it is safe to stop them.

The operator performs all actions needed when a delete of a `Pod` have been
triggered. E.g. for a dbserver it cleans out the server before it removes
the finalizers.

## Lifecycle init-container

Because the binary that is called in the `preStop` hook is not part of a standard
ArangoDB docker image, it has to be brought into the filesystem of a `Pod`.
This is done by an initial container that copies the binary to an `emptyDir` volume that
is shared between the init-container and the server container.
