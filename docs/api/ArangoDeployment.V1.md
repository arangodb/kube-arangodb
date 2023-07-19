# API Reference for ArangoDeployment V1

## Spec

### .spec.agents.[]envs.name: string

### .spec.agents.[]envs.value: string

### .spec.agents.[]volumes.configMap: core.ConfigMapVolumeSource

ConfigMap which should be mounted into pod

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#configmapvolumesource-v1-core)

### .spec.agents.[]volumes.emptyDir: core.EmptyDirVolumeSource

EmptyDir

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#emptydirvolumesource-v1-core)

### .spec.agents.[]volumes.hostPath: core.HostPathVolumeSource

HostPath

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#hostpathvolumesource-v1-core)

### .spec.agents.[]volumes.name: string

Name of volume

### .spec.agents.[]volumes.persistentVolumeClaim: core.PersistentVolumeClaimVolumeSource

PersistentVolumeClaim

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaimvolumesource-v1-core)

### .spec.agents.[]volumes.secret: core.SecretVolumeSource

Secret which should be mounted into pod

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#secretvolumesource-v1-core)

### .spec.agents.affinity: core.PodAffinity

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

### .spec.agents.allowMemberRecreation: bool

AllowMemberRecreation allows to recreate member. Value is used only for Coordinator and DBServer with default to True, for all other groups set to false.

### .spec.agents.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.

### .spec.agents.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

### .spec.agents.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

### .spec.agents.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

### .spec.agents.args: []string

Args holds additional commandline arguments

### .spec.agents.count: int

Count holds the requested number of servers

### .spec.agents.entrypoint: string

Entrypoint overrides container executable

### .spec.agents.ephemeralVolumes.apps.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

### .spec.agents.ephemeralVolumes.temp.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

### .spec.agents.exporterPort: uint16

ExporterPort define Port used by exporter

### .spec.agents.extendedRotationCheck: bool

ExtendedRotationCheck extend checks for rotation

### .spec.agents.externalPortEnabled: bool

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

### .spec.agents.indexMethod: string

IndexMethod define group Indexing method

### .spec.agents.initContainers.containers: []core.Container

Containers contains list of containers

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

### .spec.agents.initContainers.mode: string

Mode keep container replace mode

### .spec.agents.internalPort: int

InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

### .spec.agents.internalPortProtocol: string

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

### .spec.agents.labels: map[string]string

Labels specified the labels added to Pods in this group.

### .spec.agents.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

### .spec.agents.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

### .spec.agents.maxCount: int

MaxCount specifies a upper limit for count

### .spec.agents.minCount: int

MinCount specifies a lower limit for count

### .spec.agents.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#weightedpodaffinityterm-v1-core)

### .spec.agents.nodeSelector: map[string]string

NodeSelector speficies a set of selectors for nodes

### .spec.agents.overrideDetectedNumberOfCores: bool

OverrideDetectedNumberOfCores determines if number of cores should be overrided based on values in resources.

### .spec.agents.overrideDetectedTotalMemory: bool

OverrideDetectedTotalMemory determines if memory should be overrided based on values in resources.

### .spec.agents.podModes.network: string

### .spec.agents.podModes.pid: string

### .spec.agents.port: uint16

Port define Port used by member

### .spec.agents.priorityClassName: string

PriorityClassName specifies a priority class name

### .spec.agents.probes.livenessProbeDisabled: bool

LivenessProbeDisabled if true livenessProbes are disabled

### .spec.agents.probes.livenessProbeSpec.failureThreshold: int32

### .spec.agents.probes.livenessProbeSpec.initialDelaySeconds: int32

### .spec.agents.probes.livenessProbeSpec.periodSeconds: int32

### .spec.agents.probes.livenessProbeSpec.successThreshold: int32

### .spec.agents.probes.livenessProbeSpec.timeoutSeconds: int32

### .spec.agents.probes.ReadinessProbeDisabled: bool

OldReadinessProbeDisabled if true readinessProbes are disabled

Deprecated: This field is deprecated, keept only for backward compatibility.

### .spec.agents.probes.readinessProbeDisabled: bool

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

### .spec.agents.probes.readinessProbeSpec.failureThreshold: int32

### .spec.agents.probes.readinessProbeSpec.initialDelaySeconds: int32

### .spec.agents.probes.readinessProbeSpec.periodSeconds: int32

### .spec.agents.probes.readinessProbeSpec.successThreshold: int32

### .spec.agents.probes.readinessProbeSpec.timeoutSeconds: int32

### .spec.agents.probes.startupProbeDisabled: bool

StartupProbeDisabled if true startupProbes are disabled

### .spec.agents.probes.startupProbeSpec.failureThreshold: int32

### .spec.agents.probes.startupProbeSpec.initialDelaySeconds: int32

### .spec.agents.probes.startupProbeSpec.periodSeconds: int32

### .spec.agents.probes.startupProbeSpec.successThreshold: int32

### .spec.agents.probes.startupProbeSpec.timeoutSeconds: int32

### .spec.agents.pvcResizeMode: string

VolumeResizeMode specified resize mode for pvc

### .spec.agents.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

### .spec.agents.schedulerName: string

SchedulerName define scheduler name used for group

### .spec.agents.securityContext.addCapabilities: []string

AddCapabilities add new capabilities to containers

### .spec.agents.securityContext.allowPrivilegeEscalation: bool

### .spec.agents.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers

Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

### .spec.agents.securityContext.fsGroup: int64

### .spec.agents.securityContext.privileged: bool

### .spec.agents.securityContext.readOnlyRootFilesystem: bool

### .spec.agents.securityContext.runAsGroup: int64

### .spec.agents.securityContext.runAsNonRoot: bool

### .spec.agents.securityContext.runAsUser: int64

### .spec.agents.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

### .spec.agents.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

### .spec.agents.securityContext.supplementalGroups: []int64

### .spec.agents.serviceAccountName: string

ServiceAccountName specifies the name of the service account used for Pods in this group.

### .spec.agents.shutdownDelay: int

ShutdownDelay define how long operator should delay finalizer removal after shutdown

### .spec.agents.shutdownMethod: string

