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

package v1beta1

import (
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ArangoPlatformServiceSpec struct {
	// Deployment keeps the Deployment Reference
	// +doc/required
	// +doc/skip: namespace
	// +doc/skip: uid
	// +doc/skip: checksum
	Deployment *sharedApi.Object `json:"deployment,omitempty"`

	// Chart keeps the Chart Reference
	// +doc/required
	// +doc/skip: namespace
	// +doc/skip: uid
	// +doc/skip: checksum
	Chart *sharedApi.Object `json:"chart,omitempty"`

	// Values keeps the values of the Service
	// +doc/type: Object
	Values sharedApi.Any `json:"values,omitempty,omitzero"`

	// Upgrade keeps the upgrade overrides
	Upgrade *ArangoPlatformServiceSpecUpgrade `json:"upgrade,omitempty"`

	// Install keeps the install overrides
	Install *ArangoPlatformServiceSpecInstall `json:"install,omitempty"`
}

func (c *ArangoPlatformServiceSpec) Validate() error {
	if c == nil {
		return errors.Errorf("Nil spec not allowed")
	}

	return shared.WithErrors(
		shared.ValidateRequiredInterfacePath("deployment", c.Deployment),
		shared.ValidateRequiredInterfacePath("chart", c.Chart),
		shared.ValidateOptionalInterfacePath("upgrade", c.Upgrade),
		shared.ValidateOptionalInterfacePath("install", c.Install),
	)
}
