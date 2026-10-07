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

package constants

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arangodb/kube-arangodb/pkg/apis/permission"
	permissionApi "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1"
)

// ArangoPermissionRoleGroupBinding
const (
	ArangoPermissionRoleGroupBindingGroup           = permission.ArangoPermissionGroupName
	ArangoPermissionRoleGroupBindingResource        = permission.ArangoPermissionRoleGroupBindingResourcePlural
	ArangoPermissionRoleGroupBindingKind            = permission.ArangoPermissionRoleGroupBindingResourceKind
	ArangoPermissionRoleGroupBindingVersionV1Alpha1 = permissionApi.ArangoPermissionVersion
)

func init() {
	register[*permissionApi.ArangoPermissionRoleGroupBinding](ArangoPermissionRoleGroupBindingGKv1Alpha1(), ArangoPermissionRoleGroupBindingGRv1Alpha1())
}

func ArangoPermissionRoleGroupBindingGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoPermissionRoleGroupBindingGroup,
		Kind:  ArangoPermissionRoleGroupBindingKind,
	}
}

func ArangoPermissionRoleGroupBindingGKv1Alpha1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoPermissionRoleGroupBindingGroup,
		Kind:    ArangoPermissionRoleGroupBindingKind,
		Version: ArangoPermissionRoleGroupBindingVersionV1Alpha1,
	}
}

func ArangoPermissionRoleGroupBindingGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoPermissionRoleGroupBindingGroup,
		Resource: ArangoPermissionRoleGroupBindingResource,
	}
}

func ArangoPermissionRoleGroupBindingGRv1Alpha1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoPermissionRoleGroupBindingGroup,
		Resource: ArangoPermissionRoleGroupBindingResource,
		Version:  ArangoPermissionRoleGroupBindingVersionV1Alpha1,
	}
}