ShutdownMethod describe procedure of member shutdown taken by Operator

### .spec.agents.sidecarCoreNames: []string

SidecarCoreNames is a list of sidecar containers which must run in the pod.

Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

### .spec.agents.sidecars: []core.Container

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

### .spec.agents.storageClassName: string

StorageClassName specifies the classname for storage of the servers.

### .spec.agents.terminationGracePeriodSeconds: int64

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

### .spec.agents.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

### .spec.agents.volumeAllowShrink: bool

Deprecated: VolumeAllowShrink allows shrink the volume

### .spec.agents.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a template for volume claims

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

### .spec.agents.volumeMounts: []ServerGroupSpecVolumeMount

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

### .spec.allowUnsafeUpgrade: bool

AllowUnsafeUpgrade determines if upgrade on missing member or with not in sync shards is allowed

### .spec.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.

### .spec.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

### .spec.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

### .spec.architecture: []string

Architecture definition of supported architectures

### .spec.auth.jwtSecretName: string

### .spec.bootstrap.passwordSecretNames: map[string]string

PasswordSecretNames contains a map of username to password-secret-name

### .spec.chaos.enabled: bool

Enabled switches the chaos monkey for a deployment on or off.

### .spec.chaos.interval: int64

Interval is the time between events

### .spec.chaos.kill-pod-probability: int

KillPodProbability is the chance of a pod being killed during an event

### .spec.ClusterDomain: string

ClusterDomain define domain used in the kubernetes cluster.

Required only of domain is not set to default (cluster.local)

Default Value: cluster.local

### .spec.communicationMethod: string

CommunicationMethod define communication method used in deployment

### .spec.coordinators.[]envs.name: string

### .spec.coordinators.[]envs.value: string

### .spec.coordinators.[]volumes.configMap: core.ConfigMapVolumeSource

ConfigMap which should be mounted into pod

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#configmapvolumesource-v1-core)

### .spec.coordinators.[]volumes.emptyDir: core.EmptyDirVolumeSource

EmptyDir

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#emptydirvolumesource-v1-core)

### .spec.coordinators.[]volumes.hostPath: core.HostPathVolumeSource

HostPath

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#hostpathvolumesource-v1-core)

### .spec.coordinators.[]volumes.name: string

Name of volume

### .spec.coordinators.[]volumes.persistentVolumeClaim: core.PersistentVolumeClaimVolumeSource

PersistentVolumeClaim

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaimvolumesource-v1-core)

### .spec.coordinators.[]volumes.secret: core.SecretVolumeSource

Secret which should be mounted into pod

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#secretvolumesource-v1-core)

### .spec.coordinators.affinity: core.PodAffinity

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

### .spec.coordinators.allowMemberRecreation: bool

AllowMemberRecreation allows to recreate member. Value is used only for Coordinator and DBServer with default to True, for all other groups set to false.

### .spec.coordinators.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.

### .spec.coordinators.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

### .spec.coordinators.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

### .spec.coordinators.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

### .spec.coordinators.args: []string

Args holds additional commandline arguments

### .spec.coordinators.count: int

Count holds the requested number of servers

### .spec.coordinators.entrypoint: string

Entrypoint overrides container executable

### .spec.coordinators.ephemeralVolumes.apps.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

### .spec.coordinators.ephemeralVolumes.temp.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

### .spec.coordinators.exporterPort: uint16

ExporterPort define Port used by exporter

### .spec.coordinators.extendedRotationCheck: bool

ExtendedRotationCheck extend checks for rotation

### .spec.coordinators.externalPortEnabled: bool

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

### .spec.coordinators.indexMethod: string

IndexMethod define group Indexing method

### .spec.coordinators.initContainers.containers: []core.Container

Containers contains list of containers

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

### .spec.coordinators.initContainers.mode: string

Mode keep container replace mode

### .spec.coordinators.internalPort: int

InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

### .spec.coordinators.internalPortProtocol: string

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

### .spec.coordinators.labels: map[string]string

Labels specified the labels added to Pods in this group.

### .spec.coordinators.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

### .spec.coordinators.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

### .spec.coordinators.maxCount: int

MaxCount specifies a upper limit for count

### .spec.coordinators.minCount: int

MinCount specifies a lower limit for count

### .spec.coordinators.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#weightedpodaffinityterm-v1-core)

### .spec.coordinators.nodeSelector: map[string]string

NodeSelector speficies a set of selectors for nodes

### .spec.coordinators.overrideDetectedNumberOfCores: bool

OverrideDetectedNumberOfCores determines if number of cores should be overrided based on values in resources.

### .spec.coordinators.overrideDetectedTotalMemory: bool

OverrideDetectedTotalMemory determines if memory should be overrided based on values in resources.

### .spec.coordinators.podModes.network: string

### .spec.coordinators.podModes.pid: string

### .spec.coordinators.port: uint16

Port define Port used by member

### .spec.coordinators.priorityClassName: string

PriorityClassName specifies a priority class name

### .spec.coordinators.probes.livenessProbeDisabled: bool

LivenessProbeDisabled if true livenessProbes are disabled

### .spec.coordinators.probes.livenessProbeSpec.failureThreshold: int32

### .spec.coordinators.probes.livenessProbeSpec.initialDelaySeconds: int32

### .spec.coordinators.probes.livenessProbeSpec.periodSeconds: int32

### .spec.coordinators.probes.livenessProbeSpec.successThreshold: int32

### .spec.coordinators.probes.livenessProbeSpec.timeoutSeconds: int32

### .spec.coordinators.probes.readinessProbeDisabled: bool

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

### .spec.coordinators.probes.ReadinessProbeDisabled: bool

OldReadinessProbeDisabled if true readinessProbes are disabled

Deprecated: This field is deprecated, keept only for backward compatibility.

### .spec.coordinators.probes.readinessProbeSpec.failureThreshold: int32

### .spec.coordinators.probes.readinessProbeSpec.initialDelaySeconds: int32

### .spec.coordinators.probes.readinessProbeSpec.periodSeconds: int32

### .spec.coordinators.probes.readinessProbeSpec.successThreshold: int32

