# API Reference for ArangoDeployment V1

## Spec

### .spec.agents.affinity

Type: `core.PodAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L180)</sup>

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.PodAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

***

### .spec.agents.allowMemberRecreation

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L222)</sup>

AllowMemberRecreation allows to recreate member.
This setting changes the member recreation logic based on group:
- For Sync Masters, Sync Workers, Coordinator and DB-Servers it determines if a member can be recreated in case of failure (default `true`)
- For Agents and Single this value is hardcoded to `false` and the value provided in spec is ignored.

***

### .spec.agents.annotations

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L125)</sup>

Annotations specified the annotations added to Pods in this group.
Annotations are merged with `spec.annotations`.

***

### .spec.agents.annotationsIgnoreList

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L127)</sup>

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

***

### .spec.agents.annotationsMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L129)</sup>

AnnotationsMode Define annotations mode which should be use while overriding annotations

***

### .spec.agents.antiAffinity

Type: `core.PodAntiAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L176)</sup>

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

***

### .spec.agents.args

Type: `[]string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L85)</sup>

Args setting specifies additional command-line arguments passed to all servers of this group.

Default Value: `[]`

***

### .spec.agents.count

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L77)</sup>

Count setting specifies the number of servers to start for the given group.
For the Agent group, this value must be a positive, odd number.
The default value is `3` for all groups except `single` (there the default is `1`
for `spec.mode: Single` and `2` for `spec.mode: ActiveFailover`).
For the `syncworkers` group, it is highly recommended to use the same number
as for the `dbservers` group.

***

### .spec.agents.entrypoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L87)</sup>

Entrypoint overrides container executable

***

### .spec.agents.envs\[int\].name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_env_var.go#L26)</sup>

***

### .spec.agents.envs\[int\].value

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_env_var.go#L27)</sup>

***

### .spec.agents.ephemeralVolumes.apps.size

Type: `resource.Quantity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)</sup>

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

***

### .spec.agents.ephemeralVolumes.temp.size

Type: `resource.Quantity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)</sup>

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

***

### .spec.agents.exporterPort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L235)</sup>

ExporterPort define Port used by exporter

***

### .spec.agents.extendedRotationCheck

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L205)</sup>

ExtendedRotationCheck extend checks for rotation

***

### .spec.agents.externalPortEnabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L217)</sup>

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

***

### .spec.agents.indexMethod

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L228)</sup>

IndexMethod define group Indexing method

Possible Values: 
* random (default) - Pick random ID for member. Enforced on the Community Operator.
* ordered - Use sequential number as Member ID, starting from 0. Enterprise Operator required.

***

### .spec.agents.initContainers.containers

Type: `[]core.Container` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_init_containers.go#L91)</sup>

Containers contains list of containers

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

***

### .spec.agents.initContainers.mode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_init_containers.go#L94)</sup>

Mode keep container replace mode

***

### .spec.agents.internalPort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L213)</sup>

InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

***

### .spec.agents.internalPortProtocol

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L215)</sup>

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

***

### .spec.agents.labels

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L131)</sup>

Labels specified the labels added to Pods in this group.

***

### .spec.agents.labelsIgnoreList

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L133)</sup>

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

***

### .spec.agents.labelsMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L135)</sup>

LabelsMode Define labels mode which should be use while overriding labels

***

### .spec.agents.maxCount

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L81)</sup>

MaxCount specifies a maximum for the count of servers. If set, a specification is invalid if `count > maxCount`.

***

### .spec.agents.memoryReservation

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L107)</sup>

MemoryReservation determines the system reservation of memory while calculating `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` value.
If this field is set, `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` is reduced by a specified value in percent.
Accepted Range <0, 50>. If the value is outside the accepted range, it is adjusted to the closest value.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: `0`

***

### .spec.agents.minCount

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L79)</sup>

MinCount specifies a minimum for the count of servers. If set, a specification is invalid if `count < minCount`.

***

### .spec.agents.nodeAffinity

Type: `core.NodeAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L184)</sup>

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

***

### .spec.agents.nodeSelector

Type: `map[string]string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L150)</sup>

NodeSelector setting specifies a set of labels to be used as `nodeSelector` for Pods of this node.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/)

***

### .spec.agents.numactl.args

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_numactl_spec.go#L38)</sup>

Args define list of the numactl process

Default Value: `[]`

***

### .spec.agents.numactl.enabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_numactl_spec.go#L30)</sup>

Enabled define if numactl should be enabled

Default Value: `false`

***

### .spec.agents.numactl.path

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_numactl_spec.go#L34)</sup>

Path define numactl path within the container

Default Value: `/usr/bin/numactl`

***

### .spec.agents.overrideDetectedNumberOfCores

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L113)</sup>

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` Container Environment Variable

OverrideDetectedNumberOfCores determines if number of cores should be overridden based on values in resources.
If is set to true and Container CPU Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` to the value from the Container CPU Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: `true`

***

### .spec.agents.overrideDetectedTotalMemory

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L101)</sup>

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` Container Environment Variable

OverrideDetectedTotalMemory determines if memory should be overridden based on values in resources.
If is set to true and Container Memory Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` to the value from the Container Memory Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: `true`

***

### .spec.agents.podModes.network

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L31)</sup>

***

### .spec.agents.podModes.pid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L32)</sup>

***

### .spec.agents.port

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L233)</sup>

Port define Port used by member

***

### .spec.agents.priorityClassName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L156)</sup>

PriorityClassName specifies a priority class name
Will be forwarded to the pod spec.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/)

***

### .spec.agents.probes.livenessProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L245)</sup>

LivenessProbeDisabled if set to true, the operator does not generate a liveness probe for new pods belonging to this group

Default Value: `false`

***

### .spec.agents.probes.livenessProbeSpec.failureThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L295)</sup>

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: `3`

***

### .spec.agents.probes.livenessProbeSpec.initialDelaySeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L278)</sup>

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: `2`

***

### .spec.agents.probes.livenessProbeSpec.periodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L282)</sup>

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: `10`

***

### .spec.agents.probes.livenessProbeSpec.successThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L290)</sup>

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: `1`

***

### .spec.agents.probes.livenessProbeSpec.timeoutSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L286)</sup>

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: `2`

***

### .spec.agents.probes.ReadinessProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L252)</sup>

OldReadinessProbeDisabled if true readinessProbes are disabled
Deprecated: This field is deprecated, kept only for backward compatibility.

***

### .spec.agents.probes.readinessProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L254)</sup>

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

***

### .spec.agents.probes.readinessProbeSpec.failureThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L295)</sup>

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: `3`

***

### .spec.agents.probes.readinessProbeSpec.initialDelaySeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L278)</sup>

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: `2`

***

### .spec.agents.probes.readinessProbeSpec.periodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L282)</sup>

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: `10`

***

### .spec.agents.probes.readinessProbeSpec.successThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L290)</sup>

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: `1`

***

### .spec.agents.probes.readinessProbeSpec.timeoutSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L286)</sup>

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: `2`

***

### .spec.agents.probes.startupProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L259)</sup>

StartupProbeDisabled if true startupProbes are disabled

***

### .spec.agents.probes.startupProbeSpec.failureThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L295)</sup>

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: `3`

***

### .spec.agents.probes.startupProbeSpec.initialDelaySeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L278)</sup>

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: `2`

***

### .spec.agents.probes.startupProbeSpec.periodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L282)</sup>

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: `10`

***

### .spec.agents.probes.startupProbeSpec.successThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L290)</sup>

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: `1`

***

### .spec.agents.probes.startupProbeSpec.timeoutSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L286)</sup>

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: `2`

***

### .spec.agents.pvcResizeMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L170)</sup>

VolumeResizeMode specified resize mode for PVCs and PVs

Possible Values: 
* runtime (default) - PVC will be resized in Pod runtime (EKS, GKE)
* rotate - Pod will be shutdown and PVC will be resized (AKS)

***

### .spec.agents.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L95)</sup>

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

***

### .spec.agents.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L89)</sup>

SchedulerName define scheduler name used for group

***

### .spec.agents.securityContext.addCapabilities

Type: `[]core.Capability` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L41)</sup>

AddCapabilities add new capabilities to containers

***

### .spec.agents.securityContext.allowPrivilegeEscalation

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L44)</sup>

AllowPrivilegeEscalation Controls whether a process can gain more privileges than its parent process.

***

### .spec.agents.securityContext.dropAllCapabilities

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L38)</sup>

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

***

### .spec.agents.securityContext.fsGroup

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L61)</sup>

FSGroup is a special supplemental group that applies to all containers in a pod.

***

### .spec.agents.securityContext.privileged

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L47)</sup>

Privileged If true, runs container in privileged mode. Processes in privileged containers are
essentially equivalent to root on the host.

***

### .spec.agents.securityContext.readOnlyRootFilesystem

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)</sup>

ReadOnlyRootFilesystem if true, mounts the container's root filesystem as read-only.

***

### .spec.agents.securityContext.runAsGroup

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L55)</sup>

RunAsGroup is the GID to run the entrypoint of the container process.

***

### .spec.agents.securityContext.runAsNonRoot

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L51)</sup>

RunAsNonRoot if true, indicates that the container must run as a non-root user.

***

### .spec.agents.securityContext.runAsUser

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L53)</sup>

RunAsUser is the UID to run the entrypoint of the container process.

***

### .spec.agents.securityContext.seccompProfile

Type: `core.SeccompProfile` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L77)</sup>

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

***

### .spec.agents.securityContext.seLinuxOptions

Type: `core.SELinuxOptions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L82)</sup>

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

***

### .spec.agents.securityContext.supplementalGroups

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L59)</sup>

SupplementalGroups is a list of groups applied to the first process run in each container, in addition to the container's primary GID,
the fsGroup (if specified), and group memberships defined in the container image for the uid of the container process.

***

### .spec.agents.securityContext.sysctls

Type: `map[string]intstr.IntOrString` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L72)</sup>

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

***

### .spec.agents.serviceAccountName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L146)</sup>

ServiceAccountName setting specifies the `serviceAccountName` for the `Pods` created
for each server of this group. If empty, it defaults to using the
`default` service account.
Using an alternative `ServiceAccount` is typically used to separate access rights.
The ArangoDB deployments need some very minimal access rights. With the
deployment of the operator, we grant the rights to 'get' all 'pod' resources.
If you are using a different service account, please grant these rights
to that service account.

***

### .spec.agents.shutdownDelay

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L211)</sup>

ShutdownDelay define how long operator should delay finalizer removal after shutdown

***

### .spec.agents.shutdownMethod

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L209)</sup>

ShutdownMethod describe procedure of member shutdown taken by Operator

***

### .spec.agents.sidecarCoreNames

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L187)</sup>

SidecarCoreNames is a list of sidecar containers which must run in the pod.
Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

***

### .spec.agents.sidecars

Type: `[]core.Container` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L191)</sup>

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

***

### .spec.agents.storageClassName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L91)</sup>

StorageClassName specifies the classname for storage of the servers.

***

### .spec.agents.terminationGracePeriodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L224)</sup>

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

***

### .spec.agents.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L122)</sup>

Tolerations specifies the tolerations added to Pods in this group.
By default, suitable tolerations are set for the following keys with the `NoExecute` effect:
- `node.kubernetes.io/not-ready`
- `node.kubernetes.io/unreachable`
- `node.alpha.kubernetes.io/unreachable` (will be removed in future version)
For more information on tolerations, consult the https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

***

### .spec.agents.volumeAllowShrink

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L172)</sup>

Deprecated: VolumeAllowShrink allows shrink the volume

***

### .spec.agents.volumeClaimTemplate

Type: `core.PersistentVolumeClaim` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L165)</sup>

VolumeClaimTemplate specifies a volumeClaimTemplate used by operator to create to volume claims for pods of this group.
This setting is not available for group `coordinators`, `syncmasters` & `syncworkers`.
The default value describes a volume with `8Gi` storage, `ReadWriteOnce` access mode and volume mode set to `PersistentVolumeFilesystem`.
If this field is not set and `spec.<group>.resources.requests.storage` is set, then a default volume claim
with size as specified by `spec.<group>.resources.requests.storage` will be created. In that case `storage`
and `iops` is not forwarded to the pods resource requirements.

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

***

### .spec.agents.volumeMounts

Type: `[]ServerGroupSpecVolumeMount` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L201)</sup>

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation of ServerGroupSpecVolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

***

### .spec.agents.volumes\[int\].configMap

Type: `core.ConfigMapVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L138)</sup>

ConfigMap which should be mounted into pod

Links:
* [Documentation of core.ConfigMapVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#configmapvolumesource-v1-core)

***

### .spec.agents.volumes\[int\].emptyDir

Type: `core.EmptyDirVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L143)</sup>

EmptyDir

Links:
* [Documentation of core.EmptyDirVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#emptydirvolumesource-v1-core)

***

### .spec.agents.volumes\[int\].hostPath

Type: `core.HostPathVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L148)</sup>

HostPath

Links:
* [Documentation of core.HostPathVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#hostpathvolumesource-v1-core)

***

### .spec.agents.volumes\[int\].name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L128)</sup>

Name of volume

***

### .spec.agents.volumes\[int\].persistentVolumeClaim

Type: `core.PersistentVolumeClaimVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L153)</sup>

PersistentVolumeClaim

Links:
* [Documentation of core.PersistentVolumeClaimVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaimvolumesource-v1-core)

***

### .spec.agents.volumes\[int\].secret

