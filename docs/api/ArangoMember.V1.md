# API Reference for ArangoMember V1

## Spec

### .spec.deletion_priority

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/arango_member_spec.go#L47)</sup>

DeletionPriority define Deletion Priority.
Higher value means higher priority. Default is 0.
Example: set 1 for Coordinator which should be deleted first and scale down coordinators by one.

***

### .spec.deploymentUID

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/arango_member_spec.go#L36)</sup>

DeploymentUID define Deployment UID.

***

### .spec.group

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/arango_member_spec.go#L31)</sup>

Group define Member Groups.

***

### .spec.id

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/arango_member_spec.go#L33)</sup>

***

### .spec.overrides.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/arango_member_spec_overrides.go#L38)</sup>

Resources holds resource requests & limits. Overrides template provided on the group level.

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

***

### .spec.overrides.volumeClaimTemplate

Type: `core.PersistentVolumeClaim` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/arango_member_spec_overrides.go#L33)</sup>

VolumeClaimTemplate specifies a template for volume claims. Overrides template provided on the group level.

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

***

### .spec.template.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/arango_member_pod_template.go#L60)</sup>

Checksum keep the Pod Spec Checksum (with ignored fields).

***

### .spec.template.endpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/arango_member_pod_template.go#L63)</sup>

Deprecated: Endpoint is not saved into the template

***

### .spec.template.podSpec

Type: `core.PodTemplateSpec` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/arango_member_pod_template.go#L54)</sup>

PodSpec specifies the Pod Spec used for this Member.

Links:
* [Documentation of core.PodTemplateSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podtemplatespec-v1-core)

***

### .spec.template.podSpecChecksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/arango_member_pod_template.go#L57)</sup>

PodSpecChecksum keep the Pod Spec Checksum (without ignored fields).

