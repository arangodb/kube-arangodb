//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

var _ interfaces.ContainerCreator = &ArangoGatewayContainer{}

type ArangoGatewayContainer struct {
	member       *MemberGatewayPod
	resources    *Resources
	groupSpec    api.ServerGroupSpec
	spec         api.DeploymentSpec
	group        api.ServerGroup
	arangoMember api.ArangoMember
	imageInfo    api.ImageInfo
	cachedStatus interfaces.Inspector
	input        pod.Input
	status       api.MemberStatus
}

func (a *ArangoGatewayContainer) GetCommand() ([]string, error) {
	cmd := make([]string, 0, 128)
	cmd = append(cmd, a.GetExecutor())
	cmd = append(cmd, createArangoGatewayArgs(a.input)...)
	return cmd, nil
}

func (a *ArangoGatewayContainer) GetName() string {
	return shared.ServerContainerName
}

func (a *ArangoGatewayContainer) GetPorts() []core.ContainerPort {
	port := shared.ArangoPort

	return []core.ContainerPort{
		{
			Name:          shared.ServerContainerName,
			ContainerPort: int32(port),
			Protocol:      core.ProtocolTCP,
		},
	}
}

func (a *ArangoGatewayContainer) GetExecutor() string {
	return a.groupSpec.GetEntrypoint(ArangoGatewayExecutor)
}

func (a *ArangoGatewayContainer) GetSecurityContext() *core.SecurityContext {
	return k8sutil.CreateSecurityContext(a.groupSpec.SecurityContext)
}

func (a *ArangoGatewayContainer) GetProbes() (*core.Probe, *core.Probe, *core.Probe, error) {
	var liveness, readiness, startup *core.Probe

	probeLivenessConfig, err := a.resources.getLivenessProbe(a.spec, a.group, a.imageInfo)
	if err != nil {
		return nil, nil, nil, err
	}

	probeReadinessConfig, err := a.resources.getReadinessProbe(a.spec, a.group, a.imageInfo)
	if err != nil {
		return nil, nil, nil, err
	}

	probeStartupConfig, err := a.resources.getReadinessProbe(a.spec, a.group, a.imageInfo)
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

func (a *ArangoGatewayContainer) GetResourceRequirements() core.ResourceRequirements {
	return kresources.ExtractPodAcceptedResourceRequirement(a.arangoMember.Spec.Overrides.GetResources(&a.groupSpec))
}

func (a *ArangoGatewayContainer) GetLifecycle() (*core.Lifecycle, error) {
	return k8sutil.NewLifecycleFinalizers()
}

func (a *ArangoGatewayContainer) GetImagePullPolicy() core.PullPolicy {
	return a.spec.GetImagePullPolicy()
}

func (a *ArangoGatewayContainer) GetImage() string {
	return a.imageInfo.Image
}

func (a *ArangoGatewayContainer) GetEnvs() ([]core.EnvVar, []core.EnvFromSource) {
	envs := NewEnvBuilder()

	envs.Add(true, k8sutil.GetLifecycleEnv()...)

	var cmChecksum = ""

	if cm, ok := a.cachedStatus.ConfigMap().V1().GetSimple(GetGatewayConfigMapName(a.input.ApiObject.GetName())); ok {
		if v, ok := cm.Data[GatewayConfigChecksumFileName]; ok {
			cmChecksum = v
		}
	}

	envs.Add(true, core.EnvVar{
		Name:  GatewayConfigChecksumENV,
		Value: cmChecksum,
	})

	if len(a.groupSpec.Envs) > 0 {
		for _, env := range a.groupSpec.Envs {
			// Do not override preset envs
			envs.Add(false, core.EnvVar{
				Name:  env.Name,
				Value: env.Value,
			})
		}
	}

	return envs.GetEnvList(), nil
}

func (a *ArangoGatewayContainer) GetVolumeMounts() []core.VolumeMount {
	return createGatewayVolumes(a.input).VolumeMounts()
}
