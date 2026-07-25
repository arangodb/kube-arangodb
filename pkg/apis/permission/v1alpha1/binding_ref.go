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
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// ArangoPermissionBindingRef defines a reference to a permission resource, either an
// ArangoPermission CRD by name or an existing authorization sidecar object by its direct name.
type ArangoPermissionBindingRef struct {
	// Name references an ArangoPermission CRD by name. The operator resolves it to the sidecar name.
	Name string `json:"name,omitempty"`

	// Direct references an existing authorization object (role or policy) by its exact name, without
	// a backing ArangoPermission CRD - e.g. an operator-managed predefined role
	// "managed:predefined:coredb-reader". The value is used as-is. Exactly one of Name or Direct
	// must be set.
	Direct string `json:"direct,omitempty"`
}

func (r *ArangoPermissionBindingRef) Hash() string {
	if r == nil {
		return ""
	}
	return util.SHA256FromStringArray(r.Name, r.Direct)
}

// IsDirect reports whether this reference targets an authorization object directly by name (no CRD lookup).
func (r *ArangoPermissionBindingRef) IsDirect() bool {
	return r != nil && r.Direct != ""
}

func (r *ArangoPermissionBindingRef) Validate() error {
	if r == nil {
		return errors.Errorf("is required")
	}

	if r.Name != "" && r.Direct != "" {
		return errors.Errorf("name and direct are mutually exclusive")
	}

	if r.Name == "" && r.Direct == "" {
		return errors.Errorf("one of name or direct is required")
	}

	return nil
}

// ArangoPermissionBindingRefList is a list of ArangoPermissionBindingRef.
type ArangoPermissionBindingRefList []ArangoPermissionBindingRef

func (l ArangoPermissionBindingRefList) Hash() string {
	var hashes []string
	for _, r := range l {
		hashes = append(hashes, r.Hash())
	}
	return util.SHA256FromStringArray(hashes...)
}

// GetReference returns the effective reference: the direct name when Direct is set (used as-is),
// otherwise the CRD name.
func (r *ArangoPermissionBindingRef) GetReference() string {
	if r == nil {
		return ""
	}
	if r.Direct != "" {
		return r.Direct
	}
	return r.Name
}

// ArangoPermissionScope defines a scope boundary as either an inline policy
// or a reference to an ArangoPermissionPolicy CRD. Exactly one must be set.
type ArangoPermissionScope struct {
	// Policy defines the boundary policy inline
	Policy *permissionApiPolicy.Policy `json:"policy,omitempty"`

	// Ref references an ArangoPermissionPolicy CRD to use as boundary
	Ref *ArangoPermissionBindingRef `json:"ref,omitempty"`
}

func (s *ArangoPermissionScope) Hash() string {
	if s == nil {
		return ""
	}
	return util.SHA256FromStringArray(s.Policy.Hash(), s.Ref.Hash())
}

func (s *ArangoPermissionScope) Validate() error {
	if s == nil {
		return errors.Errorf("is required")
	}

	if s.Policy != nil && s.Ref != nil {
		return errors.Errorf("policy and ref are mutually exclusive")
	}

	if s.Policy == nil && s.Ref == nil {
		return errors.Errorf("one of policy or ref is required")
	}

	return shared.WithErrors(
		shared.ValidateOptionalInterfacePath("policy", s.Policy),
		shared.ValidateOptionalInterfacePath("ref", s.Ref),
	)
}

// ArangoPermissionScopedBindingRef binds a role reference with a required scope boundary.
type ArangoPermissionScopedBindingRef struct {
	// Role references an ArangoPermissionRole CRD by name
	// +doc/required
	Role *ArangoPermissionBindingRef `json:"role,omitempty"`

	// Scope defines the boundary for this role binding
	// +doc/required
	Scope *ArangoPermissionScope `json:"scope,omitempty"`
}

func (r *ArangoPermissionScopedBindingRef) Hash() string {
	if r == nil {
		return ""
	}
	return util.SHA256FromStringArray(r.Role.Hash(), r.Scope.Hash())
}

func (r *ArangoPermissionScopedBindingRef) Validate() error {
	if r == nil {
		return errors.Errorf("is required")
	}

	return shared.WithErrors(
		shared.ValidateRequiredInterfacePath("role", r.Role),
		shared.ValidateRequiredInterfacePath("scope", r.Scope),
	)
}

// ArangoPermissionScopedBindingRefList is a list of ArangoPermissionScopedBindingRef.
type ArangoPermissionScopedBindingRefList []ArangoPermissionScopedBindingRef

func (l ArangoPermissionScopedBindingRefList) Hash() string {
	var hashes []string
	for _, r := range l {
		hashes = append(hashes, r.Hash())
	}
	return util.SHA256FromStringArray(hashes...)
}