### .spec.coordinators.probes.readinessProbeSpec.timeoutSeconds: int32

### .spec.coordinators.probes.startupProbeDisabled: bool

StartupProbeDisabled if true startupProbes are disabled

### .spec.coordinators.probes.startupProbeSpec.failureThreshold: int32

### .spec.coordinators.probes.startupProbeSpec.initialDelaySeconds: int32

### .spec.coordinators.probes.startupProbeSpec.periodSeconds: int32

### .spec.coordinators.probes.startupProbeSpec.successThreshold: int32

### .spec.coordinators.probes.startupProbeSpec.timeoutSeconds: int32

### .spec.coordinators.pvcResizeMode: string

VolumeResizeMode specified resize mode for pvc

### .spec.coordinators.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

### .spec.coordinators.schedulerName: string

SchedulerName define scheduler name used for group

### .spec.coordinators.securityContext.addCapabilities: []string

AddCapabilities add new capabilities to containers

### .spec.coordinators.securityContext.allowPrivilegeEscalation: bool

### .spec.coordinators.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers

Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

### .spec.coordinators.securityContext.fsGroup: int64

### .spec.coordinators.securityContext.privileged: bool

### .spec.coordinators.securityContext.readOnlyRootFilesystem: bool

### .spec.coordinators.securityContext.runAsGroup: int64

### .spec.coordinators.securityContext.runAsNonRoot: bool

### .spec.coordinators.securityContext.runAsUser: int64

### .spec.coordinators.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

### .spec.coordinators.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

### .spec.coordinators.securityContext.supplementalGroups: []int64

### .spec.coordinators.serviceAccountName: string

ServiceAccountName specifies the name of the service account used for Pods in this group.

### .spec.coordinators.shutdownDelay: int

ShutdownDelay define how long operator should delay finalizer removal after shutdown

### .spec.coordinators.shutdownMethod: string

ShutdownMethod describe procedure of member shutdown taken by Operator

### .spec.coordinators.sidecarCoreNames: []string

SidecarCoreNames is a list of sidecar containers which must run in the pod.

Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

### .spec.coordinators.sidecars: []core.Container

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

### .spec.coordinators.storageClassName: string

StorageClassName specifies the classname for storage of the servers.

### .spec.coordinators.terminationGracePeriodSeconds: int64

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

### .spec.coordinators.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

### .spec.coordinators.volumeAllowShrink: bool

Deprecated: VolumeAllowShrink allows shrink the volume

### .spec.coordinators.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a template for volume claims

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

### .spec.coordinators.volumeMounts: []ServerGroupSpecVolumeMount

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

### .spec.database.maintenance: bool

Maintenance manage maintenance mode on Cluster side. Requires maintenance feature to be enabled

### .spec.dbservers.[]envs.name: string

### .spec.dbservers.[]envs.value: string

### .spec.dbservers.[]volumes.configMap: core.ConfigMapVolumeSource

ConfigMap which should be mounted into pod

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#configmapvolumesource-v1-core)

### .spec.dbservers.[]volumes.emptyDir: core.EmptyDirVolumeSource

EmptyDir

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#emptydirvolumesource-v1-core)

### .spec.dbservers.[]volumes.hostPath: core.HostPathVolumeSource

HostPath

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#hostpathvolumesource-v1-core)

### .spec.dbservers.[]volumes.name: string

Name of volume

### .spec.dbservers.[]volumes.persistentVolumeClaim: core.PersistentVolumeClaimVolumeSource

PersistentVolumeClaim

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaimvolumesource-v1-core)

### .spec.dbservers.[]volumes.secret: core.SecretVolumeSource

Secret which should be mounted into pod

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#secretvolumesource-v1-core)

### .spec.dbservers.affinity: core.PodAffinity

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

### .spec.dbservers.allowMemberRecreation: bool

AllowMemberRecreation allows to recreate member. Value is used only for Coordinator and DBServer with default to True, for all other groups set to false.

### .spec.dbservers.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.

### .spec.dbservers.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

### .spec.dbservers.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

### .spec.dbservers.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

### .spec.dbservers.args: []string

Args holds additional commandline arguments

### .spec.dbservers.count: int

Count holds the requested number of servers

### .spec.dbservers.entrypoint: string

Entrypoint overrides container executable

### .spec.dbservers.ephemeralVolumes.apps.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

### .spec.dbservers.ephemeralVolumes.temp.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

### .spec.dbservers.exporterPort: uint16

ExporterPort define Port used by exporter

### .spec.dbservers.extendedRotationCheck: bool

ExtendedRotationCheck extend checks for rotation

### .spec.dbservers.externalPortEnabled: bool

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

### .spec.dbservers.indexMethod: string

IndexMethod define group Indexing method

### .spec.dbservers.initContainers.containers: []core.Container

Containers contains list of containers

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

### .spec.dbservers.initContainers.mode: string

Mode keep container replace mode

### .spec.dbservers.internalPort: int

InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

### .spec.dbservers.internalPortProtocol: string

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

### .spec.dbservers.labels: map[string]string

Labels specified the labels added to Pods in this group.

### .spec.dbservers.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

### .spec.dbservers.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

### .spec.dbservers.maxCount: int

MaxCount specifies a upper limit for count

### .spec.dbservers.minCount: int

MinCount specifies a lower limit for count

### .spec.dbservers.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#weightedpodaffinityterm-v1-core)

### .spec.dbservers.nodeSelector: map[string]string

NodeSelector speficies a set of selectors for nodes

### .spec.dbservers.overrideDetectedNumberOfCores: bool

OverrideDetectedNumberOfCores determines if number of cores should be overrided based on values in resources.

### .spec.dbservers.overrideDetectedTotalMemory: bool

OverrideDetectedTotalMemory determines if memory should be overrided based on values in resources.

### .spec.dbservers.podModes.network: string

### .spec.dbservers.podModes.pid: string

### .spec.dbservers.port: uint16

Port define Port used by member

### .spec.dbservers.priorityClassName: string

PriorityClassName specifies a priority class name

### .spec.dbservers.probes.livenessProbeDisabled: bool

