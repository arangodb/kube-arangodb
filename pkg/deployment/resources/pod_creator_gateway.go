//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//

package resources

import (
	"context"
	"fmt"
	"math"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/collection"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

const (
	ArangoGatewayExecutor string = "/usr/local/bin/envoy"
	GatewayVolumeMountDir        = "/etc/gateway/"
	GatewayVolumeName            = "gateway"
	GatewayConfigFileName        = "gateway.yaml"
	GatewayConfigFilePath        = GatewayVolumeMountDir + GatewayConfigFileName
)

type ArangoGatewayContainer struct {
	groupSpec    api.ServerGroupSpec
	spec         api.DeploymentSpec
	group        api.ServerGroup
	resources    *Resources
	imageInfo    api.ImageInfo
	apiObject    meta.Object
	memberStatus api.MemberStatus
	arangoMember api.ArangoMember
}

var _ interfaces.PodCreator = &MemberGatewayPod{}
var _ interfaces.ContainerCreator = &ArangoGatewayContainer{}

type MemberGatewayPod struct {
	podName string

	groupSpec    api.ServerGroupSpec
	spec         api.DeploymentSpec
	group        api.ServerGroup
	arangoMember api.ArangoMember
	resources    *Resources
	imageInfo    api.ImageInfo
	apiObject    meta.Object
	memberStatus api.MemberStatus
	cachedStatus interfaces.Inspector
}

func GetGatewayConfigMapName(name string) string {
	return fmt.Sprintf("%s-gateway", name)
}

func (a *ArangoGatewayContainer) GetCommand() ([]string, error) {
	cmd := make([]string, 0, 128)
	cmd = append(cmd, a.GetExecutor())
	cmd = append(cmd, createArangoGatewayArgs(a.apiObject, a.spec, a.group, a.groupSpec, a.memberStatus)...)
	return cmd, nil
}

func (a *ArangoGatewayContainer) GetName() string {
	return shared.ServerContainerName
}

func (a *ArangoGatewayContainer) GetPorts() []core.ContainerPort {
	port := shared.ArangoPort

	return []core.ContainerPort{
		{
			Name:          shared.ServerContainerName,
			ContainerPort: int32(port),
			Protocol:      core.ProtocolTCP,
		},
	}
}

func (a *ArangoGatewayContainer) GetExecutor() string {
	return a.groupSpec.GetEntrypoint(ArangoGatewayExecutor)
}

func (a *ArangoGatewayContainer) GetSecurityContext() *core.SecurityContext {
	return k8sutil.CreateSecurityContext(a.groupSpec.SecurityContext)
}

func (a *ArangoGatewayContainer) GetProbes() (*core.Probe, *core.Probe, *core.Probe, error) {
	var liveness, readiness, startup *core.Probe

	probeLivenessConfig, err := a.resources.getLivenessProbe(a.spec, a.group, a.imageInfo)
	if err != nil {
		return nil, nil, nil, err
	}

	probeReadinessConfig, err := a.resources.getReadinessProbe(a.spec, a.group, a.imageInfo)
	if err != nil {
		return nil, nil, nil, err
	}

	probeStartupConfig, err := a.resources.getReadinessProbe(a.spec, a.group, a.imageInfo)
	if err != nil {
		return nil, nil, nil, err
	}

	if probeLivenessConfig != nil {
		liveness = probeLivenessConfig.Create()
	}

	if probeReadinessConfig != nil {
		readiness = probeReadinessConfig.Create()
	}

	if probeStartupConfig != nil {
		startup = probeStartupConfig.Create()
	}

	return liveness, readiness, startup, nil
}

func (a *ArangoGatewayContainer) GetResourceRequirements() core.ResourceRequirements {
	return kresources.ExtractPodAcceptedResourceRequirement(a.arangoMember.Spec.Overrides.GetResources(&a.groupSpec))
}

func (a *ArangoGatewayContainer) GetLifecycle() (*core.Lifecycle, error) {
	return k8sutil.NewLifecycleFinalizers()
}

func (a *ArangoGatewayContainer) GetImagePullPolicy() core.PullPolicy {
	return a.spec.GetImagePullPolicy()
}

func (a *ArangoGatewayContainer) GetImage() string {
	return a.imageInfo.Image
}

func (a *ArangoGatewayContainer) GetEnvs() ([]core.EnvVar, []core.EnvFromSource) {
	envs := NewEnvBuilder()

	envs.Add(true, k8sutil.GetLifecycleEnv()...)

	if len(a.groupSpec.Envs) > 0 {
		for _, env := range a.groupSpec.Envs {
			// Do not override preset envs
			envs.Add(false, core.EnvVar{
				Name:  env.Name,
				Value: env.Value,
			})
		}
	}

	return envs.GetEnvList(), nil
}

func (a *ArangoGatewayContainer) GetVolumeMounts() []core.VolumeMount {
	return createGatewayVolumes(a.apiObject.GetName()).VolumeMounts()
}

func (m *MemberGatewayPod) GetName() string {
	return m.resources.context.GetAPIObject().GetName()
}

func (m *MemberGatewayPod) GetRole() string {
	return m.group.AsRole()
}

func (m *MemberGatewayPod) GetImagePullSecrets() []string {
	return m.spec.ImagePullSecrets
}

func (m *MemberGatewayPod) GetPodAntiAffinity() *core.PodAntiAffinity {
	a := &core.PodAntiAffinity{}

	pod.AppendPodAntiAffinityDefault(m, a)

	a = kresources.MergePodAntiAffinity(a, m.groupSpec.AntiAffinity)

	return kresources.OptionalPodAntiAffinity(a)
}

func (m *MemberGatewayPod) GetPodAffinity() *core.PodAffinity {
	a := &core.PodAffinity{}

	pod.AppendAffinityWithRole(m, a, api.ServerGroupDBServers.AsRole())

	a = kresources.MergePodAffinity(a, m.groupSpec.Affinity)

	return kresources.OptionalPodAffinity(a)
}

func (m *MemberGatewayPod) GetNodeAffinity() *core.NodeAffinity {
	a := &core.NodeAffinity{}

	pod.AppendArchSelector(a, m.memberStatus.Architecture.Default(m.spec.Architecture.GetDefault()).AsNodeSelectorRequirement())

	a = kresources.MergeNodeAffinity(a, m.groupSpec.NodeAffinity)

	return kresources.OptionalNodeAffinity(a)
}

func (m *MemberGatewayPod) GetNodeSelector() map[string]string {
	return m.groupSpec.GetNodeSelector()
}

func (m *MemberGatewayPod) GetServiceAccountName() string {
	return m.groupSpec.GetServiceAccountName()
}

func (m *MemberGatewayPod) GetSidecars(pod *core.Pod) error {
	// A sidecar provided by the user
	sidecars := m.groupSpec.GetSidecars()
	if len(sidecars) > 0 {
		addLifecycleSidecar(m.groupSpec.SidecarCoreNames, sidecars)
		pod.Spec.Containers = append(pod.Spec.Containers, sidecars...)
	}

	return nil
}

func (m *MemberGatewayPod) GetVolumes() []core.Volume {
	return createGatewayVolumes(m.apiObject.GetName()).Volumes()
}

func (m *MemberGatewayPod) IsDeploymentMode() bool {
	return m.spec.IsDevelopment()
}

func (m *MemberGatewayPod) GetInitContainers(cachedStatus interfaces.Inspector) ([]core.Container, error) {
	var initContainers []core.Container
	if c := m.groupSpec.InitContainers.GetContainers(); len(c) > 0 {
		initContainers = append(initContainers, c...)
	}

	res := kresources.ExtractPodInitContainerAcceptedResourceRequirement(m.GetContainerCreator().GetResourceRequirements())

	initContainers = applyInitContainersResourceResources(initContainers, res)
	initContainers = upscaleInitContainersResourceResources(initContainers, res)

	return initContainers, nil
}

func (m *MemberGatewayPod) GetFinalizers() []string {
	return nil
}

func (m *MemberGatewayPod) GetTolerations() []core.Toleration {
	return m.resources.CreatePodTolerations(m.group, m.groupSpec)
}

func (m *MemberGatewayPod) GetContainerCreator() interfaces.ContainerCreator {
	return &ArangoGatewayContainer{
		groupSpec:    m.groupSpec,
		spec:         m.spec,
		group:        m.group,
		resources:    m.resources,
		imageInfo:    m.imageInfo,
		apiObject:    m.apiObject,
		memberStatus: m.memberStatus,
		arangoMember: m.arangoMember,
	}
}

func (m *MemberGatewayPod) GetRestartPolicy() core.RestartPolicy {
	if features.RestartPolicyAlways().Enabled() {
		return core.RestartPolicyAlways
	}
	return core.RestartPolicyNever
}

func (m *MemberGatewayPod) Init(ctx context.Context, cachedStatus interfaces.Inspector, pod *core.Pod) error {
	terminationGracePeriodSeconds := int64(math.Ceil(m.groupSpec.GetTerminationGracePeriod(m.group).Seconds()))
	pod.Spec.TerminationGracePeriodSeconds = &terminationGracePeriodSeconds
	pod.Spec.PriorityClassName = m.groupSpec.PriorityClassName

	return nil
}

func (m *MemberGatewayPod) Validate(_ interfaces.Inspector) error {
	if err := validateSidecars(m.groupSpec.SidecarCoreNames, m.groupSpec.GetSidecars()); err != nil {
		return err
	}

	return nil
}

func (m *MemberGatewayPod) ApplyPodSpec(spec *core.PodSpec) error {
	if s := m.groupSpec.SchedulerName; s != nil {
		spec.SchedulerName = *s
	}

	m.groupSpec.PodModes.Apply(spec)

	return nil
}

func (m *MemberGatewayPod) Annotations() map[string]string {
	return collection.MergeAnnotations(m.spec.Annotations, m.groupSpec.Annotations)
}

func (m *MemberGatewayPod) Labels() map[string]string {
	return collection.ReservedLabels().Filter(collection.MergeAnnotations(m.spec.Labels, m.groupSpec.Labels))
}

func createGatewayVolumes(memberName string) pod.Volumes {
	volumes := pod.NewVolumes()

	volumes.AddVolume(k8sutil.LifecycleVolume())
	volumes.AddVolumeMount(k8sutil.LifecycleVolumeMount())

	volumes.AddVolume(k8sutil.CreateVolumeWithConfigMap(GatewayVolumeName, GetGatewayConfigMapName(memberName)))
	volumes.AddVolumeMount(GatewayVolumeMount())

	return volumes
}

func GatewayVolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:      GatewayVolumeName,
		MountPath: GatewayVolumeMountDir,
		ReadOnly:  true,
	}
}
