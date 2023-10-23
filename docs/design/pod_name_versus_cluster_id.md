# Kubernetes Pod name versus cluster ID

All resources being created will get a name that contains
the user provided cluster name and a unique part.

The unique part will be difference for every pod that
is being created.
E.g. when upgrading to a new version, we generate a new
unique pod name.

The servers in the ArangoDB cluster will be assigned
a persistent, unique ID.
When a Pod changes (e.g. because of an upgrade) the
Pod name changes, but the cluster ID remains the same.

As a result, the status part of the customer resource
must list the current Pod name and cluster ID for
every server.