LivenessProbeDisabled if true livenessProbes are disabled

### .spec.dbservers.probes.livenessProbeSpec.failureThreshold: int32

### .spec.dbservers.probes.livenessProbeSpec.initialDelaySeconds: int32

### .spec.dbservers.probes.livenessProbeSpec.periodSeconds: int32

### .spec.dbservers.probes.livenessProbeSpec.successThreshold: int32

### .spec.dbservers.probes.livenessProbeSpec.timeoutSeconds: int32

### .spec.dbservers.probes.ReadinessProbeDisabled: bool

OldReadinessProbeDisabled if true readinessProbes are disabled

Deprecated: This field is deprecated, keept only for backward compatibility.

### .spec.dbservers.probes.readinessProbeDisabled: bool

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

### .spec.dbservers.probes.readinessProbeSpec.failureThreshold: int32

### .spec.dbservers.probes.readinessProbeSpec.initialDelaySeconds: int32

### .spec.dbservers.probes.readinessProbeSpec.periodSeconds: int32

### .spec.dbservers.probes.readinessProbeSpec.successThreshold: int32

### .spec.dbservers.probes.readinessProbeSpec.timeoutSeconds: int32

### .spec.dbservers.probes.startupProbeDisabled: bool

StartupProbeDisabled if true startupProbes are disabled

### .spec.dbservers.probes.startupProbeSpec.failureThreshold: int32

### .spec.dbservers.probes.startupProbeSpec.initialDelaySeconds: int32

### .spec.dbservers.probes.startupProbeSpec.periodSeconds: int32

### .spec.dbservers.probes.startupProbeSpec.successThreshold: int32

### .spec.dbservers.probes.startupProbeSpec.timeoutSeconds: int32

### .spec.dbservers.pvcResizeMode: string

VolumeResizeMode specified resize mode for pvc

### .spec.dbservers.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

### .spec.dbservers.schedulerName: string

SchedulerName define scheduler name used for group

### .spec.dbservers.securityContext.addCapabilities: []string

AddCapabilities add new capabilities to containers

### .spec.dbservers.securityContext.allowPrivilegeEscalation: bool

### .spec.dbservers.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers

Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

### .spec.dbservers.securityContext.fsGroup: int64

### .spec.dbservers.securityContext.privileged: bool

### .spec.dbservers.securityContext.readOnlyRootFilesystem: bool

### .spec.dbservers.securityContext.runAsGroup: int64

### .spec.dbservers.securityContext.runAsNonRoot: bool

### .spec.dbservers.securityContext.runAsUser: int64

### .spec.dbservers.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

### .spec.dbservers.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

### .spec.dbservers.securityContext.supplementalGroups: []int64

### .spec.dbservers.serviceAccountName: string

ServiceAccountName specifies the name of the service account used for Pods in this group.

### .spec.dbservers.shutdownDelay: int

ShutdownDelay define how long operator should delay finalizer removal after shutdown

### .spec.dbservers.shutdownMethod: string

ShutdownMethod describe procedure of member shutdown taken by Operator

### .spec.dbservers.sidecarCoreNames: []string

SidecarCoreNames is a list of sidecar containers which must run in the pod.

Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

### .spec.dbservers.sidecars: []core.Container

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

### .spec.dbservers.storageClassName: string

StorageClassName specifies the classname for storage of the servers.

### .spec.dbservers.terminationGracePeriodSeconds: int64

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

### .spec.dbservers.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

### .spec.dbservers.volumeAllowShrink: bool

Deprecated: VolumeAllowShrink allows shrink the volume

### .spec.dbservers.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a template for volume claims

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

### .spec.dbservers.volumeMounts: []ServerGroupSpecVolumeMount

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

### .spec.disableIPv6: bool

### .spec.downtimeAllowed: bool

### .spec.environment: string

### .spec.externalAccess.advertisedEndpoint: string

AdvertisedEndpoint is passed to the coordinators/single servers for advertising a specific endpoint

### .spec.externalAccess.loadBalancerIP: string

Optional IP used to configure a load-balancer on, in case of Auto or LoadBalancer type.

### .spec.externalAccess.loadBalancerSourceRanges: []string

If specified and supported by the platform, this will restrict traffic through the cloud-provider

load-balancer will be restricted to the specified client IPs. This field will be ignored if the

cloud-provider does not support the feature.

More info: https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/

### .spec.externalAccess.managedServiceNames: []string

ManagedServiceNames keeps names of services which are not managed by KubeArangoDB.

It is only relevant when type of service is `managed`.

### .spec.externalAccess.nodePort: int

Optional port used in case of Auto or NodePort type.

### .spec.externalAccess.type: string

Type of external access

### .spec.features.foxx.queues: bool

### .spec.id.affinity: core.PodAffinity

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

### .spec.id.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

### .spec.id.entrypoint: string

Entrypoint overrides container executable

### .spec.id.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#weightedpodaffinityterm-v1-core)

### .spec.id.nodeSelector: map[string]string

NodeSelector speficies a set of selectors for nodes

### .spec.id.priorityClassName: string

PriorityClassName specifies a priority class name

### .spec.id.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

### .spec.id.securityContext.addCapabilities: []string

AddCapabilities add new capabilities to containers

### .spec.id.securityContext.allowPrivilegeEscalation: bool

### .spec.id.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers

Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

### .spec.id.securityContext.fsGroup: int64

### .spec.id.securityContext.privileged: bool

### .spec.id.securityContext.readOnlyRootFilesystem: bool

### .spec.id.securityContext.runAsGroup: int64

### .spec.id.securityContext.runAsNonRoot: bool

### .spec.id.securityContext.runAsUser: int64

### .spec.id.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

### .spec.id.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

### .spec.id.securityContext.supplementalGroups: []int64

### .spec.id.serviceAccountName: string

ServiceAccountName specifies the name of the service account used for Pods in this group.

### .spec.id.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

### .spec.image: string

### .spec.imageDiscoveryMode: string

### .spec.imagePullPolicy: string

### .spec.imagePullSecrets: []string

### .spec.labels: map[string]string

Labels specified the labels added to Pods in this group.

