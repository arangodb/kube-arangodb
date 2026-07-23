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
	"github.com/arangodb/kube-arangodb/pkg/apis/permission"
	permissionApiPolicy "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1/policy"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ArangoPermissionRoleUserBindingSpec struct {
	// Deployment keeps the Deployment Reference
	// +doc/required
	// +doc/skip: namespace
	// +doc/skip: uid
	// +doc/skip: checksum
	Deployment *sharedApi.Object `json:"deployment"`

	// Role defines the role to bind, either by CRD name or direct sidecar name
	// +doc/required
	Role *ArangoPermissionBindingRef `json:"role"`

	// UserName is the name of the user to bind the role to
	// +doc/required
	UserName string `json:"userName"`

	// Scope defines the inline scope policy for this binding
	// +doc/required
	Scope *permissionApiPolicy.Policy `json:"scope"`
}

func (c *ArangoPermissionRoleUserBindingSpec) Hash() string {
	if c == nil {
		return ""
	}
	return util.SHA256FromStringArray(
		c.Deployment.GetName(),
		c.Role.Hash(),
		c.UserName,
		c.Scope.Hash(),
	)
}

func (c *ArangoPermissionRoleUserBindingSpec) Validate() error {
	if c == nil {
		return errors.Errorf("Nil spec not allowed")
	}

	return shared.WithErrors(
		shared.ValidateRequiredInterfacePath("deployment", c.Deployment),
		shared.ValidateRequiredInterfacePath("role", c.Role),
		func() error {
			// The super-admin role is reserved: it grants full access and is bound to the root user
			// automatically, so it must not be assignable to a user by a customer binding.
			if permission.IsReservedRoleName(c.Role.GetReference()) {
				return errors.Errorf("role %q is reserved and cannot be assigned", c.Role.GetReference())
			}
			return nil
		}(),
		func() error {
			if c.UserName == "" {
				return errors.Errorf("userName is required")
			}
			return nil
		}(),
		shared.ValidateRequiredInterfacePath("scope", c.Scope),
	)
}
