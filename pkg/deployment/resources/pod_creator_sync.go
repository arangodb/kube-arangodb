//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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
	"net"
	"net/url"
	"os"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/handlers/utils"
	"github.com/arangodb/kube-arangodb/pkg/util/collection"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

const (
	ArangoSyncExecutor string = "/usr/sbin/arangosync"
)

type ArangoSyncContainer struct {
	*MemberSyncPod
}

var _ interfaces.PodCreator = &MemberSyncPod{}
var _ interfaces.ContainerCreator = &ArangoSyncContainer{}

type MemberSyncPod struct {
	pod.Input

	podName                string
	tlsKeyfileSecretName   string
	clientAuthCASecretName string
	masterJWTSecretName    string
	clusterJWTSecretName   string

	cachedStatus interfaces.Inspector

	resources *Resources
}

func (a *ArangoSyncContainer) GetCommand() ([]string, error) {
	cmd := make([]string, 0, 128)
	cmd = append(cmd, a.GetExecutor())
	cmd = append(cmd, createArangoSyncArgs(a.Input)...)
	return cmd, nil
}

func (a *ArangoSyncContainer) GetName() string {
	return shared.ServerContainerName
}

func (a *ArangoSyncContainer) GetPorts() []core.ContainerPort {
	port := shared.ArangoSyncMasterPort

	if a.Group == api.ServerGroupSyncWorkers {
		port = shared.ArangoSyncWorkerPort
	}

	return []core.ContainerPort{
		{
			Name:          shared.ServerContainerName,
			ContainerPort: int32(port),
			Protocol:      core.ProtocolTCP,
		},
	}
}

func (a *ArangoSyncContainer) GetExecutor() string {
	return a.GroupSpec.GetEntrypoint(ArangoSyncExecutor)
}

func (a *ArangoSyncContainer) GetSecurityContext() *core.SecurityContext {
	return k8sutil.CreateSecurityContext(a.GroupSpec.SecurityContext)
}

