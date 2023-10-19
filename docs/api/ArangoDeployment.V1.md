# API Reference for ArangoDeployment V1

## Spec

### .spec.agents.affinity: core.PodAffinity

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.PodAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L134)

### .spec.agents.allowMemberRecreation: bool

AllowMemberRecreation allows to recreate member. Value is used only for Coordinator and DBServer with default to True, for all other groups set to false.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L171)

### .spec.agents.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L98)

### .spec.agents.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L100)

### .spec.agents.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L102)

### .spec.agents.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L130)

### .spec.agents.args: []string

Args holds additional commandline arguments

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L78)

### .spec.agents.count: int

Count holds the requested number of servers

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L72)

### .spec.agents.entrypoint: string

Entrypoint overrides container executable

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L80)

### .spec.agents.envs\[int\].name: string

[Code Reference](/pkg/apis/deployment/v1/server_group_env_var.go#L26)

### .spec.agents.envs\[int\].value: string

[Code Reference](/pkg/apis/deployment/v1/server_group_env_var.go#L27)

### .spec.agents.ephemeralVolumes.apps.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)

### .spec.agents.ephemeralVolumes.temp.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)

### .spec.agents.exporterPort: uint16

ExporterPort define Port used by exporter

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L184)

### .spec.agents.extendedRotationCheck: bool

ExtendedRotationCheck extend checks for rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L157)

### .spec.agents.externalPortEnabled: bool

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L169)

### .spec.agents.indexMethod: string

IndexMethod define group Indexing method

Possible Values: 
* random (default) - Pick random ID for member. Enforced on the Community Operator.
* ordered - Use sequential number as Member ID, starting from 0. Enterprise Operator required.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L177)

### .spec.agents.initContainers.containers: []core.Container

Containers contains list of containers

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_init_containers.go#L91)

### .spec.agents.initContainers.mode: string

Mode keep container replace mode

[Code Reference](/pkg/apis/deployment/v1/server_group_init_containers.go#L94)

### .spec.agents.internalPort: int

InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L165)

### .spec.agents.internalPortProtocol: string

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L167)

### .spec.agents.labels: map[string]string

Labels specified the labels added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L104)

### .spec.agents.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L106)

### .spec.agents.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L108)

### .spec.agents.maxCount: int

MaxCount specifies a upper limit for count

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L76)

### .spec.agents.minCount: int

MinCount specifies a lower limit for count

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L74)

### .spec.agents.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L138)

### .spec.agents.nodeSelector: map[string]string

NodeSelector speficies a set of selectors for nodes

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L114)

### .spec.agents.numactl.args: []string

Args define list of the numactl process

Default Value: []

[Code Reference](/pkg/apis/deployment/v1/server_group_numactl_spec.go#L38)

### .spec.agents.numactl.enabled: bool

Enabled define if numactl should be enabled

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/server_group_numactl_spec.go#L30)

### .spec.agents.numactl.path: string

Path define numactl path within the container

Default Value: /usr/bin/numactl

[Code Reference](/pkg/apis/deployment/v1/server_group_numactl_spec.go#L34)

### .spec.agents.overrideDetectedNumberOfCores: bool

OverrideDetectedNumberOfCores determines if number of cores should be overrided based on values in resources.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L92)

### .spec.agents.overrideDetectedTotalMemory: bool

OverrideDetectedTotalMemory determines if memory should be overrided based on values in resources.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L90)

### .spec.agents.podModes.network: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L31)

### .spec.agents.podModes.pid: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L32)

### .spec.agents.port: uint16

Port define Port used by member

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L182)

### .spec.agents.priorityClassName: string

PriorityClassName specifies a priority class name

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L118)

### .spec.agents.probes.livenessProbeDisabled: bool

LivenessProbeDisabled if true livenessProbes are disabled

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L193)

### .spec.agents.probes.livenessProbeSpec.failureThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L227)

### .spec.agents.probes.livenessProbeSpec.initialDelaySeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L223)

### .spec.agents.probes.livenessProbeSpec.periodSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.agents.probes.livenessProbeSpec.successThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L226)

### .spec.agents.probes.livenessProbeSpec.timeoutSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L225)

### .spec.agents.probes.ReadinessProbeDisabled: bool

OldReadinessProbeDisabled if true readinessProbes are disabled
Deprecated: This field is deprecated, keept only for backward compatibility.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L200)

### .spec.agents.probes.readinessProbeDisabled: bool

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L202)

### .spec.agents.probes.readinessProbeSpec.failureThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L227)

### .spec.agents.probes.readinessProbeSpec.initialDelaySeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L223)

### .spec.agents.probes.readinessProbeSpec.periodSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.agents.probes.readinessProbeSpec.successThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L226)

### .spec.agents.probes.readinessProbeSpec.timeoutSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L225)

### .spec.agents.probes.startupProbeDisabled: bool

StartupProbeDisabled if true startupProbes are disabled

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L207)

### .spec.agents.probes.startupProbeSpec.failureThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L227)

### .spec.agents.probes.startupProbeSpec.initialDelaySeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L223)

### .spec.agents.probes.startupProbeSpec.periodSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.agents.probes.startupProbeSpec.successThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L226)

### .spec.agents.probes.startupProbeSpec.timeoutSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L225)

### .spec.agents.pvcResizeMode: string

VolumeResizeMode specified resize mode for pvc

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L124)

### .spec.agents.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L88)

### .spec.agents.schedulerName: string

SchedulerName define scheduler name used for group

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L82)

### .spec.agents.securityContext.addCapabilities: []string

AddCapabilities add new capabilities to containers

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L45)

### .spec.agents.securityContext.allowPrivilegeEscalation: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L47)

### .spec.agents.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L43)

### .spec.agents.securityContext.fsGroup: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L55)

### .spec.agents.securityContext.privileged: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L48)

### .spec.agents.securityContext.readOnlyRootFilesystem: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)

### .spec.agents.securityContext.runAsGroup: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L52)

### .spec.agents.securityContext.runAsNonRoot: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L50)

### .spec.agents.securityContext.runAsUser: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L51)

### .spec.agents.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L71)

### .spec.agents.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L76)

### .spec.agents.securityContext.supplementalGroups: []int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L54)

### .spec.agents.securityContext.sysctls: map[string]intstr.IntOrString

Sysctls hold a list of namespaced sysctls used for the pod. Pods with unsupported
sysctls (by the container runtime) might fail to launch.
Map Value can be String or Int

Links:
* [Documentation](https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster/)

Example:
```yaml
sysctls:
  "kernel.shm_rmid_forced": "0"
  "net.core.somaxconn": 1024
  "kernel.msgmax": "65536"
```

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L66)

### .spec.agents.serviceAccountName: string

ServiceAccountName specifies the name of the service account used for Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L112)

### .spec.agents.shutdownDelay: int

ShutdownDelay define how long operator should delay finalizer removal after shutdown

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L163)

### .spec.agents.shutdownMethod: string

ShutdownMethod describe procedure of member shutdown taken by Operator

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L161)

### .spec.agents.sidecarCoreNames: []string

SidecarCoreNames is a list of sidecar containers which must run in the pod.
Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L141)

### .spec.agents.sidecars: []core.Container

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L145)

### .spec.agents.storageClassName: string

StorageClassName specifies the classname for storage of the servers.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L84)

### .spec.agents.terminationGracePeriodSeconds: int64

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L173)

### .spec.agents.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L96)

### .spec.agents.volumeAllowShrink: bool

Deprecated: VolumeAllowShrink allows shrink the volume

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L126)

### .spec.agents.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a template for volume claims

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L122)

### .spec.agents.volumeMounts: []ServerGroupSpecVolumeMount

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation of ServerGroupSpecVolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L153)

### .spec.agents.volumes\[int\].configMap: core.ConfigMapVolumeSource

ConfigMap which should be mounted into pod

Links:
* [Documentation of core.ConfigMapVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#configmapvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L138)

### .spec.agents.volumes\[int\].emptyDir: core.EmptyDirVolumeSource

EmptyDir

Links:
* [Documentation of core.EmptyDirVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#emptydirvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L143)

### .spec.agents.volumes\[int\].hostPath: core.HostPathVolumeSource

HostPath

Links:
* [Documentation of core.HostPathVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#hostpathvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L148)

### .spec.agents.volumes\[int\].name: string

Name of volume

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L128)

### .spec.agents.volumes\[int\].persistentVolumeClaim: core.PersistentVolumeClaimVolumeSource

PersistentVolumeClaim

Links:
* [Documentation of core.PersistentVolumeClaimVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaimvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L153)

### .spec.agents.volumes\[int\].secret: core.SecretVolumeSource

Secret which should be mounted into pod

Links:
* [Documentation of core.SecretVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#secretvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L133)

### .spec.allowUnsafeUpgrade: bool

AllowUnsafeUpgrade determines if upgrade on missing member or with not in sync shards is allowed

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L163)

### .spec.annotations: map[string]string

Annotations specifies the annotations added to all ArangoDeployment owned resources (pods, services, PVC’s, PDB’s).

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L126)

### .spec.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L129)

### .spec.annotationsMode: string

AnnotationsMode defines annotations mode which should be use while overriding annotations.

Possible Values: 
* disabled (default) - Disable annotations/labels override. Default if there is no annotations/labels set in ArangoDeployment
* append - Add new annotations/labels without affecting old ones
* replace - Replace existing annotations/labels

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L135)

