//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Tomasz Mielech <tomasz@arangodb.com>
//

package resources

import (
	"fmt"
	"math"
	"os"

	"github.com/arangodb/kube-arangodb/pkg/deployment/topology"

	"github.com/arangodb/kube-arangodb/pkg/deployment/features"

	"github.com/arangodb/kube-arangodb/pkg/util/collection"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"

	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"

	"github.com/arangodb/kube-arangodb/pkg/util"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	core "k8s.io/api/core/v1"
)

const (
	ArangoDExecutor                          = "/usr/sbin/arangod"
	ArangoDBOverrideDetectedTotalMemoryEnv   = "ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY"
	ArangoDBOverrideDetectedNumberOfCoresEnv = "ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES"
)

var _ interfaces.PodCreator = &MemberArangoDPod{}
var _ interfaces.ContainerCreator = &ArangoDContainer{}

type MemberArangoDPod struct {
	status           api.MemberStatus
	groupSpec        api.ServerGroupSpec
	spec             api.DeploymentSpec
	deploymentStatus api.DeploymentStatus
	group            api.ServerGroup
	arangoMember     api.ArangoMember
	context          Context
	resources        *Resources
	imageInfo        api.ImageInfo
	autoUpgrade      bool
	id               string
}

type ArangoDContainer struct {
	member    *MemberArangoDPod
	resources *Resources
	groupSpec api.ServerGroupSpec
	spec      api.DeploymentSpec
	group     api.ServerGroup
	imageInfo api.ImageInfo
}

func (a *ArangoDContainer) GetPorts() []core.ContainerPort {
	ports := []core.ContainerPort{
		{
			Name:          "server",
			ContainerPort: int32(k8sutil.ArangoPort),
			Protocol:      core.ProtocolTCP,
		},
	}

	if a.spec.Metrics.IsEnabled() {
		switch a.spec.Metrics.Mode.Get() {
		case api.MetricsModeInternal:
			ports = append(ports, core.ContainerPort{
				Name:          "exporter",
				ContainerPort: int32(k8sutil.ArangoPort),
				Protocol:      core.ProtocolTCP,
			})
		}
	}

	return ports
}

func (a *ArangoDContainer) GetExecutor() string {
	return a.groupSpec.GetEntrypoint(ArangoDExecutor)
}

func (a *ArangoDContainer) GetSecurityContext() *core.SecurityContext {
	return a.groupSpec.SecurityContext.NewSecurityContext()
}

func (a *ArangoDContainer) GetProbes() (*core.Probe, *core.Probe, error) {
	var liveness, readiness *core.Probe

	probeLivenessConfig, err := a.resources.getLivenessProbe(a.spec, a.group, a.imageInfo.ArangoDBVersion)
	if err != nil {
		return nil, nil, err
	}

	probeReadinessConfig, err := a.resources.getReadinessProbe(a.spec, a.group, a.imageInfo.ArangoDBVersion)
	if err != nil {
		return nil, nil, err
	}

	if probeLivenessConfig != nil {
		liveness = probeLivenessConfig.Create()
	}

	if probeReadinessConfig != nil {
		readiness = probeReadinessConfig.Create()
	}

	return liveness, readiness, nil
}

func (a *ArangoDContainer) GetImage() string {
	switch a.spec.ImageDiscoveryMode.Get() {
	case api.DeploymentImageDiscoveryDirectMode:
		// In case of direct mode ignore discovery
		return util.StringOrDefault(a.spec.Image, a.imageInfo.ImageID)
	default:
		return a.imageInfo.ImageID
	}
}

