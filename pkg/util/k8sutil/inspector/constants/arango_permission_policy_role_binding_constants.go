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

// ArangoPermissionPolicyRoleBinding
const (
	ArangoPermissionPolicyRoleBindingGroup           = permission.ArangoPermissionGroupName
	ArangoPermissionPolicyRoleBindingResource        = permission.ArangoPermissionPolicyRoleBindingResourcePlural
	ArangoPermissionPolicyRoleBindingKind            = permission.ArangoPermissionPolicyRoleBindingResourceKind
	ArangoPermissionPolicyRoleBindingVersionV1Alpha1 = permissionApi.ArangoPermissionVersion
)

func init() {
	register[*permissionApi.ArangoPermissionPolicyRoleBinding](ArangoPermissionPolicyRoleBindingGKv1Alpha1(), ArangoPermissionPolicyRoleBindingGRv1Alpha1())
}

func ArangoPermissionPolicyRoleBindingGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoPermissionPolicyRoleBindingGroup,
		Kind:  ArangoPermissionPolicyRoleBindingKind,
	}
}

func ArangoPermissionPolicyRoleBindingGKv1Alpha1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoPermissionPolicyRoleBindingGroup,
		Kind:    ArangoPermissionPolicyRoleBindingKind,
		Version: ArangoPermissionPolicyRoleBindingVersionV1Alpha1,
	}
}

func ArangoPermissionPolicyRoleBindingGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoPermissionPolicyRoleBindingGroup,
		Resource: ArangoPermissionPolicyRoleBindingResource,
	}
}

func ArangoPermissionPolicyRoleBindingGRv1Alpha1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoPermissionPolicyRoleBindingGroup,
		Resource: ArangoPermissionPolicyRoleBindingResource,
		Version:  ArangoPermissionPolicyRoleBindingVersionV1Alpha1,
	}
}
