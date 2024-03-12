---
layout: page
parent: CRD reference
title: ArangoMLExtension V1Alpha1
---

# API Reference for ArangoMLExtension V1Alpha1

## Spec

### .spec.deployment.affinity

Type: `core.Affinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L44)</sup>

Affinity defines scheduling constraints for workload

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity)

***

### .spec.deployment.hostIPC

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L42)</sup>

HostIPC defines to use the host's ipc namespace.

Default Value: `false`

***

### .spec.deployment.hostNetwork

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L36)</sup>

HostNetwork requests Host network for this pod. Use the host's network namespace.
If this option is set, the ports that will be used must be specified.

Default Value: `false`

***

### .spec.deployment.hostPID

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L39)</sup>

HostPID define to use the host's pid namespace.

Default Value: `false`

***

### .spec.deployment.nodeSelector

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L39)</sup>

NodeSelector is a selector that must be true for the workload to fit on a node.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector)

***

### .spec.deployment.podSecurityContext

Type: `core.PodSecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/security.go#L35)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.deployment.prediction.args

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L50)</sup>

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

### .spec.deployment.prediction.command

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L40)</sup>

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

### .spec.deployment.prediction.env

Type: `core.EnvVar` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L36)</sup>

Env keeps the information about environment variables provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core)

***

### .spec.deployment.prediction.envFrom

Type: `core.EnvFromSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L41)</sup>

EnvFrom keeps the information about environment variable sources provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envfromsource-v1-core)

***

### .spec.deployment.prediction.gpu

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/ml/v1alpha1/extension_spec_deployment_component.go#L31)</sup>

GPU defined if GPU Jobs are enabled for component. In use only for ArangoMLExtensionSpecDeploymentComponentPrediction and ArangoMLExtensionSpecDeploymentComponentTraining

Default Value: `false`

***

### .spec.deployment.prediction.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L37)</sup>

Image define image details

***

### .spec.deployment.prediction.imagePullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L41)</sup>

ImagePullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.deployment.prediction.imagePullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L44)</sup>

ImagePullSecrets define Secrets used to pull Image from registry

***

### .spec.deployment.prediction.lifecycle

Type: `core.Lifecycle` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/lifecycle.go#L35)</sup>

Lifecycle keeps actions that the management system should take in response to container lifecycle events.

***

### .spec.deployment.prediction.livenessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L37)</sup>

LivenessProbe keeps configuration of periodic probe of container liveness.
Container will be restarted if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.deployment.prediction.port

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/ml/v1alpha1/extension_spec_deployment_component.go#L34)</sup>

Port defines on which port the container will be listening for connections

***

### .spec.deployment.prediction.ports

Type: `[]core.ContainerPort` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/networking.go#L39)</sup>

Ports contains list of ports to expose from the container. Not specifying a port here
DOES NOT prevent that port from being exposed. Any port which is
listening on the default "0.0.0.0" address inside a container will be
accessible from the network.

***

### .spec.deployment.prediction.readinessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L42)</sup>

ReadinessProbe keeps configuration of periodic probe of container service readiness.
Container will be removed from service endpoints if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.deployment.prediction.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/resources.go#L37)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core)

***

### .spec.deployment.prediction.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/security.go#L35)</sup>

SecurityContext holds container-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.deployment.prediction.startupProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L50)</sup>

StartupProbe indicates that the Pod has successfully initialized.
If specified, no other probes are executed until this completes successfully.
If this probe fails, the Pod will be restarted, just as if the livenessProbe failed.
This can be used to provide different probe parameters at the beginning of a Pod's lifecycle,
when it might take a long time to load data or warm a cache, than during steady-state operation.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.deployment.prediction.volumeMounts

Type: `[]core.VolumeMount` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/volume_mounts.go#L35)</sup>

VolumeMounts keeps list of pod volumes to mount into the container's filesystem.

***

### .spec.deployment.prediction.workingDir

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L55)</sup>

Container's working directory.
If not specified, the container runtime's default will be used, which
might be configured in the container image.

***

### .spec.deployment.project.args

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L50)</sup>

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

### .spec.deployment.project.command

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L40)</sup>

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

### .spec.deployment.project.env

Type: `core.EnvVar` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L36)</sup>

Env keeps the information about environment variables provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core)

***

### .spec.deployment.project.envFrom

Type: `core.EnvFromSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L41)</sup>

EnvFrom keeps the information about environment variable sources provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envfromsource-v1-core)

***

### .spec.deployment.project.gpu

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/ml/v1alpha1/extension_spec_deployment_component.go#L31)</sup>

GPU defined if GPU Jobs are enabled for component. In use only for ArangoMLExtensionSpecDeploymentComponentPrediction and ArangoMLExtensionSpecDeploymentComponentTraining

Default Value: `false`

***

### .spec.deployment.project.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L37)</sup>

Image define image details

***

### .spec.deployment.project.imagePullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L41)</sup>

ImagePullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.deployment.project.imagePullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L44)</sup>

ImagePullSecrets define Secrets used to pull Image from registry

***

### .spec.deployment.project.lifecycle

Type: `core.Lifecycle` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/lifecycle.go#L35)</sup>

Lifecycle keeps actions that the management system should take in response to container lifecycle events.

***

### .spec.deployment.project.livenessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L37)</sup>

LivenessProbe keeps configuration of periodic probe of container liveness.
Container will be restarted if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.deployment.project.port

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/ml/v1alpha1/extension_spec_deployment_component.go#L34)</sup>

Port defines on which port the container will be listening for connections

***

### .spec.deployment.project.ports

Type: `[]core.ContainerPort` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/networking.go#L39)</sup>

Ports contains list of ports to expose from the container. Not specifying a port here
DOES NOT prevent that port from being exposed. Any port which is
listening on the default "0.0.0.0" address inside a container will be
accessible from the network.

