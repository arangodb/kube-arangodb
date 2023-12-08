//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package v1

import (
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

type Image struct {
	// Image define image details
	Image *string `json:"image,omitempty"`

	// PullPolicy define Image pull policy
	// +doc/default: IfNotPresent
	PullPolicy *core.PullPolicy `json:"pullPolicy,omitempty"`

	// PullSecrets define Secrets used to pull Image from registry
	PullSecrets []string `json:"pullSecrets,omitempty"`
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

func (i *Image) GetImage() string {
	if i == nil || i.Image == nil {
		return ""
	}

	return *i.Image
}

func (i *Image) Validate() error {
	if i == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceErrors("image", shared.ValidateRequired(i.Image, shared.ValidateImage)),
		shared.PrefixResourceErrors("pullPolicy", shared.ValidateOptional(i.PullPolicy, shared.ValidatePullPolicy)),
		shared.PrefixResourceErrors("pullSecrets", shared.ValidateList(i.PullSecrets, shared.ValidateResourceName)),
	)
}
