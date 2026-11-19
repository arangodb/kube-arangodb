---
layout: page
title: ALB
grand_parent: ArangoDBPlatform
parent: SSO
---

# Platform SSO with AWS ALB

With the `ALB` authentication type the AWS Application Load Balancer (ALB) terminates OIDC in front of
the Platform. Instead of running the OIDC flow itself, the gateway **trusts the identity forwarded by
the ALB**: the ALB authenticates the user, signs a JWT with a per-region key and injects it into the
`x-amzn-oidc-data` request header. The gateway verifies that signature and maps the token claims onto
the ArangoDB identity.

> This feature is **Alpha** and under active development.

## How it works

1. The ALB (configured with an OIDC authenticate action) authenticates the user and forwards the
   request with the `x-amzn-oidc-data` header, a JWT signed by the ALB.
2. The gateway reads the `signer` and `kid` headers of the token, fetches the matching ALB public key
   for the configured region, and verifies the token signature (signing method `ES256`).
3. When the token is valid, the `username` claim (default `sub`) becomes the ArangoDB user and the
   optional `groups` claim is mapped to groups.
4. A request **without** the `x-amzn-oidc-data` header is left to the rest of the authentication chain.
   A request **with** an invalid or unverifiable token is rejected with `401 Unauthorized`
   (fail-closed) — it is never passed through as unauthenticated.

> The ALB must be configured to terminate OIDC and forward requests to the Platform gateway. Only
> traffic that reaches the gateway through the ALB carries a trusted `x-amzn-oidc-data` token, so the
> gateway should not be reachable on a path that bypasses the ALB.

## ALB Configuration

[Full Configuration reference ->](./api/ArangoPlatform.V1Beta1.Authentication.ALB.md)

Example:

```yaml
---

# AWS Region of the Application Load Balancer. Used to resolve the ALB public key endpoint.
region: eu-central-1

# Optional: pin the expected ALB ARN carried in the token `signer` header.
# When set, tokens signed by any other load balancer are rejected.
signer: arn:aws:elasticloadbalancing:eu-central-1:123456789012:loadbalancer/app/my-alb/0123456789abcdef

# Optional: map the token claims onto the ArangoDB identity.
claims:
  username: sub   # Claim to read the username from (default: sub)
  groups: groups  # Claim to read the groups from (optional; omit to extract no groups)
```

| Field | Description |
| --- | --- |
| `region` | **Required.** AWS Region of the ALB. Resolves the key endpoint `public-keys.auth.elb.<region>.amazonaws.com`, or, for AWS GovCloud (US) regions (`us-gov-*`), `s3-<region>.amazonaws.com/aws-elb-public-keys-prod-<region>`. |
| `signer` | Optional. Expected ALB ARN in the token `signer` header. When set, tokens from any other load balancer are rejected. |
| `claims.username` | Optional. Claim key to read the username from. Defaults to `sub`. |
| `claims.groups` | Optional. Claim key to read the groups from. The value may be a single string or a JSON array of strings. When empty, no groups are extracted. |
| `http.insecure` | Optional. Skip TLS verification when fetching the ALB signing keys. |
| `http.proxy` | Optional. HTTP proxy used when fetching the ALB signing keys. |

## Setup

In order to enable ALB authentication on the Platform, a secret with the ALB Configuration needs to be
created.

Secret Creation:

```shell
echo "---

region: eu-central-1

claims:
  username: sub
  groups: groups" > ./config.yaml

kubectl create secret generic alb-secret --from-file=config=./config.yaml
```

Once the Secret has been created, the ArangoDeployment can be configured to work with the new
authentication:

```yaml
apiVersion: "database.arangodb.com/v1"
kind: "ArangoDeployment"
metadata:
  name: "platform-simple-single"
spec:
  gateway:
    createUsers: true # Allows user creation by default from the SSO
    authentication:
      type: ALB # Picks the ALB Type of the authentication
      secret:
        name: alb-secret # Created Secret based on the Documentation
```