***

### .spec.deployment.project.readinessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L42)</sup>

ReadinessProbe keeps configuration of periodic probe of container service readiness.
Container will be removed from service endpoints if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.deployment.project.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/resources.go#L37)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core)

***

### .spec.deployment.project.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/security.go#L35)</sup>

SecurityContext holds container-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.deployment.project.startupProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L50)</sup>

StartupProbe indicates that the Pod has successfully initialized.
If specified, no other probes are executed until this completes successfully.
If this probe fails, the Pod will be restarted, just as if the livenessProbe failed.
This can be used to provide different probe parameters at the beginning of a Pod's lifecycle,
when it might take a long time to load data or warm a cache, than during steady-state operation.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.deployment.project.volumeMounts

Type: `[]core.VolumeMount` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/volume_mounts.go#L35)</sup>

VolumeMounts keeps list of pod volumes to mount into the container's filesystem.

***

### .spec.deployment.project.workingDir

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L55)</sup>

Container's working directory.
If not specified, the container runtime's default will be used, which
might be configured in the container image.

***

### .spec.deployment.replicas

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/ml/v1alpha1/extension_spec_deployment.go#L56)</sup>

Replicas defines the number of replicas running specified components. No replicas created if no components are defined.

Default Value: `1`

***

### .spec.deployment.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L54)</sup>

SchedulerName specifies, the pod will be dispatched by specified scheduler.
If not specified, the pod will be dispatched by default scheduler.

Default Value: `""`

***

### .spec.deployment.service.type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/ml/v1alpha1/extension_spec_deployment_service.go#L37)</sup>

Type determines how the Service is exposed

Links:
* [Kubernetes Documentation](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types)

Possible Values: 
* `"ClusterIP"` (default) - service will only be accessible inside the cluster, via the cluster IP
* `"NodePort"` - service will be exposed on one port of every node, in addition to 'ClusterIP' type
* `"LoadBalancer"` - service will be exposed via an external load balancer (if the cloud provider supports it), in addition to 'NodePort' type
* `"ExternalName"` - service consists of only a reference to an external name that kubedns or equivalent will return as a CNAME record, with no exposing or proxying of any pods involved

***

### .spec.deployment.shareProcessNamespace

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L48)</sup>

ShareProcessNamespace defines to share a single process namespace between all of the containers in a pod.
When this is set containers will be able to view and signal processes from other containers
in the same pod, and the first process in each container will not be assigned PID 1.
HostPID and ShareProcessNamespace cannot both be set.

Default Value: `false`

***

### .spec.deployment.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L49)</sup>

Tolerations defines tolerations

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)

***

### .spec.deployment.training.args

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L50)</sup>

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

### .spec.deployment.training.command

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L40)</sup>

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

### .spec.deployment.training.env

Type: `core.EnvVar` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L36)</sup>

Env keeps the information about environment variables provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core)

***

### .spec.deployment.training.envFrom

Type: `core.EnvFromSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L41)</sup>

EnvFrom keeps the information about environment variable sources provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envfromsource-v1-core)

***

### .spec.deployment.training.gpu

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/ml/v1alpha1/extension_spec_deployment_component.go#L31)</sup>

GPU defined if GPU Jobs are enabled for component. In use only for ArangoMLExtensionSpecDeploymentComponentPrediction and ArangoMLExtensionSpecDeploymentComponentTraining

Default Value: `false`

***

### .spec.deployment.training.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L37)</sup>

Image define image details

***

### .spec.deployment.training.imagePullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L41)</sup>

ImagePullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.deployment.training.imagePullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L44)</sup>

ImagePullSecrets define Secrets used to pull Image from registry

***

### .spec.deployment.training.lifecycle

Type: `core.Lifecycle` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/lifecycle.go#L35)</sup>

Lifecycle keeps actions that the management system should take in response to container lifecycle events.

***

### .spec.deployment.training.livenessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L37)</sup>

LivenessProbe keeps configuration of periodic probe of container liveness.
Container will be restarted if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.deployment.training.port

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/ml/v1alpha1/extension_spec_deployment_component.go#L34)</sup>

Port defines on which port the container will be listening for connections

***

### .spec.deployment.training.ports

Type: `[]core.ContainerPort` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/networking.go#L39)</sup>

Ports contains list of ports to expose from the container. Not specifying a port here
DOES NOT prevent that port from being exposed. Any port which is
listening on the default "0.0.0.0" address inside a container will be
accessible from the network.

***

### .spec.deployment.training.readinessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L42)</sup>

ReadinessProbe keeps configuration of periodic probe of container service readiness.
Container will be removed from service endpoints if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.deployment.training.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/resources.go#L37)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core)

***

### .spec.deployment.training.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/security.go#L35)</sup>

SecurityContext holds container-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.deployment.training.startupProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L50)</sup>

StartupProbe indicates that the Pod has successfully initialized.
If specified, no other probes are executed until this completes successfully.
If this probe fails, the Pod will be restarted, just as if the livenessProbe failed.
This can be used to provide different probe parameters at the beginning of a Pod's lifecycle,
when it might take a long time to load data or warm a cache, than during steady-state operation.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.deployment.training.volumeMounts

Type: `[]core.VolumeMount` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/volume_mounts.go#L35)</sup>

VolumeMounts keeps list of pod volumes to mount into the container's filesystem.

***

### .spec.deployment.training.workingDir

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L55)</sup>

Container's working directory.
If not specified, the container runtime's default will be used, which
might be configured in the container image.

***

### .spec.deployment.volumes

Type: `[]core.Volume` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/volumes.go#L36)</sup>

Volumes keeps list of volumes that can be mounted by containers belonging to the pod.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/storage/volumes)

***

### .spec.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L37)</sup>

Image define image details

***

### .spec.imagePullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L41)</sup>

ImagePullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.imagePullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L44)</sup>

ImagePullSecrets define Secrets used to pull Image from registry

***

### .spec.init.affinity

Type: `core.Affinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L44)</sup>

