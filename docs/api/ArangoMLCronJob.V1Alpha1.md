---
layout: page
parent: CRD reference
title: ArangoMLCronJob V1Alpha1
---

# API Reference for ArangoMLCronJob V1Alpha1

## Spec

### .spec

Type: `batch.CronJobSpec` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.48/pkg/apis/ml/v1alpha1/cronjob_spec.go#L33)</sup>

Links:
* [Kubernetes Documentation](https://godoc.org/k8s.io/api/batch/v1#CronJobSpec)

## Status

### .status

Type: `batch.CronJobStatus` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.48/pkg/apis/ml/v1alpha1/cronjob_status.go#L37)</sup>

Links:
* [Kubernetes Documentation](https://godoc.org/k8s.io/api/batch/v1#CronJobStatus)

***

### .status.mlConditions

Type: `api.MLConditions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.48/pkg/apis/ml/v1alpha1/cronjob_status.go#L33)</sup>

MLConditions specific to the entire cron job

***

### .status.ref.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.48/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .status.ref.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.48/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .status.ref.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.48/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.ref.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.48/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

