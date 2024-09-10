---
layout: page
parent: CRD reference
title: GraphAnalyticsEngine V1Alpha1
---

# API Reference for GraphAnalyticsEngine V1Alpha1

## Spec

### .spec.deployment.affinity

Type: `core.Affinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/pod/resources/scheduling.go#L44)</sup>

Affinity defines scheduling constraints for workload

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity)

***

### .spec.deployment.annotations

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/pod/resources/metadata.go#L45)</sup>

Annotations is an unstructured key value map stored with a resource that may be
set by external tools to store and retrieve arbitrary metadata. They are not
queryable and should be preserved when modifying objects.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations)

***

### .spec.deployment.args

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/core.go#L50)</sup>

Arguments to the entrypoint.
The container image's CMD is used if this is not provided.
Variable references $(VAR_NAME) are expanded using the container's environment. If a variable
cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced
to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will
produce the string literal "$(VAR_NAME)". Escaped references will never be expanded, regardless
of whether the variable exists or not. Cannot be updated.

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell)

***

### .spec.deployment.automountServiceAccountToken

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/pod/resources/service_account.go#L38)</sup>

AutomountServiceAccountToken indicates whether a service account token should be automatically mounted.

***

### .spec.deployment.command

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/core.go#L40)</sup>

Entrypoint array. Not executed within a shell.
The container image's ENTRYPOINT is used if this is not provided.
Variable references $(VAR_NAME) are expanded using the container's environment. If a variable
cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced
to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will
produce the string literal "$(VAR_NAME)". Escaped references will never be expanded, regardless
of whether the variable exists or not. Cannot be updated.

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell)

***

### .spec.deployment.env

Type: `core.EnvVar` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/environments.go#L36)</sup>

Env keeps the information about environment variables provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core)

***

### .spec.deployment.envFrom

Type: `core.EnvFromSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/environments.go#L41)</sup>

EnvFrom keeps the information about environment variable sources provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envfromsource-v1-core)

***

### .spec.deployment.hostIPC

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/pod/resources/namespace.go#L42)</sup>

HostIPC defines to use the host's ipc namespace.

Default Value: `false`

***

### .spec.deployment.hostNetwork

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/pod/resources/namespace.go#L36)</sup>

HostNetwork requests Host network for this pod. Use the host's network namespace.
If this option is set, the ports that will be used must be specified.

Default Value: `false`

***

### .spec.deployment.hostPID

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/pod/resources/namespace.go#L39)</sup>

HostPID define to use the host's pid namespace.

Default Value: `false`

***

### .spec.deployment.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/image.go#L35)</sup>

Image define image details

***

### .spec.deployment.imagePullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/image.go#L39)</sup>

ImagePullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.deployment.imagePullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/pod/resources/image.go#L36)</sup>

ImagePullSecrets define Secrets used to pull Image from registry

***

### .spec.deployment.labels

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/pod/resources/metadata.go#L39)</sup>

Map of string keys and values that can be used to organize and categorize
(scope and select) objects. May match selectors of replication controllers
and services.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels)

***

### .spec.deployment.lifecycle

Type: `core.Lifecycle` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/lifecycle.go#L35)</sup>

Lifecycle keeps actions that the management system should take in response to container lifecycle events.

***

### .spec.deployment.livenessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/probes.go#L37)</sup>

LivenessProbe keeps configuration of periodic probe of container liveness.
Container will be restarted if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.deployment.nodeSelector

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/pod/resources/scheduling.go#L39)</sup>

NodeSelector is a selector that must be true for the workload to fit on a node.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector)

***

### .spec.deployment.ownerReferences

Type: `meta.OwnerReference` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/pod/resources/metadata.go#L52)</sup>

List of objects depended by this object. If ALL objects in the list have
been deleted, this object will be garbage collected. If this object is managed by a controller,
then an entry in this list will point to this controller, with the controller field set to true.
There cannot be more than one managing controller.

***

### .spec.deployment.podSecurityContext

