//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/integrations/sidecar"
	"github.com/arangodb/kube-arangodb/pkg/util/collection"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

var _ interfaces.PodCreator = &MemberGatewayPod{}

type MemberGatewayPod struct {
	pod.Input

	podName      string
	context      Context
	resources    *Resources
	cachedStatus interfaces.Inspector
}

func (m *MemberGatewayPod) GetName() string {
	return m.resources.context.GetAPIObject().GetName()
}

func (m *MemberGatewayPod) GetRole() string {
	return m.Group.AsRole()
}

func (m *MemberGatewayPod) GetImagePullSecrets() []string {
	return m.Deployment.ImagePullSecrets
}

func (m *MemberGatewayPod) GetPodAntiAffinity() *core.PodAntiAffinity {
	a := &core.PodAntiAffinity{}

	pod.AppendPodAntiAffinityDefault(m, a)

	a = kresources.MergePodAntiAffinity(a, m.GroupSpec.AntiAffinity)

	return kresources.OptionalPodAntiAffinity(a)
}

func (m *MemberGatewayPod) GetPodAffinity() *core.PodAffinity {
	a := &core.PodAffinity{}

	pod.AppendAffinityWithRole(m, a, api.ServerGroupDBServers.AsRole())

	a = kresources.MergePodAffinity(a, m.GroupSpec.Affinity)

	return kresources.OptionalPodAffinity(a)
}

func (m *MemberGatewayPod) GetNodeAffinity() *core.NodeAffinity {
	a := &core.NodeAffinity{}

	pod.AppendArchSelector(a, m.Member.Architecture.Default(m.Deployment.Architecture.GetDefault()).AsNodeSelectorRequirement())

	a = kresources.MergeNodeAffinity(a, m.GroupSpec.NodeAffinity)

	return kresources.OptionalNodeAffinity(a)
}

func (m *MemberGatewayPod) GetNodeSelector() map[string]string {
	return m.GroupSpec.GetNodeSelector()
}

func (m *MemberGatewayPod) GetServiceAccountName() string {
	return m.GroupSpec.GetServiceAccountName()
}

func (m *MemberGatewayPod) GetSidecars(pod *core.PodTemplateSpec) error {
	// A sidecar provided by the user
	sidecars := m.GroupSpec.GetSidecars()
	if len(sidecars) > 0 {
		addLifecycleSidecar(m.GroupSpec.SidecarCoreNames, sidecars)
		pod.Spec.Containers = append(pod.Spec.Containers, sidecars...)
	}

	return nil
}

func (m *MemberGatewayPod) GetVolumes() []core.Volume {
	return createGatewayVolumes(m.Input).Volumes()
}

func (m *MemberGatewayPod) IsDeploymentMode() bool {
	return m.Deployment.IsDevelopment()
}

func (m *MemberGatewayPod) GetInitContainers(cachedStatus interfaces.Inspector) ([]core.Container, error) {
	var initContainers []core.Container
	if c := m.GroupSpec.InitContainers.GetContainers(); len(c) > 0 {
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
	return m.resources.CreatePodTolerations(m.Group, m.GroupSpec)
}

func (m *MemberGatewayPod) GetContainerCreator() interfaces.ContainerCreator {
	return &MemberGatewayContainer{
		MemberGatewayPod: m,
		resources:        m.resources,
	}
}

func (m *MemberGatewayPod) GetRestartPolicy() core.RestartPolicy {
	return getDefaultRestartPolicy(m.GroupSpec)
}

func (m *MemberGatewayPod) Init(ctx context.Context, cachedStatus interfaces.Inspector, pod *core.PodTemplateSpec) error {
	terminationGracePeriodSeconds := int64(math.Ceil(m.GroupSpec.GetTerminationGracePeriod(m.Group).Seconds()))
	pod.Spec.TerminationGracePeriodSeconds = &terminationGracePeriodSeconds
	pod.Spec.PriorityClassName = m.GroupSpec.PriorityClassName

	return nil
}

func (m *MemberGatewayPod) Validate(cachedStatus interfaces.Inspector) error {
	if err := pod.SNI().Verify(m.Input, cachedStatus); err != nil {
		return err
	}

	if err := validateSidecars(m.GroupSpec.SidecarCoreNames, m.GroupSpec.GetSidecars()); err != nil {
		return err
	}

	if c, err := cachedStatus.ArangoProfile().V1Beta1(); err != nil {
		return err
	} else {
		if _, ok := c.GetSimple(fmt.Sprintf("%s-int", m.context.GetName())); !ok {
			return errors.Errorf("Unable to find deployment integration")
		}
	}

	return nil
}

func (m *MemberGatewayPod) ApplyPodSpec(spec *core.PodSpec) error {
	if s := m.GroupSpec.SchedulerName; s != nil {
		spec.SchedulerName = *s
	}

	m.GroupSpec.PodModes.Apply(spec)

	return nil
}

func (m *MemberGatewayPod) Annotations() map[string]string {
	return collection.MergeAnnotations(m.Deployment.Annotations, m.GroupSpec.Annotations)
}

func (m *MemberGatewayPod) Labels() map[string]string {
	l := collection.ReservedLabels().Filter(collection.MergeAnnotations(m.Deployment.Labels, m.GroupSpec.Labels))

	if m.Member.Topology != nil && m.Status.Topology.Enabled() && m.Status.Topology.ID == m.Member.Topology.ID {
		if l == nil {
			l = map[string]string{}
		}

		l[k8sutil.LabelKeyArangoZone] = fmt.Sprintf("%d", m.Member.Topology.Zone)
		l[k8sutil.LabelKeyArangoTopology] = string(m.Member.Topology.ID)
	}

	return l
}

func (m *MemberGatewayPod) Profiles() (schedulerApi.ProfileTemplates, error) {
	c, err := m.cachedStatus.ArangoProfile().V1Beta1()
	if err != nil {
		return nil, err
	}

	integration, ok := c.GetSimple(fmt.Sprintf("%s-int", m.context.GetName()))
	if !ok {
		return nil, errors.Errorf("Unable to find deployment integration")
	}

	if integration.Status.Accepted == nil {
		return nil, errors.Errorf("Unable to find accepted integration")
	}

	integrations, err := sidecar.NewIntegrationEnablement(
		sidecar.IntegrationEnvoyV3{
			DeploymentName: m.context.GetName(),
			Spec:           m.Deployment,
		}, sidecar.IntegrationAuthenticationV1{
			DeploymentName: m.context.GetName(),
			Spec:           m.Deployment,
		})

	if err != nil {
		return nil, err
	}

	shutdownAnnotation := sidecar.NewShutdownAnnotations([]string{shared.ServerContainerName})

	return []*schedulerApi.ProfileTemplate{integration.Status.Accepted.Template, integrations, shutdownAnnotation}, nil
}
