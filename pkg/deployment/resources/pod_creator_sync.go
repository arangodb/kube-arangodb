//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
	"math"

	"github.com/arangodb/kube-arangodb/pkg/util/collection"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
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
	groupSpec api.ServerGroupSpec
	spec      api.DeploymentSpec
	group     api.ServerGroup
	resources *Resources
	imageInfo api.ImageInfo
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
	resources              *Resources
	imageInfo              api.ImageInfo
}

func (a *ArangoSyncContainer) GetPorts() []core.ContainerPort {
	return []core.ContainerPort{
		{
			Name:          "server",
			ContainerPort: int32(k8sutil.ArangoPort),
			Protocol:      core.ProtocolTCP,
		},
	}
}

func (a *ArangoSyncContainer) GetExecutor() string {
	return ArangoSyncExecutor
}

func (a *ArangoSyncContainer) GetSecurityContext() *core.SecurityContext {
	return a.groupSpec.SecurityContext.NewSecurityContext()
}

func (a *ArangoSyncContainer) GetProbes() (*core.Probe, *core.Probe, error) {
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

func (a *ArangoSyncContainer) GetResourceRequirements() core.ResourceRequirements {
	return k8sutil.ExtractPodResourceRequirement(a.groupSpec.Resources)
}

func (a *ArangoSyncContainer) GetLifecycle() (*core.Lifecycle, error) {
	if a.resources.context.GetLifecycleImage() != "" {
		return k8sutil.NewLifecycle()
	}
	return nil, nil
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

	if a.resources.context.GetLifecycleImage() != "" {
		envs.Add(true, k8sutil.GetLifecycleEnv()...)
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

	return envs.GetEnvList()
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

func (m *MemberSyncPod) GetSidecars(pod *core.Pod) {
	// A sidecar provided by the user
	sidecars := m.groupSpec.GetSidecars()
	if len(sidecars) > 0 {
		pod.Spec.Containers = append(pod.Spec.Containers, sidecars...)
	}
}

func (m *MemberSyncPod) GetVolumes() ([]core.Volume, []core.VolumeMount) {
	var volumes []core.Volume
	var volumeMounts []core.VolumeMount

	if m.resources.context.GetLifecycleImage() != "" {
		volumes = append(volumes, k8sutil.LifecycleVolume())
		volumeMounts = append(volumeMounts, k8sutil.LifecycleVolumeMount())
	}

	if m.tlsKeyfileSecretName != "" {
		vol := k8sutil.CreateVolumeWithSecret(k8sutil.TlsKeyfileVolumeName, m.tlsKeyfileSecretName)
		volumes = append(volumes, vol)
		volumeMounts = append(volumeMounts, k8sutil.TlsKeyfileVolumeMount())
	}

	// Client Authentication certificate secret mount (if any)
	if m.clientAuthCASecretName != "" {
		vol := k8sutil.CreateVolumeWithSecret(k8sutil.ClientAuthCAVolumeName, m.clientAuthCASecretName)
		volumes = append(volumes, vol)
		volumeMounts = append(volumeMounts, k8sutil.ClientAuthCACertificateVolumeMount())
	}

	// Master JWT secret mount (if any)
	if m.masterJWTSecretName != "" {
		vol := k8sutil.CreateVolumeWithSecret(k8sutil.MasterJWTSecretVolumeName, m.masterJWTSecretName)
		volumes = append(volumes, vol)
		volumeMounts = append(volumeMounts, k8sutil.MasterJWTVolumeMount())
	}

	// Cluster JWT secret mount (if any)
	if m.clusterJWTSecretName != "" {
		vol := k8sutil.CreateVolumeWithSecret(k8sutil.ClusterJWTSecretVolumeName, m.clusterJWTSecretName)
		volumes = append(volumes, vol)
		volumeMounts = append(volumeMounts, k8sutil.ClusterJWTVolumeMount())
	}

	return volumes, volumeMounts
}

func (m *MemberSyncPod) IsDeploymentMode() bool {
	return m.spec.IsDevelopment()
}

func (m *MemberSyncPod) GetInitContainers() ([]core.Container, error) {
	var initContainers []core.Container

	lifecycleImage := m.resources.context.GetLifecycleImage()
	if lifecycleImage != "" {
		c, err := k8sutil.InitLifecycleContainer(lifecycleImage, &m.spec.Lifecycle.Resources,
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
		groupSpec: m.groupSpec,
		spec:      m.spec,
		group:     m.group,
		resources: m.resources,
		imageInfo: m.imageInfo,
	}
}

func (m *MemberSyncPod) Init(pod *core.Pod) {
	terminationGracePeriodSeconds := int64(math.Ceil(m.group.DefaultTerminationGracePeriod().Seconds()))
	pod.Spec.TerminationGracePeriodSeconds = &terminationGracePeriodSeconds
	pod.Spec.PriorityClassName = m.groupSpec.PriorityClassName
}

func (m *MemberSyncPod) Validate(cachedStatus inspector.Inspector) error {
	return nil
}

func (m *MemberSyncPod) ApplyPodSpec(spec *core.PodSpec) error {
	return nil
}

func (m *MemberSyncPod) Annotations() map[string]string {
	return collection.MergeAnnotations(m.spec.Annotations, m.groupSpec.Annotations)
}

func (m *MemberSyncPod) Labels() map[string]string {
	return collection.ReservedLabels().Filter(collection.MergeAnnotations(m.spec.Labels, m.groupSpec.Labels))
}
