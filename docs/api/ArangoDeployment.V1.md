# API Reference for ArangoDeployment V1

## Spec

### .spec.agents.affinity: core.PodAffinity

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.PodAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L180)

### .spec.agents.allowMemberRecreation: bool

AllowMemberRecreation allows to recreate member.
This setting changes the member recreation logic based on group:
- For Sync Masters, Sync Workers, Coordinator and DB-Servers it determines if a member can be recreated in case of failure (default `true`)
- For Agents and Single this value is hardcoded to `false` and the value provided in spec is ignored.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L222)

### .spec.agents.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.
Annotations are merged with `spec.annotations`.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L125)

### .spec.agents.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L127)

### .spec.agents.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L129)

### .spec.agents.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L176)

### .spec.agents.args: []string

Args setting specifies additional command-line arguments passed to all servers of this group.

Default Value: []

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L85)

### .spec.agents.count: int

Count setting specifies the number of servers to start for the given group.
For the Agent group, this value must be a positive, odd number.
The default value is `3` for all groups except `single` (there the default is `1`
for `spec.mode: Single` and `2` for `spec.mode: ActiveFailover`).
For the `syncworkers` group, it is highly recommended to use the same number
as for the `dbservers` group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L77)

### .spec.agents.entrypoint: string

Entrypoint overrides container executable

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L87)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L235)

### .spec.agents.extendedRotationCheck: bool

ExtendedRotationCheck extend checks for rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L205)

### .spec.agents.externalPortEnabled: bool

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L217)

### .spec.agents.indexMethod: string

IndexMethod define group Indexing method

Possible Values: 
* random (default) - Pick random ID for member. Enforced on the Community Operator.
* ordered - Use sequential number as Member ID, starting from 0. Enterprise Operator required.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L228)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L213)

### .spec.agents.internalPortProtocol: string

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L215)

### .spec.agents.labels: map[string]string

Labels specified the labels added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L131)

### .spec.agents.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L133)

### .spec.agents.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L135)

### .spec.agents.maxCount: int

MaxCount specifies a maximum for the count of servers. If set, a specification is invalid if `count > maxCount`.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L81)

### .spec.agents.memoryReservation: int64

MemoryReservation determines system reservation of memory while calculating `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` value.
If this field is set, `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` is reduced by specified value in percents.
Accepted Range <0, 50>. If value is outside accepted range, it is adjusted to the closest value.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: 0

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L107)

### .spec.agents.minCount: int

MinCount specifies a minimum for the count of servers. If set, a specification is invalid if `count < minCount`.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L79)

### .spec.agents.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L184)

### .spec.agents.nodeSelector: map[string]string

NodeSelector setting specifies a set of labels to be used as `nodeSelector` for Pods of this node.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L150)

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

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` Container Environment Variable

OverrideDetectedNumberOfCores determines if number of cores should be overridden based on values in resources.
If is set to true and Container CPU Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` to the value from the Container CPU Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: true

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L113)

### .spec.agents.overrideDetectedTotalMemory: bool

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` Container Environment Variable

OverrideDetectedTotalMemory determines if memory should be overridden based on values in resources.
If is set to true and Container Memory Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` to the value from the Container Memory Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: true

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L101)

### .spec.agents.podModes.network: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L31)

### .spec.agents.podModes.pid: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L32)

### .spec.agents.port: uint16

Port define Port used by member

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L233)

### .spec.agents.priorityClassName: string

PriorityClassName specifies a priority class name
Will be forwarded to the pod spec.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L156)

### .spec.agents.probes.livenessProbeDisabled: bool

LivenessProbeDisabled if set to true, the operator does not generate a liveness probe for new pods belonging to this group

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L245)

### .spec.agents.probes.livenessProbeSpec.failureThreshold: int32

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: 3

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L295)

### .spec.agents.probes.livenessProbeSpec.initialDelaySeconds: int32

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L278)

### .spec.agents.probes.livenessProbeSpec.periodSeconds: int32

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: 10

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L282)

### .spec.agents.probes.livenessProbeSpec.successThreshold: int32

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: 1

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L290)

### .spec.agents.probes.livenessProbeSpec.timeoutSeconds: int32

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L286)

### .spec.agents.probes.ReadinessProbeDisabled: bool

OldReadinessProbeDisabled if true readinessProbes are disabled
Deprecated: This field is deprecated, keept only for backward compatibility.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L252)

### .spec.agents.probes.readinessProbeDisabled: bool

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L254)

### .spec.agents.probes.readinessProbeSpec.failureThreshold: int32

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: 3

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L295)

### .spec.agents.probes.readinessProbeSpec.initialDelaySeconds: int32

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L278)

### .spec.agents.probes.readinessProbeSpec.periodSeconds: int32

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: 10

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L282)

### .spec.agents.probes.readinessProbeSpec.successThreshold: int32

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: 1

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L290)

### .spec.agents.probes.readinessProbeSpec.timeoutSeconds: int32

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L286)

### .spec.agents.probes.startupProbeDisabled: bool

StartupProbeDisabled if true startupProbes are disabled

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L259)

### .spec.agents.probes.startupProbeSpec.failureThreshold: int32

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: 3

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L295)

### .spec.agents.probes.startupProbeSpec.initialDelaySeconds: int32

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L278)

### .spec.agents.probes.startupProbeSpec.periodSeconds: int32

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: 10

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L282)

### .spec.agents.probes.startupProbeSpec.successThreshold: int32

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: 1

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L290)

### .spec.agents.probes.startupProbeSpec.timeoutSeconds: int32

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L286)

### .spec.agents.pvcResizeMode: string

VolumeResizeMode specified resize mode for PVCs and PVs

Possible Values: 
* runtime (default) - PVC will be resized in Pod runtime (EKS, GKE)
* rotate - Pod will be shutdown and PVC will be resized (AKS)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L170)

### .spec.agents.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L95)

### .spec.agents.schedulerName: string

SchedulerName define scheduler name used for group

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L89)

### .spec.agents.securityContext.addCapabilities: []core.Capability

AddCapabilities add new capabilities to containers

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L46)

### .spec.agents.securityContext.allowPrivilegeEscalation: bool

AllowPrivilegeEscalation Controls whether a process can gain more privileges than its parent process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)

### .spec.agents.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L43)

### .spec.agents.securityContext.fsGroup: int64

FSGroup is a special supplemental group that applies to all containers in a pod.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L66)

### .spec.agents.securityContext.privileged: bool

Privileged If true, runs container in privileged mode. Processes in privileged containers are
essentially equivalent to root on the host.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L52)

### .spec.agents.securityContext.readOnlyRootFilesystem: bool

ReadOnlyRootFilesystem if true, mounts the container's root filesystem as read-only.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L54)

### .spec.agents.securityContext.runAsGroup: int64

RunAsGroup is the GID to run the entrypoint of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L60)

### .spec.agents.securityContext.runAsNonRoot: bool

RunAsNonRoot if true, indicates that the container must run as a non-root user.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L56)

### .spec.agents.securityContext.runAsUser: int64

RunAsUser is the UID to run the entrypoint of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L58)

### .spec.agents.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L82)

### .spec.agents.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L87)

### .spec.agents.securityContext.supplementalGroups: []int64

SupplementalGroups is a list of groups applied to the first process run in each container, in addition to the container's primary GID,
the fsGroup (if specified), and group memberships defined in the container image for the uid of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L64)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L77)

### .spec.agents.serviceAccountName: string

ServiceAccountName setting specifies the `serviceAccountName` for the `Pods` created
for each server of this group. If empty, it defaults to using the
`default` service account.
Using an alternative `ServiceAccount` is typically used to separate access rights.
The ArangoDB deployments need some very minimal access rights. With the
deployment of the operator, we grant the rights to 'get' all 'pod' resources.
If you are using a different service account, please grant these rights
to that service account.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L146)

### .spec.agents.shutdownDelay: int

ShutdownDelay define how long operator should delay finalizer removal after shutdown

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L211)

### .spec.agents.shutdownMethod: string

ShutdownMethod describe procedure of member shutdown taken by Operator

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L209)

### .spec.agents.sidecarCoreNames: []string

SidecarCoreNames is a list of sidecar containers which must run in the pod.
Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L187)

### .spec.agents.sidecars: []core.Container

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L191)

### .spec.agents.storageClassName: string

StorageClassName specifies the classname for storage of the servers.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L91)

### .spec.agents.terminationGracePeriodSeconds: int64

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.agents.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.
By default, suitable tolerations are set for the following keys with the `NoExecute` effect:
- `node.kubernetes.io/not-ready`
- `node.kubernetes.io/unreachable`
- `node.alpha.kubernetes.io/unreachable` (will be removed in future version)
For more information on tolerations, consult the https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L122)

### .spec.agents.volumeAllowShrink: bool

Deprecated: VolumeAllowShrink allows shrink the volume

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L172)

### .spec.agents.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a volumeClaimTemplate used by operator to create to volume claims for pods of this group.
This setting is not available for group `coordinators`, `syncmasters` & `syncworkers`.
The default value describes a volume with `8Gi` storage, `ReadWriteOnce` access mode and volume mode set to `PersistentVolumeFilesystem`.
If this field is not set and `spec.<group>.resources.requests.storage` is set, then a default volume claim
with size as specified by `spec.<group>.resources.requests.storage` will be created. In that case `storage`
and `iops` is not forwarded to the pods resource requirements.

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L165)

