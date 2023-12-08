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
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

type ContainerTemplate struct {
	// Image define default image used for the job
	*Image `json:",inline"`

	// Resources define resources assigned to the pod
	*Resources `json:",inline"`

	// SecurityContainer keeps the security settings for Container
	*SecurityContainer `json:",inline"`
}

func (a *ContainerTemplate) With(other *ContainerTemplate) *ContainerTemplate {
	if a == nil && other == nil {
		return nil
	}

	if a == nil {
		return other.DeepCopy()
	}

	if other == nil {
		return a.DeepCopy()
	}

	return &ContainerTemplate{
		Image:             a.GetImage().With(other.GetImage()),
		Resources:         a.GetResources().With(other.GetResources()),
		SecurityContainer: a.GetSecurityContainer().With(other.GetSecurityContainer()),
	}
}

func (a *ContainerTemplate) GetImage() *Image {
	if a == nil || a.Image == nil {
		return nil
	}

	return a.Image
}

func (a *ContainerTemplate) GetSecurityContainer() *SecurityContainer {
	if a == nil || a.SecurityContainer == nil {
		return nil
	}

	return a.SecurityContainer
}

func (a *ContainerTemplate) GetResources() *Resources {
	if a == nil || a.Resources == nil {
		return nil
	}

	return a.Resources
}

func (a *ContainerTemplate) Validate() error {
	if a == nil {
		return nil
	}
	return shared.WithErrors(
		a.GetImage().Validate(),
		a.GetResources().Validate(),
		a.GetSecurityContainer().Validate(),
	)
}