func (a *ArangoDContainer) GetEnvs() []core.EnvVar {
	envs := NewEnvBuilder()

	if env := pod.JWT().Envs(a.member.AsInput()); len(env) > 0 {
		envs.Add(true, env...)
	}

	if a.spec.License.HasSecretName() {
		env := k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoLicenseKey, a.spec.License.GetSecretName(),
			constants.SecretKeyToken)

		envs.Add(true, env)
	}

	if a.resources.context.GetLifecycleImage() != "" {
		envs.Add(true, k8sutil.GetLifecycleEnv()...)
	}

	if a.groupSpec.Resources.Limits != nil {
		if a.groupSpec.GetOverrideDetectedTotalMemory() {
			if limits, ok := a.groupSpec.Resources.Limits[core.ResourceMemory]; ok {
				envs.Add(true, core.EnvVar{
					Name:  ArangoDBOverrideDetectedTotalMemoryEnv,
					Value: fmt.Sprintf("%d", limits.Value()),
				})
			}
		}

		if a.groupSpec.GetOverrideDetectedNumberOfCores() {
			if limits, ok := a.groupSpec.Resources.Limits[core.ResourceCPU]; ok {
				envs.Add(true, core.EnvVar{
					Name:  ArangoDBOverrideDetectedNumberOfCoresEnv,
					Value: fmt.Sprintf("%d", limits.Value()),
				})
			}
		}
	}

	if len(a.groupSpec.Envs) > 0 {
		for _, env := range a.groupSpec.Envs {
			// Do not override preset envs
			envs.Add(false, core.EnvVar{
				Name:  env.Name,
				Value: env.Value,
			})
		}
	}

	envs.Add(true, pod.Topology().Envs(a.member.AsInput())...)

	return envs.GetEnvList()
}

func (a *ArangoDContainer) GetResourceRequirements() core.ResourceRequirements {
	return k8sutil.ExtractPodResourceRequirement(a.groupSpec.Resources)
}

func (a *ArangoDContainer) GetLifecycle() (*core.Lifecycle, error) {
	if a.resources.context.GetLifecycleImage() != "" {
		return k8sutil.NewLifecycle()
	}
	return nil, nil
}

func (a *ArangoDContainer) GetImagePullPolicy() core.PullPolicy {
	return a.spec.GetImagePullPolicy()
}

func (m *MemberArangoDPod) AsInput() pod.Input {
	return pod.Input{
		ApiObject:    m.context.GetAPIObject(),
		Deployment:   m.spec,
		Status:       m.deploymentStatus,
		Group:        m.group,
		GroupSpec:    m.groupSpec,
		Version:      m.imageInfo.ArangoDBVersion,
		Enterprise:   m.imageInfo.Enterprise,
		AutoUpgrade:  m.autoUpgrade,
		Member:       m.status,
		ArangoMember: m.arangoMember,
	}
}

func (m *MemberArangoDPod) Init(pod *core.Pod) {
	terminationGracePeriodSeconds := int64(math.Ceil(m.group.DefaultTerminationGracePeriod().Seconds()))
	pod.Spec.TerminationGracePeriodSeconds = &terminationGracePeriodSeconds
	pod.Spec.PriorityClassName = m.groupSpec.PriorityClassName
}

func (m *MemberArangoDPod) Validate(cachedStatus interfaces.Inspector) error {
	i := m.AsInput()

	if err := pod.SNI().Verify(i, cachedStatus); err != nil {
		return err
	}

	if err := pod.Encryption().Verify(i, cachedStatus); err != nil {
		return err
	}

	if err := pod.JWT().Verify(i, cachedStatus); err != nil {
		return err
	}

	if err := pod.TLS().Verify(i, cachedStatus); err != nil {
		return err
	}

	return nil
}

func (m *MemberArangoDPod) GetName() string {
	return m.resources.context.GetAPIObject().GetName()
}

func (m *MemberArangoDPod) GetRole() string {
	return m.group.AsRole()
}

func (m *MemberArangoDPod) GetImagePullSecrets() []string {
	return m.spec.ImagePullSecrets
}