func (a *ArangoSyncContainer) GetProbes() (*core.Probe, *core.Probe, *core.Probe, error) {
	var liveness, readiness, startup *core.Probe

	probeLivenessConfig, err := a.resources.getLivenessProbe(a.Deployment, a.Group, a.Image)
	if err != nil {
		return nil, nil, nil, err
	}

	probeReadinessConfig, err := a.resources.getReadinessProbe(a.Deployment, a.Group, a.Image)
	if err != nil {
		return nil, nil, nil, err
	}

	probeStartupConfig, err := a.resources.getReadinessProbe(a.Deployment, a.Group, a.Image)
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

func (a *ArangoSyncContainer) GetResourceRequirements(scale float64) core.ResourceRequirements {
	return kresources.ScaleResources(kresources.ExtractPodAcceptedResourceRequirement(a.ArangoMember.Spec.Overrides.GetResources(&a.GroupSpec)), scale)
}

func (a *ArangoSyncContainer) GetResourceRequirementsDefaultScale() float64 {
	return 0.75
}

func (a *ArangoSyncContainer) GetLifecycle() (*core.Lifecycle, error) {
	return k8sutil.NewLifecycleFinalizers()
}

func (a *ArangoSyncContainer) GetImagePullPolicy() core.PullPolicy {
	return a.Deployment.GetImagePullPolicy()
}

func (a *ArangoSyncContainer) GetImage() string {
	return a.Image.Image
}

func (a *ArangoSyncContainer) GetEnvs() ([]core.EnvVar, []core.EnvFromSource) {
	envs := NewEnvBuilder()

	if a.Deployment.Sync.Monitoring.GetTokenSecretName() != "" {
		env := k8sutil.CreateEnvSecretKeySelector(utilConstants.EnvArangoSyncMonitoringToken,
			a.Deployment.Sync.Monitoring.GetTokenSecretName(), utilConstants.SecretKeyToken)

		envs.Add(true, env)
	}

	if a.Deployment.License.HasSecretName() {
		env := k8sutil.CreateEnvSecretKeySelector(utilConstants.EnvArangoLicenseKey, a.Deployment.License.GetSecretName(),
			utilConstants.SecretKeyToken)

		envs.Add(true, env)
	}

	envs.Add(true, k8sutil.GetLifecycleEnv()...)

	if p := a.GroupSpec.Port; p != nil {
		envs.Add(true, core.EnvVar{
			Name:  ArangoDBServerPortEnv,
			Value: fmt.Sprintf("%d", *p),
		})
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

	return envs.GetEnvList(), nil
}

func (a *ArangoSyncContainer) GetVolumeMounts() []core.VolumeMount {
	volumes := createArangoSyncVolumes(a.tlsKeyfileSecretName, a.clientAuthCASecretName, a.masterJWTSecretName,
		a.clusterJWTSecretName)

	return volumes.VolumeMounts()
}

func (m *MemberSyncPod) GetName() string {
	return m.resources.context.GetAPIObject().GetName()
}

func (m *MemberSyncPod) GetRole() string {
	return m.Group.AsRole()
}

func (m *MemberSyncPod) GetImagePullSecrets() []string {
	return m.Deployment.ImagePullSecrets
}

func (m *MemberSyncPod) GetPodAntiAffinity() *core.PodAntiAffinity {
	a := &core.PodAntiAffinity{}

	pod.AppendPodAntiAffinityDefault(m, a)

	a = kresources.MergePodAntiAffinity(a, m.GroupSpec.AntiAffinity)

	return kresources.OptionalPodAntiAffinity(a)
}

func (m *MemberSyncPod) GetPodAffinity() *core.PodAffinity {
	a := &core.PodAffinity{}

	if m.Group == api.ServerGroupSyncWorkers {
		pod.AppendAffinityWithRole(m, a, api.ServerGroupDBServers.AsRole())
	}

	a = kresources.MergePodAffinity(a, m.GroupSpec.Affinity)

	return kresources.OptionalPodAffinity(a)
}

func (m *MemberSyncPod) GetNodeAffinity() *core.NodeAffinity {
	a := &core.NodeAffinity{}

	pod.AppendArchSelector(a, m.Member.Architecture.Default(m.Deployment.Architecture.GetDefault()).AsNodeSelectorRequirement())

	a = kresources.MergeNodeAffinity(a, m.GroupSpec.NodeAffinity)

	return kresources.OptionalNodeAffinity(a)
}

func (m *MemberSyncPod) GetNodeSelector() map[string]string {
	return m.GroupSpec.GetNodeSelector()
}

func (m *MemberSyncPod) GetServiceAccountName() string {
	return m.GroupSpec.GetServiceAccountName()
}

func (m *MemberSyncPod) GetSidecars(pod *core.PodTemplateSpec) error {
	// A sidecar provided by the user
	sidecars := m.GroupSpec.GetSidecars()
	if len(sidecars) > 0 {
		addLifecycleSidecar(m.GroupSpec.SidecarCoreNames, sidecars)
		pod.Spec.Containers = append(pod.Spec.Containers, sidecars...)
	}

	return nil
}

// GetVolumes returns volumes for the ArangoSync container.
func (m *MemberSyncPod) GetVolumes() []core.Volume {
	volumes := createArangoSyncVolumes(m.tlsKeyfileSecretName, m.clientAuthCASecretName, m.masterJWTSecretName,
		m.clusterJWTSecretName)

	return volumes.Volumes()
}

func (m *MemberSyncPod) IsDeploymentMode() bool {
	return m.Deployment.IsDevelopment()
}

func (m *MemberSyncPod) GetInitContainers(cachedStatus interfaces.Inspector) ([]core.Container, error) {
	binaryPath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	var initContainers []core.Container
	if c := m.GroupSpec.InitContainers.GetContainers(); len(c) > 0 {
		initContainers = append(initContainers, c...)
	}

	{
		sc := k8sutil.CreateSecurityContext(m.GroupSpec.SecurityContext)
		c, err := k8sutil.InitLifecycleContainer(m.resources.context.GetOperatorImage(), binaryPath, &m.Deployment.Lifecycle.Resources, sc)
		if err != nil {
			return nil, err
		}
		initContainers = append(initContainers, c)
	}

	res := kresources.ExtractPodInitContainerAcceptedResourceRequirement(m.GetContainerCreator().GetResourceRequirements(m.GetContainerCreator().GetResourceRequirementsDefaultScale()))

	initContainers = applyInitContainersResourceResources(initContainers, res)
	initContainers = upscaleInitContainersResourceResources(initContainers, res)

	return initContainers, nil
}

func (m *MemberSyncPod) GetFinalizers() []string {
	return nil
}

func (m *MemberSyncPod) GetTolerations() []core.Toleration {
	return m.resources.CreatePodTolerations(m.Group, m.GroupSpec)
}

func (m *MemberSyncPod) GetContainerCreator() interfaces.ContainerCreator {
	return &ArangoSyncContainer{
		MemberSyncPod: m,
	}
}

func (m *MemberSyncPod) GetRestartPolicy() core.RestartPolicy {
	if features.RestartPolicyAlways().Enabled() {
		return core.RestartPolicyAlways
	}
	return core.RestartPolicyNever
}

// Init initializes the arangosync pod.
func (m *MemberSyncPod) Init(ctx context.Context, cachedStatus interfaces.Inspector, pod *core.PodTemplateSpec) error {
	terminationGracePeriodSeconds := int64(math.Ceil(m.GroupSpec.GetTerminationGracePeriod(m.Group).Seconds()))
	pod.Spec.TerminationGracePeriodSeconds = &terminationGracePeriodSeconds
	pod.Spec.PriorityClassName = m.GroupSpec.PriorityClassName

	if alias := m.syncHostAlias(); alias != nil {
		pod.Spec.HostAliases = append(pod.Spec.HostAliases, *alias)
	}

	m.masterJWTSecretName = m.Deployment.Sync.Authentication.GetJWTSecretName()
	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return k8sutil.ValidateTokenSecret(ctxChild, cachedStatus.Secret().V1().Read(), m.masterJWTSecretName)
	})
	if err != nil {
		return errors.Wrapf(err, "Master JWT secret validation failed")
	}

	monitoringTokenSecretName := m.Deployment.Sync.Monitoring.GetTokenSecretName()
	err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return k8sutil.ValidateTokenSecret(ctxChild, cachedStatus.Secret().V1().Read(), monitoringTokenSecretName)
	})
	if err != nil {
		return errors.Wrapf(err, "Monitoring token secret validation failed")
	}

	if m.Group == api.ServerGroupSyncMasters {
		// Create TLS secret
		m.tlsKeyfileSecretName = k8sutil.CreateTLSKeyfileSecretName(m.ApiObject.GetName(), m.Group.AsRole(), m.Member.ID)
		// Check cluster JWT secret
		if m.Deployment.IsAuthenticated() {
			m.clusterJWTSecretName = m.Deployment.Authentication.GetJWTSecretName()
			err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				return k8sutil.ValidateTokenSecret(ctxChild, cachedStatus.Secret().V1().Read(), m.clusterJWTSecretName)
			})
			if err != nil {
				return errors.Wrapf(err, "Cluster JWT secret validation failed")
			}
		}
		// Check client-auth CA certificate secret
		m.clientAuthCASecretName = m.Deployment.Sync.Authentication.GetClientCASecretName()
		err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return k8sutil.ValidateCACertificateSecret(ctxChild, cachedStatus.Secret().V1().Read(), m.clientAuthCASecretName)
		})
		if err != nil {
			return errors.Wrapf(err, "Client authentication CA certificate secret validation failed")
		}
	}

	return nil
}