### .spec.agents.volumeMounts: []ServerGroupSpecVolumeMount

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation of ServerGroupSpecVolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L201)

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
Possible values are:
- `amd64`: Use processors with the x86-64 architecture.
- `arm64`: Use processors with the 64-bit ARM architecture.
The setting expects a list of strings, but you should only specify a single
list item for the architecture, except when you want to migrate from one
architecture to the other. The first list item defines the new default
architecture for the deployment that you want to migrate to.

Links:
* [Architecture Change](/docs/how-to/arch_change.md)

Default Value: ['amd64']

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L265)

### .spec.auth.jwtSecretName: string

JWTSecretName setting specifies the name of a kubernetes `Secret` that contains
the JWT token used for accessing all ArangoDB servers.
When no name is specified, it defaults to `<deployment-name>-jwt`.
To disable authentication, set this value to `None`.
If you specify a name of a `Secret`, that secret must have the token
in a data field named `token`.
If you specify a name of a `Secret` that does not exist, a random token is created
and stored in a `Secret` with given name.
Changing a JWT token results in restarting of a whole cluster.

[Code Reference](/pkg/apis/deployment/v1/authentication_spec.go#L40)

### .spec.bootstrap.passwordSecretNames: map[string]string

PasswordSecretNames contains a map of username to password-secret-name
This setting specifies a secret name for the credentials per specific users.
When a deployment is created the operator will setup the user accounts
according to the credentials given by the secret. If the secret doesn't exist
the operator creates a secret with a random password.
There are two magic values for the secret name:
- `None` specifies no action. This disables root password randomization. This is the default value. (Thus the root password is empty - not recommended)
- `Auto` specifies automatic name generation, which is `<deploymentname>-root-password`.

Links:
* [How to set root user password](/docs/how-to/set_root_user_password.md)

[Code Reference](/pkg/apis/deployment/v1/bootstrap.go#L62)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L180)

### .spec.coordinators.allowMemberRecreation: bool

AllowMemberRecreation allows to recreate member.
This setting changes the member recreation logic based on group:
- For Sync Masters, Sync Workers, Coordinator and DB-Servers it determines if a member can be recreated in case of failure (default `true`)
- For Agents and Single this value is hardcoded to `false` and the value provided in spec is ignored.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L222)

### .spec.coordinators.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.
Annotations are merged with `spec.annotations`.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L125)

### .spec.coordinators.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L127)

### .spec.coordinators.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L129)

### .spec.coordinators.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L176)

### .spec.coordinators.args: []string

Args setting specifies additional command-line arguments passed to all servers of this group.

Default Value: []

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L85)

### .spec.coordinators.count: int

Count setting specifies the number of servers to start for the given group.
For the Agent group, this value must be a positive, odd number.
The default value is `3` for all groups except `single` (there the default is `1`
for `spec.mode: Single` and `2` for `spec.mode: ActiveFailover`).
For the `syncworkers` group, it is highly recommended to use the same number
as for the `dbservers` group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L77)

### .spec.coordinators.entrypoint: string

Entrypoint overrides container executable

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L87)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L235)

### .spec.coordinators.extendedRotationCheck: bool

ExtendedRotationCheck extend checks for rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L205)

### .spec.coordinators.externalPortEnabled: bool

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L217)

### .spec.coordinators.indexMethod: string

IndexMethod define group Indexing method

Possible Values: 
* random (default) - Pick random ID for member. Enforced on the Community Operator.
* ordered - Use sequential number as Member ID, starting from 0. Enterprise Operator required.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L228)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L213)

### .spec.coordinators.internalPortProtocol: string

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L215)

### .spec.coordinators.labels: map[string]string

Labels specified the labels added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L131)

### .spec.coordinators.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L133)

### .spec.coordinators.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L135)

### .spec.coordinators.maxCount: int

MaxCount specifies a maximum for the count of servers. If set, a specification is invalid if `count > maxCount`.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L81)

### .spec.coordinators.memoryReservation: int64

MemoryReservation determines system reservation of memory while calculating `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` value.
If this field is set, `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` is reduced by specified value in percents.
Accepted Range <0, 50>. If value is outside accepted range, it is adjusted to the closest value.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: 0

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L107)

### .spec.coordinators.minCount: int

MinCount specifies a minimum for the count of servers. If set, a specification is invalid if `count < minCount`.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L79)

### .spec.coordinators.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L184)

### .spec.coordinators.nodeSelector: map[string]string

NodeSelector setting specifies a set of labels to be used as `nodeSelector` for Pods of this node.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L150)

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

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` Container Environment Variable

OverrideDetectedNumberOfCores determines if number of cores should be overridden based on values in resources.
If is set to true and Container CPU Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` to the value from the Container CPU Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: true

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L113)

### .spec.coordinators.overrideDetectedTotalMemory: bool

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` Container Environment Variable

OverrideDetectedTotalMemory determines if memory should be overridden based on values in resources.
If is set to true and Container Memory Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` to the value from the Container Memory Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: true

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L101)

### .spec.coordinators.podModes.network: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L31)

### .spec.coordinators.podModes.pid: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L32)

### .spec.coordinators.port: uint16

Port define Port used by member

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L233)

### .spec.coordinators.priorityClassName: string

PriorityClassName specifies a priority class name
Will be forwarded to the pod spec.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L156)

### .spec.coordinators.probes.livenessProbeDisabled: bool

LivenessProbeDisabled if set to true, the operator does not generate a liveness probe for new pods belonging to this group

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L245)

### .spec.coordinators.probes.livenessProbeSpec.failureThreshold: int32

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: 3

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L295)

### .spec.coordinators.probes.livenessProbeSpec.initialDelaySeconds: int32

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L278)

### .spec.coordinators.probes.livenessProbeSpec.periodSeconds: int32

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: 10

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L282)

### .spec.coordinators.probes.livenessProbeSpec.successThreshold: int32

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: 1

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L290)

### .spec.coordinators.probes.livenessProbeSpec.timeoutSeconds: int32

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L286)

### .spec.coordinators.probes.ReadinessProbeDisabled: bool

OldReadinessProbeDisabled if true readinessProbes are disabled
Deprecated: This field is deprecated, keept only for backward compatibility.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L252)

### .spec.coordinators.probes.readinessProbeDisabled: bool

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L254)

### .spec.coordinators.probes.readinessProbeSpec.failureThreshold: int32

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: 3

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L295)

### .spec.coordinators.probes.readinessProbeSpec.initialDelaySeconds: int32

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L278)

### .spec.coordinators.probes.readinessProbeSpec.periodSeconds: int32

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: 10

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L282)

### .spec.coordinators.probes.readinessProbeSpec.successThreshold: int32

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: 1

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L290)

### .spec.coordinators.probes.readinessProbeSpec.timeoutSeconds: int32

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L286)

### .spec.coordinators.probes.startupProbeDisabled: bool

StartupProbeDisabled if true startupProbes are disabled

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L259)

### .spec.coordinators.probes.startupProbeSpec.failureThreshold: int32

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: 3

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L295)

### .spec.coordinators.probes.startupProbeSpec.initialDelaySeconds: int32

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L278)

### .spec.coordinators.probes.startupProbeSpec.periodSeconds: int32

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: 10

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L282)

### .spec.coordinators.probes.startupProbeSpec.successThreshold: int32

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: 1

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L290)

### .spec.coordinators.probes.startupProbeSpec.timeoutSeconds: int32

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L286)

### .spec.coordinators.pvcResizeMode: string

VolumeResizeMode specified resize mode for PVCs and PVs

Possible Values: 
* runtime (default) - PVC will be resized in Pod runtime (EKS, GKE)
* rotate - Pod will be shutdown and PVC will be resized (AKS)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L170)

### .spec.coordinators.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L95)

### .spec.coordinators.schedulerName: string

SchedulerName define scheduler name used for group

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L89)

### .spec.coordinators.securityContext.addCapabilities: []core.Capability

AddCapabilities add new capabilities to containers

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L46)

### .spec.coordinators.securityContext.allowPrivilegeEscalation: bool

AllowPrivilegeEscalation Controls whether a process can gain more privileges than its parent process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)

### .spec.coordinators.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L43)

### .spec.coordinators.securityContext.fsGroup: int64

FSGroup is a special supplemental group that applies to all containers in a pod.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L66)

### .spec.coordinators.securityContext.privileged: bool

Privileged If true, runs container in privileged mode. Processes in privileged containers are
essentially equivalent to root on the host.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L52)

### .spec.coordinators.securityContext.readOnlyRootFilesystem: bool

ReadOnlyRootFilesystem if true, mounts the container's root filesystem as read-only.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L54)

### .spec.coordinators.securityContext.runAsGroup: int64

RunAsGroup is the GID to run the entrypoint of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L60)

### .spec.coordinators.securityContext.runAsNonRoot: bool

RunAsNonRoot if true, indicates that the container must run as a non-root user.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L56)

### .spec.coordinators.securityContext.runAsUser: int64

