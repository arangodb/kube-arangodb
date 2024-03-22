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

package pod

import (
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/interfaces"
	schedulerPodResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/pod/resources"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

var _ interfaces.Pod[Pod] = &Pod{}

type Pod struct {
	// Metadata keeps the metadata settings for Pod
	*schedulerPodResourcesApi.Metadata `json:",inline"`

	// Image keeps the image information
	*schedulerPodResourcesApi.Image `json:",inline"`

	// Scheduling keeps the scheduling information
	*schedulerPodResourcesApi.Scheduling `json:",inline"`

	// Namespace keeps the Container layer Kernel namespace configuration
	*schedulerPodResourcesApi.Namespace `json:",inline"`

	// Security keeps the security settings for Pod
	*schedulerPodResourcesApi.Security `json:",inline"`

	// Volumes keeps the volumes settings for Pod
	*schedulerPodResourcesApi.Volumes `json:",inline"`

	// ServiceAccount keeps the service account settings for Pod
	*schedulerPodResourcesApi.ServiceAccount `json:",inline"`
}

func (a *Pod) With(other *Pod) *Pod {
	if a == nil && other == nil {
		return nil
	}

	if a == nil {
		return other.DeepCopy()
	}

	if other == nil {
		return a.DeepCopy()
	}

	return &Pod{
		Scheduling:     a.Scheduling.With(other.Scheduling),
		Image:          a.Image.With(other.Image),
		Namespace:      a.Namespace.With(other.Namespace),
		Security:       a.Security.With(other.Security),
		Volumes:        a.Volumes.With(other.Volumes),
		ServiceAccount: a.ServiceAccount.With(other.ServiceAccount),
		Metadata:       a.Metadata.With(other.Metadata),
	}
}

func (a *Pod) Apply(template *core.PodTemplateSpec) error {
	if a == nil {
		return nil
	}

	return shared.WithErrors(
		a.Scheduling.Apply(template),
		a.Image.Apply(template),
		a.Namespace.Apply(template),
		a.Security.Apply(template),
		a.Volumes.Apply(template),
		a.ServiceAccount.Apply(template),
		a.Metadata.Apply(template),
	)
}

func (a *Pod) GetSecurity() *schedulerPodResourcesApi.Security {
	if a == nil {
		return nil
	}

	return a.Security
}

func (a *Pod) GetImage() *schedulerPodResourcesApi.Image {
	if a == nil {
		return nil
	}

	return a.Image
}

func (a *Pod) GetScheduling() *schedulerPodResourcesApi.Scheduling {
	if a == nil {
		return nil
	}

	return a.Scheduling
}

func (a *Pod) GetContainerNamespace() *schedulerPodResourcesApi.Namespace {
	if a == nil {
		return nil
	}

	return a.Namespace
}

func (a *Pod) GetVolumes() *schedulerPodResourcesApi.Volumes {
	if a == nil {
		return nil
	}

	return a.Volumes
}

func (a *Pod) GetServiceAccount() *schedulerPodResourcesApi.ServiceAccount {
	if a == nil {
		return nil
	}

	return a.ServiceAccount
}

func (a *Pod) GetMetadata() *schedulerPodResourcesApi.Metadata {
	if a == nil {
		return nil
	}

	return a.Metadata
}

func (a *Pod) Validate() error {
	if a == nil {
		return nil
	}
	return shared.WithErrors(
		a.Scheduling.Validate(),
		a.Image.Validate(),
		a.Namespace.Validate(),
		a.Security.Validate(),
		a.Volumes.Validate(),
		a.ServiceAccount.Validate(),
		a.Metadata.Validate(),
	)
}
