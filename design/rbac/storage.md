# RBAC Storage

## ArangoDB Collections

All authorization objects are stored in ArangoDB system collections derived
from the `_users` collection properties:

| Collection | Content | Indexes |
|---|---|---|
| `_policies` | Named policies | Unique on `sequence`, TTL on `deleted` |
| `_roles` | Named roles | TTL on `deleted` |
| `_user_role_bindings` | User-role assignments with scope | TTL on `deleted` |

The TTL index on the `deleted` field provides soft-delete with automatic
cleanup. The TTL duration is configurable via the `--sidecar.auth.deleted-ttl`
flag (default: 30 days / `720h`).

All authorization collections reside in the `_system` database and are managed globally (not per-database). They use a single shard.

## Pooler

Each collection is managed by a `Pooler[T]` instance that:

- Maintains an in-memory map of all documents
- Tracks a monotonic sequence index for change detection
- Provides CRUD operations (`Create`, `Update`, `Delete`, `Item`)
- Supports change polling via `Pool(start)` for streaming updates
- Runs a background goroutine that periodically refreshes from the database

The pooler timeout is 10 seconds by default (`DefaultPoolerTimeout`).

Key file:
- `pkg/sidecar/services/authorization/pool/pool.go`

## Client-Side Sync

The authorization client maintains a streaming connection to the sidecar and
keeps an in-memory cache synchronized:

1. **Initial load** — Calls `GetPolicy()` / `GetRole()` / `GetUserRoleBinding()`
   stream RPCs to fetch all current items.
2. **Continuous polling** — Calls `PoolPolicyChanges()` / `PoolRoleChanges()` /
   `PoolUserRoleBindingChanges()` with the last known index and a 15-second
   long-poll timeout.
3. **Cache rebuild** — On each update, the `internalCache` is rebuilt with
   parsed policies, groups (role-to-policy mappings), and user role bindings.

**Sync delay breakdown:**

- Pooler refresh timeout: 10 seconds (default `DefaultPoolerTimeout`). This is how often the sidecar re-reads collections from ArangoDB.
- Client polling interval: 15-second long-poll timeout per `PoolPolicyChanges`/`PoolRoleChanges` call.
- End-to-end propagation: A change written to the database is visible to clients within ~10s (pooler refresh) + network latency. The streaming poll returns immediately when changes are detected, so under normal conditions the delay is dominated by the pooler refresh interval.

The client considers itself unhealthy if policies, roles, or user role bindings
haven't been updated within the last minute.

Key files:
- `pkg/sidecar/services/authorization/client/client.go` - Streaming client
- `pkg/sidecar/services/authorization/client/cache.go` - Cache with role/policy maps
- `pkg/sidecar/services/authorization/definition/pool.proto` - Pool service definition

### Failure Behavior

- If the client loses connection to the sidecar, it continues using the **last known cache**. Authorization decisions use stale data until the connection is restored.
- The client health check fails if policies, roles, or user role bindings haven't been updated within 1 minute, which surfaces in readiness probes.
- The system does NOT fall back to deny-all — it uses the last successfully cached policies. This means recently revoked permissions may still be granted until the cache refreshes.
- If the sidecar itself cannot reach ArangoDB, the pooler keeps serving from its in-memory state. New mutations fail but reads continue from cache.

## UserRoleBinding Key Format

User role bindings use a composite key: `<user>:<group>`. This allows
efficient prefix-based listing of all bindings for a given user by scanning
keys starting with `<user>:`.

Key file:
- `pkg/sidecar/services/authorization/impl_api_user_role_bindings.go` —
  `userRoleBindingKey()` and `userRoleBindingPrefix()` functions

## Initializer

The `NewAuthorizer()` function in `pkg/sidecar/services/authorization/impl.go`
creates all three poolers:

```go
policies:         pool.NewPooler[*Policy]("_policies", ...)
roles:            pool.NewPooler[*Role]("_roles", ...)
userRoleBindings: pool.NewPooler[*UserRoleBinding]("_user_role_bindings", ...)
```

Health checks require all three poolers to be ready. The `Refresh()` method
refreshes all three in sequence.