Type: `core.SecretVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L133)</sup>

Secret which should be mounted into pod

Links:
* [Documentation of core.SecretVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#secretvolumesource-v1-core)

***

### .spec.allowUnsafeUpgrade

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L163)</sup>

AllowUnsafeUpgrade determines if upgrade on missing member or with not in sync shards is allowed

***

### .spec.annotations

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L126)</sup>

Annotations specifies the annotations added to all ArangoDeployment owned resources (pods, services, PVC’s, PDB’s).

***

### .spec.annotationsIgnoreList

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L129)</sup>

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

***

### .spec.annotationsMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L135)</sup>

AnnotationsMode defines annotations mode which should be use while overriding annotations.

Possible Values: 
* disabled (default) - Disable annotations/labels override. Default if there is no annotations/labels set in ArangoDeployment
* append - Add new annotations/labels without affecting old ones
* replace - Replace existing annotations/labels

***

### .spec.architecture

Type: `[]string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L265)</sup>

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

Default Value: `['amd64']`

***

### .spec.auth.jwtSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/authentication_spec.go#L40)</sup>

JWTSecretName setting specifies the name of a kubernetes `Secret` that contains
the JWT token used for accessing all ArangoDB servers.
When no name is specified, it defaults to `<deployment-name>-jwt`.
To disable authentication, set this value to `None`.
If you specify a name of a `Secret`, that secret must have the token
in a data field named `token`.
If you specify a name of a `Secret` that does not exist, a random token is created
and stored in a `Secret` with given name.
Changing a JWT token results in restarting of a whole cluster.

***

### .spec.bootstrap.passwordSecretNames

Type: `map[string]string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/bootstrap.go#L62)</sup>

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

***

### .spec.chaos.enabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/chaos_spec.go#L33)</sup>

Enabled switches the chaos monkey for a deployment on or off.

***

### .spec.chaos.interval

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/chaos_spec.go#L35)</sup>

Interval is the time between events

***

### .spec.chaos.kill-pod-probability

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/chaos_spec.go#L37)</sup>

KillPodProbability is the chance of a pod being killed during an event

***

### .spec.ClusterDomain

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L237)</sup>

ClusterDomain define domain used in the kubernetes cluster.
Required only of domain is not set to default (cluster.local)

Default Value: `cluster.local`

***

### .spec.communicationMethod

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L245)</sup>

CommunicationMethod define communication method used in deployment

Possible Values: 
* headless (default) - Define old communication mechanism, based on headless service.
* dns - Define ClusterIP Service DNS based communication.
* short-dns - Define ClusterIP Service DNS based communication. Use namespaced short DNS (used in migration)
* headless-dns - Define Headless Service DNS based communication.
* ip - Define ClusterIP Service IP based communication.

***

### .spec.coordinators.affinity

Type: `core.PodAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L180)</sup>

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.PodAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

***

### .spec.coordinators.allowMemberRecreation

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L222)</sup>

AllowMemberRecreation allows to recreate member.
This setting changes the member recreation logic based on group:
- For Sync Masters, Sync Workers, Coordinator and DB-Servers it determines if a member can be recreated in case of failure (default `true`)
- For Agents and Single this value is hardcoded to `false` and the value provided in spec is ignored.

***

### .spec.coordinators.annotations

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L125)</sup>

Annotations specified the annotations added to Pods in this group.
Annotations are merged with `spec.annotations`.

***

### .spec.coordinators.annotationsIgnoreList

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L127)</sup>

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

***

### .spec.coordinators.annotationsMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L129)</sup>

AnnotationsMode Define annotations mode which should be use while overriding annotations

***

### .spec.coordinators.antiAffinity

Type: `core.PodAntiAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L176)</sup>

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

***

### .spec.coordinators.args

Type: `[]string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L85)</sup>

Args setting specifies additional command-line arguments passed to all servers of this group.

Default Value: `[]`

***

### .spec.coordinators.count

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L77)</sup>

Count setting specifies the number of servers to start for the given group.
For the Agent group, this value must be a positive, odd number.
The default value is `3` for all groups except `single` (there the default is `1`
for `spec.mode: Single` and `2` for `spec.mode: ActiveFailover`).
For the `syncworkers` group, it is highly recommended to use the same number
as for the `dbservers` group.

***

### .spec.coordinators.entrypoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L87)</sup>

Entrypoint overrides container executable

***

### .spec.coordinators.envs\[int\].name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_env_var.go#L26)</sup>

***

### .spec.coordinators.envs\[int\].value

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_env_var.go#L27)</sup>

***

### .spec.coordinators.ephemeralVolumes.apps.size

Type: `resource.Quantity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)</sup>

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

***

### .spec.coordinators.ephemeralVolumes.temp.size

Type: `resource.Quantity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)</sup>

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

***

### .spec.coordinators.exporterPort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L235)</sup>

ExporterPort define Port used by exporter

***

### .spec.coordinators.extendedRotationCheck

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L205)</sup>

ExtendedRotationCheck extend checks for rotation

***

### .spec.coordinators.externalPortEnabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L217)</sup>

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

***

### .spec.coordinators.indexMethod

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L228)</sup>

IndexMethod define group Indexing method

Possible Values: 
* random (default) - Pick random ID for member. Enforced on the Community Operator.
* ordered - Use sequential number as Member ID, starting from 0. Enterprise Operator required.

***

### .spec.coordinators.initContainers.containers

Type: `[]core.Container` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_init_containers.go#L91)</sup>

Containers contains list of containers

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

***

### .spec.coordinators.initContainers.mode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_init_containers.go#L94)</sup>

Mode keep container replace mode

***

### .spec.coordinators.internalPort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L213)</sup>

InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

***

### .spec.coordinators.internalPortProtocol

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L215)</sup>

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

***

### .spec.coordinators.labels

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L131)</sup>

Labels specified the labels added to Pods in this group.

***

### .spec.coordinators.labelsIgnoreList

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L133)</sup>

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

***

### .spec.coordinators.labelsMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L135)</sup>

LabelsMode Define labels mode which should be use while overriding labels

***

### .spec.coordinators.maxCount

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L81)</sup>

MaxCount specifies a maximum for the count of servers. If set, a specification is invalid if `count > maxCount`.

***

### .spec.coordinators.memoryReservation

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L107)</sup>

MemoryReservation determines the system reservation of memory while calculating `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` value.
If this field is set, `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` is reduced by a specified value in percent.
Accepted Range <0, 50>. If the value is outside the accepted range, it is adjusted to the closest value.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: `0`

***

### .spec.coordinators.minCount

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L79)</sup>

MinCount specifies a minimum for the count of servers. If set, a specification is invalid if `count < minCount`.

***

### .spec.coordinators.nodeAffinity

Type: `core.NodeAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L184)</sup>

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

***

### .spec.coordinators.nodeSelector

Type: `map[string]string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L150)</sup>

NodeSelector setting specifies a set of labels to be used as `nodeSelector` for Pods of this node.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/)

***

### .spec.coordinators.numactl.args

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_numactl_spec.go#L38)</sup>

Args define list of the numactl process

Default Value: `[]`

***

### .spec.coordinators.numactl.enabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_numactl_spec.go#L30)</sup>

Enabled define if numactl should be enabled

Default Value: `false`

***

### .spec.coordinators.numactl.path

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_numactl_spec.go#L34)</sup>

Path define numactl path within the container

Default Value: `/usr/bin/numactl`

***

### .spec.coordinators.overrideDetectedNumberOfCores

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L113)</sup>

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` Container Environment Variable

OverrideDetectedNumberOfCores determines if number of cores should be overridden based on values in resources.
If is set to true and Container CPU Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` to the value from the Container CPU Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: `true`

***

### .spec.coordinators.overrideDetectedTotalMemory

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L101)</sup>

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` Container Environment Variable

OverrideDetectedTotalMemory determines if memory should be overridden based on values in resources.
If is set to true and Container Memory Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` to the value from the Container Memory Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: `true`

***

### .spec.coordinators.podModes.network

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L31)</sup>

***

### .spec.coordinators.podModes.pid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L32)</sup>

***

### .spec.coordinators.port

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L233)</sup>

Port define Port used by member

***

### .spec.coordinators.priorityClassName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L156)</sup>

PriorityClassName specifies a priority class name
Will be forwarded to the pod spec.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/)

***

### .spec.coordinators.probes.livenessProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L245)</sup>

LivenessProbeDisabled if set to true, the operator does not generate a liveness probe for new pods belonging to this group

Default Value: `false`

***

### .spec.coordinators.probes.livenessProbeSpec.failureThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L295)</sup>

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: `3`

***

### .spec.coordinators.probes.livenessProbeSpec.initialDelaySeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L278)</sup>

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: `2`

***

### .spec.coordinators.probes.livenessProbeSpec.periodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L282)</sup>

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: `10`

***

### .spec.coordinators.probes.livenessProbeSpec.successThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L290)</sup>

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: `1`

***

### .spec.coordinators.probes.livenessProbeSpec.timeoutSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L286)</sup>

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: `2`

***

### .spec.coordinators.probes.ReadinessProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L252)</sup>

OldReadinessProbeDisabled if true readinessProbes are disabled
Deprecated: This field is deprecated, kept only for backward compatibility.

***

### .spec.coordinators.probes.readinessProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L254)</sup>

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

***

### .spec.coordinators.probes.readinessProbeSpec.failureThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L295)</sup>

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: `3`

***

### .spec.coordinators.probes.readinessProbeSpec.initialDelaySeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L278)</sup>

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: `2`

***

### .spec.coordinators.probes.readinessProbeSpec.periodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L282)</sup>

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: `10`

***

### .spec.coordinators.probes.readinessProbeSpec.successThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L290)</sup>

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: `1`

***

### .spec.coordinators.probes.readinessProbeSpec.timeoutSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L286)</sup>

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: `2`

***

### .spec.coordinators.probes.startupProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L259)</sup>

StartupProbeDisabled if true startupProbes are disabled

***

### .spec.coordinators.probes.startupProbeSpec.failureThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L295)</sup>

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: `3`

***

### .spec.coordinators.probes.startupProbeSpec.initialDelaySeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L278)</sup>

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: `2`

***

### .spec.coordinators.probes.startupProbeSpec.periodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L282)</sup>

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: `10`

***

### .spec.coordinators.probes.startupProbeSpec.successThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L290)</sup>

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: `1`

***

### .spec.coordinators.probes.startupProbeSpec.timeoutSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L286)</sup>

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: `2`

***

### .spec.coordinators.pvcResizeMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L170)</sup>

VolumeResizeMode specified resize mode for PVCs and PVs

Possible Values: 
* runtime (default) - PVC will be resized in Pod runtime (EKS, GKE)
* rotate - Pod will be shutdown and PVC will be resized (AKS)

***

### .spec.coordinators.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L95)</sup>

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

***

### .spec.coordinators.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L89)</sup>

SchedulerName define scheduler name used for group

***

### .spec.coordinators.securityContext.addCapabilities

Type: `[]core.Capability` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L41)</sup>

AddCapabilities add new capabilities to containers

***

### .spec.coordinators.securityContext.allowPrivilegeEscalation

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L44)</sup>

AllowPrivilegeEscalation Controls whether a process can gain more privileges than its parent process.

***

### .spec.coordinators.securityContext.dropAllCapabilities

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L38)</sup>

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

***

### .spec.coordinators.securityContext.fsGroup

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L61)</sup>

FSGroup is a special supplemental group that applies to all containers in a pod.

***

### .spec.coordinators.securityContext.privileged

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L47)</sup>

Privileged If true, runs container in privileged mode. Processes in privileged containers are
essentially equivalent to root on the host.

***

### .spec.coordinators.securityContext.readOnlyRootFilesystem

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)</sup>

ReadOnlyRootFilesystem if true, mounts the container's root filesystem as read-only.

***

### .spec.coordinators.securityContext.runAsGroup

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L55)</sup>

RunAsGroup is the GID to run the entrypoint of the container process.

***

### .spec.coordinators.securityContext.runAsNonRoot

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L51)</sup>

RunAsNonRoot if true, indicates that the container must run as a non-root user.

***

### .spec.coordinators.securityContext.runAsUser

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L53)</sup>

RunAsUser is the UID to run the entrypoint of the container process.

***

### .spec.coordinators.securityContext.seccompProfile

Type: `core.SeccompProfile` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L77)</sup>

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

***

### .spec.coordinators.securityContext.seLinuxOptions

