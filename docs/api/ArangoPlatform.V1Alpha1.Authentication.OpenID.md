---
layout: page
parent: CRD reference
title: ArangoPlatform V1Alpha1 Authentication OpenID
---

# API Reference for ArangoPlatform V1Alpha1 Authentication OpenID

## 

### .client.id

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.48/pkg/apis/platform/v1alpha1/authentication/openid.go#L188)</sup>

ID defines OpenID Client ID

***

### .client.secret

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.48/pkg/apis/platform/v1alpha1/authentication/openid.go#L191)</sup>

Secret defines OpenID Client Secret

***

### .endpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.48/pkg/apis/platform/v1alpha1/authentication/openid.go#L51)</sup>

Endpoint defines the OpenID callback Endpoint

***

### .http.insecure

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.48/pkg/apis/platform/v1alpha1/authentication/openid.go#L144)</sup>

Insecure defines if insecure HTTP Client is used

Default Value: `false`

***

### .provider..authorizationEndpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.48/pkg/apis/platform/v1alpha1/authentication/openid.go#L175)</sup>

AuthorizationEndpoint defines OpenID Authorization Endpoint

Links:
* [Documentation](https://www.ibm.com/docs/en/was-liberty/base?topic=connect-openid-endpoint-urls#rwlp_oidc_endpoint_urls__auth_endpoint__title__1)

***

### .provider..tokenEndpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.48/pkg/apis/platform/v1alpha1/authentication/openid.go#L179)</sup>

TokenEndpoint defines OpenID Token Endpoint

Links:
* [Documentation](https://www.ibm.com/docs/en/was-liberty/base?topic=connect-openid-endpoint-urls#rwlp_oidc_endpoint_urls__token_endpoint__title__1)

***

### .provider..userInfoEndpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.48/pkg/apis/platform/v1alpha1/authentication/openid.go#L183)</sup>

UserInfoEndpoint defines OpenID UserInfo Endpoint

Links:
* [Documentation](https://www.ibm.com/docs/en/was-liberty/base?topic=connect-openid-endpoint-urls#rwlp_oidc_endpoint_urls__userinfo_endpoint__title__1)

***

### .provider.issuer

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.48/pkg/apis/platform/v1alpha1/authentication/openid.go#L169)</sup>

Issuer defines OpenID Issuer

***

### .scope

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.48/pkg/apis/platform/v1alpha1/authentication/openid.go#L54)</sup>

Scope defines OpenID Scopes (OpenID is added by default).