func (m *MemberSyncPod) Validate(_ interfaces.Inspector) error {
	if err := validateSidecars(m.GroupSpec.SidecarCoreNames, m.GroupSpec.GetSidecars()); err != nil {
		return err
	}

	return nil
}

func (m *MemberSyncPod) ApplyPodSpec(spec *core.PodSpec) error {
	if s := m.GroupSpec.SchedulerName; s != nil {
		spec.SchedulerName = *s
	}

	m.GroupSpec.PodModes.Apply(spec)

	return nil
}

func (m *MemberSyncPod) Annotations() map[string]string {
	return collection.MergeAnnotations(m.Deployment.Annotations, m.GroupSpec.Annotations)
}

func (m *MemberSyncPod) Labels() map[string]string {
	return collection.ReservedLabels().Filter(collection.MergeAnnotations(m.Deployment.Labels, m.GroupSpec.Labels))
}

func createArangoSyncVolumes(tlsKeyfileSecretName, clientAuthCASecretName, masterJWTSecretName,
	clusterJWTSecretName string) pod.Volumes {
	volumes := pod.NewVolumes()

	volumes.AddVolume(k8sutil.LifecycleVolume())
	volumes.AddVolumeMount(k8sutil.LifecycleVolumeMount())

	if tlsKeyfileSecretName != "" {
		vol := k8sutil.CreateVolumeWithSecret(shared.TlsKeyfileVolumeName, tlsKeyfileSecretName)
		volumes.AddVolume(vol)
		volumes.AddVolumeMount(k8sutil.TlsKeyfileVolumeMount())
	}

	if clientAuthCASecretName != "" {
		vol := k8sutil.CreateVolumeWithSecret(shared.ClientAuthCAVolumeName, clientAuthCASecretName)
		volumes.AddVolume(vol)
		volumes.AddVolumeMount(k8sutil.ClientAuthCACertificateVolumeMount())
	}

	if masterJWTSecretName != "" {
		vol := k8sutil.CreateVolumeWithSecret(shared.MasterJWTSecretVolumeName, masterJWTSecretName)
		volumes.AddVolume(vol)
		volumes.AddVolumeMount(k8sutil.MasterJWTVolumeMount())
	}

	if clusterJWTSecretName != "" {
		vol := k8sutil.CreateVolumeWithSecret(shared.ClusterJWTSecretVolumeName, clusterJWTSecretName)
		volumes.AddVolume(vol)
		volumes.AddVolumeMount(k8sutil.ClusterJWTVolumeMount())
	}

	return volumes
}