func (m *MemberArangoDPod) GetPodAntiAffinity() *core.PodAntiAffinity {
	a := core.PodAntiAffinity{}

	pod.AppendPodAntiAffinityDefault(m, &a)

	pod.MergePodAntiAffinity(&a, topology.GetTopologyAffinityRules(m.context.GetName(), m.group, m.deploymentStatus.Topology, m.status.Topology).PodAntiAffinity)

	pod.MergePodAntiAffinity(&a, m.groupSpec.AntiAffinity)

	return pod.ReturnPodAntiAffinityOrNil(a)
}

func (m *MemberArangoDPod) GetPodAffinity() *core.PodAffinity {
	a := core.PodAffinity{}

	pod.MergePodAffinity(&a, m.groupSpec.Affinity)

	pod.MergePodAffinity(&a, topology.GetTopologyAffinityRules(m.context.GetName(), m.group, m.deploymentStatus.Topology, m.status.Topology).PodAffinity)

	return pod.ReturnPodAffinityOrNil(a)
}

func (m *MemberArangoDPod) GetNodeAffinity() *core.NodeAffinity {
	a := core.NodeAffinity{}

	pod.AppendNodeSelector(&a)

	pod.MergeNodeAffinity(&a, m.groupSpec.NodeAffinity)

	pod.MergeNodeAffinity(&a, topology.GetTopologyAffinityRules(m.context.GetName(), m.group, m.deploymentStatus.Topology, m.status.Topology).NodeAffinity)

	return pod.ReturnNodeAffinityOrNil(a)
}

func (m *MemberArangoDPod) GetNodeSelector() map[string]string {
	return m.groupSpec.GetNodeSelector()
}

func (m *MemberArangoDPod) GetServiceAccountName() string {
	return m.groupSpec.GetServiceAccountName()
}

func (m *MemberArangoDPod) GetSidecars(pod *core.Pod) error {
	if m.spec.Metrics.IsEnabled() {
		var c *core.Container

		if features.MetricsExporter().Enabled() {
			pod.Labels[k8sutil.LabelKeyArangoExporter] = "yes"
			if container, err := m.createMetricsExporterSidecarInternalExporter(); err != nil {
				return err
			} else {
				c = container
			}
		} else {
			switch m.spec.Metrics.Mode.Get() {
			case api.MetricsModeExporter:
				if !m.group.IsExportMetrics() {
					break
				}
				fallthrough
			case api.MetricsModeSidecar:
				c = m.createMetricsExporterSidecarExternalExporter()

				pod.Labels[k8sutil.LabelKeyArangoExporter] = "yes"
			default:
				pod.Labels[k8sutil.LabelKeyArangoExporter] = "yes"
			}

		}
		if c != nil {
			pod.Spec.Containers = append(pod.Spec.Containers, *c)
		}
	}

	// A sidecar provided by the user
	sidecars := m.groupSpec.GetSidecars()
	if len(sidecars) > 0 {
		pod.Spec.Containers = append(pod.Spec.Containers, sidecars...)
	}

	return nil
}