Type: `core.SELinuxOptions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L82)</sup>

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

***

### .spec.coordinators.securityContext.supplementalGroups

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L59)</sup>

SupplementalGroups is a list of groups applied to the first process run in each container, in addition to the container's primary GID,
the fsGroup (if specified), and group memberships defined in the container image for the uid of the container process.

***

### .spec.coordinators.securityContext.sysctls

Type: `map[string]intstr.IntOrString` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L72)</sup>

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

***

### .spec.coordinators.serviceAccountName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L146)</sup>

ServiceAccountName setting specifies the `serviceAccountName` for the `Pods` created
for each server of this group. If empty, it defaults to using the
`default` service account.
Using an alternative `ServiceAccount` is typically used to separate access rights.
The ArangoDB deployments need some very minimal access rights. With the
deployment of the operator, we grant the rights to 'get' all 'pod' resources.
If you are using a different service account, please grant these rights
to that service account.

***

### .spec.coordinators.shutdownDelay

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L211)</sup>

ShutdownDelay define how long operator should delay finalizer removal after shutdown

***

### .spec.coordinators.shutdownMethod

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L209)</sup>

ShutdownMethod describe procedure of member shutdown taken by Operator

***

### .spec.coordinators.sidecarCoreNames

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L187)</sup>

SidecarCoreNames is a list of sidecar containers which must run in the pod.
Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

***

### .spec.coordinators.sidecars

Type: `[]core.Container` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L191)</sup>

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

***

### .spec.coordinators.storageClassName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L91)</sup>

StorageClassName specifies the classname for storage of the servers.

***

### .spec.coordinators.terminationGracePeriodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L224)</sup>

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

***

### .spec.coordinators.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L122)</sup>

Tolerations specifies the tolerations added to Pods in this group.
By default, suitable tolerations are set for the following keys with the `NoExecute` effect:
- `node.kubernetes.io/not-ready`
- `node.kubernetes.io/unreachable`
- `node.alpha.kubernetes.io/unreachable` (will be removed in future version)
For more information on tolerations, consult the https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

***

### .spec.coordinators.volumeAllowShrink

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L172)</sup>

Deprecated: VolumeAllowShrink allows shrink the volume

***

### .spec.coordinators.volumeClaimTemplate

Type: `core.PersistentVolumeClaim` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L165)</sup>

VolumeClaimTemplate specifies a volumeClaimTemplate used by operator to create to volume claims for pods of this group.
This setting is not available for group `coordinators`, `syncmasters` & `syncworkers`.
The default value describes a volume with `8Gi` storage, `ReadWriteOnce` access mode and volume mode set to `PersistentVolumeFilesystem`.
If this field is not set and `spec.<group>.resources.requests.storage` is set, then a default volume claim
with size as specified by `spec.<group>.resources.requests.storage` will be created. In that case `storage`
and `iops` is not forwarded to the pods resource requirements.

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

***

### .spec.coordinators.volumeMounts

Type: `[]ServerGroupSpecVolumeMount` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L201)</sup>

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation of ServerGroupSpecVolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

***

### .spec.coordinators.volumes\[int\].configMap

Type: `core.ConfigMapVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L138)</sup>

ConfigMap which should be mounted into pod

Links:
* [Documentation of core.ConfigMapVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#configmapvolumesource-v1-core)

***

### .spec.coordinators.volumes\[int\].emptyDir

Type: `core.EmptyDirVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L143)</sup>

EmptyDir

Links:
* [Documentation of core.EmptyDirVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#emptydirvolumesource-v1-core)

***

### .spec.coordinators.volumes\[int\].hostPath

Type: `core.HostPathVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L148)</sup>

HostPath

Links:
* [Documentation of core.HostPathVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#hostpathvolumesource-v1-core)

***

### .spec.coordinators.volumes\[int\].name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L128)</sup>

Name of volume

***

### .spec.coordinators.volumes\[int\].persistentVolumeClaim

Type: `core.PersistentVolumeClaimVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L153)</sup>

PersistentVolumeClaim

Links:
* [Documentation of core.PersistentVolumeClaimVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaimvolumesource-v1-core)

***

### .spec.coordinators.volumes\[int\].secret

Type: `core.SecretVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L133)</sup>

Secret which should be mounted into pod

Links:
* [Documentation of core.SecretVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#secretvolumesource-v1-core)

***

### .spec.database.maintenance

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/database_spec.go#L25)</sup>

Maintenance manage maintenance mode on Cluster side. Requires maintenance feature to be enabled

***

### .spec.dbservers.affinity

Type: `core.PodAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L180)</sup>

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.PodAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

***

### .spec.dbservers.allowMemberRecreation

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L222)</sup>

AllowMemberRecreation allows to recreate member.
This setting changes the member recreation logic based on group:
- For Sync Masters, Sync Workers, Coordinator and DB-Servers it determines if a member can be recreated in case of failure (default `true`)
- For Agents and Single this value is hardcoded to `false` and the value provided in spec is ignored.

***

### .spec.dbservers.annotations

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L125)</sup>

Annotations specified the annotations added to Pods in this group.
Annotations are merged with `spec.annotations`.

***

### .spec.dbservers.annotationsIgnoreList

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L127)</sup>

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

***

### .spec.dbservers.annotationsMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L129)</sup>

AnnotationsMode Define annotations mode which should be use while overriding annotations

***

### .spec.dbservers.antiAffinity

Type: `core.PodAntiAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L176)</sup>

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

***

### .spec.dbservers.args

Type: `[]string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L85)</sup>

Args setting specifies additional command-line arguments passed to all servers of this group.

Default Value: `[]`

***

### .spec.dbservers.count

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L77)</sup>

Count setting specifies the number of servers to start for the given group.
For the Agent group, this value must be a positive, odd number.
The default value is `3` for all groups except `single` (there the default is `1`
for `spec.mode: Single` and `2` for `spec.mode: ActiveFailover`).
For the `syncworkers` group, it is highly recommended to use the same number
as for the `dbservers` group.

***

### .spec.dbservers.entrypoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L87)</sup>

Entrypoint overrides container executable

***

### .spec.dbservers.envs\[int\].name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_env_var.go#L26)</sup>

***

### .spec.dbservers.envs\[int\].value

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_env_var.go#L27)</sup>

***

### .spec.dbservers.ephemeralVolumes.apps.size

Type: `resource.Quantity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)</sup>

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

***

### .spec.dbservers.ephemeralVolumes.temp.size

Type: `resource.Quantity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)</sup>

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

***

### .spec.dbservers.exporterPort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L235)</sup>

ExporterPort define Port used by exporter

***

### .spec.dbservers.extendedRotationCheck

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L205)</sup>

ExtendedRotationCheck extend checks for rotation

***

### .spec.dbservers.externalPortEnabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L217)</sup>

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

***

### .spec.dbservers.indexMethod

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L228)</sup>

IndexMethod define group Indexing method

Possible Values: 
* random (default) - Pick random ID for member. Enforced on the Community Operator.
* ordered - Use sequential number as Member ID, starting from 0. Enterprise Operator required.

***

### .spec.dbservers.initContainers.containers

Type: `[]core.Container` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_init_containers.go#L91)</sup>

Containers contains list of containers

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

***

### .spec.dbservers.initContainers.mode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_init_containers.go#L94)</sup>

Mode keep container replace mode

***

### .spec.dbservers.internalPort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L213)</sup>

InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

***

### .spec.dbservers.internalPortProtocol

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L215)</sup>

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

***

### .spec.dbservers.labels

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L131)</sup>

Labels specified the labels added to Pods in this group.

***

### .spec.dbservers.labelsIgnoreList

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L133)</sup>

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

***

### .spec.dbservers.labelsMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L135)</sup>

LabelsMode Define labels mode which should be use while overriding labels

***

### .spec.dbservers.maxCount

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L81)</sup>

MaxCount specifies a maximum for the count of servers. If set, a specification is invalid if `count > maxCount`.

***

### .spec.dbservers.memoryReservation

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L107)</sup>

MemoryReservation determines the system reservation of memory while calculating `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` value.
If this field is set, `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` is reduced by a specified value in percent.
Accepted Range <0, 50>. If the value is outside the accepted range, it is adjusted to the closest value.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: `0`

***

### .spec.dbservers.minCount

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L79)</sup>

MinCount specifies a minimum for the count of servers. If set, a specification is invalid if `count < minCount`.

***

### .spec.dbservers.nodeAffinity

Type: `core.NodeAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L184)</sup>

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

***

### .spec.dbservers.nodeSelector

Type: `map[string]string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L150)</sup>

NodeSelector setting specifies a set of labels to be used as `nodeSelector` for Pods of this node.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/)

***

### .spec.dbservers.numactl.args

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_numactl_spec.go#L38)</sup>

Args define list of the numactl process

Default Value: `[]`

***

### .spec.dbservers.numactl.enabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_numactl_spec.go#L30)</sup>

Enabled define if numactl should be enabled

Default Value: `false`

***

### .spec.dbservers.numactl.path

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_numactl_spec.go#L34)</sup>

Path define numactl path within the container

Default Value: `/usr/bin/numactl`

***

### .spec.dbservers.overrideDetectedNumberOfCores

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L113)</sup>

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` Container Environment Variable

OverrideDetectedNumberOfCores determines if number of cores should be overridden based on values in resources.
If is set to true and Container CPU Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` to the value from the Container CPU Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: `true`

***

### .spec.dbservers.overrideDetectedTotalMemory

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L101)</sup>

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` Container Environment Variable

OverrideDetectedTotalMemory determines if memory should be overridden based on values in resources.
If is set to true and Container Memory Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` to the value from the Container Memory Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: `true`

***

### .spec.dbservers.podModes.network

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L31)</sup>

***

### .spec.dbservers.podModes.pid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L32)</sup>

***

### .spec.dbservers.port

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L233)</sup>

Port define Port used by member

***

### .spec.dbservers.priorityClassName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L156)</sup>

PriorityClassName specifies a priority class name
Will be forwarded to the pod spec.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/)

***

### .spec.dbservers.probes.livenessProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L245)</sup>

LivenessProbeDisabled if set to true, the operator does not generate a liveness probe for new pods belonging to this group

Default Value: `false`

***

### .spec.dbservers.probes.livenessProbeSpec.failureThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L295)</sup>

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: `3`

***

### .spec.dbservers.probes.livenessProbeSpec.initialDelaySeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L278)</sup>

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: `2`

***

### .spec.dbservers.probes.livenessProbeSpec.periodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L282)</sup>

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: `10`

***

### .spec.dbservers.probes.livenessProbeSpec.successThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L290)</sup>

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: `1`

***

### .spec.dbservers.probes.livenessProbeSpec.timeoutSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L286)</sup>

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: `2`

***

### .spec.dbservers.probes.ReadinessProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L252)</sup>

OldReadinessProbeDisabled if true readinessProbes are disabled
Deprecated: This field is deprecated, kept only for backward compatibility.

***

### .spec.dbservers.probes.readinessProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L254)</sup>

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

***

### .spec.dbservers.probes.readinessProbeSpec.failureThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L295)</sup>

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: `3`

***

### .spec.dbservers.probes.readinessProbeSpec.initialDelaySeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L278)</sup>

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: `2`

***

### .spec.dbservers.probes.readinessProbeSpec.periodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L282)</sup>

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: `10`

***

### .spec.dbservers.probes.readinessProbeSpec.successThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L290)</sup>

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: `1`

***

### .spec.dbservers.probes.readinessProbeSpec.timeoutSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L286)</sup>

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: `2`

***

### .spec.dbservers.probes.startupProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L259)</sup>

StartupProbeDisabled if true startupProbes are disabled

***

### .spec.dbservers.probes.startupProbeSpec.failureThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L295)</sup>

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: `3`

***

### .spec.dbservers.probes.startupProbeSpec.initialDelaySeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L278)</sup>

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: `2`

***

### .spec.dbservers.probes.startupProbeSpec.periodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L282)</sup>

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: `10`

***

### .spec.dbservers.probes.startupProbeSpec.successThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L290)</sup>

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: `1`

***

### .spec.dbservers.probes.startupProbeSpec.timeoutSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L286)</sup>

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: `2`

***

### .spec.dbservers.pvcResizeMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L170)</sup>

VolumeResizeMode specified resize mode for PVCs and PVs

Possible Values: 
* runtime (default) - PVC will be resized in Pod runtime (EKS, GKE)
* rotate - Pod will be shutdown and PVC will be resized (AKS)

***

### .spec.dbservers.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L95)</sup>

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

***

### .spec.dbservers.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L89)</sup>

SchedulerName define scheduler name used for group

***

### .spec.dbservers.securityContext.addCapabilities

Type: `[]core.Capability` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L41)</sup>

AddCapabilities add new capabilities to containers

***

### .spec.dbservers.securityContext.allowPrivilegeEscalation

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L44)</sup>

AllowPrivilegeEscalation Controls whether a process can gain more privileges than its parent process.

***

### .spec.dbservers.securityContext.dropAllCapabilities

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L38)</sup>

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

***

### .spec.dbservers.securityContext.fsGroup

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L61)</sup>

FSGroup is a special supplemental group that applies to all containers in a pod.

***

### .spec.dbservers.securityContext.privileged

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L47)</sup>

Privileged If true, runs container in privileged mode. Processes in privileged containers are
essentially equivalent to root on the host.

***

### .spec.dbservers.securityContext.readOnlyRootFilesystem

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)</sup>

ReadOnlyRootFilesystem if true, mounts the container's root filesystem as read-only.

***

### .spec.dbservers.securityContext.runAsGroup

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L55)</sup>

RunAsGroup is the GID to run the entrypoint of the container process.

***

### .spec.dbservers.securityContext.runAsNonRoot

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L51)</sup>

RunAsNonRoot if true, indicates that the container must run as a non-root user.

***

### .spec.dbservers.securityContext.runAsUser

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L53)</sup>

RunAsUser is the UID to run the entrypoint of the container process.

***

### .spec.dbservers.securityContext.seccompProfile

Type: `core.SeccompProfile` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L77)</sup>

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

***

### .spec.dbservers.securityContext.seLinuxOptions

Type: `core.SELinuxOptions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L82)</sup>

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

***

### .spec.dbservers.securityContext.supplementalGroups

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L59)</sup>

SupplementalGroups is a list of groups applied to the first process run in each container, in addition to the container's primary GID,
the fsGroup (if specified), and group memberships defined in the container image for the uid of the container process.

***

### .spec.dbservers.securityContext.sysctls

Type: `map[string]intstr.IntOrString` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L72)</sup>

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

***

### .spec.dbservers.serviceAccountName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L146)</sup>