### .spec.architecture: []string

Architecture defines the list of supported architectures.
First element on the list is marked as default architecture.

Links:
* [Architecture Change](/docs/how-to/arch_change.md)

Default Value: ['amd64']

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L258)

### .spec.auth.jwtSecretName: string

[Code Reference](/pkg/apis/deployment/v1/authentication_spec.go#L31)

### .spec.bootstrap.passwordSecretNames: map[string]string

PasswordSecretNames contains a map of username to password-secret-name

[Code Reference](/pkg/apis/deployment/v1/bootstrap.go#L53)

### .spec.chaos.enabled: bool

Enabled switches the chaos monkey for a deployment on or off.

[Code Reference](/pkg/apis/deployment/v1/chaos_spec.go#L33)

### .spec.chaos.interval: int64

Interval is the time between events

[Code Reference](/pkg/apis/deployment/v1/chaos_spec.go#L35)

### .spec.chaos.kill-pod-probability: int

KillPodProbability is the chance of a pod being killed during an event

[Code Reference](/pkg/apis/deployment/v1/chaos_spec.go#L37)

### .spec.ClusterDomain: string

ClusterDomain define domain used in the kubernetes cluster.
Required only of domain is not set to default (cluster.local)

Default Value: cluster.local

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L237)

### .spec.communicationMethod: string

CommunicationMethod define communication method used in deployment

Possible Values: 
* headless (default) - Define old communication mechanism, based on headless service.
* dns - Define ClusterIP Service DNS based communication.
* short-dns - Define ClusterIP Service DNS based communication. Use namespaced short DNS (used in migration)
* headless-dns - Define Headless Service DNS based communication.
* ip - Define ClusterIP Service IP based communication.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L245)

### .spec.coordinators.affinity: core.PodAffinity

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.PodAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L134)

### .spec.coordinators.allowMemberRecreation: bool

AllowMemberRecreation allows to recreate member. Value is used only for Coordinator and DBServer with default to True, for all other groups set to false.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L171)

### .spec.coordinators.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L98)

### .spec.coordinators.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L100)

### .spec.coordinators.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L102)

### .spec.coordinators.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L130)

### .spec.coordinators.args: []string

Args holds additional commandline arguments

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L78)

### .spec.coordinators.count: int

Count holds the requested number of servers

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L72)

### .spec.coordinators.entrypoint: string

Entrypoint overrides container executable

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L80)

### .spec.coordinators.envs\[int\].name: string

[Code Reference](/pkg/apis/deployment/v1/server_group_env_var.go#L26)

### .spec.coordinators.envs\[int\].value: string

[Code Reference](/pkg/apis/deployment/v1/server_group_env_var.go#L27)

### .spec.coordinators.ephemeralVolumes.apps.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)

### .spec.coordinators.ephemeralVolumes.temp.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)

### .spec.coordinators.exporterPort: uint16

ExporterPort define Port used by exporter

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L184)

### .spec.coordinators.extendedRotationCheck: bool

ExtendedRotationCheck extend checks for rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L157)

### .spec.coordinators.externalPortEnabled: bool

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L169)

### .spec.coordinators.indexMethod: string

IndexMethod define group Indexing method

Possible Values: 
* random (default) - Pick random ID for member. Enforced on the Community Operator.
* ordered - Use sequential number as Member ID, starting from 0. Enterprise Operator required.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L177)

### .spec.coordinators.initContainers.containers: []core.Container

Containers contains list of containers

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_init_containers.go#L91)

### .spec.coordinators.initContainers.mode: string

Mode keep container replace mode

[Code Reference](/pkg/apis/deployment/v1/server_group_init_containers.go#L94)

### .spec.coordinators.internalPort: int

InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L165)

### .spec.coordinators.internalPortProtocol: string

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L167)

### .spec.coordinators.labels: map[string]string

Labels specified the labels added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L104)

### .spec.coordinators.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L106)

### .spec.coordinators.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L108)

### .spec.coordinators.maxCount: int

MaxCount specifies a upper limit for count

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L76)

### .spec.coordinators.minCount: int

MinCount specifies a lower limit for count

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L74)

### .spec.coordinators.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L138)

### .spec.coordinators.nodeSelector: map[string]string

NodeSelector speficies a set of selectors for nodes

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L114)

### .spec.coordinators.numactl.args: []string

Args define list of the numactl process

Default Value: []

[Code Reference](/pkg/apis/deployment/v1/server_group_numactl_spec.go#L38)

### .spec.coordinators.numactl.enabled: bool

Enabled define if numactl should be enabled

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/server_group_numactl_spec.go#L30)

### .spec.coordinators.numactl.path: string

Path define numactl path within the container

Default Value: /usr/bin/numactl

[Code Reference](/pkg/apis/deployment/v1/server_group_numactl_spec.go#L34)

### .spec.coordinators.overrideDetectedNumberOfCores: bool

OverrideDetectedNumberOfCores determines if number of cores should be overrided based on values in resources.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L92)

### .spec.coordinators.overrideDetectedTotalMemory: bool

OverrideDetectedTotalMemory determines if memory should be overrided based on values in resources.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L90)

### .spec.coordinators.podModes.network: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L31)

### .spec.coordinators.podModes.pid: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L32)

### .spec.coordinators.port: uint16

Port define Port used by member

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L182)

### .spec.coordinators.priorityClassName: string

PriorityClassName specifies a priority class name

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L118)

### .spec.coordinators.probes.livenessProbeDisabled: bool

LivenessProbeDisabled if true livenessProbes are disabled

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L193)

### .spec.coordinators.probes.livenessProbeSpec.failureThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L227)

### .spec.coordinators.probes.livenessProbeSpec.initialDelaySeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L223)

### .spec.coordinators.probes.livenessProbeSpec.periodSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.coordinators.probes.livenessProbeSpec.successThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L226)

### .spec.coordinators.probes.livenessProbeSpec.timeoutSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L225)

### .spec.coordinators.probes.ReadinessProbeDisabled: bool

OldReadinessProbeDisabled if true readinessProbes are disabled
Deprecated: This field is deprecated, keept only for backward compatibility.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L200)

### .spec.coordinators.probes.readinessProbeDisabled: bool

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L202)

### .spec.coordinators.probes.readinessProbeSpec.failureThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L227)

### .spec.coordinators.probes.readinessProbeSpec.initialDelaySeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L223)

### .spec.coordinators.probes.readinessProbeSpec.periodSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.coordinators.probes.readinessProbeSpec.successThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L226)

### .spec.coordinators.probes.readinessProbeSpec.timeoutSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L225)

### .spec.coordinators.probes.startupProbeDisabled: bool

StartupProbeDisabled if true startupProbes are disabled

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L207)

### .spec.coordinators.probes.startupProbeSpec.failureThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L227)

### .spec.coordinators.probes.startupProbeSpec.initialDelaySeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L223)

### .spec.coordinators.probes.startupProbeSpec.periodSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.coordinators.probes.startupProbeSpec.successThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L226)

### .spec.coordinators.probes.startupProbeSpec.timeoutSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L225)

### .spec.coordinators.pvcResizeMode: string

VolumeResizeMode specified resize mode for pvc

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L124)

### .spec.coordinators.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L88)

### .spec.coordinators.schedulerName: string

SchedulerName define scheduler name used for group

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L82)

### .spec.coordinators.securityContext.addCapabilities: []string

AddCapabilities add new capabilities to containers

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L45)

### .spec.coordinators.securityContext.allowPrivilegeEscalation: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L47)

### .spec.coordinators.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L43)

### .spec.coordinators.securityContext.fsGroup: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L55)

### .spec.coordinators.securityContext.privileged: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L48)

### .spec.coordinators.securityContext.readOnlyRootFilesystem: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)

### .spec.coordinators.securityContext.runAsGroup: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L52)

### .spec.coordinators.securityContext.runAsNonRoot: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L50)

### .spec.coordinators.securityContext.runAsUser: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L51)

### .spec.coordinators.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L71)

### .spec.coordinators.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L76)

### .spec.coordinators.securityContext.supplementalGroups: []int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L54)

### .spec.coordinators.securityContext.sysctls: map[string]intstr.IntOrString

Sysctls hold a list of namespaced sysctls used for the pod. Pods with unsupported
sysctls (by the container runtime) might fail to launch.
Map Value can be String or Int

Links:
* [Documentation](https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster/)

Example:
```yaml
sysctls:
  "kernel.shm_rmid_forced": "0"
  "net.core.somaxconn": 1024
  "kernel.msgmax": "65536"
```

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L66)

### .spec.coordinators.serviceAccountName: string

ServiceAccountName specifies the name of the service account used for Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L112)

### .spec.coordinators.shutdownDelay: int

ShutdownDelay define how long operator should delay finalizer removal after shutdown

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L163)

### .spec.coordinators.shutdownMethod: string

ShutdownMethod describe procedure of member shutdown taken by Operator

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L161)

### .spec.coordinators.sidecarCoreNames: []string

