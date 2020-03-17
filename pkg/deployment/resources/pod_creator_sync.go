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

	"github.com/arangodb/kube-arangodb/pkg/util/constants"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	v1 "k8s.io/api/core/v1"
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

func (a *ArangoSyncContainer) GetExecutor() string {
	return ArangoSyncExecutor
}

func (a *ArangoSyncContainer) GetSecurityContext() *v1.SecurityContext {
	return a.groupSpec.SecurityContext.NewSecurityContext()
}

func (a *ArangoSyncContainer) GetProbes() (*v1.Probe, *v1.Probe, error) {
	var liveness, readiness *v1.Probe

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

func (a *ArangoSyncContainer) GetResourceRequirements() v1.ResourceRequirements {
	return k8sutil.ExtractPodResourceRequirement(a.groupSpec.Resources)
}

func (a *ArangoSyncContainer) GetLifecycle() (*v1.Lifecycle, error) {
	if a.resources.context.GetLifecycleImage() != "" {
		return k8sutil.NewLifecycle()
	}
	return nil, nil
}

func (a *ArangoSyncContainer) GetImagePullPolicy() v1.PullPolicy {
	return a.spec.GetImagePullPolicy()
}

func (a *ArangoSyncContainer) GetImage() string {
	return a.imageInfo.Image
}

func (a *ArangoSyncContainer) GetEnvs() []v1.EnvVar {
	envs := make([]v1.EnvVar, 0)

	if a.spec.Sync.Monitoring.GetTokenSecretName() != "" {
		env := k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoSyncMonitoringToken,
			a.spec.Sync.Monitoring.GetTokenSecretName(), constants.SecretKeyToken)

		envs = append(envs, env)
	}

	if a.spec.License.HasSecretName() {
		env := k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoLicenseKey, a.spec.License.GetSecretName(),
			constants.SecretKeyToken)

		envs = append(envs, env)
	}

	if a.resources.context.GetLifecycleImage() != "" {
		envs = append(envs, k8sutil.GetLifecycleEnv()...)
	}

	if len(envs) > 0 {
		return envs
	}

	return nil
}

func (m *MemberSyncPod) GetAffinityRole() string {
	if m.group == api.ServerGroupSyncWorkers {
		return api.ServerGroupDBServers.AsRole()
	}
	return ""
}

func (m *MemberSyncPod) GetImagePullSecrets() []string {
	return m.spec.ImagePullSecrets
}

func (m *MemberSyncPod) GetNodeSelector() map[string]string {
	return m.groupSpec.GetNodeSelector()
}

func (m *MemberSyncPod) GetServiceAccountName() string {
	return m.groupSpec.GetServiceAccountName()
}

func (m *MemberSyncPod) GetSidecars(pod *v1.Pod) {
	// A sidecar provided by the user
	sidecars := m.groupSpec.GetSidecars()
	if len(sidecars) > 0 {
		pod.Spec.Containers = append(pod.Spec.Containers, sidecars...)
	}
}

func (m *MemberSyncPod) GetVolumes() ([]v1.Volume, []v1.VolumeMount) {
	var volumes []v1.Volume
	var volumeMounts []v1.VolumeMount

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

func (m *MemberSyncPod) GetInitContainers() ([]v1.Container, error) {
	var initContainers []v1.Container

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

func (m *MemberSyncPod) GetTolerations() []v1.Toleration {
	return m.resources.CreatePodTolerations(m.group, m.groupSpec)
}

func (m *MemberSyncPod) GetContainerCreator() k8sutil.ContainerCreator {
	return &ArangoSyncContainer{
		groupSpec: m.groupSpec,
		spec:      m.spec,
		group:     m.group,
		resources: m.resources,
		imageInfo: m.imageInfo,
	}
}

func (m *MemberSyncPod) Init(pod *v1.Pod) {
	terminationGracePeriodSeconds := int64(math.Ceil(m.group.DefaultTerminationGracePeriod().Seconds()))
	pod.Spec.TerminationGracePeriodSeconds = &terminationGracePeriodSeconds
	pod.Spec.PriorityClassName = m.groupSpec.PriorityClassName
}