ServiceAccountName setting specifies the `serviceAccountName` for the `Pods` created
for each server of this group. If empty, it defaults to using the
`default` service account.
Using an alternative `ServiceAccount` is typically used to separate access rights.
The ArangoDB deployments need some very minimal access rights. With the
deployment of the operator, we grant the rights to 'get' all 'pod' resources.
If you are using a different service account, please grant these rights
to that service account.

***

### .spec.dbservers.shutdownDelay

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L211)</sup>

ShutdownDelay define how long operator should delay finalizer removal after shutdown

***

### .spec.dbservers.shutdownMethod

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L209)</sup>

ShutdownMethod describe procedure of member shutdown taken by Operator

***

### .spec.dbservers.sidecarCoreNames

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L187)</sup>

SidecarCoreNames is a list of sidecar containers which must run in the pod.
Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

***

### .spec.dbservers.sidecars

Type: `[]core.Container` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L191)</sup>

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

***

### .spec.dbservers.storageClassName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L91)</sup>

StorageClassName specifies the classname for storage of the servers.

***

### .spec.dbservers.terminationGracePeriodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L224)</sup>

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

***

### .spec.dbservers.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L122)</sup>

Tolerations specifies the tolerations added to Pods in this group.
By default, suitable tolerations are set for the following keys with the `NoExecute` effect:
- `node.kubernetes.io/not-ready`
- `node.kubernetes.io/unreachable`
- `node.alpha.kubernetes.io/unreachable` (will be removed in future version)
For more information on tolerations, consult the https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

***

### .spec.dbservers.volumeAllowShrink

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L172)</sup>

Deprecated: VolumeAllowShrink allows shrink the volume

***

### .spec.dbservers.volumeClaimTemplate

Type: `core.PersistentVolumeClaim` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L165)</sup>

VolumeClaimTemplate specifies a volumeClaimTemplate used by operator to create to volume claims for pods of this group.
This setting is not available for group `coordinators`, `syncmasters` & `syncworkers`.
The default value describes a volume with `8Gi` storage, `ReadWriteOnce` access mode and volume mode set to `PersistentVolumeFilesystem`.
If this field is not set and `spec.<group>.resources.requests.storage` is set, then a default volume claim
with size as specified by `spec.<group>.resources.requests.storage` will be created. In that case `storage`
and `iops` is not forwarded to the pods resource requirements.

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

***

### .spec.dbservers.volumeMounts

Type: `[]ServerGroupSpecVolumeMount` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L201)</sup>

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation of ServerGroupSpecVolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

***

### .spec.dbservers.volumes\[int\].configMap

Type: `core.ConfigMapVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L138)</sup>

ConfigMap which should be mounted into pod

Links:
* [Documentation of core.ConfigMapVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#configmapvolumesource-v1-core)

***

### .spec.dbservers.volumes\[int\].emptyDir

Type: `core.EmptyDirVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L143)</sup>

EmptyDir

Links:
* [Documentation of core.EmptyDirVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#emptydirvolumesource-v1-core)

***

### .spec.dbservers.volumes\[int\].hostPath

Type: `core.HostPathVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L148)</sup>

HostPath

Links:
* [Documentation of core.HostPathVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#hostpathvolumesource-v1-core)

***

### .spec.dbservers.volumes\[int\].name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L128)</sup>

Name of volume

***

### .spec.dbservers.volumes\[int\].persistentVolumeClaim

Type: `core.PersistentVolumeClaimVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L153)</sup>

PersistentVolumeClaim

Links:
* [Documentation of core.PersistentVolumeClaimVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaimvolumesource-v1-core)

***

### .spec.dbservers.volumes\[int\].secret

Type: `core.SecretVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L133)</sup>

Secret which should be mounted into pod

Links:
* [Documentation of core.SecretVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#secretvolumesource-v1-core)

***

### .spec.disableIPv6

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L109)</sup>

DisableIPv6 setting prevents the use of IPv6 addresses by ArangoDB servers.
This setting cannot be changed after the deployment has been created.

Default Value: `false`

***

### .spec.downtimeAllowed

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L104)</sup>

DowntimeAllowed setting is used to allow automatic reconciliation actions that yield some downtime of the ArangoDB deployment.
When this setting is set to false, no automatic action that may result in downtime is allowed.
If the need for such an action is detected, an event is added to the ArangoDeployment.
Once this setting is set to true, the automatic action is executed.
Operations that may result in downtime are:
- Rotating TLS CA certificate
Note: It is still possible that there is some downtime when the Kubernetes cluster is down, or in a bad state, irrespective of the value of this setting.

Default Value: `false`

***

### .spec.environment

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L65)</sup>

Environment setting specifies the type of environment in which the deployment is created.

Possible Values: 
* Development (default) - This value optimizes the deployment for development use. It is possible to run a deployment on a small number of nodes (e.g. minikube).
* Production - This value optimizes the deployment for production use. It puts required affinity constraints on all pods to avoid Agents & DB-Servers from running on the same machine.

***

### .spec.externalAccess.advertisedEndpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/external_access_spec.go#L58)</sup>

AdvertisedEndpoint is passed to the coordinators/single servers for advertising a specific endpoint

***

### .spec.externalAccess.loadBalancerIP

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/external_access_spec.go#L48)</sup>

LoadBalancerIP define optional IP used to configure a load-balancer on, in case of Auto or LoadBalancer type.
If you do not specify this setting, an IP will be chosen automatically by the load-balancer provisioner.

***

### .spec.externalAccess.loadBalancerSourceRanges

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/external_access_spec.go#L55)</sup>

LoadBalancerSourceRanges define LoadBalancerSourceRanges used for LoadBalancer Service type
If specified and supported by the platform, this will restrict traffic through the cloud-provider
load-balancer will be restricted to the specified client IPs. This field will be ignored if the
cloud-provider does not support the feature.

Links:
* [Cloud Provider Firewall](https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/)

***

### .spec.externalAccess.managedServiceNames

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/external_access_spec.go#L62)</sup>

ManagedServiceNames keeps names of services which are not managed by KubeArangoDB.
It is only relevant when type of service is `managed`.

***

### .spec.externalAccess.nodePort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/external_access_spec.go#L44)</sup>

NodePort define optional port used in case of Auto or NodePort type.
This setting is used when `spec.externalAccess.type` is set to `NodePort` or `Auto`.
If you do not specify this setting, a random port will be chosen automatically.

***

### .spec.externalAccess.type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/external_access_spec.go#L39)</sup>

Type specifies the type of Service that will be created to provide access to the ArangoDB deployment from outside the Kubernetes cluster.

Possible Values: 
* Auto (default) - Create a Service of type LoadBalancer and fallback to a Service or type NodePort when the LoadBalancer is not assigned an IP address.
* None - limit access to application running inside the Kubernetes cluster.
* LoadBalancer - Create a Service of type LoadBalancer for the ArangoDB deployment.
* NodePort - Create a Service of type NodePort for the ArangoDB deployment.

***

### .spec.features.foxx.queues

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_features.go#L24)</sup>

***

### .spec.id.affinity

Type: `core.PodAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_id_group_spec.go#L44)</sup>

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.PodAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

***

### .spec.id.antiAffinity

Type: `core.PodAntiAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_id_group_spec.go#L40)</sup>

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

***

### .spec.id.entrypoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_id_group_spec.go#L28)</sup>

Entrypoint overrides container executable

***

### .spec.id.nodeAffinity

Type: `core.NodeAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_id_group_spec.go#L48)</sup>

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

***

### .spec.id.nodeSelector

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_id_group_spec.go#L34)</sup>

NodeSelector specifies a set of selectors for nodes

***

### .spec.id.priorityClassName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_id_group_spec.go#L36)</sup>

PriorityClassName specifies a priority class name

***

### .spec.id.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_id_group_spec.go#L56)</sup>

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

***

### .spec.id.securityContext.addCapabilities

Type: `[]core.Capability` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L41)</sup>

AddCapabilities add new capabilities to containers

***

### .spec.id.securityContext.allowPrivilegeEscalation

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L44)</sup>

AllowPrivilegeEscalation Controls whether a process can gain more privileges than its parent process.

***

### .spec.id.securityContext.dropAllCapabilities

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L38)</sup>

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

***

### .spec.id.securityContext.fsGroup

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L61)</sup>

FSGroup is a special supplemental group that applies to all containers in a pod.

***

### .spec.id.securityContext.privileged

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L47)</sup>

Privileged If true, runs container in privileged mode. Processes in privileged containers are
essentially equivalent to root on the host.

***

### .spec.id.securityContext.readOnlyRootFilesystem

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)</sup>

ReadOnlyRootFilesystem if true, mounts the container's root filesystem as read-only.

***

### .spec.id.securityContext.runAsGroup

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L55)</sup>

RunAsGroup is the GID to run the entrypoint of the container process.

***

### .spec.id.securityContext.runAsNonRoot

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L51)</sup>

RunAsNonRoot if true, indicates that the container must run as a non-root user.

***

### .spec.id.securityContext.runAsUser

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L53)</sup>

RunAsUser is the UID to run the entrypoint of the container process.

***

### .spec.id.securityContext.seccompProfile

Type: `core.SeccompProfile` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L77)</sup>

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

***

### .spec.id.securityContext.seLinuxOptions

Type: `core.SELinuxOptions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L82)</sup>

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

***

### .spec.id.securityContext.supplementalGroups

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L59)</sup>

SupplementalGroups is a list of groups applied to the first process run in each container, in addition to the container's primary GID,
the fsGroup (if specified), and group memberships defined in the container image for the uid of the container process.

***

### .spec.id.securityContext.sysctls

Type: `map[string]intstr.IntOrString` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L72)</sup>

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

***

### .spec.id.serviceAccountName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_id_group_spec.go#L50)</sup>

ServiceAccountName specifies the name of the service account used for Pods in this group.

***

### .spec.id.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_id_group_spec.go#L32)</sup>

Tolerations specifies the tolerations added to Pods in this group.

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

***

### .spec.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L78)</sup>

Image specifies the docker image to use for all ArangoDB servers.
In a development environment this setting defaults to arangodb/arangodb:latest.
For production environments this is a required setting without a default value.
It is highly recommend to use explicit version (not latest) for production environments.

***

### .spec.imageDiscoveryMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L94)</sup>

ImageDiscoveryMode specifies the image discovery mode.

Possible Values: 
* kubelet (default) - Use sha256 of the discovered image in the pods
* direct - Use image provided in the spec.image directly in the pods

***

### .spec.imagePullPolicy

Type: `core.PullPolicy` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L86)</sup>

ImagePullPolicy specifies the pull policy for the docker image to use for all ArangoDB servers.

Links:
* [Documentation of core.PullPolicy](https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy)

Possible Values: 
* Always (default) - Means that kubelet always attempts to pull the latest image. Container will fail If the pull fails.
* Never - Means that kubelet never pulls an image, but only uses a local image. Container will fail if the image isn't present
* IfNotPresent - Means that kubelet pulls if the image isn't present on disk. Container will fail if the image isn't present and the pull fails.

***

### .spec.imagePullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L89)</sup>

ImagePullSecrets specifies the list of image pull secrets for the docker image to use for all ArangoDB servers.

***

### .spec.labels

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L138)</sup>

Labels specifies the labels added to Pods in this group.

***

### .spec.labelsIgnoreList

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L141)</sup>

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

***

### .spec.labelsMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L147)</sup>

LabelsMode Define labels mode which should be use while overriding labels

Possible Values: 
* disabled (default) - Disable annotations/labels override. Default if there is no annotations/labels set in ArangoDeployment
* append - Add new annotations/labels without affecting old ones
* replace - Replace existing annotations/labels

***

### .spec.license.secretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/license_spec.go#L33)</sup>

SecretName setting specifies the name of a kubernetes `Secret` that contains
the license key token used for enterprise images. This value is not used for
the Community Edition.

***

### .spec.lifecycle.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/lifecycle_spec.go#L31)</sup>

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

***

### .spec.memberPropagationMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L220)</sup>

MemberPropagationMode defines how changes to pod spec should be propogated.
Changes to a pod’s configuration require a restart of that pod in almost all cases.
Pods are restarted eagerly by default, which can cause more restarts than desired, especially when updating arangod as well as the operator.
The propagation of the configuration changes can be deferred to the next restart, either triggered manually by the user or by another operation like an upgrade.
This reduces the number of restarts for upgrading both the server and the operator from two to one.

Possible Values: 
* always (default) - Restart the member as soon as a configuration change is discovered
* on-restart - Wait until the next restart to change the member configuration

***

### .spec.metrics.authentication.jwtTokenSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_metrics_spec.go#L34)</sup>

JWTTokenSecretName contains the name of the JWT kubernetes secret used for authentication

***

### .spec.metrics.enabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_metrics_spec.go#L81)</sup>

Enabled if this is set to `true`, the operator runs a sidecar container for
every Agent, DB-Server, Coordinator and Single server.

Links:
* [Metrics collection](/docs/metrics.md)

Default Value: `false`

***

### .spec.metrics.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_metrics_spec.go#L83)</sup>

deprecated

***

### .spec.metrics.mode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_metrics_spec.go#L90)</sup>

deprecated

***

### .spec.metrics.port

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_metrics_spec.go#L99)</sup>

***

### .spec.metrics.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_metrics_spec.go#L88)</sup>

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

***

### .spec.metrics.serviceMonitor.enabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_metrics_service_monitor_spec.go#L24)</sup>

***

### .spec.metrics.serviceMonitor.labels

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_metrics_service_monitor_spec.go#L25)</sup>

***

### .spec.metrics.tls

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_metrics_spec.go#L95)</sup>

TLS defines if TLS should be enabled on Metrics exporter endpoint.
This option will enable TLS only if TLS is enabled on ArangoDeployment,
otherwise `true` value will not take any effect.

Default Value: `true`

***

### .spec.mode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L60)</sup>

Mode specifies the type of ArangoDB deployment to create.

Possible Values: 
* Cluster (default) - Full cluster. Defaults to 3 Agents, 3 DB-Servers & 3 Coordinators.
* ActiveFailover - Active-failover single pair. Defaults to 3 Agents and 2 single servers.
* Single - Single server only (note this does not provide high availability or reliability).

This field is **immutable**: Change of the ArangoDeployment Mode is not possible after creation.

***

### .spec.networkAttachedVolumes

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L123)</sup>

NetworkAttachedVolumes
If set to `true`, a ResignLeadership operation will be triggered when a DB-Server pod is evicted (rather than a CleanOutServer operation).
Furthermore, the pod will simply be redeployed on a different node, rather than cleaned and retired and replaced by a new member.
You must only set this option to true if your persistent volumes are “movable” in the sense that they can be mounted from a different k8s node, like in the case of network attached volumes.
If your persistent volumes are tied to a specific pod, you must leave this option on false.

Default Value: `true`

***

### .spec.rebalancer.enabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/rebalancer_spec.go#L26)</sup>

***

### .spec.rebalancer.optimizers.leader

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/rebalancer_spec.go#L74)</sup>

***

### .spec.rebalancer.parallelMoves

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/rebalancer_spec.go#L28)</sup>

***

### .spec.rebalancer.readers.count

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/rebalancer_spec.go#L62)</sup>

deprecated does not work in Rebalancer V2
Count Enable Shard Count machanism

***

### .spec.recovery.autoRecover

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/recovery_spec.go#L26)</sup>

***

### .spec.restoreEncryptionSecret

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L160)</sup>

RestoreEncryptionSecret specifies optional name of secret which contains encryption key used for restore

***

### .spec.restoreFrom

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L157)</sup>

RestoreFrom setting specifies a `ArangoBackup` resource name the cluster should be restored from.
After a restore or failure to do so, the status of the deployment contains information about the restore operation in the restore key.
It will contain some of the following fields:
- `requestedFrom`: name of the ArangoBackup used to restore from.
- `message`: optional message explaining why the restore failed.
- `state`: state indicating if the restore was successful or not. Possible values: Restoring, Restored, RestoreFailed
If the restoreFrom key is removed from the spec, the restore key is deleted as well.
A new restore attempt is made if and only if either in the status restore is not set or if spec.restoreFrom and status.requestedFrom are different.

***

### .spec.rocksdb.encryption.keySecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/rocksdb_spec.go#L37)</sup>

KeySecretName setting specifies the name of a Kubernetes `Secret` that contains an encryption key used for encrypting all data stored by ArangoDB servers.
When an encryption key is used, encryption of the data in the cluster is enabled, without it encryption is disabled.
The default value is empty.
This requires the Enterprise Edition.
The encryption key cannot be changed after the cluster has been created.
The secret specified by this setting, must have a data field named 'key' containing an encryption key that is exactly 32 bytes long.

***

### .spec.single.affinity

Type: `core.PodAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L180)</sup>

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.PodAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