RunAsUser is the UID to run the entrypoint of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L58)

### .spec.coordinators.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L82)

### .spec.coordinators.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L87)

### .spec.coordinators.securityContext.supplementalGroups: []int64

SupplementalGroups is a list of groups applied to the first process run in each container, in addition to the container's primary GID,
the fsGroup (if specified), and group memberships defined in the container image for the uid of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L64)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L77)

### .spec.coordinators.serviceAccountName: string

ServiceAccountName setting specifies the `serviceAccountName` for the `Pods` created
for each server of this group. If empty, it defaults to using the
`default` service account.
Using an alternative `ServiceAccount` is typically used to separate access rights.
The ArangoDB deployments need some very minimal access rights. With the
deployment of the operator, we grant the rights to 'get' all 'pod' resources.
If you are using a different service account, please grant these rights
to that service account.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L146)

### .spec.coordinators.shutdownDelay: int

ShutdownDelay define how long operator should delay finalizer removal after shutdown

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L211)

### .spec.coordinators.shutdownMethod: string

ShutdownMethod describe procedure of member shutdown taken by Operator

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L209)

### .spec.coordinators.sidecarCoreNames: []string

SidecarCoreNames is a list of sidecar containers which must run in the pod.
Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L187)

### .spec.coordinators.sidecars: []core.Container

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L191)

### .spec.coordinators.storageClassName: string

StorageClassName specifies the classname for storage of the servers.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L91)

### .spec.coordinators.terminationGracePeriodSeconds: int64

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.coordinators.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.
By default, suitable tolerations are set for the following keys with the `NoExecute` effect:
- `node.kubernetes.io/not-ready`
- `node.kubernetes.io/unreachable`
- `node.alpha.kubernetes.io/unreachable` (will be removed in future version)
For more information on tolerations, consult the https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L122)

### .spec.coordinators.volumeAllowShrink: bool

Deprecated: VolumeAllowShrink allows shrink the volume

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L172)

### .spec.coordinators.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a volumeClaimTemplate used by operator to create to volume claims for pods of this group.
This setting is not available for group `coordinators`, `syncmasters` & `syncworkers`.
The default value describes a volume with `8Gi` storage, `ReadWriteOnce` access mode and volume mode set to `PersistentVolumeFilesystem`.
If this field is not set and `spec.<group>.resources.requests.storage` is set, then a default volume claim
with size as specified by `spec.<group>.resources.requests.storage` will be created. In that case `storage`
and `iops` is not forwarded to the pods resource requirements.

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L165)

### .spec.coordinators.volumeMounts: []ServerGroupSpecVolumeMount

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation of ServerGroupSpecVolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L201)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L180)

### .spec.dbservers.allowMemberRecreation: bool

AllowMemberRecreation allows to recreate member.
This setting changes the member recreation logic based on group:
- For Sync Masters, Sync Workers, Coordinator and DB-Servers it determines if a member can be recreated in case of failure (default `true`)
- For Agents and Single this value is hardcoded to `false` and the value provided in spec is ignored.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L222)

### .spec.dbservers.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.
Annotations are merged with `spec.annotations`.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L125)

### .spec.dbservers.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L127)

### .spec.dbservers.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L129)

### .spec.dbservers.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L176)

### .spec.dbservers.args: []string

Args setting specifies additional command-line arguments passed to all servers of this group.

Default Value: []

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L85)

### .spec.dbservers.count: int

Count setting specifies the number of servers to start for the given group.
For the Agent group, this value must be a positive, odd number.
The default value is `3` for all groups except `single` (there the default is `1`
for `spec.mode: Single` and `2` for `spec.mode: ActiveFailover`).
For the `syncworkers` group, it is highly recommended to use the same number
as for the `dbservers` group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L77)

### .spec.dbservers.entrypoint: string

Entrypoint overrides container executable

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L87)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L235)

### .spec.dbservers.extendedRotationCheck: bool

ExtendedRotationCheck extend checks for rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L205)

### .spec.dbservers.externalPortEnabled: bool

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L217)

### .spec.dbservers.indexMethod: string

IndexMethod define group Indexing method

Possible Values: 
* random (default) - Pick random ID for member. Enforced on the Community Operator.
* ordered - Use sequential number as Member ID, starting from 0. Enterprise Operator required.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L228)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L213)

### .spec.dbservers.internalPortProtocol: string

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L215)

### .spec.dbservers.labels: map[string]string

Labels specified the labels added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L131)

### .spec.dbservers.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L133)

### .spec.dbservers.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L135)

### .spec.dbservers.maxCount: int

MaxCount specifies a maximum for the count of servers. If set, a specification is invalid if `count > maxCount`.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L81)

### .spec.dbservers.memoryReservation: int64

MemoryReservation determines system reservation of memory while calculating `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` value.
If this field is set, `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` is reduced by specified value in percents.
Accepted Range <0, 50>. If value is outside accepted range, it is adjusted to the closest value.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: 0

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L107)

### .spec.dbservers.minCount: int

MinCount specifies a minimum for the count of servers. If set, a specification is invalid if `count < minCount`.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L79)

### .spec.dbservers.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L184)

### .spec.dbservers.nodeSelector: map[string]string

NodeSelector setting specifies a set of labels to be used as `nodeSelector` for Pods of this node.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L150)

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

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` Container Environment Variable

OverrideDetectedNumberOfCores determines if number of cores should be overridden based on values in resources.
If is set to true and Container CPU Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` to the value from the Container CPU Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: true

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L113)

### .spec.dbservers.overrideDetectedTotalMemory: bool

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` Container Environment Variable

OverrideDetectedTotalMemory determines if memory should be overridden based on values in resources.
If is set to true and Container Memory Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` to the value from the Container Memory Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: true

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L101)

### .spec.dbservers.podModes.network: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L31)

### .spec.dbservers.podModes.pid: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L32)

### .spec.dbservers.port: uint16

Port define Port used by member

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L233)

### .spec.dbservers.priorityClassName: string

PriorityClassName specifies a priority class name
Will be forwarded to the pod spec.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L156)

### .spec.dbservers.probes.livenessProbeDisabled: bool

LivenessProbeDisabled if set to true, the operator does not generate a liveness probe for new pods belonging to this group

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L245)

### .spec.dbservers.probes.livenessProbeSpec.failureThreshold: int32

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: 3

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L295)

### .spec.dbservers.probes.livenessProbeSpec.initialDelaySeconds: int32

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L278)

### .spec.dbservers.probes.livenessProbeSpec.periodSeconds: int32

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: 10

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L282)

### .spec.dbservers.probes.livenessProbeSpec.successThreshold: int32

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: 1

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L290)

### .spec.dbservers.probes.livenessProbeSpec.timeoutSeconds: int32

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L286)

### .spec.dbservers.probes.ReadinessProbeDisabled: bool

OldReadinessProbeDisabled if true readinessProbes are disabled
Deprecated: This field is deprecated, keept only for backward compatibility.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L252)

### .spec.dbservers.probes.readinessProbeDisabled: bool

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L254)

### .spec.dbservers.probes.readinessProbeSpec.failureThreshold: int32

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: 3

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L295)

### .spec.dbservers.probes.readinessProbeSpec.initialDelaySeconds: int32

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L278)

### .spec.dbservers.probes.readinessProbeSpec.periodSeconds: int32

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: 10

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L282)

### .spec.dbservers.probes.readinessProbeSpec.successThreshold: int32

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: 1

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L290)

### .spec.dbservers.probes.readinessProbeSpec.timeoutSeconds: int32

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L286)

### .spec.dbservers.probes.startupProbeDisabled: bool

StartupProbeDisabled if true startupProbes are disabled

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L259)

### .spec.dbservers.probes.startupProbeSpec.failureThreshold: int32

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: 3

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L295)

### .spec.dbservers.probes.startupProbeSpec.initialDelaySeconds: int32

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L278)

### .spec.dbservers.probes.startupProbeSpec.periodSeconds: int32

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: 10

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L282)

### .spec.dbservers.probes.startupProbeSpec.successThreshold: int32

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: 1

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L290)

### .spec.dbservers.probes.startupProbeSpec.timeoutSeconds: int32

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L286)

### .spec.dbservers.pvcResizeMode: string

VolumeResizeMode specified resize mode for PVCs and PVs

Possible Values: 
* runtime (default) - PVC will be resized in Pod runtime (EKS, GKE)
* rotate - Pod will be shutdown and PVC will be resized (AKS)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L170)

### .spec.dbservers.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L95)

### .spec.dbservers.schedulerName: string

SchedulerName define scheduler name used for group

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L89)

### .spec.dbservers.securityContext.addCapabilities: []core.Capability

AddCapabilities add new capabilities to containers

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L46)

### .spec.dbservers.securityContext.allowPrivilegeEscalation: bool

AllowPrivilegeEscalation Controls whether a process can gain more privileges than its parent process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)

### .spec.dbservers.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L43)

### .spec.dbservers.securityContext.fsGroup: int64

FSGroup is a special supplemental group that applies to all containers in a pod.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L66)

### .spec.dbservers.securityContext.privileged: bool

Privileged If true, runs container in privileged mode. Processes in privileged containers are
essentially equivalent to root on the host.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L52)

