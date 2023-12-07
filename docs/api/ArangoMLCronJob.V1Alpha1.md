# API Reference for ArangoMLCronJob V1Alpha1

## Spec

### .spec

Type: `batch.CronJobSpec` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/cronjob_spec.go#L33)</sup>

Links:
* [Kubernetes Documentation](https://godoc.org/k8s.io/api/batch/v1beta1#CronJobSpec)

## Status

### .status

Type: `batch.CronJobStatus` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/cronjob_status.go#L38)</sup>

Links:
* [Kubernetes Documentation](https://godoc.org/k8s.io/api/batch/v1beta1#CronJobStatus)

***

### .status.conditions

Type: `api.Conditions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/cronjob_status.go#L34)</sup>

Conditions specific to the entire cron job

***

### .status.ref.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.ref.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.ref.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

