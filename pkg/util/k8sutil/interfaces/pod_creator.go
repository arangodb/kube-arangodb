//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

package interfaces

import (
	"context"

	core "k8s.io/api/core/v1"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangoprofile"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/configmap"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service"
)

type Inspector interface {
	secret.Inspector
	service.Inspector
	configmap.Inspector
	arangoprofile.Inspector
}

type PodModifier interface {
	ApplyPodSpec(spec *core.PodSpec) error
}

type PodCreator interface {
	Init(context.Context, Inspector, *core.PodTemplateSpec) error
	GetName() string
	GetRole() string
	GetVolumes() []core.Volume
	GetSidecars(*core.PodTemplateSpec) error
	GetInitContainers(cachedStatus Inspector) ([]core.Container, error)
	GetFinalizers() []string
	GetTolerations() []core.Toleration
	GetNodeSelector() map[string]string
	GetServiceAccountName() string
	GetPodAntiAffinity() *core.PodAntiAffinity
	GetPodAffinity() *core.PodAffinity
	GetNodeAffinity() *core.NodeAffinity
	GetRestartPolicy() core.RestartPolicy
	GetContainerCreator() ContainerCreator
	GetImagePullSecrets() []string
	IsDeploymentMode() bool
	Validate(cachedStatus Inspector) error

	Annotations() map[string]string
	Labels() map[string]string

	Profiles() (schedulerApi.ProfileTemplates, error)

	PodModifier
}

type ContainerCreator interface {
	GetCommand() ([]string, error)
	GetName() string
	GetExecutor() string
	GetProbes() (*core.Probe, *core.Probe, *core.Probe, error)
	GetResourceRequirements(scale float64) core.ResourceRequirements
	GetResourceRequirementsDefaultScale() float64
	GetLifecycle() (*core.Lifecycle, error)
	GetImagePullPolicy() core.PullPolicy
	GetImage() string
	GetEnvs() ([]core.EnvVar, []core.EnvFromSource)
	GetSecurityContext() *core.SecurityContext
	GetPorts() []core.ContainerPort
	GetVolumeMounts() []core.VolumeMount
}
