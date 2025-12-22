---
layout: page
parent: ArangoDBPlatform
title: Platform Installation File Schema
---

# Installation Definition

## Example

```yaml
packages:
  nginx: # OCI
    chart: "oci://ghcr.io/nginx/charts/nginx-ingress:2.3.1"
    version: 2.3.1
  prometheus: # Helm Index
    chart: "index://prometheus-community.github.io/helm-charts"
    version: 1.3.1
  alertmanager: # Remote Chart
    chart: "https://github.com/prometheus-community/helm-charts/releases/download/alertmanager-0.1.0/alertmanager-0.1.0.tgz"
    version: "0.1.0"
  local: # Local File
    chart: "file:///tmp/local-0.1.0.tgz"
    version: "0.1.0"
  inline: # Inline
    chart: "<base64 string>"
    version: "0.2.5"
  platform: # Platform LicenseManager
    version: v3.0.11
```

## Package

### .package.packages.\<string\>.chart

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/util/k8sutil/helm/package.go#L83)</sup>

Chart defines override of the PackageSpec
It supports multiple modes:
- If undefined, LicenseManager OCI Repository is used
- If starts with `file://` chart is fetched from local FileSystem
- If starts with `http://` or `https://` chart is fetched from the remote URL
- If starts with `index://` chart is fetched using Helm YAML Index File structure (using version and name)
- If Starts with `oci://` chart is fetched from Registry Compatible OCI Repository
- If none above match, chart is decoded using Base64 encoding

***

### .package.packages.\<string\>.overrides

Type: `Object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/util/k8sutil/helm/package.go#L87)</sup>

Overrides defines Values to override the Helm Chart Defaults (merged with Service Overrides)

***

### .package.packages.\<string\>.stage

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/util/k8sutil/helm/package.go#L70)</sup>

Stage defines stage used in the fetch from LicenseManager

***

### .package.packages.\<string\>.version

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/util/k8sutil/helm/package.go#L73)</sup>

Version keeps the version of the PackageSpec

***

### .package.releases.\<string\>.overrides

Type: `Object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/util/k8sutil/helm/package.go#L167)</sup>

Overrides defines Values to override the Helm Chart Defaults during installation

***

### .package.releases.\<string\>.package

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/util/k8sutil/helm/package.go#L163)</sup>

Package keeps the name of the Chart used from the installation script.
References to value provided in Packages

