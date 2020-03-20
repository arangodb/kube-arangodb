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
	"fmt"
	"math"

	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"

	"github.com/arangodb/kube-arangodb/pkg/util"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	core "k8s.io/api/core/v1"
)

const (
	ArangoDExecutor                        string = "/usr/sbin/arangod"
	ArangoDBOverrideDetectedTotalMemoryEnv        = "ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY"
)

var _ k8sutil.PodCreator = &MemberArangoDPod{}

type MemberArangoDPod struct {
	status                      api.MemberStatus
	tlsKeyfileSecretName        string
	rocksdbEncryptionSecretName string
	clusterJWTSecretName        string
	groupSpec                   api.ServerGroupSpec
	spec                        api.DeploymentSpec
	group                       api.ServerGroup
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

	if a.spec.IsAuthenticated() {
		if !versionHasJWTSecretKeyfile(a.imageInfo.ArangoDBVersion) {
			env := k8sutil.CreateEnvSecretKeySelector(constants.EnvArangodJWTSecret,
				a.spec.Authentication.GetJWTSecretName(), constants.SecretKeyToken)

			envs.Add(true, env)
		}
	}

	if a.spec.License.HasSecretName() {
		env := k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoLicenseKey, a.spec.License.GetSecretName(),
			constants.SecretKeyToken)

		envs.Add(true, env)
	}

	if a.resources.context.GetLifecycleImage() != "" {
		envs.Add(true, k8sutil.GetLifecycleEnv()...)
	}

	if util.BoolOrDefault(a.groupSpec.OverrideDetectedTotalMemory, false) {
		if a.groupSpec.Resources.Limits != nil {
			if limits, ok := a.groupSpec.Resources.Limits[core.ResourceMemory]; ok {
				envs.Add(true, core.EnvVar{
					Name:  ArangoDBOverrideDetectedTotalMemoryEnv,
					Value: fmt.Sprintf("%d", limits.Value()),
				})
			}
		}
	}

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

func (m *MemberArangoDPod) Init(pod *core.Pod) {
	terminationGracePeriodSeconds := int64(math.Ceil(m.group.DefaultTerminationGracePeriod().Seconds()))
	pod.Spec.TerminationGracePeriodSeconds = &terminationGracePeriodSeconds
	pod.Spec.PriorityClassName = m.groupSpec.PriorityClassName
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

	pod.MergePodAntiAffinity(&a, m.groupSpec.AntiAffinity)

	return pod.ReturnPodAntiAffinityOrNil(a)
}

func (m *MemberArangoDPod) GetPodAffinity() *core.PodAffinity {
	return nil
}

func (m *MemberArangoDPod) GetNodeAffinity() *core.NodeAffinity {
	a := core.NodeAffinity{}

	pod.AppendNodeSelector(&a)

	return pod.ReturnNodeAffinityOrNil(a)
}

func (m *MemberArangoDPod) GetNodeSelector() map[string]string {
	return m.groupSpec.GetNodeSelector()
}

func (m *MemberArangoDPod) GetServiceAccountName() string {
	return m.groupSpec.GetServiceAccountName()
}

func (m *MemberArangoDPod) GetSidecars(pod *core.Pod) {

	if isMetricsEnabledForGroup(m.spec, m.group) {
		image := m.context.GetMetricsExporterImage()
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

func (m *MemberArangoDPod) GetVolumes() ([]core.Volume, []core.VolumeMount) {
	var volumes []core.Volume
	var volumeMounts []core.VolumeMount

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

	if len(m.groupSpec.Volumes) > 0 {
		volumes = append(volumes, m.groupSpec.Volumes.Volumes()...)
	}

	if len(m.groupSpec.VolumeMounts) > 0 {
		volumeMounts = append(volumeMounts, m.groupSpec.VolumeMounts.VolumeMounts()...)
	}

	return volumes, volumeMounts
}

func (m *MemberArangoDPod) IsDeploymentMode() bool {
	return m.spec.IsDevelopment()
}

func (m *MemberArangoDPod) GetInitContainers() ([]core.Container, error) {
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

func (m *MemberArangoDPod) GetTolerations() []core.Toleration {
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
