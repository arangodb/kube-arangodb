//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	"os"
	"strconv"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/deployment/topology"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/collection"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
)

const (
	ArangoDExecutor                          = "/usr/sbin/arangod"
	ArangoDBOverrideDetectedTotalMemoryEnv   = "ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY"
	ArangoDBOverrideDetectedNumberOfCoresEnv = "ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES"
	ArangoDBOverrideServerGroupEnv           = "ARANGODB_OVERRIDE_SERVER_GROUP"
	ArangoDBOverrideDeploymentModeEnv        = "ARANGODB_OVERRIDE_DEPLOYMENT_MODE"
	ArangoDBOverrideVersionEnv               = "ARANGODB_OVERRIDE_VERSION"
	ArangoDBOverrideEnterpriseEnv            = "ARANGODB_OVERRIDE_ENTERPRISE"
	ArangoDBServerPortEnv                    = "ARANGODB_SERVER_PORT"
)

var _ interfaces.PodCreator = &MemberArangoDPod{}
var _ interfaces.ContainerCreator = &ArangoDContainer{}

type MemberArangoDPod struct {
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
	autoUpgrade      bool
	cachedStatus     interfaces.Inspector
}

type ArangoDContainer struct {
	member       *MemberArangoDPod
	resources    *Resources
	groupSpec    api.ServerGroupSpec
	spec         api.DeploymentSpec
	group        api.ServerGroup
	imageInfo    api.ImageInfo
	cachedStatus interfaces.Inspector
	input        pod.Input
	status       api.MemberStatus
}

// ArangoUpgradeContainer can construct ArangoD upgrade container.
type ArangoUpgradeContainer struct {
	interfaces.ContainerCreator
	cachedStatus interfaces.Inspector
	input        pod.Input
}

// ArangoVersionCheckContainer can construct ArangoD version check container.
type ArangoVersionCheckContainer struct {
	interfaces.ContainerCreator
	cachedStatus interfaces.Inspector
	input        pod.Input
	versionArgs  k8sutil.OptionPairs
}

func (a *ArangoDContainer) GetPorts() []core.ContainerPort {
	ports := []core.ContainerPort{
		{
			Name:          shared.ServerPortName,
			ContainerPort: int32(a.groupSpec.GetPort()),
			Protocol:      core.ProtocolTCP,
		},
	}

	if a.spec.Metrics.IsEnabled() {
		switch a.spec.Metrics.Mode.Get() {
		case api.MetricsModeInternal:
			ports = append(ports, core.ContainerPort{
				Name:          shared.ExporterPortName,
				ContainerPort: int32(a.groupSpec.GetPort()),
				Protocol:      core.ProtocolTCP,
			})
		}
	}

	return ports
}

func (a *ArangoDContainer) GetArgs() ([]string, error) {
	return createArangodArgs(a.cachedStatus, a.input)
}

func (a *ArangoDContainer) GetName() string {
	return shared.ServerContainerName
}

func (a *ArangoDContainer) GetExecutor() string {
	return a.groupSpec.GetEntrypoint(ArangoDExecutor)
}

func (a *ArangoDContainer) GetSecurityContext() *core.SecurityContext {
	return a.groupSpec.SecurityContext.NewSecurityContext()
}