***

### .spec.single.allowMemberRecreation

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L222)</sup>

AllowMemberRecreation allows to recreate member.
This setting changes the member recreation logic based on group:
- For Sync Masters, Sync Workers, Coordinator and DB-Servers it determines if a member can be recreated in case of failure (default `true`)
- For Agents and Single this value is hardcoded to `false` and the value provided in spec is ignored.

***

### .spec.single.annotations

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L125)</sup>

Annotations specified the annotations added to Pods in this group.
Annotations are merged with `spec.annotations`.

***

### .spec.single.annotationsIgnoreList

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L127)</sup>

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

***

### .spec.single.annotationsMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L129)</sup>

AnnotationsMode Define annotations mode which should be use while overriding annotations

***

### .spec.single.antiAffinity

Type: `core.PodAntiAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L176)</sup>

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

***

### .spec.single.args

Type: `[]string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L85)</sup>

Args setting specifies additional command-line arguments passed to all servers of this group.

Default Value: `[]`

***

### .spec.single.count

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L77)</sup>

Count setting specifies the number of servers to start for the given group.
For the Agent group, this value must be a positive, odd number.
The default value is `3` for all groups except `single` (there the default is `1`
for `spec.mode: Single` and `2` for `spec.mode: ActiveFailover`).
For the `syncworkers` group, it is highly recommended to use the same number
as for the `dbservers` group.

***

### .spec.single.entrypoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L87)</sup>

Entrypoint overrides container executable

***

### .spec.single.envs\[int\].name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_env_var.go#L26)</sup>

***

### .spec.single.envs\[int\].value

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_env_var.go#L27)</sup>

***

### .spec.single.ephemeralVolumes.apps.size

Type: `resource.Quantity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)</sup>

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

***

### .spec.single.ephemeralVolumes.temp.size

Type: `resource.Quantity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)</sup>

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

***

### .spec.single.exporterPort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L235)</sup>

ExporterPort define Port used by exporter

***

### .spec.single.extendedRotationCheck

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L205)</sup>

ExtendedRotationCheck extend checks for rotation

***

### .spec.single.externalPortEnabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L217)</sup>

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

***

### .spec.single.indexMethod

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L228)</sup>

IndexMethod define group Indexing method

Possible Values: 
* random (default) - Pick random ID for member. Enforced on the Community Operator.
* ordered - Use sequential number as Member ID, starting from 0. Enterprise Operator required.

***

### .spec.single.initContainers.containers

Type: `[]core.Container` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_init_containers.go#L91)</sup>

Containers contains list of containers

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

***

### .spec.single.initContainers.mode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_init_containers.go#L94)</sup>

Mode keep container replace mode

***

### .spec.single.internalPort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L213)</sup>

InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

***

### .spec.single.internalPortProtocol

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L215)</sup>

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

***

### .spec.single.labels

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L131)</sup>

Labels specified the labels added to Pods in this group.

***

### .spec.single.labelsIgnoreList

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L133)</sup>

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

***

### .spec.single.labelsMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L135)</sup>

LabelsMode Define labels mode which should be use while overriding labels

***

### .spec.single.maxCount

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L81)</sup>

MaxCount specifies a maximum for the count of servers. If set, a specification is invalid if `count > maxCount`.

***

### .spec.single.memoryReservation

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L107)</sup>

MemoryReservation determines the system reservation of memory while calculating `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` value.
If this field is set, `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` is reduced by a specified value in percent.
Accepted Range <0, 50>. If the value is outside the accepted range, it is adjusted to the closest value.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: `0`

***

### .spec.single.minCount

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L79)</sup>

MinCount specifies a minimum for the count of servers. If set, a specification is invalid if `count < minCount`.

***

### .spec.single.nodeAffinity

Type: `core.NodeAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L184)</sup>

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

***

### .spec.single.nodeSelector

Type: `map[string]string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L150)</sup>

NodeSelector setting specifies a set of labels to be used as `nodeSelector` for Pods of this node.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/)

***

### .spec.single.numactl.args

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_numactl_spec.go#L38)</sup>

Args define list of the numactl process

Default Value: `[]`

***

### .spec.single.numactl.enabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_numactl_spec.go#L30)</sup>

Enabled define if numactl should be enabled

Default Value: `false`

***

### .spec.single.numactl.path

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_numactl_spec.go#L34)</sup>

Path define numactl path within the container

Default Value: `/usr/bin/numactl`

***

### .spec.single.overrideDetectedNumberOfCores

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L113)</sup>

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` Container Environment Variable

OverrideDetectedNumberOfCores determines if number of cores should be overridden based on values in resources.
If is set to true and Container CPU Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` to the value from the Container CPU Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: `true`

***

### .spec.single.overrideDetectedTotalMemory

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L101)</sup>

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` Container Environment Variable

OverrideDetectedTotalMemory determines if memory should be overridden based on values in resources.
If is set to true and Container Memory Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` to the value from the Container Memory Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: `true`

***

### .spec.single.podModes.network

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L31)</sup>

***

### .spec.single.podModes.pid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L32)</sup>

***

### .spec.single.port

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L233)</sup>

Port define Port used by member

***

### .spec.single.priorityClassName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L156)</sup>

PriorityClassName specifies a priority class name
Will be forwarded to the pod spec.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/)

***

### .spec.single.probes.livenessProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L245)</sup>

LivenessProbeDisabled if set to true, the operator does not generate a liveness probe for new pods belonging to this group

Default Value: `false`

***

### .spec.single.probes.livenessProbeSpec.failureThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L295)</sup>

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: `3`

***

### .spec.single.probes.livenessProbeSpec.initialDelaySeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L278)</sup>

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: `2`

***

### .spec.single.probes.livenessProbeSpec.periodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L282)</sup>

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: `10`

***

### .spec.single.probes.livenessProbeSpec.successThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L290)</sup>

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: `1`

***

### .spec.single.probes.livenessProbeSpec.timeoutSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L286)</sup>

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: `2`

***

### .spec.single.probes.ReadinessProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L252)</sup>

OldReadinessProbeDisabled if true readinessProbes are disabled
Deprecated: This field is deprecated, kept only for backward compatibility.

***

### .spec.single.probes.readinessProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L254)</sup>

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

***

### .spec.single.probes.readinessProbeSpec.failureThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L295)</sup>

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: `3`

***

### .spec.single.probes.readinessProbeSpec.initialDelaySeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L278)</sup>

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: `2`

***

### .spec.single.probes.readinessProbeSpec.periodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L282)</sup>

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: `10`

***

### .spec.single.probes.readinessProbeSpec.successThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L290)</sup>

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: `1`

***

### .spec.single.probes.readinessProbeSpec.timeoutSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L286)</sup>

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: `2`

***

### .spec.single.probes.startupProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L259)</sup>

StartupProbeDisabled if true startupProbes are disabled

***

### .spec.single.probes.startupProbeSpec.failureThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L295)</sup>

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: `3`

***

### .spec.single.probes.startupProbeSpec.initialDelaySeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L278)</sup>

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: `2`

***

### .spec.single.probes.startupProbeSpec.periodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L282)</sup>

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: `10`

***

### .spec.single.probes.startupProbeSpec.successThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L290)</sup>

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: `1`

***

### .spec.single.probes.startupProbeSpec.timeoutSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L286)</sup>

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: `2`

***

### .spec.single.pvcResizeMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L170)</sup>

VolumeResizeMode specified resize mode for PVCs and PVs

Possible Values: 
* runtime (default) - PVC will be resized in Pod runtime (EKS, GKE)
* rotate - Pod will be shutdown and PVC will be resized (AKS)

***

### .spec.single.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L95)</sup>

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

***

### .spec.single.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L89)</sup>

SchedulerName define scheduler name used for group

***

### .spec.single.securityContext.addCapabilities

Type: `[]core.Capability` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L41)</sup>

AddCapabilities add new capabilities to containers

***

### .spec.single.securityContext.allowPrivilegeEscalation

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L44)</sup>

AllowPrivilegeEscalation Controls whether a process can gain more privileges than its parent process.

***

### .spec.single.securityContext.dropAllCapabilities

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L38)</sup>

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

***

### .spec.single.securityContext.fsGroup

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L61)</sup>

FSGroup is a special supplemental group that applies to all containers in a pod.

***

### .spec.single.securityContext.privileged

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L47)</sup>

Privileged If true, runs container in privileged mode. Processes in privileged containers are
essentially equivalent to root on the host.

***

### .spec.single.securityContext.readOnlyRootFilesystem

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)</sup>

ReadOnlyRootFilesystem if true, mounts the container's root filesystem as read-only.

***

### .spec.single.securityContext.runAsGroup

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L55)</sup>

RunAsGroup is the GID to run the entrypoint of the container process.

***

### .spec.single.securityContext.runAsNonRoot

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L51)</sup>

RunAsNonRoot if true, indicates that the container must run as a non-root user.

***

### .spec.single.securityContext.runAsUser

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L53)</sup>

RunAsUser is the UID to run the entrypoint of the container process.

***

### .spec.single.securityContext.seccompProfile

Type: `core.SeccompProfile` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L77)</sup>

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

***

### .spec.single.securityContext.seLinuxOptions