Affinity defines scheduling constraints for workload

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity)

***

### .spec.init.args

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L50)</sup>

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

### .spec.init.command

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L40)</sup>

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

### .spec.init.env

Type: `core.EnvVar` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L36)</sup>

Env keeps the information about environment variables provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core)

***

### .spec.init.envFrom

Type: `core.EnvFromSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L41)</sup>

EnvFrom keeps the information about environment variable sources provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envfromsource-v1-core)

***

### .spec.init.hostIPC

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L42)</sup>

HostIPC defines to use the host's ipc namespace.

Default Value: `false`

***

### .spec.init.hostNetwork

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L36)</sup>

HostNetwork requests Host network for this pod. Use the host's network namespace.
If this option is set, the ports that will be used must be specified.

Default Value: `false`

***

### .spec.init.hostPID

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L39)</sup>

HostPID define to use the host's pid namespace.

Default Value: `false`

***

### .spec.init.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L37)</sup>

Image define image details

***

### .spec.init.imagePullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L41)</sup>

ImagePullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.init.imagePullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L44)</sup>

ImagePullSecrets define Secrets used to pull Image from registry

***

### .spec.init.lifecycle

Type: `core.Lifecycle` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/lifecycle.go#L35)</sup>

Lifecycle keeps actions that the management system should take in response to container lifecycle events.

***

### .spec.init.livenessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L37)</sup>

LivenessProbe keeps configuration of periodic probe of container liveness.
Container will be restarted if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.init.nodeSelector

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L39)</sup>

NodeSelector is a selector that must be true for the workload to fit on a node.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector)

***

### .spec.init.podSecurityContext

Type: `core.PodSecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/security.go#L35)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.init.ports

Type: `[]core.ContainerPort` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/networking.go#L39)</sup>

Ports contains list of ports to expose from the container. Not specifying a port here
DOES NOT prevent that port from being exposed. Any port which is
listening on the default "0.0.0.0" address inside a container will be
accessible from the network.

***

### .spec.init.readinessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L42)</sup>

ReadinessProbe keeps configuration of periodic probe of container service readiness.
Container will be removed from service endpoints if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.init.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/resources.go#L37)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core)

***

### .spec.init.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L54)</sup>

SchedulerName specifies, the pod will be dispatched by specified scheduler.
If not specified, the pod will be dispatched by default scheduler.

Default Value: `""`

***

### .spec.init.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/security.go#L35)</sup>

SecurityContext holds container-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.init.shareProcessNamespace

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L48)</sup>

ShareProcessNamespace defines to share a single process namespace between all of the containers in a pod.
When this is set containers will be able to view and signal processes from other containers
in the same pod, and the first process in each container will not be assigned PID 1.
HostPID and ShareProcessNamespace cannot both be set.

Default Value: `false`

***

### .spec.init.startupProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L50)</sup>

StartupProbe indicates that the Pod has successfully initialized.
If specified, no other probes are executed until this completes successfully.
If this probe fails, the Pod will be restarted, just as if the livenessProbe failed.
This can be used to provide different probe parameters at the beginning of a Pod's lifecycle,
when it might take a long time to load data or warm a cache, than during steady-state operation.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.init.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L49)</sup>

Tolerations defines tolerations

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)

***

### .spec.init.volumeMounts

Type: `[]core.VolumeMount` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/volume_mounts.go#L35)</sup>

VolumeMounts keeps list of pod volumes to mount into the container's filesystem.

***

### .spec.init.volumes

Type: `[]core.Volume` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/volumes.go#L36)</sup>

Volumes keeps list of volumes that can be mounted by containers belonging to the pod.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/storage/volumes)

***

### .spec.init.workingDir

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L55)</sup>

Container's working directory.
If not specified, the container runtime's default will be used, which
might be configured in the container image.

***

### .spec.jobsTemplates.featurization.cpu.affinity

Type: `core.Affinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L44)</sup>

Affinity defines scheduling constraints for workload

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity)

***

### .spec.jobsTemplates.featurization.cpu.args

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L50)</sup>

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

### .spec.jobsTemplates.featurization.cpu.command

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L40)</sup>

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

### .spec.jobsTemplates.featurization.cpu.env

Type: `core.EnvVar` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L36)</sup>

Env keeps the information about environment variables provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core)

***

### .spec.jobsTemplates.featurization.cpu.envFrom

Type: `core.EnvFromSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L41)</sup>

EnvFrom keeps the information about environment variable sources provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envfromsource-v1-core)

***

### .spec.jobsTemplates.featurization.cpu.hostIPC

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L42)</sup>

HostIPC defines to use the host's ipc namespace.

Default Value: `false`

***

### .spec.jobsTemplates.featurization.cpu.hostNetwork

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L36)</sup>

HostNetwork requests Host network for this pod. Use the host's network namespace.
If this option is set, the ports that will be used must be specified.

Default Value: `false`

***

### .spec.jobsTemplates.featurization.cpu.hostPID

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L39)</sup>

HostPID define to use the host's pid namespace.

Default Value: `false`

***

### .spec.jobsTemplates.featurization.cpu.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L37)</sup>

Image define image details

***

### .spec.jobsTemplates.featurization.cpu.imagePullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L41)</sup>

ImagePullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.jobsTemplates.featurization.cpu.imagePullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L44)</sup>

ImagePullSecrets define Secrets used to pull Image from registry

***

### .spec.jobsTemplates.featurization.cpu.lifecycle

Type: `core.Lifecycle` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/lifecycle.go#L35)</sup>

Lifecycle keeps actions that the management system should take in response to container lifecycle events.

***

### .spec.jobsTemplates.featurization.cpu.livenessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L37)</sup>

