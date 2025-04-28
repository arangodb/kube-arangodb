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

// RoleBinding
const (
	RoleBindingGroup     = rbac.GroupName
	RoleBindingResource  = "rolebindings"
	RoleBindingKind      = "RoleBinding"
	RoleBindingVersionV1 = "v1"
)

func init() {
	register[*rbac.RoleBinding](RoleBindingGKv1(), RoleBindingGRv1())
}

func RoleBindingGK() schema.GroupKind {
	return schema.GroupKind{
		Group: RoleBindingGroup,
		Kind:  RoleBindingKind,
	}
}

func RoleBindingGKv1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   RoleBindingGroup,
		Kind:    RoleBindingKind,
		Version: RoleBindingVersionV1,
	}
}

func RoleBindingGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    RoleBindingGroup,
		Resource: RoleBindingResource,
	}
}

func RoleBindingGRv1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    RoleBindingGroup,
		Resource: RoleBindingResource,
		Version:  RoleBindingVersionV1,
	}
}
