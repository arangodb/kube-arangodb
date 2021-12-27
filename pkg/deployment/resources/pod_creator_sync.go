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
	"context"
	"math"

	"github.com/arangodb/kube-arangodb/pkg/util/globals"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/collection"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"

	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	core "k8s.io/api/core/v1"
)

const (
	ArangoSyncExecutor string = "/usr/sbin/arangosync"
)

type ArangoSyncContainer struct {
	groupSpec              api.ServerGroupSpec
	spec                   api.DeploymentSpec
	group                  api.ServerGroup
	resources              *Resources
	imageInfo              api.ImageInfo
	apiObject              meta.Object
	memberStatus           api.MemberStatus
	tlsKeyfileSecretName   string
	clientAuthCASecretName string
	masterJWTSecretName    string
	clusterJWTSecretName   string
}

var _ interfaces.PodCreator = &MemberSyncPod{}
var _ interfaces.ContainerCreator = &ArangoSyncContainer{}

type MemberSyncPod struct {
	tlsKeyfileSecretName   string
	clientAuthCASecretName string
	masterJWTSecretName    string
	clusterJWTSecretName   string
	groupSpec              api.ServerGroupSpec
	spec                   api.DeploymentSpec
	group                  api.ServerGroup
	arangoMember           api.ArangoMember
	resources              *Resources
	imageInfo              api.ImageInfo
	apiObject              meta.Object
	memberStatus           api.MemberStatus
}

func (a *ArangoSyncContainer) GetArgs() ([]string, error) {
	return createArangoSyncArgs(a.apiObject, a.spec, a.group, a.groupSpec, a.memberStatus), nil
}

func (a *ArangoSyncContainer) GetName() string {
	return k8sutil.ServerContainerName
}

func (a *ArangoSyncContainer) GetPorts() []core.ContainerPort {
	return []core.ContainerPort{
		{
			Name:          k8sutil.ServerContainerName,
			ContainerPort: int32(k8sutil.ArangoPort),
			Protocol:      core.ProtocolTCP,
		},
	}
}

func (a *ArangoSyncContainer) GetExecutor() string {
	return a.groupSpec.GetEntrypoint(ArangoSyncExecutor)
}

func (a *ArangoSyncContainer) GetSecurityContext() *core.SecurityContext {
	return a.groupSpec.SecurityContext.NewSecurityContext()
}

func (a *ArangoSyncContainer) GetProbes() (*core.Probe, *core.Probe, *core.Probe, error) {
	var liveness, readiness, startup *core.Probe

	probeLivenessConfig, err := a.resources.getLivenessProbe(a.spec, a.group, a.imageInfo.ArangoDBVersion)
	if err != nil {
		return nil, nil, nil, err
	}

	probeReadinessConfig, err := a.resources.getReadinessProbe(a.spec, a.group, a.imageInfo.ArangoDBVersion)
	if err != nil {
		return nil, nil, nil, err
	}

	probeStartupConfig, err := a.resources.getReadinessProbe(a.spec, a.group, a.imageInfo.ArangoDBVersion)
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

func (a *ArangoSyncContainer) GetResourceRequirements() core.ResourceRequirements {
	return k8sutil.ExtractPodResourceRequirement(a.groupSpec.Resources)
}

func (a *ArangoSyncContainer) GetLifecycle() (*core.Lifecycle, error) {
	return k8sutil.NewLifecycleFinalizers()
}

func (a *ArangoSyncContainer) GetImagePullPolicy() core.PullPolicy {
	return a.spec.GetImagePullPolicy()
}

func (a *ArangoSyncContainer) GetImage() string {
	return a.imageInfo.Image
}

func (a *ArangoSyncContainer) GetEnvs() []core.EnvVar {
	envs := NewEnvBuilder()

	if a.spec.Sync.Monitoring.GetTokenSecretName() != "" {
		env := k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoSyncMonitoringToken,
			a.spec.Sync.Monitoring.GetTokenSecretName(), constants.SecretKeyToken)

		envs.Add(true, env)
	}

	if a.spec.License.HasSecretName() {
		env := k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoLicenseKey, a.spec.License.GetSecretName(),
			constants.SecretKeyToken)

		envs.Add(true, env)
	}

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

	return envs.GetEnvList()
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
	return m.group.AsRole()
}