LivenessProbe keeps configuration of periodic probe of container liveness.
Container will be restarted if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.jobsTemplates.featurization.cpu.nodeSelector

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L39)</sup>

NodeSelector is a selector that must be true for the workload to fit on a node.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector)

***

### .spec.jobsTemplates.featurization.cpu.podSecurityContext

Type: `core.PodSecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/security.go#L35)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.jobsTemplates.featurization.cpu.ports

Type: `[]core.ContainerPort` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/networking.go#L39)</sup>

Ports contains list of ports to expose from the container. Not specifying a port here
DOES NOT prevent that port from being exposed. Any port which is
listening on the default "0.0.0.0" address inside a container will be
accessible from the network.

***

### .spec.jobsTemplates.featurization.cpu.readinessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L42)</sup>

ReadinessProbe keeps configuration of periodic probe of container service readiness.
Container will be removed from service endpoints if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.jobsTemplates.featurization.cpu.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/resources.go#L37)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core)

***

### .spec.jobsTemplates.featurization.cpu.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L54)</sup>

SchedulerName specifies, the pod will be dispatched by specified scheduler.
If not specified, the pod will be dispatched by default scheduler.

Default Value: `""`

***

### .spec.jobsTemplates.featurization.cpu.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/security.go#L35)</sup>

SecurityContext holds container-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.jobsTemplates.featurization.cpu.shareProcessNamespace

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L48)</sup>

ShareProcessNamespace defines to share a single process namespace between all of the containers in a pod.
When this is set containers will be able to view and signal processes from other containers
in the same pod, and the first process in each container will not be assigned PID 1.
HostPID and ShareProcessNamespace cannot both be set.

Default Value: `false`

***

### .spec.jobsTemplates.featurization.cpu.startupProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L50)</sup>

StartupProbe indicates that the Pod has successfully initialized.
If specified, no other probes are executed until this completes successfully.
If this probe fails, the Pod will be restarted, just as if the livenessProbe failed.
This can be used to provide different probe parameters at the beginning of a Pod's lifecycle,
when it might take a long time to load data or warm a cache, than during steady-state operation.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.jobsTemplates.featurization.cpu.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L49)</sup>

Tolerations defines tolerations

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)

***

### .spec.jobsTemplates.featurization.cpu.volumeMounts

Type: `[]core.VolumeMount` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/volume_mounts.go#L35)</sup>

VolumeMounts keeps list of pod volumes to mount into the container's filesystem.

***

### .spec.jobsTemplates.featurization.cpu.volumes

Type: `[]core.Volume` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/volumes.go#L36)</sup>

Volumes keeps list of volumes that can be mounted by containers belonging to the pod.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/storage/volumes)

***

### .spec.jobsTemplates.featurization.cpu.workingDir

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L55)</sup>

Container's working directory.
If not specified, the container runtime's default will be used, which
might be configured in the container image.

***

### .spec.jobsTemplates.featurization.gpu.affinity

Type: `core.Affinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L44)</sup>

Affinity defines scheduling constraints for workload

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity)

***

### .spec.jobsTemplates.featurization.gpu.args

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L50)</sup>

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

### .spec.jobsTemplates.featurization.gpu.command

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L40)</sup>

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

### .spec.jobsTemplates.featurization.gpu.env

Type: `core.EnvVar` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L36)</sup>

Env keeps the information about environment variables provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core)

***

### .spec.jobsTemplates.featurization.gpu.envFrom

Type: `core.EnvFromSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L41)</sup>

EnvFrom keeps the information about environment variable sources provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envfromsource-v1-core)

***

### .spec.jobsTemplates.featurization.gpu.hostIPC

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L42)</sup>

HostIPC defines to use the host's ipc namespace.

Default Value: `false`

***

### .spec.jobsTemplates.featurization.gpu.hostNetwork

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L36)</sup>

HostNetwork requests Host network for this pod. Use the host's network namespace.
If this option is set, the ports that will be used must be specified.

Default Value: `false`

***

### .spec.jobsTemplates.featurization.gpu.hostPID

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L39)</sup>

HostPID define to use the host's pid namespace.

Default Value: `false`

***

### .spec.jobsTemplates.featurization.gpu.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L37)</sup>

Image define image details

***

### .spec.jobsTemplates.featurization.gpu.imagePullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L41)</sup>

ImagePullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.jobsTemplates.featurization.gpu.imagePullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L44)</sup>

ImagePullSecrets define Secrets used to pull Image from registry

***

### .spec.jobsTemplates.featurization.gpu.lifecycle

Type: `core.Lifecycle` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/lifecycle.go#L35)</sup>

Lifecycle keeps actions that the management system should take in response to container lifecycle events.

***

### .spec.jobsTemplates.featurization.gpu.livenessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L37)</sup>

LivenessProbe keeps configuration of periodic probe of container liveness.
Container will be restarted if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.jobsTemplates.featurization.gpu.nodeSelector

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L39)</sup>

NodeSelector is a selector that must be true for the workload to fit on a node.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector)

***

### .spec.jobsTemplates.featurization.gpu.podSecurityContext

Type: `core.PodSecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/security.go#L35)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.jobsTemplates.featurization.gpu.ports

Type: `[]core.ContainerPort` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/networking.go#L39)</sup>

Ports contains list of ports to expose from the container. Not specifying a port here
DOES NOT prevent that port from being exposed. Any port which is
listening on the default "0.0.0.0" address inside a container will be
accessible from the network.

***

### .spec.jobsTemplates.featurization.gpu.readinessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L42)</sup>

ReadinessProbe keeps configuration of periodic probe of container service readiness.
Container will be removed from service endpoints if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.jobsTemplates.featurization.gpu.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/resources.go#L37)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core)

***

### .spec.jobsTemplates.featurization.gpu.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L54)</sup>