SidecarCoreNames is a list of sidecar containers which must run in the pod.
Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L141)

### .spec.coordinators.sidecars: []core.Container

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L145)

### .spec.coordinators.storageClassName: string

StorageClassName specifies the classname for storage of the servers.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L84)

### .spec.coordinators.terminationGracePeriodSeconds: int64

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L173)

### .spec.coordinators.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L96)

### .spec.coordinators.volumeAllowShrink: bool

Deprecated: VolumeAllowShrink allows shrink the volume

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L126)

### .spec.coordinators.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a template for volume claims

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L122)

### .spec.coordinators.volumeMounts: []ServerGroupSpecVolumeMount

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation of ServerGroupSpecVolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L153)

### .spec.coordinators.volumes\[int\].configMap: core.ConfigMapVolumeSource

ConfigMap which should be mounted into pod

Links:
* [Documentation of core.ConfigMapVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#configmapvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L138)

### .spec.coordinators.volumes\[int\].emptyDir: core.EmptyDirVolumeSource

EmptyDir

Links:
* [Documentation of core.EmptyDirVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#emptydirvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L143)

### .spec.coordinators.volumes\[int\].hostPath: core.HostPathVolumeSource

HostPath

Links:
* [Documentation of core.HostPathVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#hostpathvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L148)

### .spec.coordinators.volumes\[int\].name: string

Name of volume

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L128)

### .spec.coordinators.volumes\[int\].persistentVolumeClaim: core.PersistentVolumeClaimVolumeSource

PersistentVolumeClaim

Links:
* [Documentation of core.PersistentVolumeClaimVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaimvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L153)

### .spec.coordinators.volumes\[int\].secret: core.SecretVolumeSource

Secret which should be mounted into pod

Links:
* [Documentation of core.SecretVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#secretvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L133)

### .spec.database.maintenance: bool

Maintenance manage maintenance mode on Cluster side. Requires maintenance feature to be enabled

[Code Reference](/pkg/apis/deployment/v1/database_spec.go#L25)

### .spec.dbservers.affinity: core.PodAffinity

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.PodAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L134)

### .spec.dbservers.allowMemberRecreation: bool

AllowMemberRecreation allows to recreate member. Value is used only for Coordinator and DBServer with default to True, for all other groups set to false.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L171)

### .spec.dbservers.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L98)

### .spec.dbservers.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L100)

### .spec.dbservers.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L102)

### .spec.dbservers.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L130)

### .spec.dbservers.args: []string

Args holds additional commandline arguments

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L78)

### .spec.dbservers.count: int

Count holds the requested number of servers

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L72)

### .spec.dbservers.entrypoint: string

Entrypoint overrides container executable

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L80)

### .spec.dbservers.envs\[int\].name: string

[Code Reference](/pkg/apis/deployment/v1/server_group_env_var.go#L26)

### .spec.dbservers.envs\[int\].value: string

[Code Reference](/pkg/apis/deployment/v1/server_group_env_var.go#L27)

### .spec.dbservers.ephemeralVolumes.apps.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)

### .spec.dbservers.ephemeralVolumes.temp.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)

### .spec.dbservers.exporterPort: uint16

ExporterPort define Port used by exporter

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L184)

### .spec.dbservers.extendedRotationCheck: bool

ExtendedRotationCheck extend checks for rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L157)

### .spec.dbservers.externalPortEnabled: bool

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L169)

### .spec.dbservers.indexMethod: string

IndexMethod define group Indexing method

Possible Values: 
* random (default) - Pick random ID for member. Enforced on the Community Operator.
* ordered - Use sequential number as Member ID, starting from 0. Enterprise Operator required.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L177)

### .spec.dbservers.initContainers.containers: []core.Container

Containers contains list of containers

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_init_containers.go#L91)

### .spec.dbservers.initContainers.mode: string

Mode keep container replace mode

[Code Reference](/pkg/apis/deployment/v1/server_group_init_containers.go#L94)

### .spec.dbservers.internalPort: int

InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L165)

### .spec.dbservers.internalPortProtocol: string

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L167)

### .spec.dbservers.labels: map[string]string

Labels specified the labels added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L104)

### .spec.dbservers.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L106)

### .spec.dbservers.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L108)

### .spec.dbservers.maxCount: int

MaxCount specifies a upper limit for count

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L76)

### .spec.dbservers.minCount: int

MinCount specifies a lower limit for count

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L74)

### .spec.dbservers.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L138)

### .spec.dbservers.nodeSelector: map[string]string

NodeSelector speficies a set of selectors for nodes

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L114)

### .spec.dbservers.numactl.args: []string

Args define list of the numactl process

Default Value: []

[Code Reference](/pkg/apis/deployment/v1/server_group_numactl_spec.go#L38)

### .spec.dbservers.numactl.enabled: bool

Enabled define if numactl should be enabled

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/server_group_numactl_spec.go#L30)

### .spec.dbservers.numactl.path: string

Path define numactl path within the container

Default Value: /usr/bin/numactl

[Code Reference](/pkg/apis/deployment/v1/server_group_numactl_spec.go#L34)

### .spec.dbservers.overrideDetectedNumberOfCores: bool

OverrideDetectedNumberOfCores determines if number of cores should be overrided based on values in resources.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L92)

### .spec.dbservers.overrideDetectedTotalMemory: bool

OverrideDetectedTotalMemory determines if memory should be overrided based on values in resources.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L90)

### .spec.dbservers.podModes.network: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L31)

### .spec.dbservers.podModes.pid: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L32)

### .spec.dbservers.port: uint16

Port define Port used by member

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L182)

### .spec.dbservers.priorityClassName: string

PriorityClassName specifies a priority class name

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L118)

### .spec.dbservers.probes.livenessProbeDisabled: bool

LivenessProbeDisabled if true livenessProbes are disabled

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L193)

### .spec.dbservers.probes.livenessProbeSpec.failureThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L227)

### .spec.dbservers.probes.livenessProbeSpec.initialDelaySeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L223)

### .spec.dbservers.probes.livenessProbeSpec.periodSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.dbservers.probes.livenessProbeSpec.successThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L226)

### .spec.dbservers.probes.livenessProbeSpec.timeoutSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L225)

### .spec.dbservers.probes.ReadinessProbeDisabled: bool

OldReadinessProbeDisabled if true readinessProbes are disabled
Deprecated: This field is deprecated, keept only for backward compatibility.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L200)

### .spec.dbservers.probes.readinessProbeDisabled: bool

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L202)

### .spec.dbservers.probes.readinessProbeSpec.failureThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L227)

### .spec.dbservers.probes.readinessProbeSpec.initialDelaySeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L223)

### .spec.dbservers.probes.readinessProbeSpec.periodSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.dbservers.probes.readinessProbeSpec.successThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L226)

### .spec.dbservers.probes.readinessProbeSpec.timeoutSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L225)

### .spec.dbservers.probes.startupProbeDisabled: bool

StartupProbeDisabled if true startupProbes are disabled

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L207)

### .spec.dbservers.probes.startupProbeSpec.failureThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L227)

### .spec.dbservers.probes.startupProbeSpec.initialDelaySeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L223)

### .spec.dbservers.probes.startupProbeSpec.periodSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.dbservers.probes.startupProbeSpec.successThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L226)

### .spec.dbservers.probes.startupProbeSpec.timeoutSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L225)

### .spec.dbservers.pvcResizeMode: string

VolumeResizeMode specified resize mode for pvc

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L124)

### .spec.dbservers.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L88)

### .spec.dbservers.schedulerName: string

SchedulerName define scheduler name used for group

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L82)

### .spec.dbservers.securityContext.addCapabilities: []string

AddCapabilities add new capabilities to containers

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L45)

### .spec.dbservers.securityContext.allowPrivilegeEscalation: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L47)

### .spec.dbservers.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L43)

### .spec.dbservers.securityContext.fsGroup: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L55)

### .spec.dbservers.securityContext.privileged: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L48)

### .spec.dbservers.securityContext.readOnlyRootFilesystem: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)

### .spec.dbservers.securityContext.runAsGroup: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L52)

### .spec.dbservers.securityContext.runAsNonRoot: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L50)

### .spec.dbservers.securityContext.runAsUser: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L51)

### .spec.dbservers.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L71)

### .spec.dbservers.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L76)

### .spec.dbservers.securityContext.supplementalGroups: []int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L54)

### .spec.dbservers.securityContext.sysctls: map[string]intstr.IntOrString

Sysctls hold a list of namespaced sysctls used for the pod. Pods with unsupported
sysctls (by the container runtime) might fail to launch.
Map Value can be String or Int

Links:
* [Documentation](https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster/)

Example:
```yaml
sysctls:
  "kernel.shm_rmid_forced": "0"
  "net.core.somaxconn": 1024
  "kernel.msgmax": "65536"
```

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L66)

### .spec.dbservers.serviceAccountName: string

ServiceAccountName specifies the name of the service account used for Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L112)

### .spec.dbservers.shutdownDelay: int

ShutdownDelay define how long operator should delay finalizer removal after shutdown

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L163)

### .spec.dbservers.shutdownMethod: string

ShutdownMethod describe procedure of member shutdown taken by Operator

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L161)

### .spec.dbservers.sidecarCoreNames: []string