func (a *ArangoDContainer) GetProbes() (*core.Probe, *core.Probe, *core.Probe, error) {
	var liveness, readiness, startup *core.Probe

	probeLivenessConfig, err := a.resources.getLivenessProbe(a.spec, a.group, a.imageInfo)
	if err != nil {
		return nil, nil, nil, err
	}

	probeReadinessConfig, err := a.resources.getReadinessProbe(a.spec, a.group, a.imageInfo)
	if err != nil {
		return nil, nil, nil, err
	}

	probeStartupConfig, err := a.resources.getStartupProbe(a.spec, a.group, a.imageInfo)
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

func (a *ArangoDContainer) GetImage() string {
	switch a.spec.ImageDiscoveryMode.Get() {
	case api.DeploymentImageDiscoveryDirectMode:
		// In case of direct mode ignore discovery
		return util.TypeOrDefault[string](a.spec.Image, a.imageInfo.ImageID)
	default:
		return a.imageInfo.ImageID
	}
}

// GetEnvs returns environment variables for ArangoDB containers.
func (a *ArangoDContainer) GetEnvs() ([]core.EnvVar, []core.EnvFromSource) {
	envs := NewEnvBuilder()

	if a.spec.License.HasSecretName() && a.imageInfo.ArangoDBVersion.CompareTo("3.9.0") < 0 {
		env := k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoLicenseKey, a.spec.License.GetSecretName(),
			constants.SecretKeyToken)

		envs.Add(true, env)
	}

	envs.Add(true, k8sutil.GetLifecycleEnv()...)

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

	envs.Add(true, core.EnvVar{
		Name:  ArangoDBOverrideServerGroupEnv,
		Value: a.input.Group.AsRole(),
	})
	envs.Add(true, core.EnvVar{
		Name:  ArangoDBOverrideDeploymentModeEnv,
		Value: string(a.input.Deployment.GetMode()),
	})
	envs.Add(true, core.EnvVar{
		Name:  ArangoDBOverrideVersionEnv,
		Value: string(a.input.Version),
	})
	envs.Add(true, core.EnvVar{
		Name:  ArangoDBOverrideEnterpriseEnv,
		Value: strconv.FormatBool(a.input.Enterprise),
	})

	if p := a.groupSpec.Port; p != nil {
		envs.Add(true, core.EnvVar{
			Name:  ArangoDBServerPortEnv,
			Value: fmt.Sprintf("%d", *p),
		})
	}

	envFromSource := []core.EnvFromSource{
		{
			ConfigMapRef: &core.ConfigMapEnvSource{
				LocalObjectReference: core.LocalObjectReference{
					Name: features.ConfigMapName(),
				},
				// Optional in case if operator could not create it when process started.
				Optional: util.NewType[bool](true),
			},
		},
	}

	return envs.GetEnvList(), envFromSource
}

func (a *ArangoDContainer) GetResourceRequirements() core.ResourceRequirements {
	return k8sutil.ExtractPodResourceRequirement(a.groupSpec.Resources)
}

func (a *ArangoDContainer) GetLifecycle() (*core.Lifecycle, error) {
	if features.GracefulShutdown().Enabled() {
		return k8sutil.NewLifecyclePort()
	}
	return k8sutil.NewLifecycleFinalizers()
}

func (a *ArangoDContainer) GetImagePullPolicy() core.PullPolicy {
	return a.spec.GetImagePullPolicy()
}