SchedulerName specifies, the pod will be dispatched by specified scheduler.
If not specified, the pod will be dispatched by default scheduler.

Default Value: `""`

***

### .spec.jobsTemplates.featurization.gpu.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/security.go#L35)</sup>

SecurityContext holds container-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.jobsTemplates.featurization.gpu.shareProcessNamespace

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L48)</sup>

ShareProcessNamespace defines to share a single process namespace between all of the containers in a pod.
When this is set containers will be able to view and signal processes from other containers
in the same pod, and the first process in each container will not be assigned PID 1.
HostPID and ShareProcessNamespace cannot both be set.

Default Value: `false`

***

### .spec.jobsTemplates.featurization.gpu.startupProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L50)</sup>

StartupProbe indicates that the Pod has successfully initialized.
If specified, no other probes are executed until this completes successfully.
If this probe fails, the Pod will be restarted, just as if the livenessProbe failed.
This can be used to provide different probe parameters at the beginning of a Pod's lifecycle,
when it might take a long time to load data or warm a cache, than during steady-state operation.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.jobsTemplates.featurization.gpu.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L49)</sup>

Tolerations defines tolerations

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)

***

### .spec.jobsTemplates.featurization.gpu.volumeMounts

Type: `[]core.VolumeMount` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/volume_mounts.go#L35)</sup>

VolumeMounts keeps list of pod volumes to mount into the container's filesystem.

***

### .spec.jobsTemplates.featurization.gpu.volumes

Type: `[]core.Volume` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/volumes.go#L36)</sup>

Volumes keeps list of volumes that can be mounted by containers belonging to the pod.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/storage/volumes)

***

### .spec.jobsTemplates.featurization.gpu.workingDir

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L55)</sup>

Container's working directory.
If not specified, the container runtime's default will be used, which
might be configured in the container image.

***

### .spec.jobsTemplates.prediction.cpu.affinity

Type: `core.Affinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L44)</sup>

Affinity defines scheduling constraints for workload

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity)

***

### .spec.jobsTemplates.prediction.cpu.args

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L50)</sup>

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

### .spec.jobsTemplates.prediction.cpu.command

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L40)</sup>

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

### .spec.jobsTemplates.prediction.cpu.env

Type: `core.EnvVar` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L36)</sup>

Env keeps the information about environment variables provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core)

***

### .spec.jobsTemplates.prediction.cpu.envFrom

Type: `core.EnvFromSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L41)</sup>

EnvFrom keeps the information about environment variable sources provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envfromsource-v1-core)

***

### .spec.jobsTemplates.prediction.cpu.hostIPC

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L42)</sup>

HostIPC defines to use the host's ipc namespace.

Default Value: `false`

***

### .spec.jobsTemplates.prediction.cpu.hostNetwork

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L36)</sup>

HostNetwork requests Host network for this pod. Use the host's network namespace.
If this option is set, the ports that will be used must be specified.

Default Value: `false`

***

### .spec.jobsTemplates.prediction.cpu.hostPID

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L39)</sup>

HostPID define to use the host's pid namespace.

Default Value: `false`

***

### .spec.jobsTemplates.prediction.cpu.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L37)</sup>

Image define image details

***

### .spec.jobsTemplates.prediction.cpu.imagePullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L41)</sup>

ImagePullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.jobsTemplates.prediction.cpu.imagePullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L44)</sup>

ImagePullSecrets define Secrets used to pull Image from registry

***

### .spec.jobsTemplates.prediction.cpu.lifecycle

Type: `core.Lifecycle` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/lifecycle.go#L35)</sup>

Lifecycle keeps actions that the management system should take in response to container lifecycle events.

***

### .spec.jobsTemplates.prediction.cpu.livenessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L37)</sup>

LivenessProbe keeps configuration of periodic probe of container liveness.
Container will be restarted if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.jobsTemplates.prediction.cpu.nodeSelector

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L39)</sup>

NodeSelector is a selector that must be true for the workload to fit on a node.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector)

***

### .spec.jobsTemplates.prediction.cpu.podSecurityContext

Type: `core.PodSecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/security.go#L35)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.jobsTemplates.prediction.cpu.ports

Type: `[]core.ContainerPort` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/networking.go#L39)</sup>

Ports contains list of ports to expose from the container. Not specifying a port here
DOES NOT prevent that port from being exposed. Any port which is
listening on the default "0.0.0.0" address inside a container will be
accessible from the network.

***

### .spec.jobsTemplates.prediction.cpu.readinessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L42)</sup>

ReadinessProbe keeps configuration of periodic probe of container service readiness.
Container will be removed from service endpoints if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.jobsTemplates.prediction.cpu.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/resources.go#L37)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core)

***

### .spec.jobsTemplates.prediction.cpu.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L54)</sup>

SchedulerName specifies, the pod will be dispatched by specified scheduler.
If not specified, the pod will be dispatched by default scheduler.

Default Value: `""`

***

### .spec.jobsTemplates.prediction.cpu.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/security.go#L35)</sup>

SecurityContext holds container-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.jobsTemplates.prediction.cpu.shareProcessNamespace

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L48)</sup>

ShareProcessNamespace defines to share a single process namespace between all of the containers in a pod.
When this is set containers will be able to view and signal processes from other containers
in the same pod, and the first process in each container will not be assigned PID 1.
HostPID and ShareProcessNamespace cannot both be set.

Default Value: `false`

***

### .spec.jobsTemplates.prediction.cpu.startupProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L50)</sup>

StartupProbe indicates that the Pod has successfully initialized.
If specified, no other probes are executed until this completes successfully.
If this probe fails, the Pod will be restarted, just as if the livenessProbe failed.
This can be used to provide different probe parameters at the beginning of a Pod's lifecycle,
when it might take a long time to load data or warm a cache, than during steady-state operation.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.jobsTemplates.prediction.cpu.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L49)</sup>

Tolerations defines tolerations

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)

***

### .spec.jobsTemplates.prediction.cpu.volumeMounts

Type: `[]core.VolumeMount` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/volume_mounts.go#L35)</sup>

VolumeMounts keeps list of pod volumes to mount into the container's filesystem.

***

### .spec.jobsTemplates.prediction.cpu.volumes

Type: `[]core.Volume` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/volumes.go#L36)</sup>

Volumes keeps list of volumes that can be mounted by containers belonging to the pod.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/storage/volumes)

***

### .spec.jobsTemplates.prediction.cpu.workingDir

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L55)</sup>

Container's working directory.
If not specified, the container runtime's default will be used, which
might be configured in the container image.

***

### .spec.jobsTemplates.prediction.gpu.affinity

Type: `core.Affinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L44)</sup>

