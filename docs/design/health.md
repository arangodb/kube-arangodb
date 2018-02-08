# Health checks

## Liveness Probe

Liveness checks are done by Kubernetes to detect `Pods` that are still running,
but not responsive.

For agents, & dbservers a liveness probe is added for `/_api/version`.

For syncmasters a liveness probe is added for `/_api/version` with
a token in an `Authorization` header. If a monitoring token is specified,
this token is used, otherwise the syncmaster JWT token is used.

For syncworkers a liveness probe is added for `/_api/version` with
a monitoring token in an `Authorization` header.
If no monitoring token is specified, there is liveness probe added for syncworkers.

## Readiness Probe

Readiness probes are done by Kubernetes to exclude `Pods` from `Services` until
they are fully ready to handle requests.

For coordinators a readiness probe is added for `/_api/version`.
