---
layout: page
parent: CRD reference
title: ArangoPlatform V1Beta1 Authentication OpenID
---

# API Reference for ArangoPlatform V1Beta1 Authentication OpenID

## Object

### .claims.username

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.3/pkg/apis/platform/v1beta1/authentication/openid.go#L254)</sup>

Username defines the claim key to extract username

Default Value: `username`

***

### .client.id

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.3/pkg/apis/platform/v1beta1/authentication/openid.go#L229)</sup>

ID defines OpenID Client ID

***

### .client.secret

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.3/pkg/apis/platform/v1beta1/authentication/openid.go#L232)</sup>

Secret defines OpenID Client Secret

***

### .disabledPaths

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.3/pkg/apis/platform/v1beta1/authentication/openid.go#L67)</sup>

DisabledPaths keeps the list of SSO disabled paths. By default, "_logout" endpoint is passed through

***

### .endpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.3/pkg/apis/platform/v1beta1/authentication/openid.go#L61)</sup>

Endpoint defines the OpenID callback Endpoint

***

### .features.refreshEnabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.3/pkg/apis/platform/v1beta1/authentication/openid.go#L240)</sup>

> [!WARNING]
> ***ALPHA***
> 
> **Experimental Feature, in development**

RefreshEnabled defines if the Refresh OpenID Functionality is enabled

Default Value: `false`

***

### .http.insecure

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.3/pkg/apis/platform/v1beta1/authentication/openid.go#L185)</sup>

Insecure defines if insecure HTTP Client is used

Default Value: `false`

***

### .provider.authorizationEndpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.3/pkg/apis/platform/v1beta1/authentication/openid.go#L216)</sup>

AuthorizationEndpoint defines OpenID Authorization Endpoint

Links:
* [Documentation](https://www.ibm.com/docs/en/was-liberty/base?topic=connect-openid-endpoint-urls#rwlp_oidc_endpoint_urls__auth_endpoint__title__1)

***

### .provider.issuer

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.3/pkg/apis/platform/v1beta1/authentication/openid.go#L210)</sup>

Issuer defines OpenID Issuer

***

### .provider.tokenEndpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.3/pkg/apis/platform/v1beta1/authentication/openid.go#L220)</sup>

TokenEndpoint defines OpenID Token Endpoint

Links:
* [Documentation](https://www.ibm.com/docs/en/was-liberty/base?topic=connect-openid-endpoint-urls#rwlp_oidc_endpoint_urls__token_endpoint__title__1)

***

### .provider.userInfoEndpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.3/pkg/apis/platform/v1beta1/authentication/openid.go#L224)</sup>

UserInfoEndpoint defines OpenID UserInfo Endpoint

Links:
* [Documentation](https://www.ibm.com/docs/en/was-liberty/base?topic=connect-openid-endpoint-urls#rwlp_oidc_endpoint_urls__userinfo_endpoint__title__1)

***

### .scope

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.3/pkg/apis/platform/v1beta1/authentication/openid.go#L64)</sup>

Scope defines OpenID Scopes (OpenID is added by default).

