---
layout: page
title: Common Use Cases
parent: ArangoDBPlatform
nav_order: 1
---

# Custom ImagePullSecrets

To inject ImagePullSecrets to all pods across the Platform [ArangoProfile](./arango-profile-resource.md) can be used.

Example:
```yaml
apiVersion: scheduler.arangodb.com/v1beta1
kind: ArangoProfile
metadata:
  name: image-secrets
spec:
  selectors:
    label:
      matchLabels: {}
  template:
    pod:
      imagePullSecrets:
        - <Secret Name>
    priority: 129
```

# Custom CA Certificates

Platform images trust only public Certificate Authorities. When a Platform service connects over TLS to
infrastructure signed by a private CA (for example an on-prem S3 endpoint, a container registry, an
identity provider or an LLM endpoint), the connection fails with an unknown-authority error. To make
every Platform pod trust additional CAs, an [ArangoProfile](./arango-profile-resource.md) can mount a
CA bundle over the system trust store and point the common TLS libraries at it.

Create a `ConfigMap` holding the bundle (public roots **and** your custom CAs — the mount replaces the
system store, so it must be complete):

```shell
kubectl create configmap custom-ca-bundle --from-file 'ca-bundle.crt=<CA Bundle Path>'
```

Then inject it everywhere with a profile:
```yaml
apiVersion: scheduler.arangodb.com/v1beta1
kind: ArangoProfile
metadata:
  name: custom-ca
spec:
  selectors:
    label:
      matchLabels: {}
  template:
    priority: 200
    pod:
      volumes:
        - name: custom-ca
          configMap:
            name: custom-ca-bundle
            items:
              - key: ca-bundle.crt
                path: ca-bundle.crt
    container:
      all:
        volumeMounts:
          - name: custom-ca
            mountPath: /etc/ssl/certs/ca-certificates.crt
            subPath: ca-bundle.crt
            readOnly: true
          - name: custom-ca
            mountPath: /etc/pki/tls/certs/ca-bundle.crt
            subPath: ca-bundle.crt
            readOnly: true
          - name: custom-ca
            mountPath: /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem
            subPath: ca-bundle.crt
            readOnly: true
        env:
          - name: SSL_CERT_FILE
            value: /etc/ssl/certs/ca-certificates.crt
          - name: REQUESTS_CA_BUNDLE
            value: /etc/ssl/certs/ca-certificates.crt
          - name: NODE_EXTRA_CA_CERTS
            value: /etc/ssl/certs/ca-certificates.crt
```

The two mount paths cover Debian/Ubuntu (`/etc/ssl/certs`) and RHEL/UBI (`/etc/pki`) based images; the
environment variables cover OpenSSL, Python/`requests` and Node.js.

Optionally, [trust-manager](https://cert-manager.io/docs/trust/trust-manager/) can assemble and
distribute the bundle (public roots plus your CAs) to every namespace as the `custom-ca-bundle`
ConfigMap:
```yaml
apiVersion: trust.cert-manager.io/v1alpha1
kind: Bundle
metadata:
  name: custom-ca-bundle
spec:
  sources:
    - useDefaultCAs: true
    - configMap:
        name: customer-root-cas
        key: ca.crt
  target:
    configMap:
      key: ca-bundle.crt
```

Notes:

* The bundle **replaces** the system trust store, so it must include the public roots as well as the
  custom CAs (`useDefaultCAs: true` when using trust-manager).
* A pod is injected only if it carries the `profiles.arangodb.com/apply` or
  `profiles.arangodb.com/deployment` label, and only on CREATE, so existing pods must be restarted to
  pick up a new or changed profile.
* Profiles and the ConfigMap are namespaced: create one per namespace (trust-manager handles the
  ConfigMap distribution).
* `subPath` mounts do not refresh when the ConfigMap changes: restart pods after rotating the CA bundle.
* `container.all` covers `spec.containers` (including the profile-injected `integration` sidecar) but
  not `initContainers` or native sidecars unless the `apply-profile-all-to-init-containers` feature is
  enabled.
* The JVM does not read these settings; Java-based components need their own truststore
  (`-Djavax.net.ssl.trustStore`).
* For S3-backed [ArangoPlatformStorage](./platform/storage/s3.md), the storage integration reads its CA
  from `spec.backend.s3.caSecret` instead; set that in addition to the profile above.

# Scheduling Platform Pods on Specific Nodes

To pin all Platform pods to a dedicated node pool (for example accredited or isolated nodes), an
[ArangoProfile](./arango-profile-resource.md) can inject a `nodeSelector`, `tolerations` and
`affinity` into every matching pod.

Example:
```yaml
apiVersion: scheduler.arangodb.com/v1beta1
kind: ArangoProfile
metadata:
  name: node-placement
spec:
  selectors:
    label:
      matchLabels: {}
  template:
    priority: 200
    pod:
      nodeSelector:
        node-pool: arango-platform
      tolerations:
        - key: dedicated
          operator: Equal
          value: arango-platform
          effect: NoSchedule
```

Notes:

* A pod is injected only if it carries the `profiles.arangodb.com/apply` or
  `profiles.arangodb.com/deployment` label, and only on CREATE, so existing pods must be restarted to
  pick up a new or changed profile.
* Profiles are namespaced: create one per namespace.
* The `nodeSelector` is **merged into** the pod's existing selector (keys set by the chart or workload
  are kept), rather than replacing it. Across several matching profiles the keys are merged and the
  higher `priority` wins on a conflict. A value of `"-"` removes a key set by a lower-priority profile
  or by the pod. A profile that sets no `nodeSelector` leaves the pod's selector untouched.
* `tolerations` are added only if not already present. `affinity`, when set, replaces the pod's
  affinity.
* Database members (ArangoDeployment) are scheduled through `spec.<group>.nodeSelector`,
  `spec.<group>.tolerations` and `spec.<group>.affinity` on the ArangoDeployment resource itself; set
  those there in addition to the profile above.