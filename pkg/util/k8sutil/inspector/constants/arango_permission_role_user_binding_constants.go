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

// ArangoPermissionRoleUserBinding
const (
	ArangoPermissionRoleUserBindingGroup           = permission.ArangoPermissionGroupName
	ArangoPermissionRoleUserBindingResource        = permission.ArangoPermissionRoleUserBindingResourcePlural
	ArangoPermissionRoleUserBindingKind            = permission.ArangoPermissionRoleUserBindingResourceKind
	ArangoPermissionRoleUserBindingVersionV1Alpha1 = permissionApi.ArangoPermissionVersion
)

func init() {
	register[*permissionApi.ArangoPermissionRoleUserBinding](ArangoPermissionRoleUserBindingGKv1Alpha1(), ArangoPermissionRoleUserBindingGRv1Alpha1())
}

func ArangoPermissionRoleUserBindingGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoPermissionRoleUserBindingGroup,
		Kind:  ArangoPermissionRoleUserBindingKind,
	}
}

func ArangoPermissionRoleUserBindingGKv1Alpha1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoPermissionRoleUserBindingGroup,
		Kind:    ArangoPermissionRoleUserBindingKind,
		Version: ArangoPermissionRoleUserBindingVersionV1Alpha1,
	}
}

func ArangoPermissionRoleUserBindingGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoPermissionRoleUserBindingGroup,
		Resource: ArangoPermissionRoleUserBindingResource,
	}
}

func ArangoPermissionRoleUserBindingGRv1Alpha1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoPermissionRoleUserBindingGroup,
		Resource: ArangoPermissionRoleUserBindingResource,
		Version:  ArangoPermissionRoleUserBindingVersionV1Alpha1,
	}
}