### .spec.dbservers.securityContext.readOnlyRootFilesystem: bool

ReadOnlyRootFilesystem if true, mounts the container's root filesystem as read-only.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L54)

### .spec.dbservers.securityContext.runAsGroup: int64

RunAsGroup is the GID to run the entrypoint of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L60)

### .spec.dbservers.securityContext.runAsNonRoot: bool

RunAsNonRoot if true, indicates that the container must run as a non-root user.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L56)

### .spec.dbservers.securityContext.runAsUser: int64

RunAsUser is the UID to run the entrypoint of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L58)

### .spec.dbservers.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L82)

### .spec.dbservers.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L87)

### .spec.dbservers.securityContext.supplementalGroups: []int64

SupplementalGroups is a list of groups applied to the first process run in each container, in addition to the container's primary GID,
the fsGroup (if specified), and group memberships defined in the container image for the uid of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L64)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L77)

### .spec.dbservers.serviceAccountName: string

ServiceAccountName setting specifies the `serviceAccountName` for the `Pods` created
for each server of this group. If empty, it defaults to using the
`default` service account.
Using an alternative `ServiceAccount` is typically used to separate access rights.
The ArangoDB deployments need some very minimal access rights. With the
deployment of the operator, we grant the rights to 'get' all 'pod' resources.
If you are using a different service account, please grant these rights
to that service account.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L146)

### .spec.dbservers.shutdownDelay: int

ShutdownDelay define how long operator should delay finalizer removal after shutdown

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L211)

### .spec.dbservers.shutdownMethod: string

ShutdownMethod describe procedure of member shutdown taken by Operator

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L209)

### .spec.dbservers.sidecarCoreNames: []string

SidecarCoreNames is a list of sidecar containers which must run in the pod.
Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L187)

### .spec.dbservers.sidecars: []core.Container

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L191)

### .spec.dbservers.storageClassName: string

StorageClassName specifies the classname for storage of the servers.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L91)

### .spec.dbservers.terminationGracePeriodSeconds: int64

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.dbservers.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.
By default, suitable tolerations are set for the following keys with the `NoExecute` effect:
- `node.kubernetes.io/not-ready`
- `node.kubernetes.io/unreachable`
- `node.alpha.kubernetes.io/unreachable` (will be removed in future version)
For more information on tolerations, consult the https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L122)

### .spec.dbservers.volumeAllowShrink: bool

Deprecated: VolumeAllowShrink allows shrink the volume

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L172)

### .spec.dbservers.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a volumeClaimTemplate used by operator to create to volume claims for pods of this group.
This setting is not available for group `coordinators`, `syncmasters` & `syncworkers`.
The default value describes a volume with `8Gi` storage, `ReadWriteOnce` access mode and volume mode set to `PersistentVolumeFilesystem`.
If this field is not set and `spec.<group>.resources.requests.storage` is set, then a default volume claim
with size as specified by `spec.<group>.resources.requests.storage` will be created. In that case `storage`
and `iops` is not forwarded to the pods resource requirements.

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L165)

### .spec.dbservers.volumeMounts: []ServerGroupSpecVolumeMount

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation of ServerGroupSpecVolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L201)

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

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L58)

### .spec.externalAccess.loadBalancerIP: string

LoadBalancerIP define optional IP used to configure a load-balancer on, in case of Auto or LoadBalancer type.
If you do not specify this setting, an IP will be chosen automatically by the load-balancer provisioner.

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L48)

### .spec.externalAccess.loadBalancerSourceRanges: []string

LoadBalancerSourceRanges define LoadBalancerSourceRanges used for LoadBalancer Service type
If specified and supported by the platform, this will restrict traffic through the cloud-provider
load-balancer will be restricted to the specified client IPs. This field will be ignored if the
cloud-provider does not support the feature.

Links:
* [Cloud Provider Firewall](https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/)

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L55)

### .spec.externalAccess.managedServiceNames: []string

ManagedServiceNames keeps names of services which are not managed by KubeArangoDB.
It is only relevant when type of service is `managed`.

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L62)

### .spec.externalAccess.nodePort: int

NodePort define optional port used in case of Auto or NodePort type.
This setting is used when `spec.externalAccess.type` is set to `NodePort` or `Auto`.
If you do not specify this setting, a random port will be chosen automatically.

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L44)

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

NodeSelector specifies a set of selectors for nodes

[Code Reference](/pkg/apis/deployment/v1/server_id_group_spec.go#L34)

### .spec.id.priorityClassName: string

PriorityClassName specifies a priority class name

[Code Reference](/pkg/apis/deployment/v1/server_id_group_spec.go#L36)

### .spec.id.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_id_group_spec.go#L56)

### .spec.id.securityContext.addCapabilities: []core.Capability

AddCapabilities add new capabilities to containers

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L46)

### .spec.id.securityContext.allowPrivilegeEscalation: bool

AllowPrivilegeEscalation Controls whether a process can gain more privileges than its parent process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)

### .spec.id.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L43)

### .spec.id.securityContext.fsGroup: int64

FSGroup is a special supplemental group that applies to all containers in a pod.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L66)

### .spec.id.securityContext.privileged: bool

Privileged If true, runs container in privileged mode. Processes in privileged containers are
essentially equivalent to root on the host.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L52)

### .spec.id.securityContext.readOnlyRootFilesystem: bool

ReadOnlyRootFilesystem if true, mounts the container's root filesystem as read-only.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L54)

### .spec.id.securityContext.runAsGroup: int64

RunAsGroup is the GID to run the entrypoint of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L60)

### .spec.id.securityContext.runAsNonRoot: bool

RunAsNonRoot if true, indicates that the container must run as a non-root user.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L56)

### .spec.id.securityContext.runAsUser: int64

RunAsUser is the UID to run the entrypoint of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L58)

### .spec.id.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L82)

### .spec.id.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L87)

### .spec.id.securityContext.supplementalGroups: []int64

SupplementalGroups is a list of groups applied to the first process run in each container, in addition to the container's primary GID,
the fsGroup (if specified), and group memberships defined in the container image for the uid of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L64)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L77)

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

SecretName setting specifies the name of a kubernetes `Secret` that contains
the license key token used for enterprise images. This value is not used for
the Community Edition.

[Code Reference](/pkg/apis/deployment/v1/license_spec.go#L33)

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

Enabled if this is set to `true`, the operator runs a sidecar container for
every Agent, DB-Server, Coordinator and Single server.

Links:
* [Metrics collection](/docs/metrics.md)

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/deployment_metrics_spec.go#L81)

### .spec.metrics.image: string

deprecated

[Code Reference](/pkg/apis/deployment/v1/deployment_metrics_spec.go#L83)

### .spec.metrics.mode: string

deprecated

[Code Reference](/pkg/apis/deployment/v1/deployment_metrics_spec.go#L90)

### .spec.metrics.port: uint16

[Code Reference](/pkg/apis/deployment/v1/deployment_metrics_spec.go#L99)

### .spec.metrics.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

[Code Reference](/pkg/apis/deployment/v1/deployment_metrics_spec.go#L88)

### .spec.metrics.serviceMonitor.enabled: bool

[Code Reference](/pkg/apis/deployment/v1/deployment_metrics_service_monitor_spec.go#L24)

### .spec.metrics.serviceMonitor.labels: map[string]string

[Code Reference](/pkg/apis/deployment/v1/deployment_metrics_service_monitor_spec.go#L25)

### .spec.metrics.tls: bool

TLS defines if TLS should be enabled on Metrics exporter endpoint.
This option will enable TLS only if TLS is enabled on ArangoDeployment,
otherwise `true` value will not take any effect.

Default Value: true

[Code Reference](/pkg/apis/deployment/v1/deployment_metrics_spec.go#L95)

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

KeySecretName setting specifies the name of a Kubernetes `Secret` that contains an encryption key used for encrypting all data stored by ArangoDB servers.
When an encryption key is used, encryption of the data in the cluster is enabled, without it encryption is disabled.
The default value is empty.
This requires the Enterprise Edition.
The encryption key cannot be changed after the cluster has been created.
The secret specified by this setting, must have a data field named 'key' containing an encryption key that is exactly 32 bytes long.

[Code Reference](/pkg/apis/deployment/v1/rocksdb_spec.go#L37)

### .spec.single.affinity: core.PodAffinity

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.PodAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L180)

### .spec.single.allowMemberRecreation: bool

AllowMemberRecreation allows to recreate member.
This setting changes the member recreation logic based on group:
- For Sync Masters, Sync Workers, Coordinator and DB-Servers it determines if a member can be recreated in case of failure (default `true`)
- For Agents and Single this value is hardcoded to `false` and the value provided in spec is ignored.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L222)

### .spec.single.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.
Annotations are merged with `spec.annotations`.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L125)

### .spec.single.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L127)

### .spec.single.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L129)

### .spec.single.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L176)

### .spec.single.args: []string

Args setting specifies additional command-line arguments passed to all servers of this group.

Default Value: []

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L85)

### .spec.single.count: int

Count setting specifies the number of servers to start for the given group.
For the Agent group, this value must be a positive, odd number.
The default value is `3` for all groups except `single` (there the default is `1`
for `spec.mode: Single` and `2` for `spec.mode: ActiveFailover`).
For the `syncworkers` group, it is highly recommended to use the same number
as for the `dbservers` group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L77)

### .spec.single.entrypoint: string

Entrypoint overrides container executable

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L87)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L235)

### .spec.single.extendedRotationCheck: bool

ExtendedRotationCheck extend checks for rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L205)

### .spec.single.externalPortEnabled: bool

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L217)