SidecarCoreNames is a list of sidecar containers which must run in the pod.
Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L141)

### .spec.dbservers.sidecars: []core.Container

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L145)

### .spec.dbservers.storageClassName: string

StorageClassName specifies the classname for storage of the servers.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L84)

### .spec.dbservers.terminationGracePeriodSeconds: int64

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L173)

### .spec.dbservers.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L96)

### .spec.dbservers.volumeAllowShrink: bool

Deprecated: VolumeAllowShrink allows shrink the volume

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L126)

### .spec.dbservers.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a template for volume claims

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L122)

### .spec.dbservers.volumeMounts: []ServerGroupSpecVolumeMount

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation of ServerGroupSpecVolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L153)

### .spec.dbservers.volumes\[int\].configMap: core.ConfigMapVolumeSource

ConfigMap which should be mounted into pod

Links:
* [Documentation of core.ConfigMapVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#configmapvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L138)

### .spec.dbservers.volumes\[int\].emptyDir: core.EmptyDirVolumeSource

EmptyDir

Links:
* [Documentation of core.EmptyDirVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#emptydirvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L143)

### .spec.dbservers.volumes\[int\].hostPath: core.HostPathVolumeSource

HostPath

Links:
* [Documentation of core.HostPathVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#hostpathvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L148)

### .spec.dbservers.volumes\[int\].name: string

Name of volume

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L128)

### .spec.dbservers.volumes\[int\].persistentVolumeClaim: core.PersistentVolumeClaimVolumeSource

PersistentVolumeClaim

Links:
* [Documentation of core.PersistentVolumeClaimVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaimvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L153)

### .spec.dbservers.volumes\[int\].secret: core.SecretVolumeSource

Secret which should be mounted into pod

Links:
* [Documentation of core.SecretVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#secretvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L133)

### .spec.disableIPv6: bool

DisableIPv6 setting prevents the use of IPv6 addresses by ArangoDB servers.
This setting cannot be changed after the deployment has been created.

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L109)

### .spec.downtimeAllowed: bool

DowntimeAllowed setting is used to allow automatic reconciliation actions that yield some downtime of the ArangoDB deployment.
When this setting is set to false, no automatic action that may result in downtime is allowed.
If the need for such an action is detected, an event is added to the ArangoDeployment.
Once this setting is set to true, the automatic action is executed.
Operations that may result in downtime are:
- Rotating TLS CA certificate
Note: It is still possible that there is some downtime when the Kubernetes cluster is down, or in a bad state, irrespective of the value of this setting.

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L104)

### .spec.environment: string

Environment setting specifies the type of environment in which the deployment is created.

Possible Values: 
* Development (default) - This value optimizes the deployment for development use. It is possible to run a deployment on a small number of nodes (e.g. minikube).
* Production - This value optimizes the deployment for production use. It puts required affinity constraints on all pods to avoid Agents & DB-Servers from running on the same machine.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L65)

### .spec.externalAccess.advertisedEndpoint: string

AdvertisedEndpoint is passed to the coordinators/single servers for advertising a specific endpoint

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L55)

### .spec.externalAccess.loadBalancerIP: string

LoadBalancerIP define optional IP used to configure a load-balancer on, in case of Auto or LoadBalancer type.

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L45)

### .spec.externalAccess.loadBalancerSourceRanges: []string

LoadBalancerSourceRanges define LoadBalancerSourceRanges used for LoadBalancer Service type
If specified and supported by the platform, this will restrict traffic through the cloud-provider
load-balancer will be restricted to the specified client IPs. This field will be ignored if the
cloud-provider does not support the feature.

Links:
* [Cloud Provider Firewall](https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/)

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L52)

### .spec.externalAccess.managedServiceNames: []string

ManagedServiceNames keeps names of services which are not managed by KubeArangoDB.
It is only relevant when type of service is `managed`.

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L59)

### .spec.externalAccess.nodePort: int

NodePort define optional port used in case of Auto or NodePort type.

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L42)

### .spec.externalAccess.type: string

Type specifies the type of Service that will be created to provide access to the ArangoDB deployment from outside the Kubernetes cluster.

Possible Values: 
* Auto (default) - Create a Service of type LoadBalancer and fallback to a Service or type NodePort when the LoadBalancer is not assigned an IP address.
* None - limit access to application running inside the Kubernetes cluster.
* LoadBalancer - Create a Service of type LoadBalancer for the ArangoDB deployment.
* NodePort - Create a Service of type NodePort for the ArangoDB deployment.

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L39)

### .spec.features.foxx.queues: bool

[Code Reference](/pkg/apis/deployment/v1/deployment_features.go#L24)

### .spec.id.affinity: core.PodAffinity

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.PodAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_id_group_spec.go#L44)

### .spec.id.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_id_group_spec.go#L40)

### .spec.id.entrypoint: string

Entrypoint overrides container executable

[Code Reference](/pkg/apis/deployment/v1/server_id_group_spec.go#L28)

### .spec.id.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_id_group_spec.go#L48)

### .spec.id.nodeSelector: map[string]string

NodeSelector speficies a set of selectors for nodes

[Code Reference](/pkg/apis/deployment/v1/server_id_group_spec.go#L34)

### .spec.id.priorityClassName: string

PriorityClassName specifies a priority class name

[Code Reference](/pkg/apis/deployment/v1/server_id_group_spec.go#L36)

### .spec.id.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_id_group_spec.go#L56)

### .spec.id.securityContext.addCapabilities: []string

AddCapabilities add new capabilities to containers

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L45)

### .spec.id.securityContext.allowPrivilegeEscalation: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L47)

### .spec.id.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L43)

### .spec.id.securityContext.fsGroup: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L55)

### .spec.id.securityContext.privileged: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L48)

### .spec.id.securityContext.readOnlyRootFilesystem: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)

### .spec.id.securityContext.runAsGroup: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L52)

### .spec.id.securityContext.runAsNonRoot: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L50)

### .spec.id.securityContext.runAsUser: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L51)

### .spec.id.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L71)

### .spec.id.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L76)

### .spec.id.securityContext.supplementalGroups: []int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L54)

### .spec.id.securityContext.sysctls: map[string]intstr.IntOrString

Sysctls hold a list of namespaced sysctls used for the pod. Pods with unsupported
sysctls (by the container runtime) might fail to launch.
Map Value can be String or Int

Links:
* [Documentation](https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster/)

Example:
```yaml
sysctls:
  "kernel.shm_rmid_forced": "0"
  "net.core.somaxconn": 1024
  "kernel.msgmax": "65536"
```

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L66)

### .spec.id.serviceAccountName: string

ServiceAccountName specifies the name of the service account used for Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_id_group_spec.go#L50)

### .spec.id.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_id_group_spec.go#L32)

### .spec.image: string

Image specifies the docker image to use for all ArangoDB servers.
In a development environment this setting defaults to arangodb/arangodb:latest.
For production environments this is a required setting without a default value.
It is highly recommend to use explicit version (not latest) for production environments.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L78)

### .spec.imageDiscoveryMode: string

ImageDiscoveryMode specifies the image discovery mode.

Possible Values: 
* kubelet (default) - Use sha256 of the discovered image in the pods
* direct - Use image provided in the spec.image directly in the pods

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L94)

### .spec.imagePullPolicy: core.PullPolicy

ImagePullPolicy specifies the pull policy for the docker image to use for all ArangoDB servers.

Links:
* [Documentation of core.PullPolicy](https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy)

Possible Values: 
* Always (default) - Means that kubelet always attempts to pull the latest image. Container will fail If the pull fails.
* Never - Means that kubelet never pulls an image, but only uses a local image. Container will fail if the image isn't present
* IfNotPresent - Means that kubelet pulls if the image isn't present on disk. Container will fail if the image isn't present and the pull fails.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L86)

### .spec.imagePullSecrets: []string

ImagePullSecrets specifies the list of image pull secrets for the docker image to use for all ArangoDB servers.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L89)

### .spec.labels: map[string]string

Labels specifies the labels added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L138)

### .spec.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L141)

### .spec.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

Possible Values: 
* disabled (default) - Disable annotations/labels override. Default if there is no annotations/labels set in ArangoDeployment
* append - Add new annotations/labels without affecting old ones
* replace - Replace existing annotations/labels

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L147)

### .spec.license.secretName: string