func (m *MemberArangoDPod) GetVolumes() ([]core.Volume, []core.VolumeMount) {
	volumes := pod.NewVolumes()

	volumes.AddVolumeMount(k8sutil.ArangodVolumeMount())

	if m.resources.context.GetLifecycleImage() != "" {
		volumes.AddVolumeMount(k8sutil.LifecycleVolumeMount())
	}

	if m.status.PersistentVolumeClaimName != "" {
		vol := k8sutil.CreateVolumeWithPersitantVolumeClaim(k8sutil.ArangodVolumeName,
			m.status.PersistentVolumeClaimName)

		volumes.AddVolume(vol)
	} else {
		volumes.AddVolume(k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName))
	}

	// TLS
	volumes.Append(pod.TLS(), m.AsInput())

	// Encryption
	volumes.Append(pod.Encryption(), m.AsInput())

	// Security
	volumes.Append(pod.Security(), m.AsInput())

	if m.spec.Metrics.IsEnabled() {
		if features.MetricsExporter().Enabled() {
			token := m.spec.Metrics.GetJWTTokenSecretName()
			if token != "" {
				vol := k8sutil.CreateVolumeWithSecret(k8sutil.ExporterJWTVolumeName, token)
				volumes.AddVolume(vol)
			}
		} else {
			switch m.spec.Metrics.Mode.Get() {
			case api.MetricsModeExporter:
				if !m.group.IsExportMetrics() {
					break
				}
				fallthrough
			case api.MetricsModeSidecar:
				token := m.spec.Metrics.GetJWTTokenSecretName()
				if token != "" {
					vol := k8sutil.CreateVolumeWithSecret(k8sutil.ExporterJWTVolumeName, token)
					volumes.AddVolume(vol)
				}
			}
		}
	}

	volumes.Append(pod.JWT(), m.AsInput())

	if m.resources.context.GetLifecycleImage() != "" {
		volumes.AddVolume(k8sutil.LifecycleVolume())
	}

	// SNI
	volumes.Append(pod.SNI(), m.AsInput())

	if len(m.groupSpec.Volumes) > 0 {
		volumes.AddVolume(m.groupSpec.Volumes.Volumes()...)
	}

	if len(m.groupSpec.VolumeMounts) > 0 {
		volumes.AddVolumeMount(m.groupSpec.VolumeMounts.VolumeMounts()...)
	}

	return volumes.Volumes(), volumes.VolumeMounts()
}

func (m *MemberArangoDPod) IsDeploymentMode() bool {
	return m.spec.IsDevelopment()
}

func (m *MemberArangoDPod) GetInitContainers(cachedStatus interfaces.Inspector) ([]core.Container, error) {
	var initContainers []core.Container

	if c := m.groupSpec.InitContainers.GetContainers(); len(c) > 0 {
		initContainers = append(initContainers, c...)
	}

	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}

	lifecycleImage := m.resources.context.GetLifecycleImage()
	if lifecycleImage != "" {
		c, err := k8sutil.InitLifecycleContainer(lifecycleImage, &m.spec.Lifecycle.Resources,
			m.groupSpec.SecurityContext.NewSecurityContext())
		if err != nil {
			return nil, err
		}
		initContainers = append(initContainers, c)
	}

	operatorUUIDImage := m.resources.context.GetOperatorUUIDImage()
	if operatorUUIDImage != "" {
		engine := m.spec.GetStorageEngine().AsArangoArgument()
		requireUUID := m.group == api.ServerGroupDBServers && m.status.IsInitialized

		c := k8sutil.ArangodInitContainer(api.ServerGroupReservedInitContainerNameUUID, m.status.ID, engine, executable, operatorUUIDImage, requireUUID,
			m.groupSpec.SecurityContext.NewSecurityContext())
		initContainers = append(initContainers, c)
	}

	{
		// Upgrade container - run in background
		if m.autoUpgrade || m.status.Upgrade {
			args, err := createArangodArgsWithUpgrade(cachedStatus, m.AsInput())
			if err != nil {
				return nil, err
			}

			c, err := k8sutil.NewContainer(args, m.GetContainerCreator())
			if err != nil {
				return nil, err
			}

			_, c.VolumeMounts = m.GetVolumes()

			c.Name = api.ServerGroupReservedInitContainerNameUpgrade
			c.Lifecycle = nil
			c.LivenessProbe = nil
			c.ReadinessProbe = nil

			initContainers = append(initContainers, c)
		}

		// VersionCheck Container
		{
			versionArgs := pod.UpgradeVersionCheck().Args(m.AsInput())
			if len(versionArgs) > 0 {
				args, err := createArangodArgs(cachedStatus, m.AsInput(), versionArgs...)
				if err != nil {
					return nil, err
				}

				c, err := k8sutil.NewContainer(args, m.GetContainerCreator())
				if err != nil {
					return nil, err
				}

				_, c.VolumeMounts = m.GetVolumes()

				c.Name = api.ServerGroupReservedInitContainerNameVersionCheck
				c.Lifecycle = nil
				c.LivenessProbe = nil
				c.ReadinessProbe = nil

				initContainers = append(initContainers, c)
			}
		}
	}

	return initContainers, nil
}

