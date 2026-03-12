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

// ArangoPermissionRole
const (
	ArangoPermissionRoleGroup           = permission.ArangoPermissionGroupName
	ArangoPermissionRoleResource        = permission.ArangoPermissionRoleResourcePlural
	ArangoPermissionRoleKind            = permission.ArangoPermissionRoleResourceKind
	ArangoPermissionRoleVersionV1Alpha1 = permissionApi.ArangoPermissionVersion
)

func init() {
	register[*permissionApi.ArangoPermissionRole](ArangoPermissionRoleGKv1Alpha1(), ArangoPermissionRoleGRv1Alpha1())
}

func ArangoPermissionRoleGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoPermissionRoleGroup,
		Kind:  ArangoPermissionRoleKind,
	}
}

func ArangoPermissionRoleGKv1Alpha1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoPermissionRoleGroup,
		Kind:    ArangoPermissionRoleKind,
		Version: ArangoPermissionRoleVersionV1Alpha1,
	}
}

func ArangoPermissionRoleGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoPermissionRoleGroup,
		Resource: ArangoPermissionRoleResource,
	}
}

func ArangoPermissionRoleGRv1Alpha1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoPermissionRoleGroup,
		Resource: ArangoPermissionRoleResource,
		Version:  ArangoPermissionRoleVersionV1Alpha1,
	}
}