[Code Reference](/pkg/apis/deployment/v1/license_spec.go#L30)

### .spec.lifecycle.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

[Code Reference](/pkg/apis/deployment/v1/lifecycle_spec.go#L31)

### .spec.memberPropagationMode: string

MemberPropagationMode defines how changes to pod spec should be propogated.
Changes to a pod’s configuration require a restart of that pod in almost all cases.
Pods are restarted eagerly by default, which can cause more restarts than desired, especially when updating arangod as well as the operator.
The propagation of the configuration changes can be deferred to the next restart, either triggered manually by the user or by another operation like an upgrade.
This reduces the number of restarts for upgrading both the server and the operator from two to one.

Possible Values: 
* always (default) - Restart the member as soon as a configuration change is discovered
* on-restart - Wait until the next restart to change the member configuration

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L220)

### .spec.metrics.authentication.jwtTokenSecretName: string

JWTTokenSecretName contains the name of the JWT kubernetes secret used for authentication

[Code Reference](/pkg/apis/deployment/v1/deployment_metrics_spec.go#L34)

### .spec.metrics.enabled: bool

[Code Reference](/pkg/apis/deployment/v1/deployment_metrics_spec.go#L77)

### .spec.metrics.image: string

deprecated

[Code Reference](/pkg/apis/deployment/v1/deployment_metrics_spec.go#L79)

### .spec.metrics.mode: string

deprecated

[Code Reference](/pkg/apis/deployment/v1/deployment_metrics_spec.go#L86)

### .spec.metrics.port: uint16

[Code Reference](/pkg/apis/deployment/v1/deployment_metrics_spec.go#L91)

### .spec.metrics.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

[Code Reference](/pkg/apis/deployment/v1/deployment_metrics_spec.go#L84)

### .spec.metrics.serviceMonitor.enabled: bool

[Code Reference](/pkg/apis/deployment/v1/deployment_metrics_service_monitor_spec.go#L24)

### .spec.metrics.serviceMonitor.labels: map[string]string

[Code Reference](/pkg/apis/deployment/v1/deployment_metrics_service_monitor_spec.go#L25)

### .spec.metrics.tls: bool

[Code Reference](/pkg/apis/deployment/v1/deployment_metrics_spec.go#L87)

### .spec.mode: string

Mode specifies the type of ArangoDB deployment to create.

Possible Values: 
* Cluster (default) - Full cluster. Defaults to 3 Agents, 3 DB-Servers & 3 Coordinators.
* ActiveFailover - Active-failover single pair. Defaults to 3 Agents and 2 single servers.
* Single - Single server only (note this does not provide high availability or reliability).

This field is **immutable**: Change of the ArangoDeployment Mode is not possible after creation.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L60)

### .spec.networkAttachedVolumes: bool

NetworkAttachedVolumes
If set to `true`, a ResignLeadership operation will be triggered when a DB-Server pod is evicted (rather than a CleanOutServer operation).
Furthermore, the pod will simply be redeployed on a different node, rather than cleaned and retired and replaced by a new member.
You must only set this option to true if your persistent volumes are “movable” in the sense that they can be mounted from a different k8s node, like in the case of network attached volumes.
If your persistent volumes are tied to a specific pod, you must leave this option on false.

Default Value: true

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L123)

### .spec.rebalancer.enabled: bool

[Code Reference](/pkg/apis/deployment/v1/rebalancer_spec.go#L26)

### .spec.rebalancer.optimizers.leader: bool

[Code Reference](/pkg/apis/deployment/v1/rebalancer_spec.go#L74)

### .spec.rebalancer.parallelMoves: int

[Code Reference](/pkg/apis/deployment/v1/rebalancer_spec.go#L28)

### .spec.rebalancer.readers.count: bool

deprecated does not work in Rebalancer V2
Count Enable Shard Count machanism

[Code Reference](/pkg/apis/deployment/v1/rebalancer_spec.go#L62)

### .spec.recovery.autoRecover: bool

[Code Reference](/pkg/apis/deployment/v1/recovery_spec.go#L26)

### .spec.restoreEncryptionSecret: string

RestoreEncryptionSecret specifies optional name of secret which contains encryption key used for restore

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L160)

### .spec.restoreFrom: string

RestoreFrom setting specifies a `ArangoBackup` resource name the cluster should be restored from.
After a restore or failure to do so, the status of the deployment contains information about the restore operation in the restore key.
It will contain some of the following fields:
- `requestedFrom`: name of the ArangoBackup used to restore from.
- `message`: optional message explaining why the restore failed.
- `state`: state indicating if the restore was successful or not. Possible values: Restoring, Restored, RestoreFailed
If the restoreFrom key is removed from the spec, the restore key is deleted as well.
A new restore attempt is made if and only if either in the status restore is not set or if spec.restoreFrom and status.requestedFrom are different.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L157)

### .spec.rocksdb.encryption.keySecretName: string

[Code Reference](/pkg/apis/deployment/v1/rocksdb_spec.go#L31)

### .spec.single.affinity: core.PodAffinity

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.PodAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L134)

### .spec.single.allowMemberRecreation: bool

AllowMemberRecreation allows to recreate member. Value is used only for Coordinator and DBServer with default to True, for all other groups set to false.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L171)

### .spec.single.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L98)

### .spec.single.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L100)

### .spec.single.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L102)

### .spec.single.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L130)

### .spec.single.args: []string

Args holds additional commandline arguments

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L78)

### .spec.single.count: int

Count holds the requested number of servers

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L72)

### .spec.single.entrypoint: string

Entrypoint overrides container executable

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L80)

### .spec.single.envs\[int\].name: string

[Code Reference](/pkg/apis/deployment/v1/server_group_env_var.go#L26)

### .spec.single.envs\[int\].value: string

[Code Reference](/pkg/apis/deployment/v1/server_group_env_var.go#L27)

### .spec.single.ephemeralVolumes.apps.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)

### .spec.single.ephemeralVolumes.temp.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)

### .spec.single.exporterPort: uint16

ExporterPort define Port used by exporter

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L184)

### .spec.single.extendedRotationCheck: bool

ExtendedRotationCheck extend checks for rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L157)

### .spec.single.externalPortEnabled: bool

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L169)

### .spec.single.indexMethod: string

IndexMethod define group Indexing method

Possible Values: 
* random (default) - Pick random ID for member. Enforced on the Community Operator.
* ordered - Use sequential number as Member ID, starting from 0. Enterprise Operator required.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L177)

### .spec.single.initContainers.containers: []core.Container

Containers contains list of containers

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_init_containers.go#L91)

### .spec.single.initContainers.mode: string

Mode keep container replace mode

[Code Reference](/pkg/apis/deployment/v1/server_group_init_containers.go#L94)

### .spec.single.internalPort: int

InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L165)

### .spec.single.internalPortProtocol: string

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L167)

### .spec.single.labels: map[string]string

Labels specified the labels added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L104)

### .spec.single.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L106)

### .spec.single.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L108)

### .spec.single.maxCount: int

MaxCount specifies a upper limit for count

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L76)

### .spec.single.minCount: int

MinCount specifies a lower limit for count

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L74)

### .spec.single.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L138)

### .spec.single.nodeSelector: map[string]string

NodeSelector speficies a set of selectors for nodes

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L114)

### .spec.single.numactl.args: []string

Args define list of the numactl process

Default Value: []

[Code Reference](/pkg/apis/deployment/v1/server_group_numactl_spec.go#L38)

### .spec.single.numactl.enabled: bool

Enabled define if numactl should be enabled

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/server_group_numactl_spec.go#L30)

### .spec.single.numactl.path: string

Path define numactl path within the container

Default Value: /usr/bin/numactl

[Code Reference](/pkg/apis/deployment/v1/server_group_numactl_spec.go#L34)

### .spec.single.overrideDetectedNumberOfCores: bool

OverrideDetectedNumberOfCores determines if number of cores should be overrided based on values in resources.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L92)

### .spec.single.overrideDetectedTotalMemory: bool

OverrideDetectedTotalMemory determines if memory should be overrided based on values in resources.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L90)

### .spec.single.podModes.network: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L31)

### .spec.single.podModes.pid: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L32)

### .spec.single.port: uint16

Port define Port used by member

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L182)

### .spec.single.priorityClassName: string

PriorityClassName specifies a priority class name

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L118)

### .spec.single.probes.livenessProbeDisabled: bool

LivenessProbeDisabled if true livenessProbes are disabled

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L193)

### .spec.single.probes.livenessProbeSpec.failureThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L227)

### .spec.single.probes.livenessProbeSpec.initialDelaySeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L223)

### .spec.single.probes.livenessProbeSpec.periodSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.single.probes.livenessProbeSpec.successThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L226)

### .spec.single.probes.livenessProbeSpec.timeoutSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L225)

### .spec.single.probes.ReadinessProbeDisabled: bool

OldReadinessProbeDisabled if true readinessProbes are disabled
Deprecated: This field is deprecated, keept only for backward compatibility.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L200)

### .spec.single.probes.readinessProbeDisabled: bool

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L202)

### .spec.single.probes.readinessProbeSpec.failureThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L227)

### .spec.single.probes.readinessProbeSpec.initialDelaySeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L223)

### .spec.single.probes.readinessProbeSpec.periodSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.single.probes.readinessProbeSpec.successThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L226)

### .spec.single.probes.readinessProbeSpec.timeoutSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L225)

### .spec.single.probes.startupProbeDisabled: bool

StartupProbeDisabled if true startupProbes are disabled

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L207)

### .spec.single.probes.startupProbeSpec.failureThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L227)

### .spec.single.probes.startupProbeSpec.initialDelaySeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L223)

### .spec.single.probes.startupProbeSpec.periodSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.single.probes.startupProbeSpec.successThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L226)

### .spec.single.probes.startupProbeSpec.timeoutSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L225)

### .spec.single.pvcResizeMode: string

VolumeResizeMode specified resize mode for pvc

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L124)

### .spec.single.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L88)

### .spec.single.schedulerName: string

SchedulerName define scheduler name used for group

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L82)