### .spec.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

### .spec.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

### .spec.license.secretName: string

### .spec.lifecycle.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

### .spec.memberPropagationMode: string

### .spec.metrics.authentication.jwtTokenSecretName: string

JWTTokenSecretName contains the name of the JWT kubernetes secret used for authentication

### .spec.metrics.enabled: bool

### .spec.metrics.image: string

deprecated

### .spec.metrics.mode: string

deprecated

### .spec.metrics.port: uint16

### .spec.metrics.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

### .spec.metrics.serviceMonitor.enabled: bool

### .spec.metrics.serviceMonitor.labels: map[string]string

### .spec.metrics.tls: bool

### .spec.mode: string

### .spec.networkAttachedVolumes: bool

### .spec.rebalancer.enabled: bool

### .spec.rebalancer.optimizers.leader: bool

### .spec.rebalancer.parallelMoves: int

### .spec.rebalancer.readers.count: bool

deprecated does not work in Rebalancer V2

Count Enable Shard Count machanism

### .spec.recovery.autoRecover: bool

### .spec.restoreEncryptionSecret: string

### .spec.restoreFrom: string

### .spec.rocksdb.encryption.keySecretName: string

### .spec.single.[]envs.name: string

### .spec.single.[]envs.value: string

### .spec.single.[]volumes.configMap: core.ConfigMapVolumeSource

ConfigMap which should be mounted into pod

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#configmapvolumesource-v1-core)

### .spec.single.[]volumes.emptyDir: core.EmptyDirVolumeSource

EmptyDir

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#emptydirvolumesource-v1-core)

### .spec.single.[]volumes.hostPath: core.HostPathVolumeSource

HostPath

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#hostpathvolumesource-v1-core)

### .spec.single.[]volumes.name: string

Name of volume

### .spec.single.[]volumes.persistentVolumeClaim: core.PersistentVolumeClaimVolumeSource

PersistentVolumeClaim

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaimvolumesource-v1-core)

### .spec.single.[]volumes.secret: core.SecretVolumeSource

Secret which should be mounted into pod

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#secretvolumesource-v1-core)

### .spec.single.affinity: core.PodAffinity

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

### .spec.single.allowMemberRecreation: bool

AllowMemberRecreation allows to recreate member. Value is used only for Coordinator and DBServer with default to True, for all other groups set to false.

### .spec.single.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.

### .spec.single.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

### .spec.single.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

### .spec.single.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

### .spec.single.args: []string

Args holds additional commandline arguments

### .spec.single.count: int

Count holds the requested number of servers

### .spec.single.entrypoint: string

Entrypoint overrides container executable

### .spec.single.ephemeralVolumes.apps.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

### .spec.single.ephemeralVolumes.temp.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

### .spec.single.exporterPort: uint16

ExporterPort define Port used by exporter

### .spec.single.extendedRotationCheck: bool

ExtendedRotationCheck extend checks for rotation

### .spec.single.externalPortEnabled: bool

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

### .spec.single.indexMethod: string

IndexMethod define group Indexing method

### .spec.single.initContainers.containers: []core.Container

Containers contains list of containers

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

### .spec.single.initContainers.mode: string

Mode keep container replace mode

### .spec.single.internalPort: int

InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

### .spec.single.internalPortProtocol: string

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

### .spec.single.labels: map[string]string

Labels specified the labels added to Pods in this group.

### .spec.single.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

### .spec.single.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

### .spec.single.maxCount: int

MaxCount specifies a upper limit for count

### .spec.single.minCount: int

MinCount specifies a lower limit for count

### .spec.single.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#weightedpodaffinityterm-v1-core)

### .spec.single.nodeSelector: map[string]string

NodeSelector speficies a set of selectors for nodes

### .spec.single.overrideDetectedNumberOfCores: bool

OverrideDetectedNumberOfCores determines if number of cores should be overrided based on values in resources.

### .spec.single.overrideDetectedTotalMemory: bool

OverrideDetectedTotalMemory determines if memory should be overrided based on values in resources.

### .spec.single.podModes.network: string

### .spec.single.podModes.pid: string

### .spec.single.port: uint16

Port define Port used by member

### .spec.single.priorityClassName: string

PriorityClassName specifies a priority class name

### .spec.single.probes.livenessProbeDisabled: bool

LivenessProbeDisabled if true livenessProbes are disabled

### .spec.single.probes.livenessProbeSpec.failureThreshold: int32

### .spec.single.probes.livenessProbeSpec.initialDelaySeconds: int32

### .spec.single.probes.livenessProbeSpec.periodSeconds: int32

### .spec.single.probes.livenessProbeSpec.successThreshold: int32

### .spec.single.probes.livenessProbeSpec.timeoutSeconds: int32

### .spec.single.probes.ReadinessProbeDisabled: bool

OldReadinessProbeDisabled if true readinessProbes are disabled

Deprecated: This field is deprecated, keept only for backward compatibility.

### .spec.single.probes.readinessProbeDisabled: bool

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

### .spec.single.probes.readinessProbeSpec.failureThreshold: int32

### .spec.single.probes.readinessProbeSpec.initialDelaySeconds: int32

### .spec.single.probes.readinessProbeSpec.periodSeconds: int32

### .spec.single.probes.readinessProbeSpec.successThreshold: int32

### .spec.single.probes.readinessProbeSpec.timeoutSeconds: int32

### .spec.single.probes.startupProbeDisabled: bool

StartupProbeDisabled if true startupProbes are disabled

### .spec.single.probes.startupProbeSpec.failureThreshold: int32

### .spec.single.probes.startupProbeSpec.initialDelaySeconds: int32

### .spec.single.probes.startupProbeSpec.periodSeconds: int32

### .spec.single.probes.startupProbeSpec.successThreshold: int32

### .spec.single.probes.startupProbeSpec.timeoutSeconds: int32

### .spec.single.pvcResizeMode: string

VolumeResizeMode specified resize mode for pvc

### .spec.single.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

### .spec.single.schedulerName: string

SchedulerName define scheduler name used for group

### .spec.single.securityContext.addCapabilities: []string

AddCapabilities add new capabilities to containers

