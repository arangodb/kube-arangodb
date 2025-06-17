---
layout: page
title: OpenID
grand_parent: ArangoDBPlatform
parent: SSO
---

# Platform SSO with OpenID

## OpenID Configuration

[Full Configuration reference ->](./api/ArangoPlatform.V1Alpha1.Authentication.OpenID.md)

Example:

```yaml
---

client:
  id: <ID>
  secret: <SECRET>

provider:
  issuer: <ISSUER>

endpoint: https://myapp.example.com
```

## Setup

In order to enable OpenID on the Platform, secret with OpenID Configuration needs to be created.

Example setup will be followed on the example of AWS Cognito Pool.

Secret Creation:

```shell
echo "---

client:
  id: 6jomgv6104au8mm41idunxxxxx
  secret: 1uqqtp2tcrm38b31bmu756n30nrcqthisgauba3sntmm76fxxxxxx

provider:
  issuer: https://cognito-idp.eu-central-1.amazonaws.com/eu-central-1_xxxxxxxx

endpoint: https://myapp.example.com" > ./config.yaml

kubectl create secret generic openid-secret --from-file=config=./config.yaml
```

Once Secret has been created, ArangoDeployment can be configured to work with the new authentication:

```yaml
apiVersion: "database.arangodb.com/v1"
kind: "ArangoDeployment"
metadata:
  name: "platform-simple-single"
spec:
  gateway:
    createUsers: true # Allows user creation by default from the SSO
    authentication:
      type: OpenID # Picks the OpenID Type of the authentication
      secret:
        name: openid-secret # Created Secret based on the Documentation
```