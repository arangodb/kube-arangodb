//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

var _ interfaces.ContainerCreator = &MemberGatewayContainer{}

type MemberGatewayContainer struct {
	*MemberGatewayPod
	resources *Resources
}

func (a *MemberGatewayContainer) GetCommand() ([]string, error) {
	cmd := make([]string, 0, 128)
	cmd = append(cmd, a.GetExecutor())
	cmd = append(cmd, createArangoGatewayArgs(a.Input)...)
	return cmd, nil
}

func (a *MemberGatewayContainer) GetName() string {
	return shared.ServerContainerName
}

func (a *MemberGatewayContainer) GetPorts() []core.ContainerPort {
	port := shared.ArangoPort

	return []core.ContainerPort{
		{
			Name:          shared.ServerContainerName,
			ContainerPort: int32(port),
			Protocol:      core.ProtocolTCP,
		},
	}
}

func (a *MemberGatewayContainer) GetExecutor() string {
	return a.GroupSpec.GetEntrypoint(utilConstants.ArangoGatewayExecutor)
}

func (a *MemberGatewayContainer) GetSecurityContext() *core.SecurityContext {
	return k8sutil.CreateSecurityContext(a.GroupSpec.SecurityContext)
}

func (a *MemberGatewayContainer) GetProbes() (*core.Probe, *core.Probe, *core.Probe, error) {
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

func (a *MemberGatewayContainer) GetResourceRequirements() core.ResourceRequirements {
	return kresources.ScaleResources(kresources.ExtractPodAcceptedResourceRequirement(a.ArangoMember.Spec.Overrides.GetResources(&a.GroupSpec)), 0.75)
}

func (a *MemberGatewayContainer) GetLifecycle() (*core.Lifecycle, error) {
	return k8sutil.NewLifecycleFinalizers()
}

func (a *MemberGatewayContainer) GetImagePullPolicy() core.PullPolicy {
	return a.Deployment.GetImagePullPolicy()
}

func (a *MemberGatewayContainer) GetImage() string {
	return a.Image.Image
}

func (a *MemberGatewayContainer) GetEnvs() ([]core.EnvVar, []core.EnvFromSource) {
	envs := NewEnvBuilder()

	envs.Add(true, k8sutil.GetLifecycleEnv()...)

	if !a.Deployment.Gateway.IsDynamic() {
		if cm, ok := a.cachedStatus.ConfigMap().V1().GetSimple(GetGatewayConfigMapName(a.ApiObject.GetName())); ok {
			if v, ok := cm.Data[utilConstants.ConfigMapChecksumKey]; ok {
				envs.Add(true, core.EnvVar{
					Name:  utilConstants.GatewayConfigChecksumENV,
					Value: v,
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

	return envs.GetEnvList(), nil
}

func (a *MemberGatewayContainer) GetVolumeMounts() []core.VolumeMount {
	return createGatewayVolumes(a.Input).VolumeMounts()
}