func (m *MemberSyncPod) syncHostAlias() *core.HostAlias {
	svcName := k8sutil.CreateSyncMasterClientServiceName(m.ApiObject.GetName())
	svc, ok := m.cachedStatus.Service().V1().GetSimple(svcName)
	if !ok {
		return nil
	}

	endpoint := k8sutil.CreateSyncMasterClientServiceDNSNameWithDomain(m.ApiObject, m.Deployment.ClusterDomain)

	if svc.Spec.ClusterIP == "" || svc.Spec.ClusterIP == core.ClusterIPNone {
		return nil
	}

	var alias core.HostAlias

	alias.IP = svc.Spec.ClusterIP

	var aliases utils.StringList

	for _, u := range m.Deployment.Sync.ExternalAccess.ResolveMasterEndpoint(svcName, shared.ArangoSyncMasterPort) {
		url, err := url.Parse(u)
		if err != nil {
			continue
		}

		host := url.Host

		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}

		if host == endpoint {
			continue
		}

		if host == svcName {
			continue
		}

		if ip := net.ParseIP(host); ip != nil {
			continue
		}

		aliases = append(aliases, host)
	}

	if len(aliases) == 0 {
		return nil
	}

	aliases = aliases.Sort().Unique()

	alias.Hostnames = aliases

	return &alias
}

func (m *MemberSyncPod) Profiles() (schedulerApi.ProfileTemplates, error) {
	return nil, nil
}