Affinity defines scheduling constraints for workload

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity)

***

### .spec.jobsTemplates.prediction.gpu.args

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L50)</sup>

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

### .spec.jobsTemplates.prediction.gpu.command

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L40)</sup>

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

### .spec.jobsTemplates.prediction.gpu.env

Type: `core.EnvVar` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L36)</sup>

Env keeps the information about environment variables provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core)

***

### .spec.jobsTemplates.prediction.gpu.envFrom

Type: `core.EnvFromSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L41)</sup>

EnvFrom keeps the information about environment variable sources provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envfromsource-v1-core)

***

### .spec.jobsTemplates.prediction.gpu.hostIPC

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L42)</sup>

HostIPC defines to use the host's ipc namespace.

Default Value: `false`

***

### .spec.jobsTemplates.prediction.gpu.hostNetwork

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L36)</sup>

HostNetwork requests Host network for this pod. Use the host's network namespace.
If this option is set, the ports that will be used must be specified.

Default Value: `false`

***

### .spec.jobsTemplates.prediction.gpu.hostPID

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L39)</sup>

HostPID define to use the host's pid namespace.

Default Value: `false`

***

### .spec.jobsTemplates.prediction.gpu.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L37)</sup>

Image define image details

***

### .spec.jobsTemplates.prediction.gpu.imagePullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L41)</sup>

ImagePullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.jobsTemplates.prediction.gpu.imagePullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L44)</sup>

ImagePullSecrets define Secrets used to pull Image from registry

***

### .spec.jobsTemplates.prediction.gpu.lifecycle

Type: `core.Lifecycle` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/lifecycle.go#L35)</sup>

Lifecycle keeps actions that the management system should take in response to container lifecycle events.

***

### .spec.jobsTemplates.prediction.gpu.livenessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L37)</sup>

LivenessProbe keeps configuration of periodic probe of container liveness.
Container will be restarted if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.jobsTemplates.prediction.gpu.nodeSelector

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L39)</sup>

NodeSelector is a selector that must be true for the workload to fit on a node.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector)

***

### .spec.jobsTemplates.prediction.gpu.podSecurityContext

Type: `core.PodSecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/security.go#L35)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.jobsTemplates.prediction.gpu.ports

Type: `[]core.ContainerPort` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/networking.go#L39)</sup>

Ports contains list of ports to expose from the container. Not specifying a port here
DOES NOT prevent that port from being exposed. Any port which is
listening on the default "0.0.0.0" address inside a container will be
accessible from the network.

***

### .spec.jobsTemplates.prediction.gpu.readinessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L42)</sup>

ReadinessProbe keeps configuration of periodic probe of container service readiness.
Container will be removed from service endpoints if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.jobsTemplates.prediction.gpu.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/resources.go#L37)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core)

***

### .spec.jobsTemplates.prediction.gpu.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L54)</sup>

SchedulerName specifies, the pod will be dispatched by specified scheduler.
If not specified, the pod will be dispatched by default scheduler.

Default Value: `""`

***

### .spec.jobsTemplates.prediction.gpu.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/security.go#L35)</sup>

SecurityContext holds container-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.jobsTemplates.prediction.gpu.shareProcessNamespace

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L48)</sup>

ShareProcessNamespace defines to share a single process namespace between all of the containers in a pod.
When this is set containers will be able to view and signal processes from other containers
in the same pod, and the first process in each container will not be assigned PID 1.
HostPID and ShareProcessNamespace cannot both be set.

Default Value: `false`

***

### .spec.jobsTemplates.prediction.gpu.startupProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L50)</sup>

StartupProbe indicates that the Pod has successfully initialized.
If specified, no other probes are executed until this completes successfully.
If this probe fails, the Pod will be restarted, just as if the livenessProbe failed.
This can be used to provide different probe parameters at the beginning of a Pod's lifecycle,
when it might take a long time to load data or warm a cache, than during steady-state operation.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.jobsTemplates.prediction.gpu.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L49)</sup>

Tolerations defines tolerations

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)

***

### .spec.jobsTemplates.prediction.gpu.volumeMounts

Type: `[]core.VolumeMount` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/volume_mounts.go#L35)</sup>

VolumeMounts keeps list of pod volumes to mount into the container's filesystem.

***

### .spec.jobsTemplates.prediction.gpu.volumes

Type: `[]core.Volume` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/volumes.go#L36)</sup>

Volumes keeps list of volumes that can be mounted by containers belonging to the pod.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/storage/volumes)

***

### .spec.jobsTemplates.prediction.gpu.workingDir

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L55)</sup>

Container's working directory.
If not specified, the container runtime's default will be used, which
might be configured in the container image.

***

### .spec.jobsTemplates.training.cpu.affinity

Type: `core.Affinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L44)</sup>

Affinity defines scheduling constraints for workload

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity)

***

### .spec.jobsTemplates.training.cpu.args

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L50)</sup>

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

### .spec.jobsTemplates.training.cpu.command

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L40)</sup>

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

### .spec.jobsTemplates.training.cpu.env

Type: `core.EnvVar` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L36)</sup>

Env keeps the information about environment variables provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core)

***

### .spec.jobsTemplates.training.cpu.envFrom

Type: `core.EnvFromSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L41)</sup>