### .spec.single.indexMethod: string

IndexMethod define group Indexing method

Possible Values: 
* random (default) - Pick random ID for member. Enforced on the Community Operator.
* ordered - Use sequential number as Member ID, starting from 0. Enterprise Operator required.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L228)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L213)

### .spec.single.internalPortProtocol: string

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L215)

### .spec.single.labels: map[string]string

Labels specified the labels added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L131)

### .spec.single.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L133)

### .spec.single.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L135)

### .spec.single.maxCount: int

MaxCount specifies a maximum for the count of servers. If set, a specification is invalid if `count > maxCount`.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L81)

### .spec.single.memoryReservation: int64

MemoryReservation determines system reservation of memory while calculating `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` value.
If this field is set, `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` is reduced by specified value in percents.
Accepted Range <0, 50>. If value is outside accepted range, it is adjusted to the closest value.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: 0

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L107)

### .spec.single.minCount: int

MinCount specifies a minimum for the count of servers. If set, a specification is invalid if `count < minCount`.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L79)

### .spec.single.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L184)

### .spec.single.nodeSelector: map[string]string

NodeSelector setting specifies a set of labels to be used as `nodeSelector` for Pods of this node.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L150)

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

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` Container Environment Variable

OverrideDetectedNumberOfCores determines if number of cores should be overridden based on values in resources.
If is set to true and Container CPU Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` to the value from the Container CPU Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: true

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L113)

### .spec.single.overrideDetectedTotalMemory: bool

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` Container Environment Variable

OverrideDetectedTotalMemory determines if memory should be overridden based on values in resources.
If is set to true and Container Memory Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` to the value from the Container Memory Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: true

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L101)

### .spec.single.podModes.network: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L31)

### .spec.single.podModes.pid: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L32)

### .spec.single.port: uint16

Port define Port used by member

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L233)

### .spec.single.priorityClassName: string

PriorityClassName specifies a priority class name
Will be forwarded to the pod spec.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L156)

### .spec.single.probes.livenessProbeDisabled: bool

LivenessProbeDisabled if set to true, the operator does not generate a liveness probe for new pods belonging to this group

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L245)

### .spec.single.probes.livenessProbeSpec.failureThreshold: int32

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: 3

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L295)

### .spec.single.probes.livenessProbeSpec.initialDelaySeconds: int32

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L278)

### .spec.single.probes.livenessProbeSpec.periodSeconds: int32

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: 10

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L282)

### .spec.single.probes.livenessProbeSpec.successThreshold: int32

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: 1

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L290)

### .spec.single.probes.livenessProbeSpec.timeoutSeconds: int32

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L286)

### .spec.single.probes.ReadinessProbeDisabled: bool

OldReadinessProbeDisabled if true readinessProbes are disabled
Deprecated: This field is deprecated, keept only for backward compatibility.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L252)

### .spec.single.probes.readinessProbeDisabled: bool

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L254)

### .spec.single.probes.readinessProbeSpec.failureThreshold: int32

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: 3

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L295)

### .spec.single.probes.readinessProbeSpec.initialDelaySeconds: int32

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L278)

### .spec.single.probes.readinessProbeSpec.periodSeconds: int32

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: 10

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L282)

### .spec.single.probes.readinessProbeSpec.successThreshold: int32

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: 1

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L290)

### .spec.single.probes.readinessProbeSpec.timeoutSeconds: int32

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L286)

### .spec.single.probes.startupProbeDisabled: bool

StartupProbeDisabled if true startupProbes are disabled

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L259)

### .spec.single.probes.startupProbeSpec.failureThreshold: int32

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: 3

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L295)

### .spec.single.probes.startupProbeSpec.initialDelaySeconds: int32

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L278)

### .spec.single.probes.startupProbeSpec.periodSeconds: int32

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: 10

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L282)

### .spec.single.probes.startupProbeSpec.successThreshold: int32

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: 1

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L290)

### .spec.single.probes.startupProbeSpec.timeoutSeconds: int32

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L286)

### .spec.single.pvcResizeMode: string

VolumeResizeMode specified resize mode for PVCs and PVs

Possible Values: 
* runtime (default) - PVC will be resized in Pod runtime (EKS, GKE)
* rotate - Pod will be shutdown and PVC will be resized (AKS)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L170)

### .spec.single.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L95)

### .spec.single.schedulerName: string

SchedulerName define scheduler name used for group

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L89)

### .spec.single.securityContext.addCapabilities: []core.Capability

AddCapabilities add new capabilities to containers

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L46)

### .spec.single.securityContext.allowPrivilegeEscalation: bool

AllowPrivilegeEscalation Controls whether a process can gain more privileges than its parent process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)

### .spec.single.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L43)

### .spec.single.securityContext.fsGroup: int64

FSGroup is a special supplemental group that applies to all containers in a pod.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L66)

### .spec.single.securityContext.privileged: bool

Privileged If true, runs container in privileged mode. Processes in privileged containers are
essentially equivalent to root on the host.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L52)

### .spec.single.securityContext.readOnlyRootFilesystem: bool

ReadOnlyRootFilesystem if true, mounts the container's root filesystem as read-only.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L54)

### .spec.single.securityContext.runAsGroup: int64

RunAsGroup is the GID to run the entrypoint of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L60)

### .spec.single.securityContext.runAsNonRoot: bool

RunAsNonRoot if true, indicates that the container must run as a non-root user.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L56)

### .spec.single.securityContext.runAsUser: int64

RunAsUser is the UID to run the entrypoint of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L58)

### .spec.single.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L82)

### .spec.single.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L87)

### .spec.single.securityContext.supplementalGroups: []int64

SupplementalGroups is a list of groups applied to the first process run in each container, in addition to the container's primary GID,
the fsGroup (if specified), and group memberships defined in the container image for the uid of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L64)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L77)

### .spec.single.serviceAccountName: string

ServiceAccountName setting specifies the `serviceAccountName` for the `Pods` created
for each server of this group. If empty, it defaults to using the
`default` service account.
Using an alternative `ServiceAccount` is typically used to separate access rights.
The ArangoDB deployments need some very minimal access rights. With the
deployment of the operator, we grant the rights to 'get' all 'pod' resources.
If you are using a different service account, please grant these rights
to that service account.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L146)

### .spec.single.shutdownDelay: int

ShutdownDelay define how long operator should delay finalizer removal after shutdown

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L211)

### .spec.single.shutdownMethod: string

ShutdownMethod describe procedure of member shutdown taken by Operator

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L209)

### .spec.single.sidecarCoreNames: []string

SidecarCoreNames is a list of sidecar containers which must run in the pod.
Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L187)

### .spec.single.sidecars: []core.Container

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L191)

### .spec.single.storageClassName: string

StorageClassName specifies the classname for storage of the servers.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L91)

### .spec.single.terminationGracePeriodSeconds: int64

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.single.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.
By default, suitable tolerations are set for the following keys with the `NoExecute` effect:
- `node.kubernetes.io/not-ready`
- `node.kubernetes.io/unreachable`
- `node.alpha.kubernetes.io/unreachable` (will be removed in future version)
For more information on tolerations, consult the https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L122)

### .spec.single.volumeAllowShrink: bool

Deprecated: VolumeAllowShrink allows shrink the volume

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L172)

### .spec.single.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a volumeClaimTemplate used by operator to create to volume claims for pods of this group.
This setting is not available for group `coordinators`, `syncmasters` & `syncworkers`.
The default value describes a volume with `8Gi` storage, `ReadWriteOnce` access mode and volume mode set to `PersistentVolumeFilesystem`.
If this field is not set and `spec.<group>.resources.requests.storage` is set, then a default volume claim
with size as specified by `spec.<group>.resources.requests.storage` will be created. In that case `storage`
and `iops` is not forwarded to the pods resource requirements.

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L165)

### .spec.single.volumeMounts: []ServerGroupSpecVolumeMount

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation of ServerGroupSpecVolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L201)

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

ClientCASecretName setting specifies the name of a kubernetes `Secret` that contains
a PEM encoded CA certificate used for client certificate verification
in all ArangoSync master servers.
This is a required setting when `spec.sync.enabled` is `true`.