### .spec.single.securityContext.addCapabilities: []string

AddCapabilities add new capabilities to containers

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L45)

### .spec.single.securityContext.allowPrivilegeEscalation: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L47)

### .spec.single.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L43)

### .spec.single.securityContext.fsGroup: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L55)

### .spec.single.securityContext.privileged: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L48)

### .spec.single.securityContext.readOnlyRootFilesystem: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)

### .spec.single.securityContext.runAsGroup: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L52)

### .spec.single.securityContext.runAsNonRoot: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L50)

### .spec.single.securityContext.runAsUser: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L51)

### .spec.single.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L71)

### .spec.single.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L76)

### .spec.single.securityContext.supplementalGroups: []int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L54)

### .spec.single.securityContext.sysctls: map[string]intstr.IntOrString

Sysctls hold a list of namespaced sysctls used for the pod. Pods with unsupported
sysctls (by the container runtime) might fail to launch.
Map Value can be String or Int

Links:
* [Documentation](https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster/)

Example:
```yaml
sysctls:
  "kernel.shm_rmid_forced": "0"
  "net.core.somaxconn": 1024
  "kernel.msgmax": "65536"
```

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L66)

### .spec.single.serviceAccountName: string

ServiceAccountName specifies the name of the service account used for Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L112)

### .spec.single.shutdownDelay: int

ShutdownDelay define how long operator should delay finalizer removal after shutdown

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L163)

### .spec.single.shutdownMethod: string

ShutdownMethod describe procedure of member shutdown taken by Operator

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L161)

### .spec.single.sidecarCoreNames: []string

SidecarCoreNames is a list of sidecar containers which must run in the pod.
Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L141)

### .spec.single.sidecars: []core.Container

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L145)

### .spec.single.storageClassName: string

StorageClassName specifies the classname for storage of the servers.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L84)

### .spec.single.terminationGracePeriodSeconds: int64

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L173)

### .spec.single.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L96)

### .spec.single.volumeAllowShrink: bool

Deprecated: VolumeAllowShrink allows shrink the volume

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L126)

### .spec.single.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a template for volume claims

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L122)

### .spec.single.volumeMounts: []ServerGroupSpecVolumeMount

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation of ServerGroupSpecVolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L153)

### .spec.single.volumes\[int\].configMap: core.ConfigMapVolumeSource

ConfigMap which should be mounted into pod

Links:
* [Documentation of core.ConfigMapVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#configmapvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L138)

### .spec.single.volumes\[int\].emptyDir: core.EmptyDirVolumeSource

EmptyDir

Links:
* [Documentation of core.EmptyDirVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#emptydirvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L143)

### .spec.single.volumes\[int\].hostPath: core.HostPathVolumeSource

HostPath

Links:
* [Documentation of core.HostPathVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#hostpathvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L148)

### .spec.single.volumes\[int\].name: string

Name of volume

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L128)

### .spec.single.volumes\[int\].persistentVolumeClaim: core.PersistentVolumeClaimVolumeSource

PersistentVolumeClaim

Links:
* [Documentation of core.PersistentVolumeClaimVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaimvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L153)

### .spec.single.volumes\[int\].secret: core.SecretVolumeSource

Secret which should be mounted into pod

Links:
* [Documentation of core.SecretVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#secretvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L133)

### .spec.storageEngine: string

StorageEngine specifies the type of storage engine used for all servers in the cluster.
This setting cannot be changed after the cluster has been created.

Possible Values: 
* RocksDB (default) - To use the RocksDB storage engine.
* MMFiles - To use the MMFiles storage engine. Deprecated.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L72)

### .spec.sync.auth.clientCASecretName: string

[Code Reference](/pkg/apis/deployment/v1/sync_authentication_spec.go#L32)

### .spec.sync.auth.jwtSecretName: string

[Code Reference](/pkg/apis/deployment/v1/sync_authentication_spec.go#L31)

### .spec.sync.enabled: bool

[Code Reference](/pkg/apis/deployment/v1/sync_spec.go#L30)

### .spec.sync.externalAccess.accessPackageSecretNames: []string

[Code Reference](/pkg/apis/deployment/v1/sync_external_access_spec.go#L36)

### .spec.sync.externalAccess.advertisedEndpoint: string

AdvertisedEndpoint is passed to the coordinators/single servers for advertising a specific endpoint

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L55)

### .spec.sync.externalAccess.loadBalancerIP: string

LoadBalancerIP define optional IP used to configure a load-balancer on, in case of Auto or LoadBalancer type.

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L45)

### .spec.sync.externalAccess.loadBalancerSourceRanges: []string

LoadBalancerSourceRanges define LoadBalancerSourceRanges used for LoadBalancer Service type
If specified and supported by the platform, this will restrict traffic through the cloud-provider
load-balancer will be restricted to the specified client IPs. This field will be ignored if the
cloud-provider does not support the feature.

Links:
* [Cloud Provider Firewall](https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/)

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L52)

### .spec.sync.externalAccess.managedServiceNames: []string

ManagedServiceNames keeps names of services which are not managed by KubeArangoDB.
It is only relevant when type of service is `managed`.

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L59)

### .spec.sync.externalAccess.masterEndpoint: []string

[Code Reference](/pkg/apis/deployment/v1/sync_external_access_spec.go#L35)

### .spec.sync.externalAccess.nodePort: int

NodePort define optional port used in case of Auto or NodePort type.

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L42)

### .spec.sync.externalAccess.type: string

Type specifies the type of Service that will be created to provide access to the ArangoDB deployment from outside the Kubernetes cluster.

Possible Values: 
* Auto (default) - Create a Service of type LoadBalancer and fallback to a Service or type NodePort when the LoadBalancer is not assigned an IP address.
* None - limit access to application running inside the Kubernetes cluster.
* LoadBalancer - Create a Service of type LoadBalancer for the ArangoDB deployment.
* NodePort - Create a Service of type NodePort for the ArangoDB deployment.

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L39)

### .spec.sync.image: string

[Code Reference](/pkg/apis/deployment/v1/sync_spec.go#L36)

### .spec.sync.monitoring.tokenSecretName: string

[Code Reference](/pkg/apis/deployment/v1/sync_monitoring_spec.go#L31)

### .spec.sync.tls.altNames: []string

[Code Reference](/pkg/apis/deployment/v1/tls_spec.go#L58)

### .spec.sync.tls.caSecretName: string

[Code Reference](/pkg/apis/deployment/v1/tls_spec.go#L57)

### .spec.sync.tls.mode: string

[Code Reference](/pkg/apis/deployment/v1/tls_spec.go#L61)

### .spec.sync.tls.sni.mapping.\<string\>: []string

[Code Reference](/pkg/apis/deployment/v1/tls_sni_spec.go#L30)

### .spec.sync.tls.ttl: string

[Code Reference](/pkg/apis/deployment/v1/tls_spec.go#L59)

### .spec.syncmasters.affinity: core.PodAffinity

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.PodAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L134)

### .spec.syncmasters.allowMemberRecreation: bool

AllowMemberRecreation allows to recreate member. Value is used only for Coordinator and DBServer with default to True, for all other groups set to false.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L171)

### .spec.syncmasters.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L98)

### .spec.syncmasters.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L100)

### .spec.syncmasters.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L102)

### .spec.syncmasters.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L130)

### .spec.syncmasters.args: []string

Args holds additional commandline arguments

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L78)

### .spec.syncmasters.count: int

Count holds the requested number of servers

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L72)

### .spec.syncmasters.entrypoint: string

Entrypoint overrides container executable

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L80)

### .spec.syncmasters.envs\[int\].name: string

[Code Reference](/pkg/apis/deployment/v1/server_group_env_var.go#L26)

### .spec.syncmasters.envs\[int\].value: string

[Code Reference](/pkg/apis/deployment/v1/server_group_env_var.go#L27)

### .spec.syncmasters.ephemeralVolumes.apps.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)

### .spec.syncmasters.ephemeralVolumes.temp.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)

### .spec.syncmasters.exporterPort: uint16

ExporterPort define Port used by exporter

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L184)

### .spec.syncmasters.extendedRotationCheck: bool

ExtendedRotationCheck extend checks for rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L157)

### .spec.syncmasters.externalPortEnabled: bool

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L169)

### .spec.syncmasters.indexMethod: string

IndexMethod define group Indexing method

Possible Values: 
* random (default) - Pick random ID for member. Enforced on the Community Operator.
* ordered - Use sequential number as Member ID, starting from 0. Enterprise Operator required.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L177)

### .spec.syncmasters.initContainers.containers: []core.Container

Containers contains list of containers

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_init_containers.go#L91)

### .spec.syncmasters.initContainers.mode: string

Mode keep container replace mode

[Code Reference](/pkg/apis/deployment/v1/server_group_init_containers.go#L94)

### .spec.syncmasters.internalPort: int

InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L165)

### .spec.syncmasters.internalPortProtocol: string

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L167)

### .spec.syncmasters.labels: map[string]string

Labels specified the labels added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L104)

### .spec.syncmasters.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L106)

### .spec.syncmasters.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L108)

### .spec.syncmasters.maxCount: int

MaxCount specifies a upper limit for count

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L76)

### .spec.syncmasters.minCount: int