EnvFrom keeps the information about environment variable sources provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envfromsource-v1-core)

***

### .spec.jobsTemplates.training.cpu.hostIPC

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L42)</sup>

HostIPC defines to use the host's ipc namespace.

Default Value: `false`

***

### .spec.jobsTemplates.training.cpu.hostNetwork

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L36)</sup>

HostNetwork requests Host network for this pod. Use the host's network namespace.
If this option is set, the ports that will be used must be specified.

Default Value: `false`

***

### .spec.jobsTemplates.training.cpu.hostPID

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L39)</sup>

HostPID define to use the host's pid namespace.

Default Value: `false`

***

### .spec.jobsTemplates.training.cpu.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L37)</sup>

Image define image details

***

### .spec.jobsTemplates.training.cpu.imagePullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L41)</sup>

ImagePullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.jobsTemplates.training.cpu.imagePullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L44)</sup>

ImagePullSecrets define Secrets used to pull Image from registry

***

### .spec.jobsTemplates.training.cpu.lifecycle

Type: `core.Lifecycle` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/lifecycle.go#L35)</sup>

Lifecycle keeps actions that the management system should take in response to container lifecycle events.

***

### .spec.jobsTemplates.training.cpu.livenessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L37)</sup>

LivenessProbe keeps configuration of periodic probe of container liveness.
Container will be restarted if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.jobsTemplates.training.cpu.nodeSelector

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L39)</sup>

NodeSelector is a selector that must be true for the workload to fit on a node.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector)

***

### .spec.jobsTemplates.training.cpu.podSecurityContext

Type: `core.PodSecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/security.go#L35)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.jobsTemplates.training.cpu.ports

Type: `[]core.ContainerPort` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/networking.go#L39)</sup>

Ports contains list of ports to expose from the container. Not specifying a port here
DOES NOT prevent that port from being exposed. Any port which is
listening on the default "0.0.0.0" address inside a container will be
accessible from the network.

***

### .spec.jobsTemplates.training.cpu.readinessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L42)</sup>

ReadinessProbe keeps configuration of periodic probe of container service readiness.
Container will be removed from service endpoints if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.jobsTemplates.training.cpu.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/resources.go#L37)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core)

***

### .spec.jobsTemplates.training.cpu.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L54)</sup>

SchedulerName specifies, the pod will be dispatched by specified scheduler.
If not specified, the pod will be dispatched by default scheduler.

Default Value: `""`

***

### .spec.jobsTemplates.training.cpu.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/security.go#L35)</sup>

SecurityContext holds container-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.jobsTemplates.training.cpu.shareProcessNamespace

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L48)</sup>

ShareProcessNamespace defines to share a single process namespace between all of the containers in a pod.
When this is set containers will be able to view and signal processes from other containers
in the same pod, and the first process in each container will not be assigned PID 1.
HostPID and ShareProcessNamespace cannot both be set.

Default Value: `false`

***

### .spec.jobsTemplates.training.cpu.startupProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L50)</sup>

StartupProbe indicates that the Pod has successfully initialized.
If specified, no other probes are executed until this completes successfully.
If this probe fails, the Pod will be restarted, just as if the livenessProbe failed.
This can be used to provide different probe parameters at the beginning of a Pod's lifecycle,
when it might take a long time to load data or warm a cache, than during steady-state operation.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.jobsTemplates.training.cpu.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L49)</sup>

Tolerations defines tolerations

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)

***

### .spec.jobsTemplates.training.cpu.volumeMounts

Type: `[]core.VolumeMount` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/volume_mounts.go#L35)</sup>

VolumeMounts keeps list of pod volumes to mount into the container's filesystem.

***

### .spec.jobsTemplates.training.cpu.volumes

Type: `[]core.Volume` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/volumes.go#L36)</sup>

Volumes keeps list of volumes that can be mounted by containers belonging to the pod.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/storage/volumes)

***

### .spec.jobsTemplates.training.cpu.workingDir

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L55)</sup>

Container's working directory.
If not specified, the container runtime's default will be used, which
might be configured in the container image.

***

### .spec.jobsTemplates.training.gpu.affinity

Type: `core.Affinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L44)</sup>

Affinity defines scheduling constraints for workload

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity)

***

### .spec.jobsTemplates.training.gpu.args

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L50)</sup>

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

### .spec.jobsTemplates.training.gpu.command

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L40)</sup>

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

### .spec.jobsTemplates.training.gpu.env

Type: `core.EnvVar` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L36)</sup>

Env keeps the information about environment variables provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core)

***

### .spec.jobsTemplates.training.gpu.envFrom

Type: `core.EnvFromSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/environments.go#L41)</sup>

EnvFrom keeps the information about environment variable sources provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envfromsource-v1-core)

***

### .spec.jobsTemplates.training.gpu.hostIPC

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L42)</sup>

HostIPC defines to use the host's ipc namespace.

Default Value: `false`

***

### .spec.jobsTemplates.training.gpu.hostNetwork

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L36)</sup>

HostNetwork requests Host network for this pod. Use the host's network namespace.
If this option is set, the ports that will be used must be specified.

Default Value: `false`

***

### .spec.jobsTemplates.training.gpu.hostPID

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L39)</sup>

HostPID define to use the host's pid namespace.

Default Value: `false`

***

### .spec.jobsTemplates.training.gpu.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L37)</sup>

Image define image details

***

### .spec.jobsTemplates.training.gpu.imagePullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L41)</sup>

ImagePullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.jobsTemplates.training.gpu.imagePullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/image.go#L44)</sup>

ImagePullSecrets define Secrets used to pull Image from registry

***

### .spec.jobsTemplates.training.gpu.lifecycle

Type: `core.Lifecycle` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/lifecycle.go#L35)</sup>

Lifecycle keeps actions that the management system should take in response to container lifecycle events.

***

### .spec.jobsTemplates.training.gpu.livenessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L37)</sup>

LivenessProbe keeps configuration of periodic probe of container liveness.
Container will be restarted if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.jobsTemplates.training.gpu.nodeSelector

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L39)</sup>

NodeSelector is a selector that must be true for the workload to fit on a node.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector)

***

### .spec.jobsTemplates.training.gpu.podSecurityContext

Type: `core.PodSecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/security.go#L35)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.jobsTemplates.training.gpu.ports

