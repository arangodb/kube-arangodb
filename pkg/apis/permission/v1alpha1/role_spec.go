//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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
	permissionApiPolicy "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1/policy"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ArangoPermissionRoleSpec struct {
	// Deployment keeps the Deployment Reference
	// +doc/required
	// +doc/skip: namespace
	// +doc/skip: uid
	// +doc/skip: checksum
	Deployment *sharedApi.Object `json:"deployment"`

	// Policy defined the Authorization Policy
	Policy *permissionApiPolicy.Policy `json:"policy,omitempty"`
}

func (c *ArangoPermissionRoleSpec) Validate() error {
	if c == nil {
		return errors.Errorf("Nil spec not allowed")
	}

	return shared.WithErrors(
		shared.ValidateOptionalInterfacePath("policy", c.Policy),
		shared.ValidateRequiredInterfacePath("deployment", c.Deployment),
	)
}
