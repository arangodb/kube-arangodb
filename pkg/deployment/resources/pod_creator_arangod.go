//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/deployment/topology"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/collection"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
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
	pod.Input

	podName      string
	context      Context
	resources    *Resources
	cachedStatus interfaces.Inspector
}

type ArangoDContainer struct {
	*MemberArangoDPod
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
			ContainerPort: int32(a.GroupSpec.GetPort()),
			Protocol:      core.ProtocolTCP,
		},
	}

	if a.Deployment.Metrics.IsEnabled() {
		//nolint:staticcheck
		switch a.Deployment.Metrics.Mode.Get() {
		case api.MetricsModeInternal:
			ports = append(ports, core.ContainerPort{
				Name:          shared.ExporterPortName,
				ContainerPort: int32(a.GroupSpec.GetPort()),
				Protocol:      core.ProtocolTCP,
			})
		}
	}

	return ports
}

func (a *ArangoDContainer) GetCommand() ([]string, error) {
	cmd := make([]string, 0, 128)

	if args := createArangodNumactl(a.GroupSpec); len(args) > 0 {
		cmd = append(cmd, args...)
	}

	cmd = append(cmd, a.GetExecutor())

	args, err := createArangodArgs(a.cachedStatus, a.Input)
	if err != nil {
		return nil, err
	}

	cmd = append(cmd, args...)

	return cmd, nil
}

func (a *ArangoDContainer) GetName() string {
	return shared.ServerContainerName
}

func (a *ArangoDContainer) GetExecutor() string {
	return a.GroupSpec.GetEntrypoint(ArangoDExecutor)
}

func (a *ArangoDContainer) GetSecurityContext() *core.SecurityContext {
	return k8sutil.CreateSecurityContext(a.GroupSpec.SecurityContext)
}

func (a *ArangoDContainer) GetProbes() (*core.Probe, *core.Probe, *core.Probe, error) {
	var liveness, readiness, startup *core.Probe

	probeLivenessConfig, err := a.resources.getLivenessProbe(a.Deployment, a.Group, a.Image)
	if err != nil {
		return nil, nil, nil, err
	}

	probeReadinessConfig, err := a.resources.getReadinessProbe(a.Deployment, a.Group, a.Image)
	if err != nil {
		return nil, nil, nil, err
	}

	probeStartupConfig, err := a.resources.getStartupProbe(a.Deployment, a.Group, a.Image)
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
	switch a.Deployment.ImageDiscoveryMode.Get() {
	case api.DeploymentImageDiscoveryDirectMode:
		// In case of direct mode ignore discovery
		return util.TypeOrDefault[string](a.Deployment.Image, a.Image.ImageID)
	default:
		return a.Image.ImageID
	}
}