MinCount specifies a lower limit for count

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L74)

### .spec.syncmasters.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L138)

### .spec.syncmasters.nodeSelector: map[string]string

NodeSelector speficies a set of selectors for nodes

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L114)

### .spec.syncmasters.numactl.args: []string

Args define list of the numactl process

Default Value: []

[Code Reference](/pkg/apis/deployment/v1/server_group_numactl_spec.go#L38)

### .spec.syncmasters.numactl.enabled: bool

Enabled define if numactl should be enabled

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/server_group_numactl_spec.go#L30)

### .spec.syncmasters.numactl.path: string

Path define numactl path within the container

Default Value: /usr/bin/numactl

[Code Reference](/pkg/apis/deployment/v1/server_group_numactl_spec.go#L34)

### .spec.syncmasters.overrideDetectedNumberOfCores: bool

OverrideDetectedNumberOfCores determines if number of cores should be overrided based on values in resources.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L92)

### .spec.syncmasters.overrideDetectedTotalMemory: bool

OverrideDetectedTotalMemory determines if memory should be overrided based on values in resources.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L90)

### .spec.syncmasters.podModes.network: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L31)

### .spec.syncmasters.podModes.pid: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L32)

### .spec.syncmasters.port: uint16

Port define Port used by member

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L182)

### .spec.syncmasters.priorityClassName: string

PriorityClassName specifies a priority class name

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L118)

### .spec.syncmasters.probes.livenessProbeDisabled: bool

LivenessProbeDisabled if true livenessProbes are disabled

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L193)

### .spec.syncmasters.probes.livenessProbeSpec.failureThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L227)

### .spec.syncmasters.probes.livenessProbeSpec.initialDelaySeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L223)

### .spec.syncmasters.probes.livenessProbeSpec.periodSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.syncmasters.probes.livenessProbeSpec.successThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L226)

### .spec.syncmasters.probes.livenessProbeSpec.timeoutSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L225)

### .spec.syncmasters.probes.ReadinessProbeDisabled: bool

OldReadinessProbeDisabled if true readinessProbes are disabled
Deprecated: This field is deprecated, keept only for backward compatibility.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L200)

### .spec.syncmasters.probes.readinessProbeDisabled: bool

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L202)

### .spec.syncmasters.probes.readinessProbeSpec.failureThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L227)

### .spec.syncmasters.probes.readinessProbeSpec.initialDelaySeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L223)

### .spec.syncmasters.probes.readinessProbeSpec.periodSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.syncmasters.probes.readinessProbeSpec.successThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L226)

### .spec.syncmasters.probes.readinessProbeSpec.timeoutSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L225)

### .spec.syncmasters.probes.startupProbeDisabled: bool

StartupProbeDisabled if true startupProbes are disabled

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L207)

### .spec.syncmasters.probes.startupProbeSpec.failureThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L227)

### .spec.syncmasters.probes.startupProbeSpec.initialDelaySeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L223)

### .spec.syncmasters.probes.startupProbeSpec.periodSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.syncmasters.probes.startupProbeSpec.successThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L226)

### .spec.syncmasters.probes.startupProbeSpec.timeoutSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L225)

### .spec.syncmasters.pvcResizeMode: string

VolumeResizeMode specified resize mode for pvc

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L124)

### .spec.syncmasters.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L88)

### .spec.syncmasters.schedulerName: string

SchedulerName define scheduler name used for group

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L82)

### .spec.syncmasters.securityContext.addCapabilities: []string

AddCapabilities add new capabilities to containers

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L45)

### .spec.syncmasters.securityContext.allowPrivilegeEscalation: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L47)

### .spec.syncmasters.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L43)

### .spec.syncmasters.securityContext.fsGroup: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L55)

### .spec.syncmasters.securityContext.privileged: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L48)

### .spec.syncmasters.securityContext.readOnlyRootFilesystem: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)

### .spec.syncmasters.securityContext.runAsGroup: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L52)

### .spec.syncmasters.securityContext.runAsNonRoot: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L50)

### .spec.syncmasters.securityContext.runAsUser: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L51)

### .spec.syncmasters.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L71)

### .spec.syncmasters.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L76)

### .spec.syncmasters.securityContext.supplementalGroups: []int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L54)

### .spec.syncmasters.securityContext.sysctls: map[string]intstr.IntOrString

Sysctls hold a list of namespaced sysctls used for the pod. Pods with unsupported
sysctls (by the container runtime) might fail to launch.
Map Value can be String or Int

Links:
* [Documentation](https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster/)

Example:
```yaml
sysctls:
  "kernel.shm_rmid_forced": "0"
  "net.core.somaxconn": 1024
  "kernel.msgmax": "65536"
```

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L66)

### .spec.syncmasters.serviceAccountName: string

ServiceAccountName specifies the name of the service account used for Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L112)

### .spec.syncmasters.shutdownDelay: int

ShutdownDelay define how long operator should delay finalizer removal after shutdown

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L163)

### .spec.syncmasters.shutdownMethod: string

ShutdownMethod describe procedure of member shutdown taken by Operator

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L161)

### .spec.syncmasters.sidecarCoreNames: []string

SidecarCoreNames is a list of sidecar containers which must run in the pod.
Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L141)

### .spec.syncmasters.sidecars: []core.Container

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L145)

### .spec.syncmasters.storageClassName: string

StorageClassName specifies the classname for storage of the servers.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L84)

### .spec.syncmasters.terminationGracePeriodSeconds: int64

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L173)

### .spec.syncmasters.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L96)

### .spec.syncmasters.volumeAllowShrink: bool

Deprecated: VolumeAllowShrink allows shrink the volume

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L126)

### .spec.syncmasters.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a template for volume claims

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L122)

### .spec.syncmasters.volumeMounts: []ServerGroupSpecVolumeMount

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation of ServerGroupSpecVolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L153)

### .spec.syncmasters.volumes\[int\].configMap: core.ConfigMapVolumeSource

ConfigMap which should be mounted into pod

Links:
* [Documentation of core.ConfigMapVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#configmapvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L138)

### .spec.syncmasters.volumes\[int\].emptyDir: core.EmptyDirVolumeSource

EmptyDir

Links:
* [Documentation of core.EmptyDirVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#emptydirvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L143)

### .spec.syncmasters.volumes\[int\].hostPath: core.HostPathVolumeSource

HostPath

Links:
* [Documentation of core.HostPathVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#hostpathvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L148)

### .spec.syncmasters.volumes\[int\].name: string

Name of volume

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L128)

### .spec.syncmasters.volumes\[int\].persistentVolumeClaim: core.PersistentVolumeClaimVolumeSource

PersistentVolumeClaim

Links:
* [Documentation of core.PersistentVolumeClaimVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaimvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L153)

### .spec.syncmasters.volumes\[int\].secret: core.SecretVolumeSource

Secret which should be mounted into pod

Links:
* [Documentation of core.SecretVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#secretvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L133)

### .spec.syncworkers.affinity: core.PodAffinity

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.PodAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L134)

### .spec.syncworkers.allowMemberRecreation: bool

AllowMemberRecreation allows to recreate member. Value is used only for Coordinator and DBServer with default to True, for all other groups set to false.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L171)

### .spec.syncworkers.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L98)

### .spec.syncworkers.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L100)

### .spec.syncworkers.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L102)

### .spec.syncworkers.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L130)

### .spec.syncworkers.args: []string

Args holds additional commandline arguments

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L78)

### .spec.syncworkers.count: int

Count holds the requested number of servers

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L72)

### .spec.syncworkers.entrypoint: string

Entrypoint overrides container executable

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L80)

### .spec.syncworkers.envs\[int\].name: string

[Code Reference](/pkg/apis/deployment/v1/server_group_env_var.go#L26)

### .spec.syncworkers.envs\[int\].value: string

[Code Reference](/pkg/apis/deployment/v1/server_group_env_var.go#L27)

### .spec.syncworkers.ephemeralVolumes.apps.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)

### .spec.syncworkers.ephemeralVolumes.temp.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)

### .spec.syncworkers.exporterPort: uint16

ExporterPort define Port used by exporter

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L184)

### .spec.syncworkers.extendedRotationCheck: bool

ExtendedRotationCheck extend checks for rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L157)

### .spec.syncworkers.externalPortEnabled: bool

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L169)

### .spec.syncworkers.indexMethod: string

IndexMethod define group Indexing method

Possible Values: 
* random (default) - Pick random ID for member. Enforced on the Community Operator.
* ordered - Use sequential number as Member ID, starting from 0. Enterprise Operator required.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L177)

### .spec.syncworkers.initContainers.containers: []core.Container

Containers contains list of containers

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_init_containers.go#L91)

### .spec.syncworkers.initContainers.mode: string

Mode keep container replace mode

[Code Reference](/pkg/apis/deployment/v1/server_group_init_containers.go#L94)

### .spec.syncworkers.internalPort: int

InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L165)

### .spec.syncworkers.internalPortProtocol: string

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L167)

### .spec.syncworkers.labels: map[string]string

Labels specified the labels added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L104)

### .spec.syncworkers.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L106)

### .spec.syncworkers.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L108)

### .spec.syncworkers.maxCount: int

MaxCount specifies a upper limit for count

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L76)