### .spec.single.securityContext.allowPrivilegeEscalation: bool

### .spec.single.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers

Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

### .spec.single.securityContext.fsGroup: int64

### .spec.single.securityContext.privileged: bool

### .spec.single.securityContext.readOnlyRootFilesystem: bool

### .spec.single.securityContext.runAsGroup: int64

### .spec.single.securityContext.runAsNonRoot: bool

### .spec.single.securityContext.runAsUser: int64

### .spec.single.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

### .spec.single.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

### .spec.single.securityContext.supplementalGroups: []int64

### .spec.single.serviceAccountName: string

ServiceAccountName specifies the name of the service account used for Pods in this group.

### .spec.single.shutdownDelay: int

ShutdownDelay define how long operator should delay finalizer removal after shutdown

### .spec.single.shutdownMethod: string

ShutdownMethod describe procedure of member shutdown taken by Operator

### .spec.single.sidecarCoreNames: []string

SidecarCoreNames is a list of sidecar containers which must run in the pod.

Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

### .spec.single.sidecars: []core.Container

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

### .spec.single.storageClassName: string

StorageClassName specifies the classname for storage of the servers.

### .spec.single.terminationGracePeriodSeconds: int64

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

### .spec.single.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

### .spec.single.volumeAllowShrink: bool

Deprecated: VolumeAllowShrink allows shrink the volume

### .spec.single.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a template for volume claims

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

### .spec.single.volumeMounts: []ServerGroupSpecVolumeMount

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

### .spec.storageEngine: string

### .spec.sync.auth.clientCASecretName: string

### .spec.sync.auth.jwtSecretName: string

### .spec.sync.enabled: bool

### .spec.sync.externalAccess.accessPackageSecretNames: []string

### .spec.sync.externalAccess.advertisedEndpoint: string

AdvertisedEndpoint is passed to the coordinators/single servers for advertising a specific endpoint

### .spec.sync.externalAccess.loadBalancerIP: string

Optional IP used to configure a load-balancer on, in case of Auto or LoadBalancer type.

### .spec.sync.externalAccess.loadBalancerSourceRanges: []string

If specified and supported by the platform, this will restrict traffic through the cloud-provider

load-balancer will be restricted to the specified client IPs. This field will be ignored if the

cloud-provider does not support the feature.

More info: https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/

### .spec.sync.externalAccess.managedServiceNames: []string

ManagedServiceNames keeps names of services which are not managed by KubeArangoDB.

It is only relevant when type of service is `managed`.

### .spec.sync.externalAccess.masterEndpoint: []string

### .spec.sync.externalAccess.nodePort: int

Optional port used in case of Auto or NodePort type.

### .spec.sync.externalAccess.type: string

Type of external access

### .spec.sync.image: string

### .spec.sync.monitoring.tokenSecretName: string

### .spec.sync.tls.altNames: []string

### .spec.sync.tls.caSecretName: string

### .spec.sync.tls.mode: string

### .spec.sync.tls.sni.<string>mapping: []string

### .spec.sync.tls.ttl: string

### .spec.syncmasters.[]envs.name: string

### .spec.syncmasters.[]envs.value: string

### .spec.syncmasters.[]volumes.configMap: core.ConfigMapVolumeSource

ConfigMap which should be mounted into pod

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#configmapvolumesource-v1-core)

### .spec.syncmasters.[]volumes.emptyDir: core.EmptyDirVolumeSource

EmptyDir

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#emptydirvolumesource-v1-core)

### .spec.syncmasters.[]volumes.hostPath: core.HostPathVolumeSource

HostPath

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#hostpathvolumesource-v1-core)

### .spec.syncmasters.[]volumes.name: string

Name of volume

### .spec.syncmasters.[]volumes.persistentVolumeClaim: core.PersistentVolumeClaimVolumeSource

PersistentVolumeClaim

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaimvolumesource-v1-core)

### .spec.syncmasters.[]volumes.secret: core.SecretVolumeSource

Secret which should be mounted into pod

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#secretvolumesource-v1-core)

### .spec.syncmasters.affinity: core.PodAffinity

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

### .spec.syncmasters.allowMemberRecreation: bool

AllowMemberRecreation allows to recreate member. Value is used only for Coordinator and DBServer with default to True, for all other groups set to false.

### .spec.syncmasters.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.

### .spec.syncmasters.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

### .spec.syncmasters.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

### .spec.syncmasters.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

### .spec.syncmasters.args: []string

Args holds additional commandline arguments

### .spec.syncmasters.count: int

Count holds the requested number of servers

### .spec.syncmasters.entrypoint: string

Entrypoint overrides container executable

### .spec.syncmasters.ephemeralVolumes.apps.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

### .spec.syncmasters.ephemeralVolumes.temp.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

### .spec.syncmasters.exporterPort: uint16

ExporterPort define Port used by exporter

### .spec.syncmasters.extendedRotationCheck: bool

ExtendedRotationCheck extend checks for rotation

### .spec.syncmasters.externalPortEnabled: bool

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

### .spec.syncmasters.indexMethod: string

IndexMethod define group Indexing method

### .spec.syncmasters.initContainers.containers: []core.Container

Containers contains list of containers

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

### .spec.syncmasters.initContainers.mode: string

Mode keep container replace mode

### .spec.syncmasters.internalPort: int

InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

### .spec.syncmasters.internalPortProtocol: string

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

### .spec.syncmasters.labels: map[string]string

Labels specified the labels added to Pods in this group.

### .spec.syncmasters.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

### .spec.syncmasters.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

### .spec.syncmasters.maxCount: int

MaxCount specifies a upper limit for count

### .spec.syncmasters.minCount: int

MinCount specifies a lower limit for count

### .spec.syncmasters.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#weightedpodaffinityterm-v1-core)

### .spec.syncmasters.nodeSelector: map[string]string

NodeSelector speficies a set of selectors for nodes

### .spec.syncmasters.overrideDetectedNumberOfCores: bool

OverrideDetectedNumberOfCores determines if number of cores should be overrided based on values in resources.

### .spec.syncmasters.overrideDetectedTotalMemory: bool