// GetEnvs returns environment variables for ArangoDB containers.
func (a *ArangoDContainer) GetEnvs() ([]core.EnvVar, []core.EnvFromSource) {
	envs := NewEnvBuilder()

	if a.Deployment.License.HasSecretName() && a.Image.ArangoDBVersion.CompareTo("3.9.0") < 0 {
		env := k8sutil.CreateEnvSecretKeySelector(utilConstants.EnvArangoLicenseKey, a.Deployment.License.GetSecretName(),
			utilConstants.SecretKeyToken)

		envs.Add(true, env)
	}

	envs.Add(true, k8sutil.GetLifecycleEnv()...)

	resources := a.ArangoMember.Spec.Overrides.GetResources(&a.GroupSpec)

	if resources.Limits != nil {
		if a.GroupSpec.GetOverrideDetectedTotalMemory() {
			if limits, ok := resources.Limits[core.ResourceMemory]; ok {
				value := a.GroupSpec.CalculateMemoryReservation(limits.Value())

				envs.Add(true, core.EnvVar{
					Name:  ArangoDBOverrideDetectedTotalMemoryEnv,
					Value: fmt.Sprintf("%d", value),
				})
			}
		}

		if a.GroupSpec.GetOverrideDetectedNumberOfCores() {
			if limits, ok := resources.Limits[core.ResourceCPU]; ok {
				envs.Add(true, core.EnvVar{
					Name:  ArangoDBOverrideDetectedNumberOfCoresEnv,
					Value: fmt.Sprintf("%d", limits.Value()),
				})
			}
		}
	}

	if len(a.GroupSpec.Envs) > 0 {
		for _, env := range a.GroupSpec.Envs {
			// Do not override preset envs
			envs.Add(false, core.EnvVar{
				Name:  env.Name,
				Value: env.Value,
			})
		}
	}

	envs.Add(true, pod.Topology().Envs(a.Input)...)

	envs.Add(true, core.EnvVar{
		Name:  ArangoDBOverrideServerGroupEnv,
		Value: a.Group.AsRole(),
	})
	envs.Add(true, core.EnvVar{
		Name:  ArangoDBOverrideDeploymentModeEnv,
		Value: string(a.Deployment.GetMode()),
	})
	envs.Add(true, core.EnvVar{
		Name:  ArangoDBOverrideVersionEnv,
		Value: string(a.Image.ArangoDBVersion),
	})
	envs.Add(true, core.EnvVar{
		Name:  ArangoDBOverrideEnterpriseEnv,
		Value: strconv.FormatBool(a.Image.Enterprise),
	})

	if p := a.GroupSpec.Port; p != nil {
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

func (a *ArangoDContainer) GetResourceRequirements(scale float64) core.ResourceRequirements {
	return kresources.ScaleResources(kresources.ExtractPodAcceptedResourceRequirement(a.ArangoMember.Spec.Overrides.GetResources(&a.GroupSpec)), scale)
}

func (a *ArangoDContainer) GetResourceRequirementsDefaultScale() float64 {
	return 1
}

func (a *ArangoDContainer) GetLifecycle() (*core.Lifecycle, error) {
	if features.GracefulShutdown().Enabled() {
		return k8sutil.NewLifecyclePort()
	}
	return k8sutil.NewLifecycleFinalizers()
}

func (a *ArangoDContainer) GetImagePullPolicy() core.PullPolicy {
	return a.Deployment.GetImagePullPolicy()
}

func (a *ArangoDContainer) GetVolumeMounts() []core.VolumeMount {
	volumes := CreateArangoDVolumes(a.Member, a.Input, a.Deployment, a.GroupSpec)

	return volumes.VolumeMounts()
}

func (m *MemberArangoDPod) Init(_ context.Context, _ interfaces.Inspector, pod *core.PodTemplateSpec) error {
	terminationGracePeriodSeconds := int64(math.Ceil(m.GroupSpec.GetTerminationGracePeriod(m.Group).Seconds()))
	pod.Spec.TerminationGracePeriodSeconds = &terminationGracePeriodSeconds
	pod.Spec.PriorityClassName = m.GroupSpec.PriorityClassName

	return nil
}

func (m *MemberArangoDPod) Validate(cachedStatus interfaces.Inspector) error {
	if err := pod.SNI().Verify(m.Input, cachedStatus); err != nil {
		return err
	}

	if err := pod.Encryption().Verify(m.Input, cachedStatus); err != nil {
		return err
	}

	if err := pod.JWT().Verify(m.Input, cachedStatus); err != nil {
		return err
	}

	if err := pod.TLS().Verify(m.Input, cachedStatus); err != nil {
		return err
	}

	if err := validateSidecars(m.GroupSpec.SidecarCoreNames, m.GroupSpec.GetSidecars()); err != nil {
		return err
	}

	return nil
}

func (m *MemberArangoDPod) GetName() string {
	return m.resources.context.GetAPIObject().GetName()
}

func (m *MemberArangoDPod) GetRole() string {
	return m.Group.AsRole()
}

func (m *MemberArangoDPod) GetImagePullSecrets() []string {
	return m.Deployment.ImagePullSecrets
}

func (m *MemberArangoDPod) GetPodAntiAffinity() *core.PodAntiAffinity {
	a := &core.PodAntiAffinity{}

	pod.AppendPodAntiAffinityDefault(m, a)

	a = kresources.MergePodAntiAffinity(a, topology.GetTopologyAffinityRules(m.context.GetName(), m.Status, m.Group, m.Member).PodAntiAffinity)

	a = kresources.MergePodAntiAffinity(a, m.GroupSpec.AntiAffinity)

	return kresources.OptionalPodAntiAffinity(a)
}

func (m *MemberArangoDPod) GetPodAffinity() *core.PodAffinity {
	a := &core.PodAffinity{}

	a = kresources.MergePodAffinity(a, m.GroupSpec.Affinity)

	a = kresources.MergePodAffinity(a, topology.GetTopologyAffinityRules(m.context.GetName(), m.Status, m.Group, m.Member).PodAffinity)

	return kresources.OptionalPodAffinity(a)
}

func (m *MemberArangoDPod) GetNodeAffinity() *core.NodeAffinity {
	a := &core.NodeAffinity{}

	pod.AppendArchSelector(a, m.Member.Architecture.Default(m.Deployment.Architecture.GetDefault()).AsNodeSelectorRequirement())

	a = kresources.MergeNodeAffinity(a, m.GroupSpec.NodeAffinity)

	a = kresources.MergeNodeAffinity(a, topology.GetTopologyAffinityRules(m.context.GetName(), m.Status, m.Group, m.Member).NodeAffinity)

	return kresources.OptionalNodeAffinity(a)
}

func (m *MemberArangoDPod) GetNodeSelector() map[string]string {
	return m.GroupSpec.GetNodeSelector()
}

func (m *MemberArangoDPod) GetServiceAccountName() string {
	return m.GroupSpec.GetServiceAccountName()
}

func (m *MemberArangoDPod) GetSidecars(pod *core.PodTemplateSpec) error {
	//nolint:staticcheck
	if m.Deployment.Metrics.IsEnabled() && m.Deployment.Metrics.Mode.Get() != api.MetricsModeInternal {
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

	if m.Deployment.Sidecar.IsEnabled(m.Deployment.Gateway.IsEnabled()) && m.Deployment.Mode.ServingGroup() == m.Group {
		var c *core.Container

		pod.Labels[k8sutil.LabelKeyArangoSidecar] = "yes"
		if container, err := m.createServingSidecarExporter(); err != nil {
			return err
		} else {
			c = container
		}
		if c != nil {
			pod.Spec.Containers = append(pod.Spec.Containers, *c)
		}
	}

	// A sidecar provided by the user
	sidecars := m.GroupSpec.GetSidecars()
	if len(sidecars) > 0 {
		addLifecycleSidecar(m.GroupSpec.SidecarCoreNames, sidecars)
		pod.Spec.Containers = append(pod.Spec.Containers, sidecars...)
	}

	return nil
}

func (m *MemberArangoDPod) GetVolumes() []core.Volume {
	volumes := CreateArangoDVolumes(m.Member, m.Input, m.Deployment, m.GroupSpec)

	return volumes.Volumes()
}

func (m *MemberArangoDPod) IsDeploymentMode() bool {
	return m.Deployment.IsDevelopment()
}

func (m *MemberArangoDPod) GetInitContainers(cachedStatus interfaces.Inspector) ([]core.Container, error) {
	var initContainers []core.Container

	if c := m.GroupSpec.InitContainers.GetContainers(); len(c) > 0 {
		initContainers = append(initContainers, c...)
	}

	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}

	{
		sc := k8sutil.CreateSecurityContext(m.GroupSpec.SecurityContext)
		c, err := k8sutil.InitLifecycleContainer(m.resources.context.GetOperatorImage(), executable, &m.Deployment.Lifecycle.Resources, sc)
		if err != nil {
			return nil, err
		}
		initContainers = append(initContainers, c)
	}

	{
		engine := m.Deployment.GetStorageEngine().AsArangoArgument()
		requireUUID := m.Group == api.ServerGroupDBServers && m.Member.IsInitialized

		sc := k8sutil.CreateSecurityContext(m.GroupSpec.SecurityContext)
		c := k8sutil.ArangodInitContainer(api.ServerGroupReservedInitContainerNameUUID, m.Member.ID, engine, executable,
			m.resources.context.GetOperatorImage(), requireUUID, sc)
		initContainers = append(initContainers, c)
	}

	{
		// Upgrade container - run in background
		if m.AutoUpgrade || m.Member.Upgrade {
			upgradeContainer := &ArangoUpgradeContainer{
				m.GetContainerCreator(),
				cachedStatus,
				m.Input,
			}
			c, err := k8sutil.NewContainer(upgradeContainer)
			if err != nil {
				return nil, err
			}

			initContainers = append(initContainers, c)
		}

		// VersionCheck Container
		{
			switch m.Group {
			case api.ServerGroupAgents, api.ServerGroupDBServers, api.ServerGroupSingle:
				if features.UpgradeVersionCheckV2().Enabled() {
					c := k8sutil.ArangodVersionCheckInitContainer(api.ServerGroupReservedInitContainerNameVersionCheck, executable, m.resources.context.GetOperatorImage(),
						m.Image.ArangoDBVersion, m.GroupSpec.SecurityContext.NewSecurityContext())
					initContainers = append(initContainers, c)
				} else if features.UpgradeVersionCheck().Enabled() {
					upgradeContainer := &ArangoVersionCheckContainer{
						m.GetContainerCreator(),
						cachedStatus,
						m.Input,
						pod.UpgradeVersionCheck().Args(m.Input),
					}

					c, err := k8sutil.NewContainer(upgradeContainer)
					if err != nil {
						return nil, err
					}

					initContainers = append(initContainers, c)
				}
			}
		}
	}

	res := kresources.ExtractPodInitContainerAcceptedResourceRequirement(m.GetContainerCreator().GetResourceRequirements(m.GetContainerCreator().GetResourceRequirementsDefaultScale()))

	initContainers = applyInitContainersResourceResources(initContainers, res)
	initContainers = upscaleInitContainersResourceResources(initContainers, res)

	return initContainers, nil
}

func (m *MemberArangoDPod) GetFinalizers() []string {
	return k8sutil.GetFinalizers(m.Deployment.GetServerGroupSpec(m.Group), m.Group)
}

func (m *MemberArangoDPod) GetTolerations() []core.Toleration {
	return m.resources.CreatePodTolerations(m.Group, m.GroupSpec)
}

func (m *MemberArangoDPod) GetContainerCreator() interfaces.ContainerCreator {
	return &ArangoDContainer{
		MemberArangoDPod: m,
	}
}

func (m *MemberArangoDPod) GetRestartPolicy() core.RestartPolicy {
	return getDefaultRestartPolicy(m.GroupSpec)
}

func (m *MemberArangoDPod) createMetricsExporterSidecarInternalExporter() (*core.Container, error) {
	image := m.GetContainerCreator().GetImage()

	args := createInternalExporterArgs(m.Deployment, m.Group, m.GroupSpec, m.Image.ArangoDBVersion)

	c, err := ArangodbInternalExporterContainer(image, args,
		createExporterLivenessProbe(m.Deployment.IsSecure() && m.Deployment.Metrics.IsTLS()), m.Deployment.Metrics.Resources,
		m.Deployment, m.GroupSpec)
	if err != nil {
		return nil, err
	}

	if m.Deployment.Authentication.IsAuthenticated() && m.Deployment.Metrics.GetJWTTokenSecretName() != "" {
		c.VolumeMounts = append(c.VolumeMounts, k8sutil.ExporterJWTVolumeMount())
	}

	if pod.IsTLSEnabled(m.Input) {
		c.VolumeMounts = append(c.VolumeMounts, k8sutil.TlsKeyfileVolumeMount())
	}

	return &c, nil
}

func (m *MemberArangoDPod) createServingSidecarExporter() (*core.Container, error) {
	image := m.GetContainerCreator().GetImage()

	args := createInternalSidecarArgs(m.Deployment, m.GroupSpec)

	baseResources := kresources.CleanContainerResource(
		kresources.UpscaleResourceRequirements(
			*k8sutil.CreateBasicContainerResources(),
			m.Deployment.Sidecar.Resources,
		),
	)

	c, err := ArangodbInternalSidecarContainer(image, args,
		baseResources,
		m.Deployment, m.GroupSpec)
	if err != nil {
		return nil, err
	}

	if m.Deployment.Authentication.IsAuthenticated() {
		c.VolumeMounts = append(c.VolumeMounts, k8sutil.ClusterJWTVolumeMount())
	}

	if pod.IsTLSEnabled(m.Input) {
		c.VolumeMounts = append(c.VolumeMounts, k8sutil.TlsKeyfileVolumeMount())
	}

	return &c, nil
}

func (m *MemberArangoDPod) ApplyPodSpec(p *core.PodSpec) error {
	p.SecurityContext = k8sutil.CreatePodSecurityContext(m.GroupSpec.SecurityContext)
	if s := m.GroupSpec.SchedulerName; s != nil {
		p.SchedulerName = *s
	}

	m.GroupSpec.PodModes.Apply(p)

	return nil
}

func (m *MemberArangoDPod) Annotations() map[string]string {
	// Merge deployment and group annotations and add hardcoded scrape annotation for ArangoD pods
	result := collection.MergeAnnotations(m.Deployment.Annotations, m.GroupSpec.Annotations)

	// Enable scraping via platform by default for ArangoD (requires metrics sidecar to be enabled)
	if m.Deployment.Metrics.IsEnabled() {
		if result == nil {
			result = map[string]string{}

		}
		result[utilConstants.AnnotationMetricsScrapeLabel] = "true"
		result[utilConstants.AnnotationMetricsScrapePort] = fmt.Sprintf("%d", m.GroupSpec.GetExporterPort())
	}
	return result
}

func (m *MemberArangoDPod) Profiles() (schedulerApi.ProfileTemplates, error) {
	return nil, nil
}

func (m *MemberArangoDPod) Labels() map[string]string {
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

// GetCommand returns list of arguments for the ArangoD upgrade container.
func (a *ArangoUpgradeContainer) GetCommand() ([]string, error) {
	args, err := a.ContainerCreator.GetCommand()
	if err != nil {
		return nil, err
	}

	upgradeArgs := append(
		pod.AutoUpgrade().Args(a.input).Sort().AsArgs(),
		pod.UpgradeDebug().Args(a.input).Sort().AsArgs()...,
	)

	if a.input.Group == api.ServerGroupDBServers || a.input.Group == api.ServerGroupSingle {
		if a.input.GroupSpec.UpgradeMode.Get() == api.ServerGroupUpgradeModeOptionalReplace ||
			(a.input.GroupSpec.UpgradeMode.Get() == api.ServerGroupUpgradeModeManual && a.input.GroupSpec.ManualUpgradeMode.Get() == api.ServerGroupUpgradeModeOptionalReplace) {
			upgradeArgs = append(upgradeArgs, "--database.auto-upgrade-full-compaction")
		}
	}

	return append(args, upgradeArgs...), nil
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

// GetCommand returns list of arguments for the ArangoD version check container.
func (a *ArangoVersionCheckContainer) GetCommand() ([]string, error) {
	args, err := a.ContainerCreator.GetCommand()
	if err != nil {
		return nil, err
	}

	return append(args, a.versionArgs.Sort().AsArgs()...), nil
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