Type: `[]core.ContainerPort` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/networking.go#L39)</sup>

Ports contains list of ports to expose from the container. Not specifying a port here
DOES NOT prevent that port from being exposed. Any port which is
listening on the default "0.0.0.0" address inside a container will be
accessible from the network.

***

### .spec.jobsTemplates.training.gpu.readinessProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L42)</sup>

ReadinessProbe keeps configuration of periodic probe of container service readiness.
Container will be removed from service endpoints if the probe fails.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.jobsTemplates.training.gpu.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/resources.go#L37)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core)

***

### .spec.jobsTemplates.training.gpu.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L54)</sup>

SchedulerName specifies, the pod will be dispatched by specified scheduler.
If not specified, the pod will be dispatched by default scheduler.

Default Value: `""`

***

### .spec.jobsTemplates.training.gpu.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/security.go#L35)</sup>

SecurityContext holds container-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

***

### .spec.jobsTemplates.training.gpu.shareProcessNamespace

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/namespace.go#L48)</sup>

ShareProcessNamespace defines to share a single process namespace between all of the containers in a pod.
When this is set containers will be able to view and signal processes from other containers
in the same pod, and the first process in each container will not be assigned PID 1.
HostPID and ShareProcessNamespace cannot both be set.

Default Value: `false`

***

### .spec.jobsTemplates.training.gpu.startupProbe

Type: `core.Probe` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/probes.go#L50)</sup>

StartupProbe indicates that the Pod has successfully initialized.
If specified, no other probes are executed until this completes successfully.
If this probe fails, the Pod will be restarted, just as if the livenessProbe failed.
This can be used to provide different probe parameters at the beginning of a Pod's lifecycle,
when it might take a long time to load data or warm a cache, than during steady-state operation.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes)

***

### .spec.jobsTemplates.training.gpu.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/scheduling.go#L49)</sup>

Tolerations defines tolerations

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)

***

### .spec.jobsTemplates.training.gpu.volumeMounts

Type: `[]core.VolumeMount` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/volume_mounts.go#L35)</sup>

VolumeMounts keeps list of pod volumes to mount into the container's filesystem.

***

### .spec.jobsTemplates.training.gpu.volumes

Type: `[]core.Volume` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/pod/resources/volumes.go#L36)</sup>

Volumes keeps list of volumes that can be mounted by containers belonging to the pod.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/concepts/storage/volumes)

***

### .spec.jobsTemplates.training.gpu.workingDir

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/scheduler/v1alpha1/container/resources/core.go#L55)</sup>

Container's working directory.
If not specified, the container runtime's default will be used, which
might be configured in the container image.

***

### .spec.metadataService.local.arangoMLFeatureStore

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/ml/v1alpha1/extension_spec_metadata_service.go#L65)</sup>

ArangoMLFeatureStoreDatabase define Database name to be used as MetadataService Backend in ArangoMLFeatureStoreDatabase

Default Value: `arangomlfeaturestore`

***

### .spec.metadataService.local.arangoPipeDatabase

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/ml/v1alpha1/extension_spec_metadata_service.go#L61)</sup>

ArangoPipeDatabase define Database name to be used as MetadataService Backend in ArangoPipe

Default Value: `arangopipe`

***

### .spec.storage.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .spec.storage.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.storage.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

## Status

### .status.arangoDB.jwtTokenSecret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.arangoDB.jwtTokenSecret.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.arangoDB.jwtTokenSecret.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.arangoDB.secret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.arangoDB.secret.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.arangoDB.secret.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.conditions

Type: `api.Conditions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/ml/v1alpha1/extension_status.go#L31)</sup>

Conditions specific to the entire extension

***

### .status.metadataService.jwtTokenSecret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.metadataService.jwtTokenSecret.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.metadataService.jwtTokenSecret.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.metadataService.local.arangoMLFeatureStore

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/ml/v1alpha1/extension_status_metadata_service.go#L41)</sup>

ArangoMLFeatureStoreDatabase define Database name to be used as MetadataService Backend

***

### .status.metadataService.local.arangoPipe

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/ml/v1alpha1/extension_status_metadata_service.go#L38)</sup>

ArangoPipeDatabase define Database name to be used as MetadataService Backend

***

### .status.metadataService.secret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.metadataService.secret.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.metadataService.secret.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.reconciliation.serviceChecksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/ml/v1alpha1/extension_status_reconciliation.go#L25)</sup>

***

### .status.reconciliation.statefulSetChecksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/ml/v1alpha1/extension_status_reconciliation.go#L24)</sup>

***

### .status.serviceAccount.cluster.binding.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.serviceAccount.cluster.binding.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.serviceAccount.cluster.binding.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.serviceAccount.cluster.role.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.serviceAccount.cluster.role.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.serviceAccount.cluster.role.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.serviceAccount.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.serviceAccount.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.serviceAccount.namespaced.binding.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.serviceAccount.namespaced.binding.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.serviceAccount.namespaced.binding.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.serviceAccount.namespaced.role.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.serviceAccount.namespaced.role.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.serviceAccount.namespaced.role.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.serviceAccount.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.39/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