func (a *ArangoDContainer) GetVolumeMounts() []core.VolumeMount {
	volumes := CreateArangoDVolumes(a.status, a.input, a.spec, a.groupSpec)

	return volumes.VolumeMounts()
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

func (m *MemberArangoDPod) Init(_ context.Context, _ interfaces.Inspector, pod *core.Pod) error {
	terminationGracePeriodSeconds := int64(math.Ceil(m.groupSpec.GetTerminationGracePeriod(m.group).Seconds()))
	pod.Spec.TerminationGracePeriodSeconds = &terminationGracePeriodSeconds
	pod.Spec.PriorityClassName = m.groupSpec.PriorityClassName

	return nil
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

	if err := validateSidecars(m.groupSpec.SidecarCoreNames, m.groupSpec.GetSidecars()); err != nil {
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

	pod.MergePodAntiAffinity(&a, topology.GetTopologyAffinityRules(m.context.GetName(), m.deploymentStatus, m.group, m.status).PodAntiAffinity)

	pod.MergePodAntiAffinity(&a, m.groupSpec.AntiAffinity)

	return pod.ReturnPodAntiAffinityOrNil(a)
}

func (m *MemberArangoDPod) GetPodAffinity() *core.PodAffinity {
	a := core.PodAffinity{}

	pod.MergePodAffinity(&a, m.groupSpec.Affinity)

	pod.MergePodAffinity(&a, topology.GetTopologyAffinityRules(m.context.GetName(), m.deploymentStatus, m.group, m.status).PodAffinity)

	return pod.ReturnPodAffinityOrNil(a)
}

func (m *MemberArangoDPod) GetNodeAffinity() *core.NodeAffinity {
	a := core.NodeAffinity{}

	pod.AppendArchSelector(&a, m.status.Architecture.Default(m.spec.Architecture.GetDefault()).AsNodeSelectorRequirement())

	pod.MergeNodeAffinity(&a, m.groupSpec.NodeAffinity)

	pod.MergeNodeAffinity(&a, topology.GetTopologyAffinityRules(m.context.GetName(), m.deploymentStatus, m.group, m.status).NodeAffinity)

	return pod.ReturnNodeAffinityOrNil(a)
}

func (m *MemberArangoDPod) GetNodeSelector() map[string]string {
	return m.groupSpec.GetNodeSelector()
}

func (m *MemberArangoDPod) GetServiceAccountName() string {
	return m.groupSpec.GetServiceAccountName()
}

func (m *MemberArangoDPod) GetSidecars(pod *core.Pod) error {
	if m.spec.Metrics.IsEnabled() && m.spec.Metrics.Mode.Get() != api.MetricsModeInternal {
		var c *core.Container

		pod.Labels[k8sutil.LabelKeyArangoExporter] = "yes"
		if container, err := m.createMetricsExporterSidecarInternalExporter(); err != nil {
			return err
		} else {
			c = container
		}
		if c != nil {
			pod.Spec.Containers = append(pod.Spec.Containers, *c)
		}
	}

	// A sidecar provided by the user
	sidecars := m.groupSpec.GetSidecars()
	if len(sidecars) > 0 {
		addLifecycleSidecar(m.groupSpec.SidecarCoreNames, sidecars)
		pod.Spec.Containers = append(pod.Spec.Containers, sidecars...)
	}

	return nil
}

func (m *MemberArangoDPod) GetVolumes() []core.Volume {
	volumes := CreateArangoDVolumes(m.status, m.AsInput(), m.spec, m.groupSpec)

	return volumes.Volumes()
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

	{
		c, err := k8sutil.InitLifecycleContainer(m.resources.context.GetOperatorImage(), &m.spec.Lifecycle.Resources,
			m.groupSpec.SecurityContext.NewSecurityContext())
		if err != nil {
			return nil, err
		}
		initContainers = append(initContainers, c)
	}

	{
		engine := m.spec.GetStorageEngine().AsArangoArgument()
		requireUUID := m.group == api.ServerGroupDBServers && m.status.IsInitialized

		c := k8sutil.ArangodInitContainer(api.ServerGroupReservedInitContainerNameUUID, m.status.ID, engine, executable, m.resources.context.GetOperatorImage(), requireUUID,
			m.groupSpec.SecurityContext.NewSecurityContext())
		initContainers = append(initContainers, c)
	}

	{
		// Upgrade container - run in background
		if m.autoUpgrade || m.status.Upgrade {
			upgradeContainer := &ArangoUpgradeContainer{
				m.GetContainerCreator(),
				cachedStatus,
				m.AsInput(),
			}
			c, err := k8sutil.NewContainer(upgradeContainer)
			if err != nil {
				return nil, err
			}

			initContainers = append(initContainers, c)
		}

		// VersionCheck Container
		{
			versionArgs := pod.UpgradeVersionCheck().Args(m.AsInput())
			if len(versionArgs) > 0 {
				upgradeContainer := &ArangoVersionCheckContainer{
					m.GetContainerCreator(),
					cachedStatus,
					m.AsInput(),
					versionArgs,
				}

				c, err := k8sutil.NewContainer(upgradeContainer)
				if err != nil {
					return nil, err
				}

				initContainers = append(initContainers, c)
			}
		}
	}

	return initContainers, nil
}

func (m *MemberArangoDPod) GetFinalizers() []string {
	return k8sutil.GetFinalizers(m.spec.GetServerGroupSpec(m.group), m.group)
}

func (m *MemberArangoDPod) GetTolerations() []core.Toleration {
	return m.resources.CreatePodTolerations(m.group, m.groupSpec)
}

func (m *MemberArangoDPod) GetContainerCreator() interfaces.ContainerCreator {
	return &ArangoDContainer{
		member:       m,
		spec:         m.spec,
		group:        m.group,
		resources:    m.resources,
		imageInfo:    m.imageInfo,
		groupSpec:    m.groupSpec,
		cachedStatus: m.cachedStatus,
		input:        m.AsInput(),
		status:       m.status,
	}
}

func (m *MemberArangoDPod) GetRestartPolicy() core.RestartPolicy {
	if features.RestartPolicyAlways().Enabled() {
		return core.RestartPolicyAlways
	}
	return core.RestartPolicyNever
}

func (m *MemberArangoDPod) createMetricsExporterSidecarInternalExporter() (*core.Container, error) {
	image := m.GetContainerCreator().GetImage()

	args := createInternalExporterArgs(m.spec, m.groupSpec, m.imageInfo.ArangoDBVersion)

	c, err := ArangodbInternalExporterContainer(image, args,
		createExporterLivenessProbe(m.spec.IsSecure() && m.spec.Metrics.IsTLS()), m.spec.Metrics.Resources,
		m.spec, m.groupSpec)
	if err != nil {
		return nil, err
	}

	if m.spec.Authentication.IsAuthenticated() && m.spec.Metrics.GetJWTTokenSecretName() != "" {
		c.VolumeMounts = append(c.VolumeMounts, k8sutil.ExporterJWTVolumeMount())
	}

	if pod.IsTLSEnabled(m.AsInput()) {
		c.VolumeMounts = append(c.VolumeMounts, k8sutil.TlsKeyfileVolumeMount())
	}

	return &c, nil
}

func (m *MemberArangoDPod) ApplyPodSpec(p *core.PodSpec) error {
	p.SecurityContext = m.groupSpec.SecurityContext.NewPodSecurityContext()

	if s := m.groupSpec.SchedulerName; s != nil {
		p.SchedulerName = *s
	}

	m.groupSpec.PodModes.Apply(p)

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

// CreateArangoDVolumes returns wrapper with volumes for a pod and volume mounts for a container.
func CreateArangoDVolumes(status api.MemberStatus, input pod.Input, spec api.DeploymentSpec,
	groupSpec api.ServerGroupSpec) pod.Volumes {
	volumes := pod.NewVolumes()

	volumes.AddVolumeMount(k8sutil.ArangodVolumeMount())

	volumes.AddVolumeMount(k8sutil.LifecycleVolumeMount())

	if status.PersistentVolumeClaim != nil {
		vol := k8sutil.CreateVolumeWithPersitantVolumeClaim(shared.ArangodVolumeName,
			status.PersistentVolumeClaim.GetName())

		volumes.AddVolume(vol)
	} else {
		volumes.AddVolume(k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName))
	}

	// TLS
	volumes.Append(pod.TLS(), input)

	// Encryption
	volumes.Append(pod.Encryption(), input)

	// Security
	volumes.Append(pod.Security(), input)

	if spec.Metrics.IsEnabled() {
		token := spec.Metrics.GetJWTTokenSecretName()
		if spec.Authentication.IsAuthenticated() && token != "" {
			vol := k8sutil.CreateVolumeWithSecret(shared.ExporterJWTVolumeName, token)
			volumes.AddVolume(vol)
		}
	}

	volumes.Append(pod.JWT(), input)

	volumes.AddVolume(k8sutil.LifecycleVolume())

	// SNI
	volumes.Append(pod.SNI(), input)

	volumes.Append(pod.Timezone(), input)

	if len(groupSpec.Volumes) > 0 {
		volumes.AddVolume(groupSpec.Volumes.RenderVolumes(input.ApiObject, input.Group, status)...)
	}

	if len(groupSpec.VolumeMounts) > 0 {
		volumes.AddVolumeMount(groupSpec.VolumeMounts.VolumeMounts()...)
	}

	return volumes
}

// GetArgs returns list of arguments for the ArangoD upgrade container.
func (a *ArangoUpgradeContainer) GetArgs() ([]string, error) {
	return createArangodArgsWithUpgrade(a.cachedStatus, a.input)
}

// GetLifecycle returns no lifecycle for the ArangoD upgrade container.
func (a *ArangoUpgradeContainer) GetLifecycle() (*core.Lifecycle, error) {
	return nil, nil
}

// GetName returns the name of the ArangoD upgrade container.
func (a *ArangoUpgradeContainer) GetName() string {
	return api.ServerGroupReservedInitContainerNameUpgrade
}

// GetProbes returns no probes for the ArangoD upgrade container.
func (a *ArangoUpgradeContainer) GetProbes() (*core.Probe, *core.Probe, *core.Probe, error) {
	return nil, nil, nil, nil
}

// GetArgs returns list of arguments for the ArangoD version check container.
func (a *ArangoVersionCheckContainer) GetArgs() ([]string, error) {
	return createArangodArgs(a.cachedStatus, a.input, a.versionArgs...)
}

// GetLifecycle returns no lifecycle for the ArangoD version check container.
func (a *ArangoVersionCheckContainer) GetLifecycle() (*core.Lifecycle, error) {
	return nil, nil
}

// GetName returns the name of the ArangoD version check container.
func (a *ArangoVersionCheckContainer) GetName() string {
	return api.ServerGroupReservedInitContainerNameVersionCheck
}

// GetProbes returns no probes for the ArangoD version check container.
func (a *ArangoVersionCheckContainer) GetProbes() (*core.Probe, *core.Probe, *core.Probe, error) {
	return nil, nil, nil, nil
}

// validateSidecars checks if all core names are in the sidecar list.
// It returns error when at least one core name is missing.
func validateSidecars(coreNames []string, sidecars []core.Container) error {
	for _, coreName := range coreNames {
		if api.IsReservedServerGroupContainerName(coreName) {
			return fmt.Errorf("sidecar core name \"%s\" can not be used because it is reserved", coreName)
		}

		found := false
		for _, sidecar := range sidecars {
			if sidecar.Name == coreName {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("sidecar core name \"%s\" does not exist on the sidecars' list", coreName)
		}
	}

	return nil

}

// addLifecycleSidecar adds lifecycle to all core sidecar unless the sidecar contains its own custom lifecycle.
func addLifecycleSidecar(coreNames []string, sidecars []core.Container) error {
	for _, coreName := range coreNames {
		for i, sidecar := range sidecars {
			if coreName != sidecar.Name {
				continue
			}

			if sidecar.Lifecycle != nil && sidecar.Lifecycle.PreStop != nil {
				// A user provided a custom lifecycle preStop, so break and check next core name container.
				break
			}

			if !k8sutil.VolumeMountExists(sidecar.VolumeMounts, shared.LifecycleVolumeName) {
				sidecars[i].VolumeMounts = append(sidecars[i].VolumeMounts, k8sutil.LifecycleVolumeMount())
			}

			sidecars[i].Env = k8sutil.AppendLifecycleEnv(sidecars[i].Env)

			lifecycle, err := k8sutil.NewLifecycleFinalizers()
			if err != nil {
				return err
			}

			if sidecar.Lifecycle == nil {
				sidecars[i].Lifecycle = lifecycle
			} else {
				// Set only preStop, because user can provide postStart lifecycle.
				sidecars[i].Lifecycle.PreStop = lifecycle.PreStop
			}

			break
		}
	}

	return nil
}
