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
	ArangoDExecutor string = "/usr/sbin/arangod"
)

type MemberArangoDPod struct {
	status                      api.MemberStatus
	tlsKeyfileSecretName        string
	rocksdbEncryptionSecretName string
	clusterJWTSecretName        string
	groupSpec                   api.ServerGroupSpec
	spec                        api.DeploymentSpec
	group                       api.ServerGroup
	context                     Context
	resources                   *Resources
	imageInfo                   api.ImageInfo
}

type ArangoDContainer struct {
	resources *Resources
	groupSpec api.ServerGroupSpec
	spec      api.DeploymentSpec
	group     api.ServerGroup
	imageInfo api.ImageInfo
}

func (a *ArangoDContainer) GetExecutor() string {
	return ArangoDExecutor
}

func (a *ArangoDContainer) GetSecurityContext() *v1.SecurityContext {
	return a.groupSpec.SecurityContext.NewSecurityContext()
}

func (a *ArangoDContainer) GetProbes() (*v1.Probe, *v1.Probe, error) {
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

func (a *ArangoDContainer) GetImage() string {
	return a.imageInfo.ImageID
}

func (a *ArangoDContainer) GetEnvs() []v1.EnvVar {
	envs := make([]v1.EnvVar, 0)

	if a.spec.IsAuthenticated() {
		if !versionHasJWTSecretKeyfile(a.imageInfo.ArangoDBVersion) {
			env := k8sutil.CreateEnvSecretKeySelector(constants.EnvArangodJWTSecret,
				a.spec.Authentication.GetJWTSecretName(), constants.SecretKeyToken)

			envs = append(envs, env)
		}
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

func (a *ArangoDContainer) GetResourceRequirements() v1.ResourceRequirements {
	if a.groupSpec.GetVolumeClaimTemplate() != nil {
		return a.groupSpec.Resources
	}

	return k8sutil.ExtractPodResourceRequirement(a.groupSpec.Resources)
}

func (a *ArangoDContainer) GetLifecycle() (*v1.Lifecycle, error) {
	if a.resources.context.GetLifecycleImage() != "" {
		return k8sutil.NewLifecycle()
	}
	return nil, nil
}

func (a *ArangoDContainer) GetImagePullPolicy() v1.PullPolicy {
	return a.spec.GetImagePullPolicy()
}

func (m *MemberArangoDPod) Init(pod *v1.Pod) {
	terminationGracePeriodSeconds := int64(math.Ceil(m.group.DefaultTerminationGracePeriod().Seconds()))
	pod.Spec.TerminationGracePeriodSeconds = &terminationGracePeriodSeconds
	pod.Spec.PriorityClassName = m.groupSpec.PriorityClassName
}

func (m *MemberArangoDPod) GetImagePullSecrets() []string {
	return m.spec.ImagePullSecrets
}

func (m *MemberArangoDPod) GetAffinityRole() string {
	return ""
}

func (m *MemberArangoDPod) GetNodeSelector() map[string]string {
	return m.groupSpec.GetNodeSelector()
}

func (m *MemberArangoDPod) GetServiceAccountName() string {
	return m.groupSpec.GetServiceAccountName()
}

func (m *MemberArangoDPod) GetSidecars(pod *v1.Pod) {

	if isMetricsEnabledForGroup(m.spec, m.group) {
		image := m.spec.GetImage()
		if m.spec.Metrics.HasImage() {
			image = m.spec.Metrics.GetImage()
		}

		c := ArangodbExporterContainer(image, createExporterArgs(m.spec.IsSecure()),
			createExporterLivenessProbe(m.spec.IsSecure()), m.spec.Metrics.Resources,
			m.groupSpec.SecurityContext.NewSecurityContext())

		if m.spec.Metrics.GetJWTTokenSecretName() != "" {
			c.VolumeMounts = append(c.VolumeMounts, k8sutil.ExporterJWTVolumeMount())
		}

		if m.tlsKeyfileSecretName != "" {
			c.VolumeMounts = append(c.VolumeMounts, k8sutil.TlsKeyfileVolumeMount())
		}

		pod.Spec.Containers = append(pod.Spec.Containers, c)
		pod.Labels[k8sutil.LabelKeyArangoExporter] = "yes"
	}

	// A sidecar provided by the user
	sidecars := m.groupSpec.GetSidecars()
	if len(sidecars) > 0 {
		pod.Spec.Containers = append(pod.Spec.Containers, sidecars...)
	}

	return
}

func (m *MemberArangoDPod) GetVolumes() ([]v1.Volume, []v1.VolumeMount) {
	var volumes []v1.Volume
	var volumeMounts []v1.VolumeMount

	volumeMounts = append(volumeMounts, k8sutil.ArangodVolumeMount())

	if m.resources.context.GetLifecycleImage() != "" {
		volumeMounts = append(volumeMounts, k8sutil.LifecycleVolumeMount())
	}

	if m.status.PersistentVolumeClaimName != "" {
		vol := k8sutil.CreateVolumeWithPersitantVolumeClaim(k8sutil.ArangodVolumeName,
			m.status.PersistentVolumeClaimName)

		volumes = append(volumes, vol)
	} else {
		volumes = append(volumes, k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName))
	}

	if m.tlsKeyfileSecretName != "" {
		vol := k8sutil.CreateVolumeWithSecret(k8sutil.TlsKeyfileVolumeName, m.tlsKeyfileSecretName)
		volumes = append(volumes, vol)
		volumeMounts = append(volumeMounts, k8sutil.TlsKeyfileVolumeMount())
	}

	if m.rocksdbEncryptionSecretName != "" {
		vol := k8sutil.CreateVolumeWithSecret(k8sutil.RocksdbEncryptionVolumeName, m.rocksdbEncryptionSecretName)
		volumes = append(volumes, vol)
		volumeMounts = append(volumeMounts, k8sutil.RocksdbEncryptionVolumeMount())
	}

	if isMetricsEnabledForGroup(m.spec, m.group) {
		token := m.spec.Metrics.GetJWTTokenSecretName()
		if token != "" {
			vol := k8sutil.CreateVolumeWithSecret(k8sutil.ExporterJWTVolumeName, token)
			volumes = append(volumes, vol)
		}
	}

	if m.clusterJWTSecretName != "" {
		vol := k8sutil.CreateVolumeWithSecret(k8sutil.ClusterJWTSecretVolumeName, m.clusterJWTSecretName)
		volumes = append(volumes, vol)
		volumeMounts = append(volumeMounts, k8sutil.ClusterJWTVolumeMount())
	}

	if m.resources.context.GetLifecycleImage() != "" {
		volumes = append(volumes, k8sutil.LifecycleVolume())
	}

	return volumes, volumeMounts
}

func (m *MemberArangoDPod) IsDeploymentMode() bool {
	return m.spec.IsDevelopment()
}

func (m *MemberArangoDPod) GetInitContainers() ([]v1.Container, error) {
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

	alpineImage := m.resources.context.GetAlpineImage()
	if alpineImage != "" {
		engine := m.spec.GetStorageEngine().AsArangoArgument()
		requireUUID := m.group == api.ServerGroupDBServers && m.status.IsInitialized

		c := k8sutil.ArangodInitContainer("uuid", m.status.ID, engine, alpineImage, requireUUID,
			m.groupSpec.SecurityContext.NewSecurityContext())
		initContainers = append(initContainers, c)
	}

	return initContainers, nil
}

func (m *MemberArangoDPod) GetFinalizers() []string {
	return m.resources.CreatePodFinalizers(m.group)
}

func (m *MemberArangoDPod) GetTolerations() []v1.Toleration {
	return m.resources.CreatePodTolerations(m.group, m.groupSpec)
}

func (m *MemberArangoDPod) GetContainerCreator() k8sutil.ContainerCreator {
	return &ArangoDContainer{
		spec:      m.spec,
		group:     m.group,
		resources: m.resources,
		imageInfo: m.imageInfo,
		groupSpec: m.groupSpec,
	}
}

func isMetricsEnabledForGroup(spec api.DeploymentSpec, group api.ServerGroup) bool {
	return spec.Metrics.IsEnabled() && group.IsExportMetrics()
}