[Code Reference](/pkg/apis/deployment/v1/sync_authentication_spec.go#L41)

### .spec.sync.auth.jwtSecretName: string

JWTSecretName setting specifies the name of a kubernetes `Secret` that contains
the JWT token used for accessing all ArangoSync master servers.
When not specified, the `spec.auth.jwtSecretName` value is used.
If you specify a name of a `Secret` that does not exist, a random token is created
and stored in a `Secret` with given name.

[Code Reference](/pkg/apis/deployment/v1/sync_authentication_spec.go#L36)

### .spec.sync.enabled: bool

Enabled setting enables/disables support for data center 2 data center
replication in the cluster. When enabled, the cluster will contain
a number of `syncmaster` & `syncworker` servers.

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/sync_spec.go#L34)

### .spec.sync.externalAccess.accessPackageSecretNames: []string

AccessPackageSecretNames setting specifies the names of zero of more `Secrets` that will be created by the deployment
operator containing "access packages". An access package contains those `Secrets` that are needed
to access the SyncMasters of this `ArangoDeployment`.
By removing a name from this setting, the corresponding `Secret` is also deleted.
Note that to remove all access packages, leave an empty array in place (`[]`).
Completely removing the setting results in not modifying the list.

Links:
* [See the ArangoDeploymentReplication specification](deployment-replication-resource-reference.md)

[Code Reference](/pkg/apis/deployment/v1/sync_external_access_spec.go#L49)

### .spec.sync.externalAccess.advertisedEndpoint: string

AdvertisedEndpoint is passed to the coordinators/single servers for advertising a specific endpoint

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L58)

### .spec.sync.externalAccess.loadBalancerIP: string

LoadBalancerIP define optional IP used to configure a load-balancer on, in case of Auto or LoadBalancer type.
If you do not specify this setting, an IP will be chosen automatically by the load-balancer provisioner.

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L48)

### .spec.sync.externalAccess.loadBalancerSourceRanges: []string

LoadBalancerSourceRanges define LoadBalancerSourceRanges used for LoadBalancer Service type
If specified and supported by the platform, this will restrict traffic through the cloud-provider
load-balancer will be restricted to the specified client IPs. This field will be ignored if the
cloud-provider does not support the feature.

Links:
* [Cloud Provider Firewall](https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/)

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L55)

### .spec.sync.externalAccess.managedServiceNames: []string

ManagedServiceNames keeps names of services which are not managed by KubeArangoDB.
It is only relevant when type of service is `managed`.

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L62)

### .spec.sync.externalAccess.masterEndpoint: []string

MasterEndpoint setting specifies the master endpoint(s) advertised by the ArangoSync SyncMasters.
If not set, this setting defaults to:
- If `spec.sync.externalAccess.loadBalancerIP` is set, it defaults to `https://<load-balancer-ip>:<8629>`.
- Otherwise it defaults to `https://<sync-service-dns-name>:<8629>`.

[Code Reference](/pkg/apis/deployment/v1/sync_external_access_spec.go#L40)

### .spec.sync.externalAccess.nodePort: int

NodePort define optional port used in case of Auto or NodePort type.
This setting is used when `spec.externalAccess.type` is set to `NodePort` or `Auto`.
If you do not specify this setting, a random port will be chosen automatically.

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L44)

### .spec.sync.externalAccess.type: string

Type specifies the type of Service that will be created to provide access to the ArangoDB deployment from outside the Kubernetes cluster.

Possible Values: 
* Auto (default) - Create a Service of type LoadBalancer and fallback to a Service or type NodePort when the LoadBalancer is not assigned an IP address.
* None - limit access to application running inside the Kubernetes cluster.
* LoadBalancer - Create a Service of type LoadBalancer for the ArangoDB deployment.
* NodePort - Create a Service of type NodePort for the ArangoDB deployment.

[Code Reference](/pkg/apis/deployment/v1/external_access_spec.go#L39)

### .spec.sync.image: string

[Code Reference](/pkg/apis/deployment/v1/sync_spec.go#L40)

### .spec.sync.monitoring.tokenSecretName: string

TokenSecretName setting specifies the name of a kubernetes `Secret` that contains
the bearer token used for accessing all monitoring endpoints of all arangod/arangosync servers.
When not specified, no monitoring token is used.

[Code Reference](/pkg/apis/deployment/v1/sync_monitoring_spec.go#L34)

### .spec.sync.tls.altNames: []string

AltNames setting specifies a list of alternate names that will be added to all generated
certificates. These names can be DNS names or email addresses.
The default value is empty.

[Code Reference](/pkg/apis/deployment/v1/tls_spec.go#L72)

### .spec.sync.tls.caSecretName: string

CASecretName  setting specifies the name of a kubernetes `Secret` that contains
a standard CA certificate + private key used to sign certificates for individual
ArangoDB servers.
When no name is specified, it defaults to `<deployment-name>-ca`.
To disable authentication, set this value to `None`.
If you specify a name of a `Secret` that does not exist, a self-signed CA certificate + key is created
and stored in a `Secret` with given name.
The specified `Secret`, must contain the following data fields:
- `ca.crt` PEM encoded public key of the CA certificate
- `ca.key` PEM encoded private key of the CA certificate

[Code Reference](/pkg/apis/deployment/v1/tls_spec.go#L67)

### .spec.sync.tls.mode: string

[Code Reference](/pkg/apis/deployment/v1/tls_spec.go#L81)

### .spec.sync.tls.sni.mapping.\<string\>: []string

[Code Reference](/pkg/apis/deployment/v1/tls_sni_spec.go#L30)

### .spec.sync.tls.ttl: string

TTL setting specifies the time to live of all generated server certificates.
When the server certificate is about to expire, it will be automatically replaced
by a new one and the affected server will be restarted.
Note: The time to live of the CA certificate (when created automatically)
will be set to 10 years.

Default Value: "2160h" (about 3 months)

[Code Reference](/pkg/apis/deployment/v1/tls_spec.go#L79)

### .spec.syncmasters.affinity: core.PodAffinity

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.PodAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L180)

### .spec.syncmasters.allowMemberRecreation: bool

AllowMemberRecreation allows to recreate member.
This setting changes the member recreation logic based on group:
- For Sync Masters, Sync Workers, Coordinator and DB-Servers it determines if a member can be recreated in case of failure (default `true`)
- For Agents and Single this value is hardcoded to `false` and the value provided in spec is ignored.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L222)

### .spec.syncmasters.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.
Annotations are merged with `spec.annotations`.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L125)

### .spec.syncmasters.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L127)

### .spec.syncmasters.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L129)

### .spec.syncmasters.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L176)

### .spec.syncmasters.args: []string

Args setting specifies additional command-line arguments passed to all servers of this group.

Default Value: []

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L85)

### .spec.syncmasters.count: int

Count setting specifies the number of servers to start for the given group.
For the Agent group, this value must be a positive, odd number.
The default value is `3` for all groups except `single` (there the default is `1`
for `spec.mode: Single` and `2` for `spec.mode: ActiveFailover`).
For the `syncworkers` group, it is highly recommended to use the same number
as for the `dbservers` group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L77)

### .spec.syncmasters.entrypoint: string

Entrypoint overrides container executable

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L87)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L235)

### .spec.syncmasters.extendedRotationCheck: bool

ExtendedRotationCheck extend checks for rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L205)

### .spec.syncmasters.externalPortEnabled: bool

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L217)

### .spec.syncmasters.indexMethod: string

IndexMethod define group Indexing method

Possible Values: 
* random (default) - Pick random ID for member. Enforced on the Community Operator.
* ordered - Use sequential number as Member ID, starting from 0. Enterprise Operator required.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L228)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L213)

### .spec.syncmasters.internalPortProtocol: string

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L215)

### .spec.syncmasters.labels: map[string]string

Labels specified the labels added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L131)

### .spec.syncmasters.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L133)

### .spec.syncmasters.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L135)

### .spec.syncmasters.maxCount: int

MaxCount specifies a maximum for the count of servers. If set, a specification is invalid if `count > maxCount`.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L81)

### .spec.syncmasters.memoryReservation: int64

MemoryReservation determines system reservation of memory while calculating `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` value.
If this field is set, `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` is reduced by specified value in percents.
Accepted Range <0, 50>. If value is outside accepted range, it is adjusted to the closest value.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: 0

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L107)

### .spec.syncmasters.minCount: int

MinCount specifies a minimum for the count of servers. If set, a specification is invalid if `count < minCount`.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L79)

### .spec.syncmasters.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L184)

### .spec.syncmasters.nodeSelector: map[string]string

NodeSelector setting specifies a set of labels to be used as `nodeSelector` for Pods of this node.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L150)

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

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` Container Environment Variable

OverrideDetectedNumberOfCores determines if number of cores should be overridden based on values in resources.
If is set to true and Container CPU Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` to the value from the Container CPU Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: true

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L113)

### .spec.syncmasters.overrideDetectedTotalMemory: bool

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` Container Environment Variable

OverrideDetectedTotalMemory determines if memory should be overridden based on values in resources.
If is set to true and Container Memory Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` to the value from the Container Memory Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: true

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L101)

### .spec.syncmasters.podModes.network: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L31)

### .spec.syncmasters.podModes.pid: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L32)

### .spec.syncmasters.port: uint16

Port define Port used by member

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L233)

### .spec.syncmasters.priorityClassName: string

PriorityClassName specifies a priority class name
Will be forwarded to the pod spec.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L156)

### .spec.syncmasters.probes.livenessProbeDisabled: bool

LivenessProbeDisabled if set to true, the operator does not generate a liveness probe for new pods belonging to this group

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L245)

### .spec.syncmasters.probes.livenessProbeSpec.failureThreshold: int32

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: 3

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L295)

### .spec.syncmasters.probes.livenessProbeSpec.initialDelaySeconds: int32

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L278)

