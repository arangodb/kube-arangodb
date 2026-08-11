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

- [Proto](https://github.com/arangodb/kube-arangodb/blob/1.4.4/integrations/authentication/v1/definition/definition.proto)

## Endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/_integration/authn/v1/validate` | Validate a token and return its identity |
| `POST` | `/_integration/authn/v1/createToken` | Create a new signed token |
| `GET`  | `/_integration/authn/v1/identity` | Return the identity of the caller |
| `POST` | `/_integration/authn/v1/login` | Exchange credentials for a token |
| `GET`  | `/_integration/authn/v1/logout` | Invalidate the current session |

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
