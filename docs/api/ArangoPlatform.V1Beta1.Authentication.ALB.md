---
layout: page
parent: CRD reference
title: ArangoPlatform V1Beta1 Authentication ALB
---

# API Reference for ArangoPlatform V1Beta1 Authentication ALB

## Object

### .claims.groups

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.8/pkg/apis/platform/v1beta1/authentication/alb.go#L180)</sup>

Groups defines the claim key to extract the groups from. When empty, no groups are extracted.

***

### .claims.username

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.8/pkg/apis/platform/v1beta1/authentication/alb.go#L177)</sup>

Username defines the claim key to extract the username from.

Default Value: `sub`

***

### .http.insecure

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.8/pkg/apis/platform/v1beta1/authentication/openid.go#L187)</sup>

Insecure defines if insecure HTTP Client is used

Default Value: `false`

***

### .http.proxy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.8/pkg/apis/platform/v1beta1/authentication/openid.go#L191)</sup>

Proxy defines an explicit proxy URL (e.g. `http://proxy:3128`) to route outbound requests
through. When empty, no proxy is used.

***

### .region

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.8/pkg/apis/platform/v1beta1/authentication/alb.go#L64)</sup>

Region defines the AWS Region of the Application Load Balancer. It is used to resolve the
public key endpoint `public-keys.auth.elb.<region>.amazonaws.com`, or, for AWS GovCloud (US)
regions (`us-gov-*`), the S3-hosted endpoint `s3-<region>.amazonaws.com/aws-elb-public-keys-prod-<region>`.

***

### .signer

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.8/pkg/apis/platform/v1beta1/authentication/alb.go#L68)</sup>

Signer optionally pins the expected ALB ARN carried in the token `signer` header. When set,
tokens signed by any other load balancer are rejected.

