---
layout: page
title: Integration Sidecar Authentication V1
grand_parent: ArangoDBPlatform
parent: Integration Sidecars
---

# Authentication V1

The Authentication V1 integration service validates ArangoDB JWT tokens and
issues new ones on behalf of the deployment. It is used by services (and by
legacy, JWT-based clients) that need a valid ArangoDB token without having
direct access to the deployment JWT secret.

## Service Definition

- [Proto](https://github.com/arangodb/kube-arangodb/blob/1.4.5/integrations/authentication/v1/definition/definition.proto)

## Endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/_integration/authn/v1/validate` | Validate a token and return its identity |
| `POST` | `/_integration/authn/v1/createToken` | Create a new signed token |
| `GET`  | `/_integration/authn/v1/identity` | Return the identity of the caller |
| `POST` | `/_integration/authn/v1/login` | Exchange credentials for a token |
| `GET`  | `/_integration/authn/v1/logout` | Invalidate the current session |

## RBAC Permissions

Only `CreateToken` is RBAC-gated, and only when central services are enabled with an asymmetric
signing key; the caller's token must be granted the matching action via an
[`ArangoPermissionPolicy`](../platform/rbac/policies.md) bound to their role. `Validate`, `Identity`,
`Login` and `Logout` are not gated by an RBAC action.

| Action | Resource |
|---|---|
| `authentication:CreateToken` | the user the token is minted for (e.g. `root`, or `*` for any user) |

See [Authorization gating](#authorization-gating) below for exactly when this applies and how a mint
request without an explicit `user` is handled.

## Token creation

`CreateToken` signs a new token with the deployment JWT secret. The request
accepts:

- `user` — the user the token is issued for. Defaults to
  `--integration.authentication.v1.token.user` (`root`).
- `lifetime` — the token lifetime, clamped to
  `--integration.authentication.v1.token.ttl.min` /
  `--integration.authentication.v1.token.ttl.max`
  (default `--integration.authentication.v1.token.ttl.default`, 1h).
- `groups` — optional groups assigned to the token.

### Authorization gating

Token creation is gated by the [Authorization V1](authorization.v1.md) (IAM)
service when central services are enabled (that is, when asymmetric signing
keys and remote validation are in use). In that mode the authentication service
keeps running locally — so per-request `Validate` stays local — but delegates
the `CreateToken` decision to the central authorization service (SuperUser
wrapped). If the caller is not authorized to mint a token for the requested
`user`, the request is rejected.

When central services are not enabled, `CreateToken` is served locally.

> The previous static allow-list (`--integration.authentication.v1.token.allowed`)
> has been removed; who may create a token is now governed by the authorization
> service instead of a fixed list.

### Required RBAC permission

When the gate applies (central services enabled **and** an asymmetric signing key
in use), the caller of `createToken` must be granted the following permission via
an [`ArangoPermissionPolicy`](../platform/rbac/policies.md) bound to their role:

| Action | Resource |
|---|---|
| `authentication:CreateToken` | the user the token is minted for (e.g. `root`, or `*` for any user) |

Example policy statement allowing a subject to mint tokens for `root`:

```yaml
statements:
  - effect: Allow
    actions:
      - "authentication:CreateToken"
    resources:
      - "root"
```

A mint request without an explicit `user` is a privileged default mint and is
allowed by the SuperUser wrapper (no policy required). With symmetric keys, or
when central services are not enabled, no permission is required and
`createToken` behaves as before.

## Example: obtain a `root` token

This is the typical way for a service (or a legacy, JWT-based client) to get a
usable ArangoDB token without reading the deployment JWT secret. Call the
integration sidecar's `createToken` endpoint:

```bash
# Request a root token valid for 1 hour.
curl -sk -X POST \
  "https://<integration-sidecar>:<port>/_integration/authn/v1/createToken" \
  -H "Content-Type: application/json" \
  -d '{"user": "root", "lifetime": "1h"}'
```

Response:

```json
{
  "lifetime": "3600s",
  "user": "root",
  "token": "<JWT>",
  "groups": []
}
```

Use the returned `token` as the ArangoDB bearer token:

```bash
curl -sk "https://<coordinator>:8529/_api/version" \
  -H "Authorization: bearer <JWT>"
```

When central services are enabled the caller of `createToken` must itself be
authorized (by the authorization service) to mint a token for `root`.

## Pod environment variables

The deployment-wide authentication and authorization *modes* are exposed by the
integration sidecar profile:

| Env var | Values | Description |
|---|---|---|
| `INTEGRATION_AUTHENTICATION_MODE` | `None` / `Native` / `SSO` | Authentication mode: `SSO` (Gateway OpenID), `Native` (ArangoDB JWT) or `None` (disabled) |
| `INTEGRATION_AUTHORIZATION_MODE` | `None` / `Native` / `RBAC` | Effective authorization mode (`RBAC` when enforced by the platform gateway) |
| `INTEGRATION_AUTHORIZATION_MODE_COREDB` | `None` / `Native` | Authorization enforced by the ArangoDB core (RBAC reported as `Native`, since it is enforced upstream at the gateway) |

See [Integration Sidecars](../integration-sidecar.md#envs) for the full list of
integration sidecar environment variables.
