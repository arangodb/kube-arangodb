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

package v1alpha1

import (
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
)

type ArangoMLExtensionSpecInit struct {
	// Image define default image used for the init job
	*sharedApi.Image `json:",inline"`
}

func (a *ArangoMLExtensionSpecInit) GetImage() *sharedApi.Image {
	if a == nil || a.Image == nil {
		return nil
	}

	return a.Image
}

func (a *ArangoMLExtensionSpecInit) Validate() error {
	if a == nil {
		return nil
	}
	return shared.WithErrors(
		a.GetImage().Validate(),
	)
}