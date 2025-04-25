//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package constants

import (
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Role
const (
	RoleGroup     = rbac.GroupName
	RoleResource  = "roles"
	RoleKind      = "Role"
	RoleVersionV1 = "v1"
)

func init() {
	register[*rbac.Role](RoleGKv1(), RoleGRv1())
}

func RoleGK() schema.GroupKind {
	return schema.GroupKind{
		Group: RoleGroup,
		Kind:  RoleKind,
	}
}

func RoleGKv1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   RoleGroup,
		Kind:    RoleKind,
		Version: RoleVersionV1,
	}
}

func RoleGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    RoleGroup,
		Resource: RoleResource,
	}
}

func RoleGRv1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    RoleGroup,
		Resource: RoleResource,
		Version:  RoleVersionV1,
	}
}
