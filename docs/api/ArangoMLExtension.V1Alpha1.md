# API Reference for ArangoMLExtension V1Alpha1

## Spec

### .spec.deployment.affinity

Type: `core.Affinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/scheduling.go#L37)</sup>

Affinity defines scheduling constraints for workload

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity)

***

### .spec.deployment.hostIPC

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/container_namespace.go#L33)</sup>

HostIPC defines to use the host's ipc namespace.

Default Value: `false`

***

### .spec.deployment.hostNetwork

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/container_namespace.go#L27)</sup>

HostNetwork requests Host network for this pod. Use the host's network namespace.
If this option is set, the ports that will be used must be specified.

Default Value: `false`

***

### .spec.deployment.hostPID

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/container_namespace.go#L30)</sup>

HostPID define to use the host's pid namespace.

Default Value: `false`

***

### .spec.deployment.nodeSelector

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/scheduling.go#L32)</sup>

NodeSelector is a selector that must be true for the workload to fit on a node.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector)

***

### .spec.deployment.podSecurityContext

Type: `core.PodSecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/security_pod.go#L29)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.deployment.prediction.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L31)</sup>

Image define image details

***

### .spec.deployment.prediction.port

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/extension_spec_deployment_component.go#L30)</sup>

Port defines on which port the container will be listening for connections

***

### .spec.deployment.prediction.pullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L35)</sup>

PullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.deployment.prediction.pullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L38)</sup>

PullSecrets define Secrets used to pull Image from registry

***

### .spec.deployment.prediction.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/resources.go#L34)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

***

### .spec.deployment.prediction.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/security_container.go#L29)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.deployment.project.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L31)</sup>

Image define image details

***

### .spec.deployment.project.port

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/extension_spec_deployment_component.go#L30)</sup>

Port defines on which port the container will be listening for connections

***

### .spec.deployment.project.pullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L35)</sup>

PullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.deployment.project.pullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L38)</sup>

PullSecrets define Secrets used to pull Image from registry

***

### .spec.deployment.project.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/resources.go#L34)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

***

### .spec.deployment.project.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/security_container.go#L29)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.deployment.replicas

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/extension_spec_deployment.go#L56)</sup>

Replicas defines the number of replicas running specified components. No replicas created if no components are defined.

Default Value: `1`

***

### .spec.deployment.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/scheduling.go#L47)</sup>

SchedulerName specifies, the pod will be dispatched by specified scheduler.
If not specified, the pod will be dispatched by default scheduler.

Default Value: `""`

***

### .spec.deployment.service.type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/extension_spec_deployment_service.go#L37)</sup>

Type determines how the Service is exposed

Links:
* [Kubernetes Documentation](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types)

Possible Values: 
* ClusterIP (default) - service will only be accessible inside the cluster, via the cluster IP
* NodePort - service will be exposed on one port of every node, in addition to 'ClusterIP' type
* LoadBalancer - service will be exposed via an external load balancer (if the cloud provider supports it), in addition to 'NodePort' type
* ExternalName - service consists of only a reference to an external name that kubedns or equivalent will return as a CNAME record, with no exposing or proxying of any pods involved

***

### .spec.deployment.shareProcessNamespace

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/container_namespace.go#L39)</sup>

ShareProcessNamespace defines to share a single process namespace between all of the containers in a pod.
When this is set containers will be able to view and signal processes from other containers
in the same pod, and the first process in each container will not be assigned PID 1.
HostPID and ShareProcessNamespace cannot both be set.

Default Value: `false`

***

### .spec.deployment.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/scheduling.go#L42)</sup>

Tolerations defines tolerations

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)

***

### .spec.deployment.training.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L31)</sup>

Image define image details

***

### .spec.deployment.training.port

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/extension_spec_deployment_component.go#L30)</sup>

Port defines on which port the container will be listening for connections

***

### .spec.deployment.training.pullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L35)</sup>

PullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.deployment.training.pullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L38)</sup>

PullSecrets define Secrets used to pull Image from registry

***

### .spec.deployment.training.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/resources.go#L34)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

***

### .spec.deployment.training.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/security_container.go#L29)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L31)</sup>

Image define image details

***