Type: `core.PodSecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/pod/resources/security.go#L35)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.deployment.port

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/analytics/v1alpha1/gae_spec_deployment.go#L50)</sup>

Port defines on which port the container will be listening for connections

***

### .spec.deployment.ports

Type: `[]core.ContainerPort` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/networking.go#L39)</sup>

Ports contains list of ports to expose from the container. Not specifying a port here
DOES NOT prevent that port from being exposed. Any port which is
listening on the default "0.0.0.0" address inside a container will be
accessible from the network.

***

### .spec.deployment.readinessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/probes.go#L42)</sup>

ReadinessProbe keeps configuration of periodic probe of container service readiness.
Container will be removed from service endpoints if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.deployment.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/resources.go#L37)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core)

***

### .spec.deployment.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/pod/resources/scheduling.go#L54)</sup>

SchedulerName specifies, the pod will be dispatched by specified scheduler.
If not specified, the pod will be dispatched by default scheduler.

Default Value: `""`

***

### .spec.deployment.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/security.go#L35)</sup>

SecurityContext holds container-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.deployment.service.type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/analytics/v1alpha1/gae_spec_deployment_service.go#L38)</sup>

Type determines how the Service is exposed

Links:
* [Kubernetes Documentation](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types)

Possible Values: 
* `"ClusterIP"` (default) - service will only be accessible inside the cluster, via the cluster IP
* `"NodePort"` - service will be exposed on one port of every node, in addition to 'ClusterIP' type
* `"LoadBalancer"` - service will be exposed via an external load balancer (if the cloud provider supports it), in addition to 'NodePort' type
* `"ExternalName"` - service consists of only a reference to an external name that kubedns or equivalent will return as a CNAME record, with no exposing or proxying of any pods involved
* `"None"` - service is not created

***

### .spec.deployment.serviceAccountName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/pod/resources/service_account.go#L35)</sup>

ServiceAccountName is the name of the ServiceAccount to use to run this pod.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/)

***

### .spec.deployment.shareProcessNamespace

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/pod/resources/namespace.go#L48)</sup>

ShareProcessNamespace defines to share a single process namespace between all of the containers in a pod.
When this is set containers will be able to view and signal processes from other containers
in the same pod, and the first process in each container will not be assigned PID 1.
HostPID and ShareProcessNamespace cannot both be set.

Default Value: `false`

***

### .spec.deployment.startupProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/probes.go#L50)</sup>

StartupProbe indicates that the Pod has successfully initialized.
If specified, no other probes are executed until this completes successfully.
If this probe fails, the Pod will be restarted, just as if the livenessProbe failed.
This can be used to provide different probe parameters at the beginning of a Pod's lifecycle,
when it might take a long time to load data or warm a cache, than during steady-state operation.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.deployment.tls.altNames

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/tls.go#L28)</sup>

AltNames define TLS AltNames used when TLS on the ArangoDB is enabled

***

### .spec.deployment.tls.enabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/tls.go#L25)</sup>

Enabled define if TLS Should be enabled. If is not set then default is taken from ArangoDeployment settings

***

### .spec.deployment.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/pod/resources/scheduling.go#L49)</sup>

Tolerations defines tolerations

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)

***

### .spec.deployment.volumeMounts

Type: `[]core.VolumeMount` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/volume_mounts.go#L35)</sup>

VolumeMounts keeps list of pod volumes to mount into the container's filesystem.

***

### .spec.deployment.volumes

Type: `[]core.Volume` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/pod/resources/volumes.go#L36)</sup>

Volumes keeps list of volumes that can be mounted by containers belonging to the pod.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/storage/volumes)

***

### .spec.deployment.workingDir

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/core.go#L55)</sup>

Container's working directory.
If not specified, the container runtime's default will be used, which
might be configured in the container image.

***

### .spec.deploymentName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/analytics/v1alpha1/gae_spec.go#L31)</sup>

DeploymentName define deployment name used in the object. Immutable

***

### .spec.integrationSidecar.args

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/core.go#L50)</sup>

Arguments to the entrypoint.
The container image's CMD is used if this is not provided.
Variable references $(VAR_NAME) are expanded using the container's environment. If a variable
cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced
to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will
produce the string literal "$(VAR_NAME)". Escaped references will never be expanded, regardless
of whether the variable exists or not. Cannot be updated.

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell)

***

### .spec.integrationSidecar.command

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/core.go#L40)</sup>

Entrypoint array. Not executed within a shell.
The container image's ENTRYPOINT is used if this is not provided.
Variable references $(VAR_NAME) are expanded using the container's environment. If a variable
cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced
to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will
produce the string literal "$(VAR_NAME)". Escaped references will never be expanded, regardless
of whether the variable exists or not. Cannot be updated.

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell)

***

### .spec.integrationSidecar.controllerListenPort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/integration/integration.go#L16)</sup>

ControllerListenPort defines on which port the sidecar container will be listening for controller requests

Default Value: `9202`

***

### .spec.integrationSidecar.env

Type: `core.EnvVar` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/environments.go#L36)</sup>

Env keeps the information about environment variables provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core)

***

### .spec.integrationSidecar.envFrom

Type: `core.EnvFromSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/environments.go#L41)</sup>

