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
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ArangoPermissionPolicyRoleBindingSpec struct {
	// Deployment keeps the Deployment Reference
	// +doc/required
	// +doc/skip: namespace
	// +doc/skip: uid
	// +doc/skip: checksum
	Deployment *sharedApi.Object `json:"deployment"`

	// Policy defines the policy to bind, either by CRD name or direct sidecar name
	// +doc/required
	Policy *ArangoPermissionBindingRef `json:"policy"`

	// Role defines the role to bind to, either by CRD name or direct sidecar name
	// +doc/required
	Role *ArangoPermissionBindingRef `json:"role"`
}

func (c *ArangoPermissionPolicyRoleBindingSpec) Hash() string {
	if c == nil {
		return ""
	}
	return util.SHA256FromStringArray(
		c.Deployment.GetName(),
		c.Policy.Hash(),
		c.Role.Hash(),
	)
}

func (c *ArangoPermissionPolicyRoleBindingSpec) Validate() error {
	if c == nil {
		return errors.Errorf("Nil spec not allowed")
	}

	return shared.WithErrors(
		shared.ValidateRequiredInterfacePath("deployment", c.Deployment),
		shared.ValidateRequiredInterfacePath("policy", c.Policy),
		shared.ValidateRequiredInterfacePath("role", c.Role),
	)
}