### .spec.init.affinity

Type: `core.Affinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/scheduling.go#L37)</sup>

Affinity defines scheduling constraints for workload

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity)

***

### .spec.init.hostIPC

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/container_namespace.go#L33)</sup>

HostIPC defines to use the host's ipc namespace.

Default Value: `false`

***

### .spec.init.hostNetwork

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/container_namespace.go#L27)</sup>

HostNetwork requests Host network for this pod. Use the host's network namespace.
If this option is set, the ports that will be used must be specified.

Default Value: `false`

***

### .spec.init.hostPID

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/container_namespace.go#L30)</sup>

HostPID define to use the host's pid namespace.

Default Value: `false`

***

### .spec.init.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L31)</sup>

Image define image details

***

### .spec.init.nodeSelector

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/scheduling.go#L32)</sup>

NodeSelector is a selector that must be true for the workload to fit on a node.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector)

***

### .spec.init.podSecurityContext

Type: `core.PodSecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/security_pod.go#L29)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.init.pullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L35)</sup>

PullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.init.pullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L38)</sup>

PullSecrets define Secrets used to pull Image from registry

***

### .spec.init.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/resources.go#L34)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

***

### .spec.init.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/scheduling.go#L47)</sup>

SchedulerName specifies, the pod will be dispatched by specified scheduler.
If not specified, the pod will be dispatched by default scheduler.

Default Value: `""`

***

### .spec.init.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/security_container.go#L29)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.init.shareProcessNamespace

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/container_namespace.go#L39)</sup>

ShareProcessNamespace defines to share a single process namespace between all of the containers in a pod.
When this is set containers will be able to view and signal processes from other containers
in the same pod, and the first process in each container will not be assigned PID 1.
HostPID and ShareProcessNamespace cannot both be set.

Default Value: `false`

***

### .spec.init.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/scheduling.go#L42)</sup>

Tolerations defines tolerations

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)

***

### .spec.jobsTemplates.prediction.affinity

Type: `core.Affinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/scheduling.go#L37)</sup>

Affinity defines scheduling constraints for workload

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity)

***

### .spec.jobsTemplates.prediction.hostIPC

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/container_namespace.go#L33)</sup>

HostIPC defines to use the host's ipc namespace.

Default Value: `false`

***

### .spec.jobsTemplates.prediction.hostNetwork

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/container_namespace.go#L27)</sup>

HostNetwork requests Host network for this pod. Use the host's network namespace.
If this option is set, the ports that will be used must be specified.

Default Value: `false`

***

### .spec.jobsTemplates.prediction.hostPID

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/container_namespace.go#L30)</sup>

HostPID define to use the host's pid namespace.

Default Value: `false`

***

### .spec.jobsTemplates.prediction.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L31)</sup>

Image define image details

***

### .spec.jobsTemplates.prediction.nodeSelector

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/scheduling.go#L32)</sup>

NodeSelector is a selector that must be true for the workload to fit on a node.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector)

***

### .spec.jobsTemplates.prediction.podSecurityContext

Type: `core.PodSecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/security_pod.go#L29)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.jobsTemplates.prediction.pullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L35)</sup>

PullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.jobsTemplates.prediction.pullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L38)</sup>

PullSecrets define Secrets used to pull Image from registry

***

### .spec.jobsTemplates.prediction.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/resources.go#L34)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

***

### .spec.jobsTemplates.prediction.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/scheduling.go#L47)</sup>

SchedulerName specifies, the pod will be dispatched by specified scheduler.
If not specified, the pod will be dispatched by default scheduler.

Default Value: `""`

***

### .spec.jobsTemplates.prediction.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/security_container.go#L29)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.jobsTemplates.prediction.shareProcessNamespace

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/container_namespace.go#L39)</sup>

ShareProcessNamespace defines to share a single process namespace between all of the containers in a pod.
When this is set containers will be able to view and signal processes from other containers
in the same pod, and the first process in each container will not be assigned PID 1.
HostPID and ShareProcessNamespace cannot both be set.

Default Value: `false`

***

### .spec.jobsTemplates.prediction.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/scheduling.go#L42)</sup>

Tolerations defines tolerations

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)

***