OverrideDetectedTotalMemory determines if memory should be overrided based on values in resources.

### .spec.syncmasters.podModes.network: string

### .spec.syncmasters.podModes.pid: string

### .spec.syncmasters.port: uint16

Port define Port used by member

### .spec.syncmasters.priorityClassName: string

PriorityClassName specifies a priority class name

### .spec.syncmasters.probes.livenessProbeDisabled: bool

LivenessProbeDisabled if true livenessProbes are disabled

### .spec.syncmasters.probes.livenessProbeSpec.failureThreshold: int32

### .spec.syncmasters.probes.livenessProbeSpec.initialDelaySeconds: int32

### .spec.syncmasters.probes.livenessProbeSpec.periodSeconds: int32

### .spec.syncmasters.probes.livenessProbeSpec.successThreshold: int32

### .spec.syncmasters.probes.livenessProbeSpec.timeoutSeconds: int32

### .spec.syncmasters.probes.readinessProbeDisabled: bool

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

### .spec.syncmasters.probes.ReadinessProbeDisabled: bool

OldReadinessProbeDisabled if true readinessProbes are disabled

Deprecated: This field is deprecated, keept only for backward compatibility.

### .spec.syncmasters.probes.readinessProbeSpec.failureThreshold: int32

### .spec.syncmasters.probes.readinessProbeSpec.initialDelaySeconds: int32

### .spec.syncmasters.probes.readinessProbeSpec.periodSeconds: int32

### .spec.syncmasters.probes.readinessProbeSpec.successThreshold: int32

### .spec.syncmasters.probes.readinessProbeSpec.timeoutSeconds: int32

### .spec.syncmasters.probes.startupProbeDisabled: bool

StartupProbeDisabled if true startupProbes are disabled

### .spec.syncmasters.probes.startupProbeSpec.failureThreshold: int32

### .spec.syncmasters.probes.startupProbeSpec.initialDelaySeconds: int32

### .spec.syncmasters.probes.startupProbeSpec.periodSeconds: int32

### .spec.syncmasters.probes.startupProbeSpec.successThreshold: int32

### .spec.syncmasters.probes.startupProbeSpec.timeoutSeconds: int32

### .spec.syncmasters.pvcResizeMode: string

VolumeResizeMode specified resize mode for pvc

### .spec.syncmasters.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

### .spec.syncmasters.schedulerName: string

SchedulerName define scheduler name used for group

### .spec.syncmasters.securityContext.addCapabilities: []string

AddCapabilities add new capabilities to containers

### .spec.syncmasters.securityContext.allowPrivilegeEscalation: bool

### .spec.syncmasters.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers

Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

### .spec.syncmasters.securityContext.fsGroup: int64

### .spec.syncmasters.securityContext.privileged: bool

### .spec.syncmasters.securityContext.readOnlyRootFilesystem: bool

### .spec.syncmasters.securityContext.runAsGroup: int64

### .spec.syncmasters.securityContext.runAsNonRoot: bool

### .spec.syncmasters.securityContext.runAsUser: int64

### .spec.syncmasters.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

### .spec.syncmasters.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

### .spec.syncmasters.securityContext.supplementalGroups: []int64

### .spec.syncmasters.serviceAccountName: string

ServiceAccountName specifies the name of the service account used for Pods in this group.

### .spec.syncmasters.shutdownDelay: int

ShutdownDelay define how long operator should delay finalizer removal after shutdown

### .spec.syncmasters.shutdownMethod: string

ShutdownMethod describe procedure of member shutdown taken by Operator

### .spec.syncmasters.sidecarCoreNames: []string

SidecarCoreNames is a list of sidecar containers which must run in the pod.

Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

### .spec.syncmasters.sidecars: []core.Container

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

### .spec.syncmasters.storageClassName: string

StorageClassName specifies the classname for storage of the servers.

### .spec.syncmasters.terminationGracePeriodSeconds: int64

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

### .spec.syncmasters.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

### .spec.syncmasters.volumeAllowShrink: bool

Deprecated: VolumeAllowShrink allows shrink the volume

### .spec.syncmasters.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a template for volume claims

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

### .spec.syncmasters.volumeMounts: []ServerGroupSpecVolumeMount

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

### .spec.syncworkers.[]envs.name: string

### .spec.syncworkers.[]envs.value: string

### .spec.syncworkers.[]volumes.configMap: core.ConfigMapVolumeSource

ConfigMap which should be mounted into pod

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#configmapvolumesource-v1-core)

### .spec.syncworkers.[]volumes.emptyDir: core.EmptyDirVolumeSource

EmptyDir

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#emptydirvolumesource-v1-core)

### .spec.syncworkers.[]volumes.hostPath: core.HostPathVolumeSource

HostPath

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#hostpathvolumesource-v1-core)

### .spec.syncworkers.[]volumes.name: string

Name of volume

### .spec.syncworkers.[]volumes.persistentVolumeClaim: core.PersistentVolumeClaimVolumeSource

PersistentVolumeClaim

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaimvolumesource-v1-core)

### .spec.syncworkers.[]volumes.secret: core.SecretVolumeSource

Secret which should be mounted into pod

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#secretvolumesource-v1-core)

### .spec.syncworkers.affinity: core.PodAffinity

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

### .spec.syncworkers.allowMemberRecreation: bool

AllowMemberRecreation allows to recreate member. Value is used only for Coordinator and DBServer with default to True, for all other groups set to false.

### .spec.syncworkers.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.

### .spec.syncworkers.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

### .spec.syncworkers.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

### .spec.syncworkers.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

### .spec.syncworkers.args: []string

Args holds additional commandline arguments

### .spec.syncworkers.count: int

Count holds the requested number of servers

### .spec.syncworkers.entrypoint: string

Entrypoint overrides container executable

### .spec.syncworkers.ephemeralVolumes.apps.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

### .spec.syncworkers.ephemeralVolumes.temp.size: resource.Quantity

Size define size of the ephemeral volume

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

### .spec.syncworkers.exporterPort: uint16

ExporterPort define Port used by exporter

### .spec.syncworkers.extendedRotationCheck: bool