func (m *MemberSyncPod) GetImagePullSecrets() []string {
	return m.spec.ImagePullSecrets
}

func (m *MemberSyncPod) GetPodAntiAffinity() *core.PodAntiAffinity {
	a := core.PodAntiAffinity{}

	pod.AppendPodAntiAffinityDefault(m, &a)

	pod.MergePodAntiAffinity(&a, m.groupSpec.AntiAffinity)

	return pod.ReturnPodAntiAffinityOrNil(a)
}

func (m *MemberSyncPod) GetPodAffinity() *core.PodAffinity {
	a := core.PodAffinity{}

	if m.group == api.ServerGroupSyncWorkers {
		pod.AppendAffinityWithRole(m, &a, api.ServerGroupDBServers.AsRole())
	}

	pod.MergePodAffinity(&a, m.groupSpec.Affinity)

	return pod.ReturnPodAffinityOrNil(a)
}

func (m *MemberSyncPod) GetNodeAffinity() *core.NodeAffinity {
	a := core.NodeAffinity{}

	pod.AppendNodeSelector(&a)

	pod.MergeNodeAffinity(&a, m.groupSpec.NodeAffinity)

	return pod.ReturnNodeAffinityOrNil(a)
}

func (m *MemberSyncPod) GetNodeSelector() map[string]string {
	return m.groupSpec.GetNodeSelector()
}

func (m *MemberSyncPod) GetServiceAccountName() string {
	return m.groupSpec.GetServiceAccountName()
}

func (m *MemberSyncPod) GetSidecars(pod *core.Pod) error {
	// A sidecar provided by the user
	sidecars := m.groupSpec.GetSidecars()
	if len(sidecars) > 0 {
		addLifecycleSidecar(m.groupSpec.SidecarCoreNames, sidecars)
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
	return m.spec.IsDevelopment()
}

func (m *MemberSyncPod) GetInitContainers(cachedStatus interfaces.Inspector) ([]core.Container, error) {
	var initContainers []core.Container

	if c := m.groupSpec.InitContainers.GetContainers(); len(c) > 0 {
		initContainers = append(initContainers, c...)
	}

	{
		c, err := k8sutil.InitLifecycleContainer(m.resources.context.GetOperatorImage(), &m.spec.Lifecycle.Resources,
			m.groupSpec.SecurityContext.NewSecurityContext())
		if err != nil {
			return nil, err
		}
		initContainers = append(initContainers, c)
	}

	return initContainers, nil
}

func (m *MemberSyncPod) GetFinalizers() []string {
	return nil
}

func (m *MemberSyncPod) GetTolerations() []core.Toleration {
	return m.resources.CreatePodTolerations(m.group, m.groupSpec)
}

func (m *MemberSyncPod) GetContainerCreator() interfaces.ContainerCreator {
	return &ArangoSyncContainer{
		groupSpec:              m.groupSpec,
		spec:                   m.spec,
		group:                  m.group,
		resources:              m.resources,
		imageInfo:              m.imageInfo,
		apiObject:              m.apiObject,
		memberStatus:           m.memberStatus,
		tlsKeyfileSecretName:   m.tlsKeyfileSecretName,
		clientAuthCASecretName: m.clientAuthCASecretName,
		masterJWTSecretName:    m.masterJWTSecretName,
		clusterJWTSecretName:   m.clusterJWTSecretName,
	}
}

// Init initializes the arangosync pod.
func (m *MemberSyncPod) Init(ctx context.Context, cachedStatus interfaces.Inspector, pod *core.Pod) error {
	terminationGracePeriodSeconds := int64(math.Ceil(m.group.DefaultTerminationGracePeriod().Seconds()))
	pod.Spec.TerminationGracePeriodSeconds = &terminationGracePeriodSeconds
	pod.Spec.PriorityClassName = m.groupSpec.PriorityClassName

	m.masterJWTSecretName = m.spec.Sync.Authentication.GetJWTSecretName()
	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return k8sutil.ValidateTokenSecret(ctxChild, cachedStatus.SecretReadInterface(), m.masterJWTSecretName)
	})
	if err != nil {
		return errors.Wrapf(err, "Master JWT secret validation failed")
	}

	monitoringTokenSecretName := m.spec.Sync.Monitoring.GetTokenSecretName()
	err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return k8sutil.ValidateTokenSecret(ctxChild, cachedStatus.SecretReadInterface(), monitoringTokenSecretName)
	})
	if err != nil {
		return errors.Wrapf(err, "Monitoring token secret validation failed")
	}

	if m.group == api.ServerGroupSyncMasters {
		// Create TLS secret
		m.tlsKeyfileSecretName = k8sutil.CreateTLSKeyfileSecretName(m.apiObject.GetName(), m.group.AsRole(), m.memberStatus.ID)
		// Check cluster JWT secret
		if m.spec.IsAuthenticated() {
			m.clusterJWTSecretName = m.spec.Authentication.GetJWTSecretName()
			err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				return k8sutil.ValidateTokenSecret(ctxChild, cachedStatus.SecretReadInterface(), m.clusterJWTSecretName)
			})
			if err != nil {
				return errors.Wrapf(err, "Cluster JWT secret validation failed")
			}
		}
		// Check client-auth CA certificate secret
		m.clientAuthCASecretName = m.spec.Sync.Authentication.GetClientCASecretName()
		err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return k8sutil.ValidateCACertificateSecret(ctxChild, cachedStatus.SecretReadInterface(), m.clientAuthCASecretName)
		})
		if err != nil {
			return errors.Wrapf(err, "Client authentication CA certificate secret validation failed")
		}
	}

	return nil
}

