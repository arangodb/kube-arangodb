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
	"github.com/arangodb/kube-arangodb/pkg/util"
)

var _ interfaces.Container[Image] = &Image{}

type Image struct {
	// Image define image details
	Image *string `json:"image,omitempty"`

	// ImagePullPolicy define Image pull policy
	// +doc/default: IfNotPresent
	ImagePullPolicy *core.PullPolicy `json:"imagePullPolicy,omitempty"`
}

func (i *Image) Apply(pod *core.PodTemplateSpec, container *core.Container) error {
	if i == nil {
		return nil
	}

	container.Image = util.WithDefault(i.Image)
	container.ImagePullPolicy = util.WithDefault(i.ImagePullPolicy)

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
		shared.PrefixResourceErrors("imagePullPolicy", shared.ValidateOptional(i.ImagePullPolicy, shared.ValidatePullPolicy)),
	)
}