### .spec.syncmasters.probes.livenessProbeSpec.periodSeconds: int32

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: 10

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L282)

### .spec.syncmasters.probes.livenessProbeSpec.successThreshold: int32

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: 1

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L290)

### .spec.syncmasters.probes.livenessProbeSpec.timeoutSeconds: int32

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L286)

### .spec.syncmasters.probes.ReadinessProbeDisabled: bool

OldReadinessProbeDisabled if true readinessProbes are disabled
Deprecated: This field is deprecated, keept only for backward compatibility.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L252)

### .spec.syncmasters.probes.readinessProbeDisabled: bool

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L254)

### .spec.syncmasters.probes.readinessProbeSpec.failureThreshold: int32

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: 3

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L295)

### .spec.syncmasters.probes.readinessProbeSpec.initialDelaySeconds: int32

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L278)

### .spec.syncmasters.probes.readinessProbeSpec.periodSeconds: int32

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: 10

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L282)

### .spec.syncmasters.probes.readinessProbeSpec.successThreshold: int32

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: 1

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L290)

### .spec.syncmasters.probes.readinessProbeSpec.timeoutSeconds: int32

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L286)

### .spec.syncmasters.probes.startupProbeDisabled: bool

StartupProbeDisabled if true startupProbes are disabled

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L259)

### .spec.syncmasters.probes.startupProbeSpec.failureThreshold: int32

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: 3

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L295)

### .spec.syncmasters.probes.startupProbeSpec.initialDelaySeconds: int32

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L278)

### .spec.syncmasters.probes.startupProbeSpec.periodSeconds: int32

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: 10

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L282)

### .spec.syncmasters.probes.startupProbeSpec.successThreshold: int32

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: 1

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L290)

### .spec.syncmasters.probes.startupProbeSpec.timeoutSeconds: int32

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L286)

### .spec.syncmasters.pvcResizeMode: string

VolumeResizeMode specified resize mode for PVCs and PVs

Possible Values: 
* runtime (default) - PVC will be resized in Pod runtime (EKS, GKE)
* rotate - Pod will be shutdown and PVC will be resized (AKS)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L170)

### .spec.syncmasters.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L95)

### .spec.syncmasters.schedulerName: string

SchedulerName define scheduler name used for group

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L89)

### .spec.syncmasters.securityContext.addCapabilities: []core.Capability

AddCapabilities add new capabilities to containers

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L46)

### .spec.syncmasters.securityContext.allowPrivilegeEscalation: bool

AllowPrivilegeEscalation Controls whether a process can gain more privileges than its parent process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)

### .spec.syncmasters.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L43)

### .spec.syncmasters.securityContext.fsGroup: int64

FSGroup is a special supplemental group that applies to all containers in a pod.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L66)

### .spec.syncmasters.securityContext.privileged: bool

Privileged If true, runs container in privileged mode. Processes in privileged containers are
essentially equivalent to root on the host.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L52)

### .spec.syncmasters.securityContext.readOnlyRootFilesystem: bool

ReadOnlyRootFilesystem if true, mounts the container's root filesystem as read-only.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L54)

### .spec.syncmasters.securityContext.runAsGroup: int64

RunAsGroup is the GID to run the entrypoint of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L60)

### .spec.syncmasters.securityContext.runAsNonRoot: bool

RunAsNonRoot if true, indicates that the container must run as a non-root user.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L56)

### .spec.syncmasters.securityContext.runAsUser: int64

RunAsUser is the UID to run the entrypoint of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L58)

### .spec.syncmasters.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L82)

### .spec.syncmasters.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L87)

### .spec.syncmasters.securityContext.supplementalGroups: []int64

SupplementalGroups is a list of groups applied to the first process run in each container, in addition to the container's primary GID,
the fsGroup (if specified), and group memberships defined in the container image for the uid of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L64)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L77)

### .spec.syncmasters.serviceAccountName: string

ServiceAccountName setting specifies the `serviceAccountName` for the `Pods` created
for each server of this group. If empty, it defaults to using the
`default` service account.
Using an alternative `ServiceAccount` is typically used to separate access rights.
The ArangoDB deployments need some very minimal access rights. With the
deployment of the operator, we grant the rights to 'get' all 'pod' resources.
If you are using a different service account, please grant these rights
to that service account.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L146)

### .spec.syncmasters.shutdownDelay: int

ShutdownDelay define how long operator should delay finalizer removal after shutdown

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L211)

### .spec.syncmasters.shutdownMethod: string

ShutdownMethod describe procedure of member shutdown taken by Operator

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L209)

### .spec.syncmasters.sidecarCoreNames: []string

SidecarCoreNames is a list of sidecar containers which must run in the pod.
Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L187)

### .spec.syncmasters.sidecars: []core.Container

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L191)

### .spec.syncmasters.storageClassName: string

StorageClassName specifies the classname for storage of the servers.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L91)

### .spec.syncmasters.terminationGracePeriodSeconds: int64

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.syncmasters.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.
By default, suitable tolerations are set for the following keys with the `NoExecute` effect:
- `node.kubernetes.io/not-ready`
- `node.kubernetes.io/unreachable`
- `node.alpha.kubernetes.io/unreachable` (will be removed in future version)
For more information on tolerations, consult the https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L122)

### .spec.syncmasters.volumeAllowShrink: bool

Deprecated: VolumeAllowShrink allows shrink the volume

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L172)

### .spec.syncmasters.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a volumeClaimTemplate used by operator to create to volume claims for pods of this group.
This setting is not available for group `coordinators`, `syncmasters` & `syncworkers`.
The default value describes a volume with `8Gi` storage, `ReadWriteOnce` access mode and volume mode set to `PersistentVolumeFilesystem`.
If this field is not set and `spec.<group>.resources.requests.storage` is set, then a default volume claim
with size as specified by `spec.<group>.resources.requests.storage` will be created. In that case `storage`
and `iops` is not forwarded to the pods resource requirements.

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L165)

### .spec.syncmasters.volumeMounts: []ServerGroupSpecVolumeMount

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation of ServerGroupSpecVolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L201)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L180)

### .spec.syncworkers.allowMemberRecreation: bool

AllowMemberRecreation allows to recreate member.
This setting changes the member recreation logic based on group:
- For Sync Masters, Sync Workers, Coordinator and DB-Servers it determines if a member can be recreated in case of failure (default `true`)
- For Agents and Single this value is hardcoded to `false` and the value provided in spec is ignored.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L222)

### .spec.syncworkers.annotations: map[string]string

Annotations specified the annotations added to Pods in this group.
Annotations are merged with `spec.annotations`.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L125)

### .spec.syncworkers.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L127)

### .spec.syncworkers.annotationsMode: string

AnnotationsMode Define annotations mode which should be use while overriding annotations

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L129)

### .spec.syncworkers.antiAffinity: core.PodAntiAffinity

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L176)

### .spec.syncworkers.args: []string

Args setting specifies additional command-line arguments passed to all servers of this group.

Default Value: []

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L85)

### .spec.syncworkers.count: int

Count setting specifies the number of servers to start for the given group.
For the Agent group, this value must be a positive, odd number.
The default value is `3` for all groups except `single` (there the default is `1`
for `spec.mode: Single` and `2` for `spec.mode: ActiveFailover`).
For the `syncworkers` group, it is highly recommended to use the same number
as for the `dbservers` group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L77)

### .spec.syncworkers.entrypoint: string

Entrypoint overrides container executable

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L87)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L235)

### .spec.syncworkers.extendedRotationCheck: bool

ExtendedRotationCheck extend checks for rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L205)

### .spec.syncworkers.externalPortEnabled: bool

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L217)

### .spec.syncworkers.indexMethod: string

IndexMethod define group Indexing method

Possible Values: 
* random (default) - Pick random ID for member. Enforced on the Community Operator.
* ordered - Use sequential number as Member ID, starting from 0. Enterprise Operator required.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L228)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L213)

### .spec.syncworkers.internalPortProtocol: string

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L215)

### .spec.syncworkers.labels: map[string]string

Labels specified the labels added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L131)

### .spec.syncworkers.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L133)

### .spec.syncworkers.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L135)

### .spec.syncworkers.maxCount: int

MaxCount specifies a maximum for the count of servers. If set, a specification is invalid if `count > maxCount`.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L81)

### .spec.syncworkers.memoryReservation: int64

MemoryReservation determines system reservation of memory while calculating `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` value.
If this field is set, `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` is reduced by specified value in percents.
Accepted Range <0, 50>. If value is outside accepted range, it is adjusted to the closest value.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: 0

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L107)

### .spec.syncworkers.minCount: int

MinCount specifies a minimum for the count of servers. If set, a specification is invalid if `count < minCount`.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L79)

### .spec.syncworkers.nodeAffinity: core.NodeAffinity

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L184)

### .spec.syncworkers.nodeSelector: map[string]string

NodeSelector setting specifies a set of labels to be used as `nodeSelector` for Pods of this node.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L150)

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

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` Container Environment Variable

OverrideDetectedNumberOfCores determines if number of cores should be overridden based on values in resources.
If is set to true and Container CPU Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` to the value from the Container CPU Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: true

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L113)

