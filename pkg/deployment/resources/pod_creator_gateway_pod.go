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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	schedulerContainerResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container/resources"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/integrations/sidecar"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/collection"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

var _ interfaces.PodCreator = &MemberGatewayPod{}

type MemberGatewayPod struct {
	podName          string
	status           api.MemberStatus
	groupSpec        api.ServerGroupSpec
	spec             api.DeploymentSpec
	deploymentStatus api.DeploymentStatus
	group            api.ServerGroup
	arangoMember     api.ArangoMember
	context          Context
	resources        *Resources
	imageInfo        api.ImageInfo
	cachedStatus     interfaces.Inspector
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

func (m *MemberGatewayPod) AsInput() pod.Input {
	return pod.Input{
		ApiObject:    m.context.GetAPIObject(),
		Deployment:   m.spec,
		Status:       m.deploymentStatus,
		Group:        m.group,
		GroupSpec:    m.groupSpec,
		Version:      m.imageInfo.ArangoDBVersion,
		Enterprise:   m.imageInfo.Enterprise,
		Member:       m.status,
		ArangoMember: m.arangoMember,
	}
}

func (m *MemberGatewayPod) GetPodAffinity() *core.PodAffinity {
	a := &core.PodAffinity{}

	pod.AppendAffinityWithRole(m, a, api.ServerGroupDBServers.AsRole())

	a = kresources.MergePodAffinity(a, m.groupSpec.Affinity)

	return kresources.OptionalPodAffinity(a)
}

func (m *MemberGatewayPod) GetNodeAffinity() *core.NodeAffinity {
	a := &core.NodeAffinity{}

	pod.AppendArchSelector(a, m.status.Architecture.Default(m.spec.Architecture.GetDefault()).AsNodeSelectorRequirement())

	a = kresources.MergeNodeAffinity(a, m.groupSpec.NodeAffinity)

	return kresources.OptionalNodeAffinity(a)
}

func (m *MemberGatewayPod) GetNodeSelector() map[string]string {
	return m.groupSpec.GetNodeSelector()
}

func (m *MemberGatewayPod) GetServiceAccountName() string {
	return m.groupSpec.GetServiceAccountName()
}

func (m *MemberGatewayPod) GetSidecars(pod *core.PodTemplateSpec) error {
	// A sidecar provided by the user
	sidecars := m.groupSpec.GetSidecars()
	if len(sidecars) > 0 {
		addLifecycleSidecar(m.groupSpec.SidecarCoreNames, sidecars)
		pod.Spec.Containers = append(pod.Spec.Containers, sidecars...)
	}

	return nil
}

func (m *MemberGatewayPod) GetVolumes() []core.Volume {
	return createGatewayVolumes(m.AsInput()).Volumes()
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
		member:       m,
		spec:         m.spec,
		group:        m.group,
		resources:    m.resources,
		imageInfo:    m.imageInfo,
		groupSpec:    m.groupSpec,
		arangoMember: m.arangoMember,
		cachedStatus: m.cachedStatus,
		input:        m.AsInput(),
		status:       m.status,
	}
}

func (m *MemberGatewayPod) GetRestartPolicy() core.RestartPolicy {
	if features.RestartPolicyAlways().Enabled() {
		return core.RestartPolicyAlways
	}
	return core.RestartPolicyNever
}

func (m *MemberGatewayPod) Init(ctx context.Context, cachedStatus interfaces.Inspector, pod *core.PodTemplateSpec) error {
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
	l := collection.ReservedLabels().Filter(collection.MergeAnnotations(m.spec.Labels, m.groupSpec.Labels))

	if m.status.Topology != nil && m.deploymentStatus.Topology.Enabled() && m.deploymentStatus.Topology.ID == m.status.Topology.ID {
		if l == nil {
			l = map[string]string{}
		}

		l[k8sutil.LabelKeyArangoZone] = fmt.Sprintf("%d", m.status.Topology.Zone)
		l[k8sutil.LabelKeyArangoTopology] = string(m.status.Topology.ID)
	}

	return l
}

func (m *MemberGatewayPod) Profiles() (schedulerApi.ProfileTemplates, error) {
	integration, err := sidecar.NewIntegration(&schedulerContainerResourcesApi.Image{
		Image: util.NewType(m.resources.context.GetOperatorImage()),
	}, m.spec.Gateway.GetSidecar(), []string{shared.ServerContainerName})

	if err != nil {
		return nil, err
	}

	return []*schedulerApi.ProfileTemplate{integration}, nil
}