### .spec.syncworkers.minCount: int

MinCount specifies a lower limit for count

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L74)

### .spec.syncworkers.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L138)

### .spec.syncworkers.nodeSelector: map[string]string

NodeSelector speficies a set of selectors for nodes

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L114)

### .spec.syncworkers.numactl.args: []string

Args define list of the numactl process

Default Value: []

[Code Reference](/pkg/apis/deployment/v1/server_group_numactl_spec.go#L38)

### .spec.syncworkers.numactl.enabled: bool

Enabled define if numactl should be enabled

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/server_group_numactl_spec.go#L30)

### .spec.syncworkers.numactl.path: string

Path define numactl path within the container

Default Value: /usr/bin/numactl

[Code Reference](/pkg/apis/deployment/v1/server_group_numactl_spec.go#L34)

### .spec.syncworkers.overrideDetectedNumberOfCores: bool

OverrideDetectedNumberOfCores determines if number of cores should be overrided based on values in resources.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L92)

### .spec.syncworkers.overrideDetectedTotalMemory: bool

OverrideDetectedTotalMemory determines if memory should be overrided based on values in resources.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L90)

### .spec.syncworkers.podModes.network: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L31)

### .spec.syncworkers.podModes.pid: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L32)

### .spec.syncworkers.port: uint16

Port define Port used by member

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L182)

### .spec.syncworkers.priorityClassName: string

PriorityClassName specifies a priority class name

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L118)

### .spec.syncworkers.probes.livenessProbeDisabled: bool

LivenessProbeDisabled if true livenessProbes are disabled

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L193)

### .spec.syncworkers.probes.livenessProbeSpec.failureThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L227)

### .spec.syncworkers.probes.livenessProbeSpec.initialDelaySeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L223)

### .spec.syncworkers.probes.livenessProbeSpec.periodSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.syncworkers.probes.livenessProbeSpec.successThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L226)

### .spec.syncworkers.probes.livenessProbeSpec.timeoutSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L225)

### .spec.syncworkers.probes.ReadinessProbeDisabled: bool

OldReadinessProbeDisabled if true readinessProbes are disabled
Deprecated: This field is deprecated, keept only for backward compatibility.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L200)

### .spec.syncworkers.probes.readinessProbeDisabled: bool

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L202)

### .spec.syncworkers.probes.readinessProbeSpec.failureThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L227)

### .spec.syncworkers.probes.readinessProbeSpec.initialDelaySeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L223)

### .spec.syncworkers.probes.readinessProbeSpec.periodSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.syncworkers.probes.readinessProbeSpec.successThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L226)

### .spec.syncworkers.probes.readinessProbeSpec.timeoutSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L225)

### .spec.syncworkers.probes.startupProbeDisabled: bool

StartupProbeDisabled if true startupProbes are disabled

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L207)

### .spec.syncworkers.probes.startupProbeSpec.failureThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L227)

### .spec.syncworkers.probes.startupProbeSpec.initialDelaySeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L223)

### .spec.syncworkers.probes.startupProbeSpec.periodSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.syncworkers.probes.startupProbeSpec.successThreshold: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L226)

### .spec.syncworkers.probes.startupProbeSpec.timeoutSeconds: int32

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L225)

### .spec.syncworkers.pvcResizeMode: string

VolumeResizeMode specified resize mode for pvc

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L124)

### .spec.syncworkers.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L88)

### .spec.syncworkers.schedulerName: string

SchedulerName define scheduler name used for group

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L82)

### .spec.syncworkers.securityContext.addCapabilities: []string

AddCapabilities add new capabilities to containers

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L45)

### .spec.syncworkers.securityContext.allowPrivilegeEscalation: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L47)

### .spec.syncworkers.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L43)

### .spec.syncworkers.securityContext.fsGroup: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L55)

### .spec.syncworkers.securityContext.privileged: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L48)

### .spec.syncworkers.securityContext.readOnlyRootFilesystem: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)

### .spec.syncworkers.securityContext.runAsGroup: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L52)

### .spec.syncworkers.securityContext.runAsNonRoot: bool

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L50)

### .spec.syncworkers.securityContext.runAsUser: int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L51)

### .spec.syncworkers.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L71)

### .spec.syncworkers.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L76)

### .spec.syncworkers.securityContext.supplementalGroups: []int64

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L54)

### .spec.syncworkers.securityContext.sysctls: map[string]intstr.IntOrString

Sysctls hold a list of namespaced sysctls used for the pod. Pods with unsupported
sysctls (by the container runtime) might fail to launch.
Map Value can be String or Int

Links:
* [Documentation](https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster/)

Example:
```yaml
sysctls:
  "kernel.shm_rmid_forced": "0"
  "net.core.somaxconn": 1024
  "kernel.msgmax": "65536"
```

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L66)

### .spec.syncworkers.serviceAccountName: string

ServiceAccountName specifies the name of the service account used for Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L112)

### .spec.syncworkers.shutdownDelay: int

ShutdownDelay define how long operator should delay finalizer removal after shutdown

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L163)

### .spec.syncworkers.shutdownMethod: string

ShutdownMethod describe procedure of member shutdown taken by Operator

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L161)

### .spec.syncworkers.sidecarCoreNames: []string

SidecarCoreNames is a list of sidecar containers which must run in the pod.
Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L141)

### .spec.syncworkers.sidecars: []core.Container

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L145)

### .spec.syncworkers.storageClassName: string

StorageClassName specifies the classname for storage of the servers.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L84)

### .spec.syncworkers.terminationGracePeriodSeconds: int64

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L173)

### .spec.syncworkers.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L96)

### .spec.syncworkers.volumeAllowShrink: bool

Deprecated: VolumeAllowShrink allows shrink the volume

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L126)

### .spec.syncworkers.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a template for volume claims

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L122)

### .spec.syncworkers.volumeMounts: []ServerGroupSpecVolumeMount

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation of ServerGroupSpecVolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L153)

### .spec.syncworkers.volumes\[int\].configMap: core.ConfigMapVolumeSource

ConfigMap which should be mounted into pod

Links:
* [Documentation of core.ConfigMapVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#configmapvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L138)

### .spec.syncworkers.volumes\[int\].emptyDir: core.EmptyDirVolumeSource

EmptyDir

Links:
* [Documentation of core.EmptyDirVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#emptydirvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L143)

### .spec.syncworkers.volumes\[int\].hostPath: core.HostPathVolumeSource

HostPath

Links:
* [Documentation of core.HostPathVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#hostpathvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L148)

### .spec.syncworkers.volumes\[int\].name: string

Name of volume

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L128)

### .spec.syncworkers.volumes\[int\].persistentVolumeClaim: core.PersistentVolumeClaimVolumeSource

PersistentVolumeClaim

Links:
* [Documentation of core.PersistentVolumeClaimVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaimvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L153)

### .spec.syncworkers.volumes\[int\].secret: core.SecretVolumeSource

Secret which should be mounted into pod

Links:
* [Documentation of core.SecretVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#secretvolumesource-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_volume.go#L133)

### .spec.timeouts.actions: map[string]meta.Duration

Actions keep map of the actions timeouts.

Links:
* [List of supported action names](/docs/generated/actions.md)
* [Definition of meta.Duration](https://github.com/kubernetes/apimachinery/blob/v0.26.6/pkg/apis/meta/v1/duration.go)

Example:
```yaml
actions:
  AddMember: 30m
```

[Code Reference](/pkg/apis/deployment/v1/timeouts.go#L44)

### .spec.timeouts.maintenanceGracePeriod: int64

MaintenanceGracePeriod action timeout

[Code Reference](/pkg/apis/deployment/v1/timeouts.go#L36)

### .spec.timezone: string

Timezone if specified, will set a timezone for deployment.
Must be in format accepted by "tzdata", e.g. `America/New_York` or `Europe/London`

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L262)

### .spec.tls.altNames: []string

[Code Reference](/pkg/apis/deployment/v1/tls_spec.go#L58)

### .spec.tls.caSecretName: string

[Code Reference](/pkg/apis/deployment/v1/tls_spec.go#L57)

### .spec.tls.mode: string

[Code Reference](/pkg/apis/deployment/v1/tls_spec.go#L61)

### .spec.tls.sni.mapping.\<string\>: []string

[Code Reference](/pkg/apis/deployment/v1/tls_sni_spec.go#L30)

### .spec.tls.ttl: string

[Code Reference](/pkg/apis/deployment/v1/tls_spec.go#L59)

### .spec.topology.enabled: bool

[Code Reference](/pkg/apis/deployment/v1/topology_spec.go#L26)

### .spec.topology.label: string

[Code Reference](/pkg/apis/deployment/v1/topology_spec.go#L28)

### .spec.topology.zones: int

[Code Reference](/pkg/apis/deployment/v1/topology_spec.go#L27)

### .spec.upgrade.autoUpgrade: bool

AutoUpgrade flag specifies if upgrade should be auto-injected, even if is not required (in case of stuck)

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/deployment_upgrade_spec.go#L26)

### .spec.upgrade.debugLog: bool

DebugLog flag specifies if containers running upgrade process should print more debugging information.
This applies only to init containers.

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/deployment_upgrade_spec.go#L30)