### .spec.syncworkers.overrideDetectedTotalMemory: bool

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` Container Environment Variable

OverrideDetectedTotalMemory determines if memory should be overridden based on values in resources.
If is set to true and Container Memory Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` to the value from the Container Memory Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: true

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L101)

### .spec.syncworkers.podModes.network: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L31)

### .spec.syncworkers.podModes.pid: string

[Code Reference](/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L32)

### .spec.syncworkers.port: uint16

Port define Port used by member

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L233)

### .spec.syncworkers.priorityClassName: string

PriorityClassName specifies a priority class name
Will be forwarded to the pod spec.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L156)

### .spec.syncworkers.probes.livenessProbeDisabled: bool

LivenessProbeDisabled if set to true, the operator does not generate a liveness probe for new pods belonging to this group

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L245)

### .spec.syncworkers.probes.livenessProbeSpec.failureThreshold: int32

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: 3

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L295)

### .spec.syncworkers.probes.livenessProbeSpec.initialDelaySeconds: int32

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L278)

### .spec.syncworkers.probes.livenessProbeSpec.periodSeconds: int32

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: 10

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L282)

### .spec.syncworkers.probes.livenessProbeSpec.successThreshold: int32

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: 1

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L290)

### .spec.syncworkers.probes.livenessProbeSpec.timeoutSeconds: int32

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L286)

### .spec.syncworkers.probes.ReadinessProbeDisabled: bool

OldReadinessProbeDisabled if true readinessProbes are disabled
Deprecated: This field is deprecated, keept only for backward compatibility.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L252)

### .spec.syncworkers.probes.readinessProbeDisabled: bool

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L254)

### .spec.syncworkers.probes.readinessProbeSpec.failureThreshold: int32

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: 3

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L295)

### .spec.syncworkers.probes.readinessProbeSpec.initialDelaySeconds: int32

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L278)

### .spec.syncworkers.probes.readinessProbeSpec.periodSeconds: int32

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: 10

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L282)

### .spec.syncworkers.probes.readinessProbeSpec.successThreshold: int32

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: 1

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L290)

### .spec.syncworkers.probes.readinessProbeSpec.timeoutSeconds: int32

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L286)

### .spec.syncworkers.probes.startupProbeDisabled: bool

StartupProbeDisabled if true startupProbes are disabled

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L259)

### .spec.syncworkers.probes.startupProbeSpec.failureThreshold: int32

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: 3

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L295)

### .spec.syncworkers.probes.startupProbeSpec.initialDelaySeconds: int32

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L278)

### .spec.syncworkers.probes.startupProbeSpec.periodSeconds: int32

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: 10

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L282)

### .spec.syncworkers.probes.startupProbeSpec.successThreshold: int32

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: 1

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L290)

### .spec.syncworkers.probes.startupProbeSpec.timeoutSeconds: int32

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: 2

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L286)

### .spec.syncworkers.pvcResizeMode: string

VolumeResizeMode specified resize mode for PVCs and PVs

Possible Values: 
* runtime (default) - PVC will be resized in Pod runtime (EKS, GKE)
* rotate - Pod will be shutdown and PVC will be resized (AKS)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L170)

### .spec.syncworkers.resources: core.ResourceRequirements

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L95)

### .spec.syncworkers.schedulerName: string

SchedulerName define scheduler name used for group

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L89)

### .spec.syncworkers.securityContext.addCapabilities: []core.Capability

AddCapabilities add new capabilities to containers

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L46)

### .spec.syncworkers.securityContext.allowPrivilegeEscalation: bool

AllowPrivilegeEscalation Controls whether a process can gain more privileges than its parent process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)

### .spec.syncworkers.securityContext.dropAllCapabilities: bool

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L43)

### .spec.syncworkers.securityContext.fsGroup: int64

FSGroup is a special supplemental group that applies to all containers in a pod.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L66)

### .spec.syncworkers.securityContext.privileged: bool

Privileged If true, runs container in privileged mode. Processes in privileged containers are
essentially equivalent to root on the host.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L52)

### .spec.syncworkers.securityContext.readOnlyRootFilesystem: bool

ReadOnlyRootFilesystem if true, mounts the container's root filesystem as read-only.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L54)

### .spec.syncworkers.securityContext.runAsGroup: int64

RunAsGroup is the GID to run the entrypoint of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L60)

### .spec.syncworkers.securityContext.runAsNonRoot: bool

RunAsNonRoot if true, indicates that the container must run as a non-root user.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L56)

### .spec.syncworkers.securityContext.runAsUser: int64

RunAsUser is the UID to run the entrypoint of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L58)

### .spec.syncworkers.securityContext.seccompProfile: core.SeccompProfile

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L82)

### .spec.syncworkers.securityContext.seLinuxOptions: core.SELinuxOptions

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L87)

### .spec.syncworkers.securityContext.supplementalGroups: []int64

SupplementalGroups is a list of groups applied to the first process run in each container, in addition to the container's primary GID,
the fsGroup (if specified), and group memberships defined in the container image for the uid of the container process.

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L64)

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

[Code Reference](/pkg/apis/deployment/v1/server_group_security_context_spec.go#L77)

### .spec.syncworkers.serviceAccountName: string

ServiceAccountName setting specifies the `serviceAccountName` for the `Pods` created
for each server of this group. If empty, it defaults to using the
`default` service account.
Using an alternative `ServiceAccount` is typically used to separate access rights.
The ArangoDB deployments need some very minimal access rights. With the
deployment of the operator, we grant the rights to 'get' all 'pod' resources.
If you are using a different service account, please grant these rights
to that service account.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L146)

### .spec.syncworkers.shutdownDelay: int

ShutdownDelay define how long operator should delay finalizer removal after shutdown

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L211)

### .spec.syncworkers.shutdownMethod: string

ShutdownMethod describe procedure of member shutdown taken by Operator

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L209)

### .spec.syncworkers.sidecarCoreNames: []string

SidecarCoreNames is a list of sidecar containers which must run in the pod.
Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L187)

### .spec.syncworkers.sidecars: []core.Container

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L191)

### .spec.syncworkers.storageClassName: string

StorageClassName specifies the classname for storage of the servers.

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L91)

### .spec.syncworkers.terminationGracePeriodSeconds: int64

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L224)

### .spec.syncworkers.tolerations: []core.Toleration

Tolerations specifies the tolerations added to Pods in this group.
By default, suitable tolerations are set for the following keys with the `NoExecute` effect:
- `node.kubernetes.io/not-ready`
- `node.kubernetes.io/unreachable`
- `node.alpha.kubernetes.io/unreachable` (will be removed in future version)
For more information on tolerations, consult the https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L122)

### .spec.syncworkers.volumeAllowShrink: bool

Deprecated: VolumeAllowShrink allows shrink the volume

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L172)

### .spec.syncworkers.volumeClaimTemplate: core.PersistentVolumeClaim

VolumeClaimTemplate specifies a volumeClaimTemplate used by operator to create to volume claims for pods of this group.
This setting is not available for group `coordinators`, `syncmasters` & `syncworkers`.
The default value describes a volume with `8Gi` storage, `ReadWriteOnce` access mode and volume mode set to `PersistentVolumeFilesystem`.
If this field is not set and `spec.<group>.resources.requests.storage` is set, then a default volume claim
with size as specified by `spec.<group>.resources.requests.storage` will be created. In that case `storage`
and `iops` is not forwarded to the pods resource requirements.

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L165)

### .spec.syncworkers.volumeMounts: []ServerGroupSpecVolumeMount

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation of ServerGroupSpecVolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

[Code Reference](/pkg/apis/deployment/v1/server_group_spec.go#L201)

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

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L269)

### .spec.tls.altNames: []string

AltNames setting specifies a list of alternate names that will be added to all generated
certificates. These names can be DNS names or email addresses.
The default value is empty.

[Code Reference](/pkg/apis/deployment/v1/tls_spec.go#L72)

### .spec.tls.caSecretName: string

CASecretName  setting specifies the name of a kubernetes `Secret` that contains
a standard CA certificate + private key used to sign certificates for individual
ArangoDB servers.
When no name is specified, it defaults to `<deployment-name>-ca`.
To disable authentication, set this value to `None`.
If you specify a name of a `Secret` that does not exist, a self-signed CA certificate + key is created
and stored in a `Secret` with given name.
The specified `Secret`, must contain the following data fields:
- `ca.crt` PEM encoded public key of the CA certificate
- `ca.key` PEM encoded private key of the CA certificate

[Code Reference](/pkg/apis/deployment/v1/tls_spec.go#L67)

### .spec.tls.mode: string

[Code Reference](/pkg/apis/deployment/v1/tls_spec.go#L81)

### .spec.tls.sni.mapping.\<string\>: []string

[Code Reference](/pkg/apis/deployment/v1/tls_sni_spec.go#L30)

### .spec.tls.ttl: string

TTL setting specifies the time to live of all generated server certificates.
When the server certificate is about to expire, it will be automatically replaced
by a new one and the affected server will be restarted.
Note: The time to live of the CA certificate (when created automatically)
will be set to 10 years.

Default Value: "2160h" (about 3 months)

[Code Reference](/pkg/apis/deployment/v1/tls_spec.go#L79)

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