Type: `core.SELinuxOptions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L82)</sup>

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

***

### .spec.single.securityContext.supplementalGroups

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L59)</sup>

SupplementalGroups is a list of groups applied to the first process run in each container, in addition to the container's primary GID,
the fsGroup (if specified), and group memberships defined in the container image for the uid of the container process.

***

### .spec.single.securityContext.sysctls

Type: `map[string]intstr.IntOrString` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L72)</sup>

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

***

### .spec.single.serviceAccountName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L146)</sup>

ServiceAccountName setting specifies the `serviceAccountName` for the `Pods` created
for each server of this group. If empty, it defaults to using the
`default` service account.
Using an alternative `ServiceAccount` is typically used to separate access rights.
The ArangoDB deployments need some very minimal access rights. With the
deployment of the operator, we grant the rights to 'get' all 'pod' resources.
If you are using a different service account, please grant these rights
to that service account.

***

### .spec.single.shutdownDelay

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L211)</sup>

ShutdownDelay define how long operator should delay finalizer removal after shutdown

***

### .spec.single.shutdownMethod

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L209)</sup>

ShutdownMethod describe procedure of member shutdown taken by Operator

***

### .spec.single.sidecarCoreNames

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L187)</sup>

SidecarCoreNames is a list of sidecar containers which must run in the pod.
Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

***

### .spec.single.sidecars

Type: `[]core.Container` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L191)</sup>

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

***

### .spec.single.storageClassName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L91)</sup>

StorageClassName specifies the classname for storage of the servers.

***

### .spec.single.terminationGracePeriodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L224)</sup>

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

***

### .spec.single.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L122)</sup>

Tolerations specifies the tolerations added to Pods in this group.
By default, suitable tolerations are set for the following keys with the `NoExecute` effect:
- `node.kubernetes.io/not-ready`
- `node.kubernetes.io/unreachable`
- `node.alpha.kubernetes.io/unreachable` (will be removed in future version)
For more information on tolerations, consult the https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

***

### .spec.single.volumeAllowShrink

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L172)</sup>

Deprecated: VolumeAllowShrink allows shrink the volume

***

### .spec.single.volumeClaimTemplate

Type: `core.PersistentVolumeClaim` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L165)</sup>

VolumeClaimTemplate specifies a volumeClaimTemplate used by operator to create to volume claims for pods of this group.
This setting is not available for group `coordinators`, `syncmasters` & `syncworkers`.
The default value describes a volume with `8Gi` storage, `ReadWriteOnce` access mode and volume mode set to `PersistentVolumeFilesystem`.
If this field is not set and `spec.<group>.resources.requests.storage` is set, then a default volume claim
with size as specified by `spec.<group>.resources.requests.storage` will be created. In that case `storage`
and `iops` is not forwarded to the pods resource requirements.

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

***

### .spec.single.volumeMounts

Type: `[]ServerGroupSpecVolumeMount` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L201)</sup>

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation of ServerGroupSpecVolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

***

### .spec.single.volumes\[int\].configMap

Type: `core.ConfigMapVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L138)</sup>

ConfigMap which should be mounted into pod

Links:
* [Documentation of core.ConfigMapVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#configmapvolumesource-v1-core)

***

### .spec.single.volumes\[int\].emptyDir

Type: `core.EmptyDirVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L143)</sup>

EmptyDir

Links:
* [Documentation of core.EmptyDirVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#emptydirvolumesource-v1-core)

***

### .spec.single.volumes\[int\].hostPath

Type: `core.HostPathVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L148)</sup>

HostPath

Links:
* [Documentation of core.HostPathVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#hostpathvolumesource-v1-core)

***

### .spec.single.volumes\[int\].name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L128)</sup>

Name of volume

***

### .spec.single.volumes\[int\].persistentVolumeClaim

Type: `core.PersistentVolumeClaimVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L153)</sup>

PersistentVolumeClaim

Links:
* [Documentation of core.PersistentVolumeClaimVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaimvolumesource-v1-core)

***

### .spec.single.volumes\[int\].secret

Type: `core.SecretVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L133)</sup>

Secret which should be mounted into pod

Links:
* [Documentation of core.SecretVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#secretvolumesource-v1-core)

***

### .spec.storageEngine

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L72)</sup>

StorageEngine specifies the type of storage engine used for all servers in the cluster.

Possible Values: 
* RocksDB (default) - To use the RocksDB storage engine.
* MMFiles - To use the MMFiles storage engine. Deprecated.

This field is **immutable**: This setting cannot be changed after the cluster has been created.

***

### .spec.sync.auth.clientCASecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/sync_authentication_spec.go#L41)</sup>

ClientCASecretName setting specifies the name of a kubernetes `Secret` that contains
a PEM encoded CA certificate used for client certificate verification
in all ArangoSync master servers.
This is a required setting when `spec.sync.enabled` is `true`.

***

### .spec.sync.auth.jwtSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/sync_authentication_spec.go#L36)</sup>

JWTSecretName setting specifies the name of a kubernetes `Secret` that contains
the JWT token used for accessing all ArangoSync master servers.
When not specified, the `spec.auth.jwtSecretName` value is used.
If you specify a name of a `Secret` that does not exist, a random token is created
and stored in a `Secret` with given name.

***

### .spec.sync.enabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/sync_spec.go#L34)</sup>

Enabled setting enables/disables support for data center 2 data center
replication in the cluster. When enabled, the cluster will contain
a number of `syncmaster` & `syncworker` servers.

Default Value: `false`

***

### .spec.sync.externalAccess.accessPackageSecretNames

Type: `[]string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/sync_external_access_spec.go#L49)</sup>

AccessPackageSecretNames setting specifies the names of zero of more `Secrets` that will be created by the deployment
operator containing "access packages". An access package contains those `Secrets` that are needed
to access the SyncMasters of this `ArangoDeployment`.
By removing a name from this setting, the corresponding `Secret` is also deleted.
Note that to remove all access packages, leave an empty array in place (`[]`).
Completely removing the setting results in not modifying the list.

Links:
* [See the ArangoDeploymentReplication specification](deployment-replication-resource-reference.md)

***

### .spec.sync.externalAccess.advertisedEndpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/external_access_spec.go#L58)</sup>

AdvertisedEndpoint is passed to the coordinators/single servers for advertising a specific endpoint

***

### .spec.sync.externalAccess.loadBalancerIP

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/external_access_spec.go#L48)</sup>

LoadBalancerIP define optional IP used to configure a load-balancer on, in case of Auto or LoadBalancer type.
If you do not specify this setting, an IP will be chosen automatically by the load-balancer provisioner.

***

### .spec.sync.externalAccess.loadBalancerSourceRanges

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/external_access_spec.go#L55)</sup>

LoadBalancerSourceRanges define LoadBalancerSourceRanges used for LoadBalancer Service type
If specified and supported by the platform, this will restrict traffic through the cloud-provider
load-balancer will be restricted to the specified client IPs. This field will be ignored if the
cloud-provider does not support the feature.

Links:
* [Cloud Provider Firewall](https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/)

***

### .spec.sync.externalAccess.managedServiceNames

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/external_access_spec.go#L62)</sup>

ManagedServiceNames keeps names of services which are not managed by KubeArangoDB.
It is only relevant when type of service is `managed`.

***

### .spec.sync.externalAccess.masterEndpoint

Type: `[]string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/sync_external_access_spec.go#L40)</sup>

MasterEndpoint setting specifies the master endpoint(s) advertised by the ArangoSync SyncMasters.
If not set, this setting defaults to:
- If `spec.sync.externalAccess.loadBalancerIP` is set, it defaults to `https://<load-balancer-ip>:<8629>`.
- Otherwise it defaults to `https://<sync-service-dns-name>:<8629>`.

***

### .spec.sync.externalAccess.nodePort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/external_access_spec.go#L44)</sup>

NodePort define optional port used in case of Auto or NodePort type.
This setting is used when `spec.externalAccess.type` is set to `NodePort` or `Auto`.
If you do not specify this setting, a random port will be chosen automatically.

***

### .spec.sync.externalAccess.type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/external_access_spec.go#L39)</sup>

Type specifies the type of Service that will be created to provide access to the ArangoDB deployment from outside the Kubernetes cluster.

Possible Values: 
* Auto (default) - Create a Service of type LoadBalancer and fallback to a Service or type NodePort when the LoadBalancer is not assigned an IP address.
* None - limit access to application running inside the Kubernetes cluster.
* LoadBalancer - Create a Service of type LoadBalancer for the ArangoDB deployment.
* NodePort - Create a Service of type NodePort for the ArangoDB deployment.

***

### .spec.sync.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/sync_spec.go#L40)</sup>

***

### .spec.sync.monitoring.tokenSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/sync_monitoring_spec.go#L34)</sup>

TokenSecretName setting specifies the name of a kubernetes `Secret` that contains
the bearer token used for accessing all monitoring endpoints of all arangod/arangosync servers.
When not specified, no monitoring token is used.

***

### .spec.sync.tls.altNames

Type: `[]string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/tls_spec.go#L72)</sup>

AltNames setting specifies a list of alternate names that will be added to all generated
certificates. These names can be DNS names or email addresses.
The default value is empty.

***

### .spec.sync.tls.caSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/tls_spec.go#L67)</sup>

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

***

### .spec.sync.tls.mode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/tls_spec.go#L81)</sup>

***

### .spec.sync.tls.sni.mapping.\<string\>

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/tls_sni_spec.go#L30)</sup>

***

### .spec.sync.tls.ttl

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/tls_spec.go#L79)</sup>

TTL setting specifies the time to live of all generated server certificates.
When the server certificate is about to expire, it will be automatically replaced
by a new one and the affected server will be restarted.
Note: The time to live of the CA certificate (when created automatically)
will be set to 10 years.

Default Value: `"2160h" (about 3 months)`

***

### .spec.syncmasters.affinity

Type: `core.PodAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L180)</sup>

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.PodAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

***

### .spec.syncmasters.allowMemberRecreation

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L222)</sup>

AllowMemberRecreation allows to recreate member.
This setting changes the member recreation logic based on group:
- For Sync Masters, Sync Workers, Coordinator and DB-Servers it determines if a member can be recreated in case of failure (default `true`)
- For Agents and Single this value is hardcoded to `false` and the value provided in spec is ignored.

***

### .spec.syncmasters.annotations

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L125)</sup>

Annotations specified the annotations added to Pods in this group.
Annotations are merged with `spec.annotations`.

***

### .spec.syncmasters.annotationsIgnoreList

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L127)</sup>

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

***

### .spec.syncmasters.annotationsMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L129)</sup>

AnnotationsMode Define annotations mode which should be use while overriding annotations

***

### .spec.syncmasters.antiAffinity

Type: `core.PodAntiAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L176)</sup>

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

***

### .spec.syncmasters.args

Type: `[]string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L85)</sup>

Args setting specifies additional command-line arguments passed to all servers of this group.

Default Value: `[]`

***

### .spec.syncmasters.count

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L77)</sup>

Count setting specifies the number of servers to start for the given group.
For the Agent group, this value must be a positive, odd number.
The default value is `3` for all groups except `single` (there the default is `1`
for `spec.mode: Single` and `2` for `spec.mode: ActiveFailover`).
For the `syncworkers` group, it is highly recommended to use the same number
as for the `dbservers` group.

***

### .spec.syncmasters.entrypoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L87)</sup>

Entrypoint overrides container executable

***

### .spec.syncmasters.envs\[int\].name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_env_var.go#L26)</sup>

***

### .spec.syncmasters.envs\[int\].value

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_env_var.go#L27)</sup>

***

### .spec.syncmasters.ephemeralVolumes.apps.size

Type: `resource.Quantity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)</sup>

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

***

### .spec.syncmasters.ephemeralVolumes.temp.size

Type: `resource.Quantity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)</sup>

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

***

### .spec.syncmasters.exporterPort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L235)</sup>

ExporterPort define Port used by exporter

***

### .spec.syncmasters.extendedRotationCheck

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L205)</sup>

ExtendedRotationCheck extend checks for rotation

***

### .spec.syncmasters.externalPortEnabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L217)</sup>

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

***

### .spec.syncmasters.indexMethod

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L228)</sup>

IndexMethod define group Indexing method

Possible Values: 
* random (default) - Pick random ID for member. Enforced on the Community Operator.
* ordered - Use sequential number as Member ID, starting from 0. Enterprise Operator required.

***

### .spec.syncmasters.initContainers.containers

Type: `[]core.Container` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_init_containers.go#L91)</sup>

Containers contains list of containers

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

***

### .spec.syncmasters.initContainers.mode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_init_containers.go#L94)</sup>

Mode keep container replace mode

***

### .spec.syncmasters.internalPort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L213)</sup>

InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

***

### .spec.syncmasters.internalPortProtocol

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L215)</sup>

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

***

### .spec.syncmasters.labels

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L131)</sup>

Labels specified the labels added to Pods in this group.

***

### .spec.syncmasters.labelsIgnoreList

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L133)</sup>

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

***

### .spec.syncmasters.labelsMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L135)</sup>

LabelsMode Define labels mode which should be use while overriding labels

***

### .spec.syncmasters.maxCount

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L81)</sup>

MaxCount specifies a maximum for the count of servers. If set, a specification is invalid if `count > maxCount`.

***

### .spec.syncmasters.memoryReservation

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L107)</sup>

MemoryReservation determines the system reservation of memory while calculating `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` value.
If this field is set, `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` is reduced by a specified value in percent.
Accepted Range <0, 50>. If the value is outside the accepted range, it is adjusted to the closest value.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: `0`

***

### .spec.syncmasters.minCount

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L79)</sup>

MinCount specifies a minimum for the count of servers. If set, a specification is invalid if `count < minCount`.

***

### .spec.syncmasters.nodeAffinity

Type: `core.NodeAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L184)</sup>

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

***

### .spec.syncmasters.nodeSelector

Type: `map[string]string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L150)</sup>

NodeSelector setting specifies a set of labels to be used as `nodeSelector` for Pods of this node.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/)

***

### .spec.syncmasters.numactl.args

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_numactl_spec.go#L38)</sup>

Args define list of the numactl process

Default Value: `[]`

***

### .spec.syncmasters.numactl.enabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_numactl_spec.go#L30)</sup>

Enabled define if numactl should be enabled

Default Value: `false`

***

### .spec.syncmasters.numactl.path

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_numactl_spec.go#L34)</sup>

Path define numactl path within the container

Default Value: `/usr/bin/numactl`

***

### .spec.syncmasters.overrideDetectedNumberOfCores

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L113)</sup>

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` Container Environment Variable

OverrideDetectedNumberOfCores determines if number of cores should be overridden based on values in resources.
If is set to true and Container CPU Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` to the value from the Container CPU Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: `true`