func (m *MemberSyncPod) Validate(_ interfaces.Inspector) error {
	if err := validateSidecars(m.groupSpec.SidecarCoreNames, m.groupSpec.GetSidecars()); err != nil {
		return err
	}

	return nil
}

func (m *MemberSyncPod) ApplyPodSpec(spec *core.PodSpec) error {
	if s := m.groupSpec.SchedulerName; s != nil {
		spec.SchedulerName = *s
	}

	return nil
}

func (m *MemberSyncPod) Annotations() map[string]string {
	return collection.MergeAnnotations(m.spec.Annotations, m.groupSpec.Annotations)
}

func (m *MemberSyncPod) Labels() map[string]string {
	return collection.ReservedLabels().Filter(collection.MergeAnnotations(m.spec.Labels, m.groupSpec.Labels))
}

func createArangoSyncVolumes(tlsKeyfileSecretName, clientAuthCASecretName, masterJWTSecretName,
	clusterJWTSecretName string) pod.Volumes {
	volumes := pod.NewVolumes()

	volumes.AddVolume(k8sutil.LifecycleVolume())
	volumes.AddVolumeMount(k8sutil.LifecycleVolumeMount())

	if tlsKeyfileSecretName != "" {
		vol := k8sutil.CreateVolumeWithSecret(k8sutil.TlsKeyfileVolumeName, tlsKeyfileSecretName)
		volumes.AddVolume(vol)
		volumes.AddVolumeMount(k8sutil.TlsKeyfileVolumeMount())
	}

	if clientAuthCASecretName != "" {
		vol := k8sutil.CreateVolumeWithSecret(k8sutil.ClientAuthCAVolumeName, clientAuthCASecretName)
		volumes.AddVolume(vol)
		volumes.AddVolumeMount(k8sutil.ClientAuthCACertificateVolumeMount())
	}

	if masterJWTSecretName != "" {
		vol := k8sutil.CreateVolumeWithSecret(k8sutil.MasterJWTSecretVolumeName, masterJWTSecretName)
		volumes.AddVolume(vol)
		volumes.AddVolumeMount(k8sutil.MasterJWTVolumeMount())
	}

	if clusterJWTSecretName != "" {
		vol := k8sutil.CreateVolumeWithSecret(k8sutil.ClusterJWTSecretVolumeName, clusterJWTSecretName)
		volumes.AddVolume(vol)
		volumes.AddVolumeMount(k8sutil.ClusterJWTVolumeMount())
	}

	return volumes
}