EnvFrom keeps the information about environment variable sources provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envfromsource-v1-core)

***

### .spec.integrationSidecar.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/image.go#L35)</sup>

Image define image details

***

### .spec.integrationSidecar.imagePullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/image.go#L39)</sup>

ImagePullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.integrationSidecar.lifecycle

Type: `core.Lifecycle` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/lifecycle.go#L35)</sup>

Lifecycle keeps actions that the management system should take in response to container lifecycle events.

***

### .spec.integrationSidecar.listenPort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/integration/integration.go#L12)</sup>

ListenPort defines on which port the sidecar container will be listening for connections

Default Value: `9201`

***

### .spec.integrationSidecar.livenessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/probes.go#L37)</sup>

LivenessProbe keeps configuration of periodic probe of container liveness.
Container will be restarted if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.integrationSidecar.ports

Type: `[]core.ContainerPort` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/networking.go#L39)</sup>

Ports contains list of ports to expose from the container. Not specifying a port here
DOES NOT prevent that port from being exposed. Any port which is
listening on the default "0.0.0.0" address inside a container will be
accessible from the network.

***

### .spec.integrationSidecar.readinessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/probes.go#L42)</sup>

ReadinessProbe keeps configuration of periodic probe of container service readiness.
Container will be removed from service endpoints if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.integrationSidecar.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/resources.go#L37)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core)

***

### .spec.integrationSidecar.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/security.go#L35)</sup>

SecurityContext holds container-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.integrationSidecar.startupProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/probes.go#L50)</sup>

StartupProbe indicates that the Pod has successfully initialized.
If specified, no other probes are executed until this completes successfully.
If this probe fails, the Pod will be restarted, just as if the livenessProbe failed.
This can be used to provide different probe parameters at the beginning of a Pod's lifecycle,
when it might take a long time to load data or warm a cache, than during steady-state operation.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.integrationSidecar.volumeMounts

Type: `[]core.VolumeMount` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/volume_mounts.go#L35)</sup>

VolumeMounts keeps list of pod volumes to mount into the container's filesystem.

***

### .spec.integrationSidecar.workingDir

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/scheduler/v1beta1/container/resources/core.go#L55)</sup>

Container's working directory.
If not specified, the container runtime's default will be used, which
might be configured in the container image.

## Status

### .status.arangoDB.deployment.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .status.arangoDB.deployment.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .status.arangoDB.deployment.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.arangoDB.deployment.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

***

### .status.arangoDB.secret.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .status.arangoDB.secret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .status.arangoDB.secret.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.arangoDB.secret.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

***

### .status.arangoDB.tls.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .status.arangoDB.tls.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .status.arangoDB.tls.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.arangoDB.tls.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

***

### .status.conditions

Type: `api.Conditions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/analytics/v1alpha1/gae_status.go#L30)</sup>

Conditions specific to the entire extension

***

### .status.reconciliation.service.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .status.reconciliation.service.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .status.reconciliation.service.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.reconciliation.service.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

***

### .status.reconciliation.statefulSet.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .status.reconciliation.statefulSet.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .status.reconciliation.statefulSet.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.reconciliation.statefulSet.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

