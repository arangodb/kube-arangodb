# RBAC Runtime Behavior

How the RBAC system behaves in practice — propagation timing, session
impact, and degraded-state handling.

## Where RBAC Data Is Stored

RBAC is managed centrally per ArangoDeployment. The authorization service runs
in the integration sidecar and stores its data in system collections of the
deployment's `_system` database:

| Collection             | Content  |
|------------------------|----------|
| `_policies`            | Policies |
| `_roles`               | Roles    |
| `_user_role_bindings`  | User role bindings with their scopes  |
| `_group_role_bindings` | Group role bindings with their scopes |

Every sidecar instance reads the same collections, so changes made through the
management API on any gateway, or through CRDs, apply to all services of the
deployment. Separate deployments have separate RBAC data.

Deletions are soft: a tombstone is kept so that all sidecars observe the
delete, and it is removed after the sidecar's `--sidecar.auth.deleted-ttl`
(default 30 days).

Platform services do not store permissions themselves. They only define the
action names they check and ask their local sidecar to evaluate them (see
[Making a Service RBAC-Ready](service_integration.md)).

## Permission Change Propagation

### When roles or policies are modified, how long until it takes effect?

Changes propagate through a pipeline:

1. **CRD update** — immediate (Kubernetes API)
2. **Operator reconciliation** — seconds (operator watches CRD changes)
3. **Sidecar pool refresh** — up to 10 seconds (pooler reads from ArangoDB).
   `POST /_management/permissions/refresh` triggers an immediate refresh on
   the sidecar that receives the request.
4. **Client cache sync** — up to 15 seconds (streaming long-poll)

**Total worst-case latency: ~30 seconds.**

In practice, most changes are visible within 10–15 seconds. The system does
not provide a real-time guarantee — there is always a window where the old
permissions are still cached.

### When I modify my own roles, do I get an immediate update?

No. Permission changes are **not session-aware**. The same propagation delay
applies whether you modify your own roles or someone else's.

You do **not** need to log out and log in. The RBAC evaluation happens on
every request using the current server-side state, not from the JWT token.
Once the change propagates through the cache pipeline, subsequent requests
use the updated permissions automatically.

### Do I need to re-issue the JWT token?

No. Permissions are evaluated server-side using role and policy data stored
in the sidecar, not from JWT claims. The JWT identifies the user; the sidecar
resolves the user's roles and policies at evaluation time. Changing a role
or policy takes effect without re-issuing the token.

## Degraded State Behavior

### What happens if the RBAC system is unreachable?

The behavior depends on the authorization mode:

| Mode | Behavior when degraded |
|------|----------------------|
| `Central` (strict) | Requests that require authorization **fail** with an error. The system does not silently allow or deny — it returns an explicit failure. |
| `CentralPermissive` (default) | Requests that fail authorization evaluation are **allowed** with a warning logged. This prevents outages but may temporarily grant broader access. |

### What happens during a network split?

If a pod loses connection to the sidecar:

- The **streaming cache client** continues using the last known data. It does
  not fall back to deny-all — it serves from the stale cache.
- Recently revoked permissions **may still be granted** until the connection
  is restored and the cache refreshes.
- The client health check fails after 1 minute without updates, which
  surfaces in readiness probes.

If the sidecar loses connection to ArangoDB:

- The **sidecar pooler** keeps serving from its in-memory state.
- New mutations (create/update/delete policies or roles) **fail**.
- Read operations (permission evaluation) continue from the cached data.

### What happens if the operator is down?

- Existing pods and sidecars continue running with their current configuration.
- CRD changes are queued by Kubernetes and reconciled when the operator restarts.
- No new tokens or role bindings are created until the operator recovers.
- Running services are **not affected** — they use the sidecar directly.

## Summary

| Scenario | Impact | Recovery |
|----------|--------|----------|
| Role/policy changed | ~30s propagation delay | Automatic |
| User modifies own roles | Same delay, no re-login needed | Automatic |
| Sidecar unreachable | Stale cache used, no deny-all | Automatic on reconnect |
| ArangoDB unreachable | Reads from cache, writes fail | Automatic on reconnect |
| Operator down | No new reconciliation | Automatic on restart |
| Network split | Stale permissions in affected pods | Automatic on network restore |
