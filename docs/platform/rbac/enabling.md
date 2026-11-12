---
layout: page
title: Enabling RBAC
parent: RBAC
grand_parent: ArangoDBPlatform
nav_order: 1
---

# Enabling RBAC

> **Alpha Feature** - RBAC is currently in alpha (`v1alpha1`). APIs, CRD schemas,
> and behavior may change in future releases without notice.

## Prerequisites

RBAC authorization requires:
- Gateway integration enabled on the ArangoDeployment
- Integration sidecar running alongside the deployment

## Feature Flags

RBAC is controlled by two operator feature flags:

| Feature | Default | Description |
|---|---|---|
| `central-services` | Disabled | Enables the central authorization service in the sidecar |
| `rbac-enforced` | Disabled | Switches from permissive to enforced mode |

Both features depend on `gateway-integration` being enabled.

### Enabling via Helm

Set the feature flags in the operator Helm values:

```yaml
operator:
  args:
    - "--deployment.feature.central-services=true"
    - "--deployment.feature.rbac-enforced=true"
```

Or via `--set`:

```bash
helm install kube-arangodb arangodb/kube-arangodb \
  --set "operator.args={--deployment.feature.central-services=true,--deployment.feature.rbac-enforced=true}"
```

## Authorization Modes

The authorization mode is determined automatically based on enabled features:

| central-services | rbac-enforced | Mode | Behavior |
|---|---|---|---|
| Disabled | - | `always` | All requests allowed (no authorization) |
| Enabled | Disabled | `central-permissive` | Policies evaluated but denials are logged, not enforced |
| Enabled | Enabled | `central` | Full enforcement, denied requests are blocked |

### Recommended Rollout

1. **Start with permissive mode** - Enable `central-services` without
   `rbac-enforced`. This evaluates policies and logs denials without blocking
   requests. Use this to verify policies are correct.

2. **Switch to enforced mode** - Once policies are validated, enable
   `rbac-enforced` to start blocking unauthorized requests.

## Verifying RBAC is Active

Check the sidecar condition on the ArangoDeployment:

```bash
kubectl get arangodeployment <name> -o jsonpath='{.status.conditions}' | jq '.[] | select(.type == "GatewaySidecarEnabled")'
```

The sidecar exposes the authorization service on the integration port
(default 9201). The authorization mode is passed via environment variable
`INTEGRATION_AUTHORIZATION_V1_TYPE`.

You can also check via the inventory endpoint. `security.authorization` is
`AuthorizationRBAC` when RBAC is active:

```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://<gateway>/_inventory
```

The `/_management` route is only added to the gateway while the
`GatewaySidecarEnabled` condition is `True`.

### Operator Logs

The operator logs RBAC-related events at Info and Debug level under the
`sidecar-authz` and `platform-storage-operator` loggers. To view them:

```bash
# Operator pod logs
kubectl logs -n <namespace> deployment/arango-operator-operator -c operator | grep -i "authz\|permission\|rbac"

# Sidecar logs (from ArangoDB pods)
kubectl logs -n <namespace> <pod-name> -c sidecar | grep -i "authz\|permission"
```

In permissive mode, denied requests are logged at Info level with the user,
action, resource, and reason. Use these logs to validate policies before
switching to enforced mode.

## Authentication Requirement

RBAC requires authentication to be enabled on the ArangoDeployment. If
`spec.authentication.jwtSecretName` is set to `"None"`, authorization is
skipped regardless of feature flags.

## Accessing the Management API

The RBAC management API is served by the sidecar under
`/_management/permissions/*` and is reached through the deployment's gateway,
for example:

```
https://<deployment-name>-ea.<namespace>.svc:8529/_management/permissions/role
```

The gateway forwards the `Authorization` header unchanged. The sidecar
authenticates the request with a bearer JWT signed with the deployment's JWT
secret (`spec.authentication.jwtSecretName`). It reads these claims:

| Claim | Use |
|---|---|
| `preferred_username` | The user whose role bindings are evaluated |
| `groups` | Group names used for [group role bindings](group_bindings.md) |
| `exp` | Token lifetime |

Roles are not taken from the JWT. They are resolved from the user and group
role bindings stored in the sidecar at evaluation time.

Each management endpoint checks a specific RBAC action, for example
`rbac:ListRole` or `rbac:AssignUserRole`. Requests pass this check when:

- the token has **no `preferred_username` and no `groups`** claim (a superuser
  token, such as the one the operator uses internally), or
- the token's user is `root`, which the operator binds to the reserved
  `managed:predefined:super-admin` role, or
- the user's role bindings grant the required `rbac:*` action.

Example of creating a superuser token from the deployment's JWT secret
(HS256, which is the default):

```bash
SECRET=$(kubectl get secret <jwtSecretName> -o jsonpath='{.data.token}' | base64 -d)
TOKEN=$(python3 -c 'import jwt,sys,time; n=int(time.time()); print(jwt.encode({"iss":"arangodb","server_id":"admin","iat":n,"exp":n+3600}, sys.argv[1], algorithm="HS256"))' "$SECRET")

curl -H "Authorization: Bearer $TOKEN" https://<gateway>/_management/permissions/role
```

> **Warning:** In `central-permissive` mode, requests **without** credentials
> are not rejected. They are evaluated as the user `system:anonymous`, and
> because permissive mode allows denied requests, the management API can be
> used without a token. Requests with an invalid token are still rejected.
> Only `central` (enforced) mode requires a valid token and the matching
> permission.
