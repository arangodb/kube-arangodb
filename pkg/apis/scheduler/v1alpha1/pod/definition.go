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
	schedulerPodResourcesApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/pod/resources"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

var _ interfaces.Pod[Pod] = &Pod{}

type Pod struct {
	// Metadata keeps the metadata settings for Pod
	*schedulerPodResourcesApiv1alpha1.Metadata `json:",inline"`

	// Image keeps the image information
	*schedulerPodResourcesApiv1alpha1.Image `json:",inline"`

	// Scheduling keeps the scheduling information
	*schedulerPodResourcesApiv1alpha1.Scheduling `json:",inline"`

	// Namespace keeps the Container layer Kernel namespace configuration
	*schedulerPodResourcesApiv1alpha1.Namespace `json:",inline"`

	// Security keeps the security settings for Pod
	*schedulerPodResourcesApiv1alpha1.Security `json:",inline"`

	// Volumes keeps the volumes settings for Pod
	*schedulerPodResourcesApiv1alpha1.Volumes `json:",inline"`

	// ServiceAccount keeps the service account settings for Pod
	*schedulerPodResourcesApiv1alpha1.ServiceAccount `json:",inline"`
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

func (a *Pod) GetSecurity() *schedulerPodResourcesApiv1alpha1.Security {
	if a == nil {
		return nil
	}

	return a.Security
}

func (a *Pod) GetImage() *schedulerPodResourcesApiv1alpha1.Image {
	if a == nil {
		return nil
	}

	return a.Image
}

func (a *Pod) GetScheduling() *schedulerPodResourcesApiv1alpha1.Scheduling {
	if a == nil {
		return nil
	}

	return a.Scheduling
}

func (a *Pod) GetContainerNamespace() *schedulerPodResourcesApiv1alpha1.Namespace {
	if a == nil {
		return nil
	}

	return a.Namespace
}

func (a *Pod) GetVolumes() *schedulerPodResourcesApiv1alpha1.Volumes {
	if a == nil {
		return nil
	}

	return a.Volumes
}

func (a *Pod) GetServiceAccount() *schedulerPodResourcesApiv1alpha1.ServiceAccount {
	if a == nil {
		return nil
	}

	return a.ServiceAccount
}

func (a *Pod) GetMetadata() *schedulerPodResourcesApiv1alpha1.Metadata {
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