### .spec.jobsTemplates.training.affinity

Type: `core.Affinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/scheduling.go#L37)</sup>

Affinity defines scheduling constraints for workload

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity)

***

### .spec.jobsTemplates.training.hostIPC

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/container_namespace.go#L33)</sup>

HostIPC defines to use the host's ipc namespace.

Default Value: `false`

***

### .spec.jobsTemplates.training.hostNetwork

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/container_namespace.go#L27)</sup>

HostNetwork requests Host network for this pod. Use the host's network namespace.
If this option is set, the ports that will be used must be specified.

Default Value: `false`

***

### .spec.jobsTemplates.training.hostPID

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/container_namespace.go#L30)</sup>

HostPID define to use the host's pid namespace.

Default Value: `false`

***

### .spec.jobsTemplates.training.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L31)</sup>

Image define image details

***

### .spec.jobsTemplates.training.nodeSelector

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/scheduling.go#L32)</sup>

NodeSelector is a selector that must be true for the workload to fit on a node.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector)

***

### .spec.jobsTemplates.training.podSecurityContext

Type: `core.PodSecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/security_pod.go#L29)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.jobsTemplates.training.pullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L35)</sup>

PullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.jobsTemplates.training.pullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L38)</sup>

PullSecrets define Secrets used to pull Image from registry

***

### .spec.jobsTemplates.training.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/resources.go#L34)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

***

### .spec.jobsTemplates.training.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/scheduling.go#L47)</sup>

SchedulerName specifies, the pod will be dispatched by specified scheduler.
If not specified, the pod will be dispatched by default scheduler.

Default Value: `""`

***

### .spec.jobsTemplates.training.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/security_container.go#L29)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.jobsTemplates.training.shareProcessNamespace

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/container_namespace.go#L39)</sup>

ShareProcessNamespace defines to share a single process namespace between all of the containers in a pod.
When this is set containers will be able to view and signal processes from other containers
in the same pod, and the first process in each container will not be assigned PID 1.
HostPID and ShareProcessNamespace cannot both be set.

Default Value: `false`

***

### .spec.jobsTemplates.training.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/scheduling.go#L42)</sup>

Tolerations defines tolerations

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)

***

### .spec.metadataService.local.arangoMLFeatureStore

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/extension_spec_metadata_service.go#L65)</sup>

ArangoMLFeatureStoreDatabase define Database name to be used as MetadataService Backend in ArangoMLFeatureStoreDatabase

Default Value: `arangomlfeaturestore`

***

### .spec.metadataService.local.arangoPipeDatabase

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/extension_spec_metadata_service.go#L61)</sup>

ArangoPipeDatabase define Database name to be used as MetadataService Backend in ArangoPipe

Default Value: `arangopipe`

***

### .spec.pullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L35)</sup>

PullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.pullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L38)</sup>

PullSecrets define Secrets used to pull Image from registry

***

### .spec.storage.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .spec.storage.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.storage.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

## Status

### .status.arangoDB.secret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.arangoDB.secret.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.arangoDB.secret.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.conditions

Type: `api.Conditions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/extension_status.go#L31)</sup>

Conditions specific to the entire extension

***

### .status.metadataService.local.arangoMLFeatureStore

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/extension_status_metadata_service.go#L38)</sup>

ArangoMLFeatureStoreDatabase define Database name to be used as MetadataService Backend

***

### .status.metadataService.local.arangoPipe

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/extension_status_metadata_service.go#L35)</sup>

ArangoPipeDatabase define Database name to be used as MetadataService Backend

***

### .status.metadataService.secret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.metadataService.secret.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.metadataService.secret.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.serviceAccount.cluster.binding.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.serviceAccount.cluster.binding.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.serviceAccount.cluster.binding.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.serviceAccount.cluster.role.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.serviceAccount.cluster.role.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.serviceAccount.cluster.role.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.serviceAccount.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.serviceAccount.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.serviceAccount.namespaced.binding.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.serviceAccount.namespaced.binding.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.serviceAccount.namespaced.binding.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.serviceAccount.namespaced.role.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.serviceAccount.namespaced.role.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.serviceAccount.namespaced.role.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.serviceAccount.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