***

### .spec.syncmasters.overrideDetectedTotalMemory

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L101)</sup>

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` Container Environment Variable

OverrideDetectedTotalMemory determines if memory should be overridden based on values in resources.
If is set to true and Container Memory Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` to the value from the Container Memory Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: `true`

***

### .spec.syncmasters.podModes.network

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L31)</sup>

***

### .spec.syncmasters.podModes.pid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L32)</sup>

***

### .spec.syncmasters.port

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L233)</sup>

Port define Port used by member

***

### .spec.syncmasters.priorityClassName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L156)</sup>

PriorityClassName specifies a priority class name
Will be forwarded to the pod spec.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/)

***

### .spec.syncmasters.probes.livenessProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L245)</sup>

LivenessProbeDisabled if set to true, the operator does not generate a liveness probe for new pods belonging to this group

Default Value: `false`

***

### .spec.syncmasters.probes.livenessProbeSpec.failureThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L295)</sup>

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: `3`

***

### .spec.syncmasters.probes.livenessProbeSpec.initialDelaySeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L278)</sup>

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: `2`

***

### .spec.syncmasters.probes.livenessProbeSpec.periodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L282)</sup>

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: `10`

***

### .spec.syncmasters.probes.livenessProbeSpec.successThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L290)</sup>

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: `1`

***

### .spec.syncmasters.probes.livenessProbeSpec.timeoutSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L286)</sup>

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: `2`

***

### .spec.syncmasters.probes.ReadinessProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L252)</sup>

OldReadinessProbeDisabled if true readinessProbes are disabled
Deprecated: This field is deprecated, kept only for backward compatibility.

***

### .spec.syncmasters.probes.readinessProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L254)</sup>

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

***

### .spec.syncmasters.probes.readinessProbeSpec.failureThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L295)</sup>

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: `3`

***

### .spec.syncmasters.probes.readinessProbeSpec.initialDelaySeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L278)</sup>

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: `2`

***

### .spec.syncmasters.probes.readinessProbeSpec.periodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L282)</sup>

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: `10`

***

### .spec.syncmasters.probes.readinessProbeSpec.successThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L290)</sup>

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: `1`

***

### .spec.syncmasters.probes.readinessProbeSpec.timeoutSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L286)</sup>

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: `2`

***

### .spec.syncmasters.probes.startupProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L259)</sup>

StartupProbeDisabled if true startupProbes are disabled

***

### .spec.syncmasters.probes.startupProbeSpec.failureThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L295)</sup>

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: `3`

***

### .spec.syncmasters.probes.startupProbeSpec.initialDelaySeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L278)</sup>

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: `2`

***

### .spec.syncmasters.probes.startupProbeSpec.periodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L282)</sup>

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: `10`

***

### .spec.syncmasters.probes.startupProbeSpec.successThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L290)</sup>

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: `1`

***

### .spec.syncmasters.probes.startupProbeSpec.timeoutSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L286)</sup>

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: `2`

***

### .spec.syncmasters.pvcResizeMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L170)</sup>

VolumeResizeMode specified resize mode for PVCs and PVs

Possible Values: 
* runtime (default) - PVC will be resized in Pod runtime (EKS, GKE)
* rotate - Pod will be shutdown and PVC will be resized (AKS)

***

### .spec.syncmasters.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L95)</sup>

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

***

### .spec.syncmasters.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L89)</sup>

SchedulerName define scheduler name used for group

***

### .spec.syncmasters.securityContext.addCapabilities

Type: `[]core.Capability` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L41)</sup>

AddCapabilities add new capabilities to containers

***

### .spec.syncmasters.securityContext.allowPrivilegeEscalation

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L44)</sup>

AllowPrivilegeEscalation Controls whether a process can gain more privileges than its parent process.

***

### .spec.syncmasters.securityContext.dropAllCapabilities

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L38)</sup>

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

***

### .spec.syncmasters.securityContext.fsGroup

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L61)</sup>

FSGroup is a special supplemental group that applies to all containers in a pod.

***

### .spec.syncmasters.securityContext.privileged

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L47)</sup>

Privileged If true, runs container in privileged mode. Processes in privileged containers are
essentially equivalent to root on the host.

***

### .spec.syncmasters.securityContext.readOnlyRootFilesystem

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)</sup>

ReadOnlyRootFilesystem if true, mounts the container's root filesystem as read-only.

***

### .spec.syncmasters.securityContext.runAsGroup

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L55)</sup>

RunAsGroup is the GID to run the entrypoint of the container process.

***

### .spec.syncmasters.securityContext.runAsNonRoot

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L51)</sup>

RunAsNonRoot if true, indicates that the container must run as a non-root user.

***

### .spec.syncmasters.securityContext.runAsUser

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L53)</sup>

RunAsUser is the UID to run the entrypoint of the container process.

***

### .spec.syncmasters.securityContext.seccompProfile

Type: `core.SeccompProfile` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L77)</sup>

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

***

### .spec.syncmasters.securityContext.seLinuxOptions

Type: `core.SELinuxOptions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L82)</sup>

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

***

### .spec.syncmasters.securityContext.supplementalGroups

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L59)</sup>

SupplementalGroups is a list of groups applied to the first process run in each container, in addition to the container's primary GID,
the fsGroup (if specified), and group memberships defined in the container image for the uid of the container process.

***

### .spec.syncmasters.securityContext.sysctls

Type: `map[string]intstr.IntOrString` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L72)</sup>

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

***

### .spec.syncmasters.serviceAccountName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L146)</sup>

ServiceAccountName setting specifies the `serviceAccountName` for the `Pods` created
for each server of this group. If empty, it defaults to using the
`default` service account.
Using an alternative `ServiceAccount` is typically used to separate access rights.
The ArangoDB deployments need some very minimal access rights. With the
deployment of the operator, we grant the rights to 'get' all 'pod' resources.
If you are using a different service account, please grant these rights
to that service account.

***

### .spec.syncmasters.shutdownDelay

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L211)</sup>

ShutdownDelay define how long operator should delay finalizer removal after shutdown

***

### .spec.syncmasters.shutdownMethod

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L209)</sup>

ShutdownMethod describe procedure of member shutdown taken by Operator

***

### .spec.syncmasters.sidecarCoreNames

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L187)</sup>

SidecarCoreNames is a list of sidecar containers which must run in the pod.
Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

***

### .spec.syncmasters.sidecars

Type: `[]core.Container` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L191)</sup>

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

***

### .spec.syncmasters.storageClassName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L91)</sup>

StorageClassName specifies the classname for storage of the servers.

***

### .spec.syncmasters.terminationGracePeriodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L224)</sup>

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

***

### .spec.syncmasters.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L122)</sup>

Tolerations specifies the tolerations added to Pods in this group.
By default, suitable tolerations are set for the following keys with the `NoExecute` effect:
- `node.kubernetes.io/not-ready`
- `node.kubernetes.io/unreachable`
- `node.alpha.kubernetes.io/unreachable` (will be removed in future version)
For more information on tolerations, consult the https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

***

### .spec.syncmasters.volumeAllowShrink

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L172)</sup>

Deprecated: VolumeAllowShrink allows shrink the volume

***

### .spec.syncmasters.volumeClaimTemplate

Type: `core.PersistentVolumeClaim` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L165)</sup>

VolumeClaimTemplate specifies a volumeClaimTemplate used by operator to create to volume claims for pods of this group.
This setting is not available for group `coordinators`, `syncmasters` & `syncworkers`.
The default value describes a volume with `8Gi` storage, `ReadWriteOnce` access mode and volume mode set to `PersistentVolumeFilesystem`.
If this field is not set and `spec.<group>.resources.requests.storage` is set, then a default volume claim
with size as specified by `spec.<group>.resources.requests.storage` will be created. In that case `storage`
and `iops` is not forwarded to the pods resource requirements.

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

***

### .spec.syncmasters.volumeMounts

Type: `[]ServerGroupSpecVolumeMount` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L201)</sup>

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation of ServerGroupSpecVolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

***

### .spec.syncmasters.volumes\[int\].configMap

Type: `core.ConfigMapVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L138)</sup>

ConfigMap which should be mounted into pod

Links:
* [Documentation of core.ConfigMapVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#configmapvolumesource-v1-core)

***

### .spec.syncmasters.volumes\[int\].emptyDir

Type: `core.EmptyDirVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L143)</sup>

EmptyDir

Links:
* [Documentation of core.EmptyDirVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#emptydirvolumesource-v1-core)

***

### .spec.syncmasters.volumes\[int\].hostPath

Type: `core.HostPathVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L148)</sup>

HostPath

Links:
* [Documentation of core.HostPathVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#hostpathvolumesource-v1-core)

***

### .spec.syncmasters.volumes\[int\].name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L128)</sup>

Name of volume

***

### .spec.syncmasters.volumes\[int\].persistentVolumeClaim

Type: `core.PersistentVolumeClaimVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L153)</sup>

PersistentVolumeClaim

Links:
* [Documentation of core.PersistentVolumeClaimVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaimvolumesource-v1-core)

***

### .spec.syncmasters.volumes\[int\].secret

Type: `core.SecretVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L133)</sup>

Secret which should be mounted into pod

Links:
* [Documentation of core.SecretVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#secretvolumesource-v1-core)

***

### .spec.syncworkers.affinity

Type: `core.PodAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L180)</sup>

Affinity specified additional affinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.PodAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core)

***

### .spec.syncworkers.allowMemberRecreation

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L222)</sup>

AllowMemberRecreation allows to recreate member.
This setting changes the member recreation logic based on group:
- For Sync Masters, Sync Workers, Coordinator and DB-Servers it determines if a member can be recreated in case of failure (default `true`)
- For Agents and Single this value is hardcoded to `false` and the value provided in spec is ignored.

***

### .spec.syncworkers.annotations

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L125)</sup>

Annotations specified the annotations added to Pods in this group.
Annotations are merged with `spec.annotations`.

***

### .spec.syncworkers.annotationsIgnoreList

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L127)</sup>

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

***

### .spec.syncworkers.annotationsMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L129)</sup>

AnnotationsMode Define annotations mode which should be use while overriding annotations

***

### .spec.syncworkers.antiAffinity

Type: `core.PodAntiAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L176)</sup>

AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of core.Pod.AntiAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core)

***

### .spec.syncworkers.args

Type: `[]string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L85)</sup>

Args setting specifies additional command-line arguments passed to all servers of this group.

Default Value: `[]`

***

### .spec.syncworkers.count

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L77)</sup>

Count setting specifies the number of servers to start for the given group.
For the Agent group, this value must be a positive, odd number.
The default value is `3` for all groups except `single` (there the default is `1`
for `spec.mode: Single` and `2` for `spec.mode: ActiveFailover`).
For the `syncworkers` group, it is highly recommended to use the same number
as for the `dbservers` group.

***

### .spec.syncworkers.entrypoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L87)</sup>

Entrypoint overrides container executable

***

### .spec.syncworkers.envs\[int\].name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_env_var.go#L26)</sup>

***

### .spec.syncworkers.envs\[int\].value

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_env_var.go#L27)</sup>

***

### .spec.syncworkers.ephemeralVolumes.apps.size

Type: `resource.Quantity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)</sup>

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

***

### .spec.syncworkers.ephemeralVolumes.temp.size

Type: `resource.Quantity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_ephemeral_volumes.go#L64)</sup>

Size define size of the ephemeral volume

Links:
* [Documentation of resource.Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#quantity-resource-core)

***

### .spec.syncworkers.exporterPort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L235)</sup>

ExporterPort define Port used by exporter

***

### .spec.syncworkers.extendedRotationCheck

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L205)</sup>

ExtendedRotationCheck extend checks for rotation

***

### .spec.syncworkers.externalPortEnabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L217)</sup>

ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members

***

### .spec.syncworkers.indexMethod

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L228)</sup>

IndexMethod define group Indexing method

Possible Values: 
* random (default) - Pick random ID for member. Enforced on the Community Operator.
* ordered - Use sequential number as Member ID, starting from 0. Enterprise Operator required.

***

### .spec.syncworkers.initContainers.containers

Type: `[]core.Container` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_init_containers.go#L91)</sup>

Containers contains list of containers

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

***

### .spec.syncworkers.initContainers.mode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_init_containers.go#L94)</sup>

Mode keep container replace mode

***

### .spec.syncworkers.internalPort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L213)</sup>

InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

***

### .spec.syncworkers.internalPortProtocol

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L215)</sup>

InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members

***

### .spec.syncworkers.labels

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L131)</sup>

Labels specified the labels added to Pods in this group.

***

### .spec.syncworkers.labelsIgnoreList

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L133)</sup>

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

***

### .spec.syncworkers.labelsMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L135)</sup>

LabelsMode Define labels mode which should be use while overriding labels

***

### .spec.syncworkers.maxCount

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L81)</sup>

MaxCount specifies a maximum for the count of servers. If set, a specification is invalid if `count > maxCount`.

***

### .spec.syncworkers.memoryReservation

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L107)</sup>

MemoryReservation determines the system reservation of memory while calculating `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` value.
If this field is set, `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` is reduced by a specified value in percent.
Accepted Range <0, 50>. If the value is outside the accepted range, it is adjusted to the closest value.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: `0`

***

### .spec.syncworkers.minCount

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L79)</sup>

MinCount specifies a minimum for the count of servers. If set, a specification is invalid if `count < minCount`.

***

### .spec.syncworkers.nodeAffinity

Type: `core.NodeAffinity` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L184)</sup>

NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions

Links:
* [Documentation of code.NodeAffinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core)

***

### .spec.syncworkers.nodeSelector

Type: `map[string]string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L150)</sup>

