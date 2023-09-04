# API Reference for ArangoMember V1

## Spec

### .spec.deploymentUID: string

DeploymentUID define Deployment UID.

[Code Reference](/pkg/apis/deployment/v1/arango_member_spec.go#L34)

### .spec.group: int

Group define Member Groups.

[Code Reference](/pkg/apis/deployment/v1/arango_member_spec.go#L29)

### .spec.id: string

[Code Reference](/pkg/apis/deployment/v1/arango_member_spec.go#L31)

### .spec.overrides.resources: core.ResourceRequirements

Resources holds resource requests & limits. Overrides template provided on the group level.

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

[Code Reference](/pkg/apis/deployment/v1/arango_member_spec_overrides.go#L38)

### .spec.overrides.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a template for volume claims. Overrides template provided on the group level.

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

[Code Reference](/pkg/apis/deployment/v1/arango_member_spec_overrides.go#L33)

### .spec.template.checksum: string

Checksum keep the Pod Spec Checksum (with ignored fields).

[Code Reference](/pkg/apis/deployment/v1/arango_member_pod_template.go#L60)

### .spec.template.endpoint: string

Deprecated: Endpoint is not saved into the template

[Code Reference](/pkg/apis/deployment/v1/arango_member_pod_template.go#L63)

### .spec.template.podSpec: core.PodTemplateSpec

PodSpec specifies the Pod Spec used for this Member.

Links:
* [Documentation of core.PodTemplateSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podtemplatespec-v1-core)

[Code Reference](/pkg/apis/deployment/v1/arango_member_pod_template.go#L54)

### .spec.template.podSpecChecksum: string

PodSpecChecksum keep the Pod Spec Checksum (without ignored fields).

[Code Reference](/pkg/apis/deployment/v1/arango_member_pod_template.go#L57)

