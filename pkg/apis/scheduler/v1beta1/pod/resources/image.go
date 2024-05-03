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

	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/interfaces"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

type ImagePullSecrets []string

var _ interfaces.Pod[Image] = &Image{}

type Image struct {
	// ImagePullSecrets define Secrets used to pull Image from registry
	ImagePullSecrets ImagePullSecrets `json:"imagePullSecrets,omitempty"`
}

func (i *Image) Apply(pod *core.PodTemplateSpec) error {
	if i == nil {
		return nil
	}

	for _, secret := range i.ImagePullSecrets {
		if hasImagePullSecret(pod.Spec.ImagePullSecrets, secret) {
			continue
		}

		pod.Spec.ImagePullSecrets = append(pod.Spec.ImagePullSecrets, core.LocalObjectReference{
			Name: secret,
		})
	}

	return nil
}

func (i *Image) With(other *Image) *Image {
	if i == nil && other == nil {
		return nil
	}

	if other == nil {
		return i.DeepCopy()
	}

	return other.DeepCopy()
}

func (i *Image) Validate() error {
	if i == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceErrors("pullSecrets", shared.ValidateList(i.ImagePullSecrets, shared.ValidateResourceName)),
	)
}

func hasImagePullSecret(secrets []core.LocalObjectReference, secret string) bool {
	for _, sec := range secrets {
		if sec.Name == secret {
			return true
		}
	}

	return false
}