func (m *MemberArangoDPod) GetFinalizers() []string {
	return m.resources.CreatePodFinalizers(m.group)
}

func (m *MemberArangoDPod) GetTolerations() []core.Toleration {
	return m.resources.CreatePodTolerations(m.group, m.groupSpec)
}

func (m *MemberArangoDPod) GetContainerCreator() interfaces.ContainerCreator {
	return &ArangoDContainer{
		member:    m,
		spec:      m.spec,
		group:     m.group,
		resources: m.resources,
		imageInfo: m.imageInfo,
		groupSpec: m.groupSpec,
	}
}

func (m *MemberArangoDPod) createMetricsExporterSidecarInternalExporter() (*core.Container, error) {
	image := m.GetContainerCreator().GetImage()

	args := createInternalExporterArgs(m.spec, m.groupSpec, m.imageInfo.ArangoDBVersion)

	c, err := ArangodbInternalExporterContainer(image, args,
		createExporterLivenessProbe(m.spec.IsSecure() && m.spec.Metrics.IsTLS()), m.spec.Metrics.Resources,
		m.groupSpec.SecurityContext.NewSecurityContext(),
		m.spec)
	if err != nil {
		return nil, err
	}

	if m.spec.Metrics.GetJWTTokenSecretName() != "" {
		c.VolumeMounts = append(c.VolumeMounts, k8sutil.ExporterJWTVolumeMount())
	}

	if pod.IsTLSEnabled(m.AsInput()) {
		c.VolumeMounts = append(c.VolumeMounts, k8sutil.TlsKeyfileVolumeMount())
	}

	return &c, nil
}

func (m *MemberArangoDPod) createMetricsExporterSidecarExternalExporter() *core.Container {
	image := m.context.GetMetricsExporterImage()
	if m.spec.Metrics.HasImage() {
		image = m.spec.Metrics.GetImage()
	}

	args := createExporterArgs(m.spec, m.groupSpec)
	if m.spec.Metrics.Mode.Get() == api.MetricsModeSidecar {
		args = append(args, "--mode=passthru")
	}

	c := ArangodbExporterContainer(image, args,
		createExporterLivenessProbe(m.spec.IsSecure() && m.spec.Metrics.IsTLS()), m.spec.Metrics.Resources,
		m.groupSpec.SecurityContext.NewSecurityContext(),
		m.spec)

	if m.spec.Metrics.GetJWTTokenSecretName() != "" {
		c.VolumeMounts = append(c.VolumeMounts, k8sutil.ExporterJWTVolumeMount())
	}

	if pod.IsTLSEnabled(m.AsInput()) {
		c.VolumeMounts = append(c.VolumeMounts, k8sutil.TlsKeyfileVolumeMount())
	}

	return &c
}

func (m *MemberArangoDPod) ApplyPodSpec(p *core.PodSpec) error {
	p.SecurityContext = m.groupSpec.SecurityContext.NewPodSecurityContext()

	if s := m.groupSpec.SchedulerName; s != nil {
		p.SchedulerName = *s
	}

	return nil
}

func (m *MemberArangoDPod) Annotations() map[string]string {
	return collection.MergeAnnotations(m.spec.Annotations, m.groupSpec.Annotations)
}

func (m *MemberArangoDPod) Labels() map[string]string {
	l := collection.ReservedLabels().Filter(collection.MergeAnnotations(m.spec.Labels, m.groupSpec.Labels))

	if m.group.IsArangod() && m.status.Topology != nil && m.deploymentStatus.Topology.Enabled() && m.deploymentStatus.Topology.ID == m.status.Topology.ID {
		if l == nil {
			l = map[string]string{}
		}

		l[k8sutil.LabelKeyArangoZone] = fmt.Sprintf("%d", m.status.Topology.Zone)
		l[k8sutil.LabelKeyArangoTopology] = string(m.status.Topology.ID)
	}

	return l
}
