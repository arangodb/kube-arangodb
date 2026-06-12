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
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	permissionApiPolicy "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1/policy"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	ArangoPermissionTokenMinTTL = 15 * time.Minute
)

type ArangoPermissionTokenSpec struct {
	// Deployment keeps the Deployment Reference
	// +doc/required
	// +doc/skip: namespace
	// +doc/skip: uid
	// +doc/skip: checksum
	Deployment *sharedApi.Object `json:"deployment"`

	// Roles keeps the scoped role references assigned to the token.
	// Each entry binds a role with a required scope boundary.
	Roles ArangoPermissionScopedBindingRefList `json:"roles,omitempty"`

	// TTL Defines the TTL of the token.
	// +doc/type: string
	// +doc/default: 1h
	TTL *meta.Duration `json:"ttl,omitempty"`

	// Policy defines an inline authorization policy to create for this token
	Policy *permissionApiPolicy.Policy `json:"policy,omitempty"`

	// Scope defines the boundary policy that constrains what the referenced policy can grant on this token's managed role
	Scope *permissionApiPolicy.Policy `json:"scope,omitempty"`
}

func (c *ArangoPermissionTokenSpec) Hash() string {
	return util.SHA256FromStringArray(c.Deployment.GetName(), c.Roles.Hash(), c.Policy.Hash(), c.Scope.Hash(), c.GetTTL().String())
}

func (c *ArangoPermissionTokenSpec) GetTTL() time.Duration {
	if t := c.TTL; t != nil {
		return t.Duration
	}

	return time.Hour
}

func (c *ArangoPermissionTokenSpec) Validate() error {
	if c == nil {
		return errors.Errorf("Nil spec not allowed")
	}

	return shared.WithErrors(
		shared.ValidateRequiredInterfacePath("deployment", c.Deployment),
		shared.ValidateOptionalInterfacePath("policy", c.Policy),
		shared.ValidateOptionalInterfacePath("scope", c.Scope),
		shared.ValidateOptionalPath("ttl", c.TTL, func(duration meta.Duration) error {
			if duration.Duration < ArangoPermissionTokenMinTTL {
				return errors.Errorf("MinTTL is %s, while TTL has been set to %s", ArangoPermissionTokenMinTTL.String(), duration.Duration.String())
			}

			return nil
		}),
	)
}