NodeSelector setting specifies a set of labels to be used as `nodeSelector` for Pods of this node.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/)

***

### .spec.syncworkers.numactl.args

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_numactl_spec.go#L38)</sup>

Args define list of the numactl process

Default Value: `[]`

***

### .spec.syncworkers.numactl.enabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_numactl_spec.go#L30)</sup>

Enabled define if numactl should be enabled

Default Value: `false`

***

### .spec.syncworkers.numactl.path

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_numactl_spec.go#L34)</sup>

Path define numactl path within the container

Default Value: `/usr/bin/numactl`

***

### .spec.syncworkers.overrideDetectedNumberOfCores

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L113)</sup>

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` Container Environment Variable

OverrideDetectedNumberOfCores determines if number of cores should be overridden based on values in resources.
If is set to true and Container CPU Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` to the value from the Container CPU Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: `true`

***

### .spec.syncworkers.overrideDetectedTotalMemory

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L101)</sup>

**Important**: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` Container Environment Variable

OverrideDetectedTotalMemory determines if memory should be overridden based on values in resources.
If is set to true and Container Memory Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` to the value from the Container Memory Limits.

Links:
* [Docs of the ArangoDB Envs](https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/)

Default Value: `true`

***

### .spec.syncworkers.podModes.network

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L31)</sup>

***

### .spec.syncworkers.podModes.pid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec_pod_modes.go#L32)</sup>

***

### .spec.syncworkers.port

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L233)</sup>

Port define Port used by member

***

### .spec.syncworkers.priorityClassName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L156)</sup>

PriorityClassName specifies a priority class name
Will be forwarded to the pod spec.

Links:
* [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/)

***

### .spec.syncworkers.probes.livenessProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L245)</sup>

LivenessProbeDisabled if set to true, the operator does not generate a liveness probe for new pods belonging to this group

Default Value: `false`

***

### .spec.syncworkers.probes.livenessProbeSpec.failureThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L295)</sup>

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: `3`

***

### .spec.syncworkers.probes.livenessProbeSpec.initialDelaySeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L278)</sup>

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: `2`

***

### .spec.syncworkers.probes.livenessProbeSpec.periodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L282)</sup>

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: `10`

***

### .spec.syncworkers.probes.livenessProbeSpec.successThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L290)</sup>

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: `1`

***

### .spec.syncworkers.probes.livenessProbeSpec.timeoutSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L286)</sup>

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: `2`

***

### .spec.syncworkers.probes.ReadinessProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L252)</sup>

OldReadinessProbeDisabled if true readinessProbes are disabled
Deprecated: This field is deprecated, kept only for backward compatibility.

***

### .spec.syncworkers.probes.readinessProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L254)</sup>

ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility

***

### .spec.syncworkers.probes.readinessProbeSpec.failureThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L295)</sup>

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: `3`

***

### .spec.syncworkers.probes.readinessProbeSpec.initialDelaySeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L278)</sup>

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: `2`

***

### .spec.syncworkers.probes.readinessProbeSpec.periodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L282)</sup>

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: `10`

***

### .spec.syncworkers.probes.readinessProbeSpec.successThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L290)</sup>

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: `1`

***

### .spec.syncworkers.probes.readinessProbeSpec.timeoutSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L286)</sup>

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: `2`

***

### .spec.syncworkers.probes.startupProbeDisabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L259)</sup>

StartupProbeDisabled if true startupProbes are disabled

***

### .spec.syncworkers.probes.startupProbeSpec.failureThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L295)</sup>

FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
Giving up means restarting the container.
Minimum value is 1.

Default Value: `3`

***

### .spec.syncworkers.probes.startupProbeSpec.initialDelaySeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L278)</sup>

InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
Minimum value is 0.

Default Value: `2`

***

### .spec.syncworkers.probes.startupProbeSpec.periodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L282)</sup>

PeriodSeconds How often (in seconds) to perform the probe.
Minimum value is 1.

Default Value: `10`

***

### .spec.syncworkers.probes.startupProbeSpec.successThreshold

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L290)</sup>

SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
Minimum value is 1.

Default Value: `1`

***

### .spec.syncworkers.probes.startupProbeSpec.timeoutSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L286)</sup>

TimeoutSeconds specifies number of seconds after which the probe times out
Minimum value is 1.

Default Value: `2`

***

### .spec.syncworkers.pvcResizeMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L170)</sup>

VolumeResizeMode specified resize mode for PVCs and PVs

Possible Values: 
* runtime (default) - PVC will be resized in Pod runtime (EKS, GKE)
* rotate - Pod will be shutdown and PVC will be resized (AKS)

***

### .spec.syncworkers.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L95)</sup>

Resources holds resource requests & limits

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

***

### .spec.syncworkers.schedulerName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L89)</sup>

SchedulerName define scheduler name used for group

***

### .spec.syncworkers.securityContext.addCapabilities

Type: `[]core.Capability` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L41)</sup>

AddCapabilities add new capabilities to containers

***

### .spec.syncworkers.securityContext.allowPrivilegeEscalation

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L44)</sup>

AllowPrivilegeEscalation Controls whether a process can gain more privileges than its parent process.

***

### .spec.syncworkers.securityContext.dropAllCapabilities

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L38)</sup>

DropAllCapabilities specifies if capabilities should be dropped for this pod containers
Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.

***

### .spec.syncworkers.securityContext.fsGroup

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L61)</sup>

FSGroup is a special supplemental group that applies to all containers in a pod.

***

### .spec.syncworkers.securityContext.privileged

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L47)</sup>

Privileged If true, runs container in privileged mode. Processes in privileged containers are
essentially equivalent to root on the host.

***

### .spec.syncworkers.securityContext.readOnlyRootFilesystem

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L49)</sup>

ReadOnlyRootFilesystem if true, mounts the container's root filesystem as read-only.

***

### .spec.syncworkers.securityContext.runAsGroup

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L55)</sup>

RunAsGroup is the GID to run the entrypoint of the container process.

***

### .spec.syncworkers.securityContext.runAsNonRoot

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L51)</sup>

RunAsNonRoot if true, indicates that the container must run as a non-root user.

***

### .spec.syncworkers.securityContext.runAsUser

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L53)</sup>

RunAsUser is the UID to run the entrypoint of the container process.

***

### .spec.syncworkers.securityContext.seccompProfile

Type: `core.SeccompProfile` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L77)</sup>

SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.

Links:
* [Documentation of core.SeccompProfile](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core)

***

### .spec.syncworkers.securityContext.seLinuxOptions

Type: `core.SELinuxOptions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L82)</sup>

SELinuxOptions are the labels to be applied to the container

Links:
* [Documentation of core.SELinuxOptions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core)

***

### .spec.syncworkers.securityContext.supplementalGroups

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L59)</sup>

SupplementalGroups is a list of groups applied to the first process run in each container, in addition to the container's primary GID,
the fsGroup (if specified), and group memberships defined in the container image for the uid of the container process.

***

### .spec.syncworkers.securityContext.sysctls

Type: `map[string]intstr.IntOrString` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_security_context_spec.go#L72)</sup>

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

***

### .spec.syncworkers.serviceAccountName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L146)</sup>

ServiceAccountName setting specifies the `serviceAccountName` for the `Pods` created
for each server of this group. If empty, it defaults to using the
`default` service account.
Using an alternative `ServiceAccount` is typically used to separate access rights.
The ArangoDB deployments need some very minimal access rights. With the
deployment of the operator, we grant the rights to 'get' all 'pod' resources.
If you are using a different service account, please grant these rights
to that service account.

***

### .spec.syncworkers.shutdownDelay

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L211)</sup>

ShutdownDelay define how long operator should delay finalizer removal after shutdown

***

### .spec.syncworkers.shutdownMethod

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L209)</sup>

ShutdownMethod describe procedure of member shutdown taken by Operator

***

### .spec.syncworkers.sidecarCoreNames

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L187)</sup>

SidecarCoreNames is a list of sidecar containers which must run in the pod.
Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.

***

### .spec.syncworkers.sidecars

Type: `[]core.Container` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L191)</sup>

Sidecars specifies a list of additional containers to be started

Links:
* [Documentation of core.Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core)

***

### .spec.syncworkers.storageClassName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L91)</sup>

StorageClassName specifies the classname for storage of the servers.

***

### .spec.syncworkers.terminationGracePeriodSeconds

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L224)</sup>

TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation

***

### .spec.syncworkers.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L122)</sup>

Tolerations specifies the tolerations added to Pods in this group.
By default, suitable tolerations are set for the following keys with the `NoExecute` effect:
- `node.kubernetes.io/not-ready`
- `node.kubernetes.io/unreachable`
- `node.alpha.kubernetes.io/unreachable` (will be removed in future version)
For more information on tolerations, consult the https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

***

### .spec.syncworkers.volumeAllowShrink

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L172)</sup>

Deprecated: VolumeAllowShrink allows shrink the volume

***

### .spec.syncworkers.volumeClaimTemplate

Type: `core.PersistentVolumeClaim` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L165)</sup>

VolumeClaimTemplate specifies a volumeClaimTemplate used by operator to create to volume claims for pods of this group.
This setting is not available for group `coordinators`, `syncmasters` & `syncworkers`.
The default value describes a volume with `8Gi` storage, `ReadWriteOnce` access mode and volume mode set to `PersistentVolumeFilesystem`.
If this field is not set and `spec.<group>.resources.requests.storage` is set, then a default volume claim
with size as specified by `spec.<group>.resources.requests.storage` will be created. In that case `storage`
and `iops` is not forwarded to the pods resource requirements.

Links:
* [Documentation of core.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core)

***

### .spec.syncworkers.volumeMounts

Type: `[]ServerGroupSpecVolumeMount` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_spec.go#L201)</sup>

VolumeMounts define list of volume mounts mounted into server container

Links:
* [Documentation of ServerGroupSpecVolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core)

***

### .spec.syncworkers.volumes\[int\].configMap

Type: `core.ConfigMapVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L138)</sup>

ConfigMap which should be mounted into pod

Links:
* [Documentation of core.ConfigMapVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#configmapvolumesource-v1-core)

***

### .spec.syncworkers.volumes\[int\].emptyDir

Type: `core.EmptyDirVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L143)</sup>

EmptyDir

Links:
* [Documentation of core.EmptyDirVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#emptydirvolumesource-v1-core)

***

### .spec.syncworkers.volumes\[int\].hostPath

Type: `core.HostPathVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L148)</sup>

HostPath

Links:
* [Documentation of core.HostPathVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#hostpathvolumesource-v1-core)

***

### .spec.syncworkers.volumes\[int\].name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L128)</sup>

Name of volume

***

### .spec.syncworkers.volumes\[int\].persistentVolumeClaim

Type: `core.PersistentVolumeClaimVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L153)</sup>

PersistentVolumeClaim

Links:
* [Documentation of core.PersistentVolumeClaimVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaimvolumesource-v1-core)

***

### .spec.syncworkers.volumes\[int\].secret

Type: `core.SecretVolumeSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/server_group_volume.go#L133)</sup>

Secret which should be mounted into pod

Links:
* [Documentation of core.SecretVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#secretvolumesource-v1-core)

***

### .spec.timeouts.actions

Type: `map[string]meta.Duration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/timeouts.go#L44)</sup>

Actions keep map of the actions timeouts.

Links:
* [List of supported action names](/docs/generated/actions.md)
* [Definition of meta.Duration](https://github.com/kubernetes/apimachinery/blob/v0.26.6/pkg/apis/meta/v1/duration.go)

Example:
```yaml
actions:
  AddMember: 30m
```

***

### .spec.timeouts.maintenanceGracePeriod

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/timeouts.go#L36)</sup>

MaintenanceGracePeriod action timeout

***

### .spec.timezone

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_spec.go#L269)</sup>

Timezone if specified, will set a timezone for deployment.
Must be in format accepted by "tzdata", e.g. `America/New_York` or `Europe/London`

***

### .spec.tls.altNames

Type: `[]string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/tls_spec.go#L72)</sup>

AltNames setting specifies a list of alternate names that will be added to all generated
certificates. These names can be DNS names or email addresses.
The default value is empty.

***

### .spec.tls.caSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/tls_spec.go#L67)</sup>

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

***

### .spec.tls.mode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/tls_spec.go#L81)</sup>

***

### .spec.tls.sni.mapping.\<string\>

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/tls_sni_spec.go#L30)</sup>

***

### .spec.tls.ttl

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/tls_spec.go#L79)</sup>

TTL setting specifies the time to live of all generated server certificates.
When the server certificate is about to expire, it will be automatically replaced
by a new one and the affected server will be restarted.
Note: The time to live of the CA certificate (when created automatically)
will be set to 10 years.

Default Value: `"2160h" (about 3 months)`

***

### .spec.topology.enabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/topology_spec.go#L26)</sup>

***

### .spec.topology.label

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/topology_spec.go#L28)</sup>

***

### .spec.topology.zones

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/topology_spec.go#L27)</sup>

***

### .spec.upgrade.autoUpgrade

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_upgrade_spec.go#L26)</sup>

AutoUpgrade flag specifies if upgrade should be auto-injected, even if is not required (in case of stuck)

Default Value: `false`

***

### .spec.upgrade.debugLog

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/deployment/v1/deployment_upgrade_spec.go#L30)</sup>

DebugLog flag specifies if containers running upgrade process should print more debugging information.
This applies only to init containers.

Default Value: `false`