ExtendedRotationCheck extend checks for rotation

### .spec.syncworkers.externalPortEnabled: bool

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

### .spec.syncworkers.indexMethod: string

IndexMethod define group Indexing method

### .spec.syncworkers.initContainers.containers: []core.Container

Containers contains list of containers

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

### .spec.syncworkers.initContainers.mode: string

Mode keep container replace mode

### .spec.syncworkers.internalPort: int

InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

### .spec.syncworkers.internalPortProtocol: string

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

### .spec.syncworkers.labels: map[string]string

Labels specified the labels added to Pods in this group.

### .spec.syncworkers.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

### .spec.syncworkers.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

### .spec.syncworkers.maxCount: int

MaxCount specifies a upper limit for count

### .spec.syncworkers.minCount: int

MinCount specifies a lower limit for count

### .spec.syncworkers.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#weightedpodaffinityterm-v1-core)

### .spec.syncworkers.nodeSelector: map[string]string

NodeSelector speficies a set of selectors for nodes

### .spec.syncworkers.overrideDetectedNumberOfCores: bool

OverrideDetectedNumberOfCores determines if number of cores should be overrided based on values in resources.

### .spec.syncworkers.overrideDetectedTotalMemory: bool

OverrideDetectedTotalMemory determines if memory should be overrided based on values in resources.

### .spec.syncworkers.podModes.network: string

### .spec.syncworkers.podModes.pid: string

### .spec.syncworkers.port: uint16

Port define Port used by member

### .spec.syncworkers.priorityClassName: string

PriorityClassName specifies a priority class name

### .spec.syncworkers.probes.livenessProbeDisabled: bool

LivenessProbeDisabled if true livenessProbes are disabled

### .spec.syncworkers.probes.livenessProbeSpec.failureThreshold: int32

### .spec.syncworkers.probes.livenessProbeSpec.initialDelaySeconds: int32

### .spec.syncworkers.probes.livenessProbeSpec.periodSeconds: int32

### .spec.syncworkers.probes.livenessProbeSpec.successThreshold: int32

### .spec.syncworkers.probes.livenessProbeSpec.timeoutSeconds: int32

### .spec.syncworkers.probes.ReadinessProbeDisabled: bool

OldReadinessProbeDisabled if true readinessProbes are disabled

Deprecated: This field is deprecated, keept only for backward compatibility.

### .spec.syncworkers.probes.readinessProbeDisabled: bool

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

### .spec.syncworkers.probes.readinessProbeSpec.failureThreshold: int32

### .spec.syncworkers.probes.readinessProbeSpec.initialDelaySeconds: int32

### .spec.syncworkers.probes.readinessProbeSpec.periodSeconds: int32

### .spec.syncworkers.probes.readinessProbeSpec.successThreshold: int32

### .spec.syncworkers.probes.readinessProbeSpec.timeoutSeconds: int32

### .spec.syncworkers.probes.startupProbeDisabled: bool

StartupProbeDisabled if true startupProbes are disabled

### .spec.syncworkers.probes.startupProbeSpec.failureThreshold: int32

### .spec.syncworkers.probes.startupProbeSpec.initialDelaySeconds: int32

### .spec.syncworkers.probes.startupProbeSpec.periodSeconds: int32

### .spec.syncworkers.probes.startupProbeSpec.successThreshold: int32

### .spec.syncworkers.probes.startupProbeSpec.timeoutSeconds: int32

### .spec.syncworkers.pvcResizeMode: string

VolumeResizeMode specified resize mode for pvc

### .spec.syncworkers.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

### .spec.syncworkers.schedulerName: string

SchedulerName define scheduler name used for group

### .spec.syncworkers.securityContext.addCapabilities: []string

AddCapabilities add new capabilities to containers

### .spec.syncworkers.securityContext.allowPrivilegeEscalation: bool

### .spec.syncworkers.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers

Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

### .spec.syncworkers.securityContext.fsGroup: int64

### .spec.syncworkers.securityContext.privileged: bool

### .spec.syncworkers.securityContext.readOnlyRootFilesystem: bool

### .spec.syncworkers.securityContext.runAsGroup: int64

### .spec.syncworkers.securityContext.runAsNonRoot: bool

### .spec.syncworkers.securityContext.runAsUser: int64

### .spec.syncworkers.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

### .spec.syncworkers.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

### .spec.syncworkers.securityContext.supplementalGroups: []int64

### .spec.syncworkers.serviceAccountName: string

ServiceAccountName specifies the name of the service account used for Pods in this group.

### .spec.syncworkers.shutdownDelay: int

ShutdownDelay define how long operator should delay finalizer removal after shutdown

### .spec.syncworkers.shutdownMethod: string

ShutdownMethod describe procedure of member shutdown taken by Operator

### .spec.syncworkers.sidecarCoreNames: []string

SidecarCoreNames is a list of sidecar containers which must run in the pod.

Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

### .spec.syncworkers.sidecars: []core.Container

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

### .spec.syncworkers.storageClassName: string

StorageClassName specifies the classname for storage of the servers.

### .spec.syncworkers.terminationGracePeriodSeconds: int64

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

### .spec.syncworkers.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

### .spec.syncworkers.volumeAllowShrink: bool

Deprecated: VolumeAllowShrink allows shrink the volume

### .spec.syncworkers.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a template for volume claims

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

### .spec.syncworkers.volumeMounts: []ServerGroupSpecVolumeMount

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

### .spec.timeouts.-: int64

deprecated

### .spec.timeouts.<v1.ActionType>actions: int64

Actions

### .spec.timeouts.maintenanceGracePeriod: int64

MaintenanceGracePeriod action timeout

### .spec.timezone: string

### .spec.tls.altNames: []string

### .spec.tls.caSecretName: string

### .spec.tls.mode: string

### .spec.tls.sni.<string>mapping: []string

### .spec.tls.ttl: string

### .spec.topology.enabled: bool

### .spec.topology.label: string

### .spec.topology.zones: int

### .spec.upgrade.autoUpgrade: bool

Flag specify if upgrade should be auto-injected, even if is not required (in case of stuck)

