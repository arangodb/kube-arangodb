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

// ArangoPermissionPolicy
const (
	ArangoPermissionPolicyGroup           = permission.ArangoPermissionGroupName
	ArangoPermissionPolicyResource        = permission.ArangoPermissionPolicyResourcePlural
	ArangoPermissionPolicyKind            = permission.ArangoPermissionPolicyResourceKind
	ArangoPermissionPolicyVersionV1Alpha1 = permissionApi.ArangoPermissionVersion
)

func init() {
	register[*permissionApi.ArangoPermissionPolicy](ArangoPermissionPolicyGKv1Alpha1(), ArangoPermissionPolicyGRv1Alpha1())
}

func ArangoPermissionPolicyGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoPermissionPolicyGroup,
		Kind:  ArangoPermissionPolicyKind,
	}
}

func ArangoPermissionPolicyGKv1Alpha1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoPermissionPolicyGroup,
		Kind:    ArangoPermissionPolicyKind,
		Version: ArangoPermissionPolicyVersionV1Alpha1,
	}
}

func ArangoPermissionPolicyGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoPermissionPolicyGroup,
		Resource: ArangoPermissionPolicyResource,
	}
}

func ArangoPermissionPolicyGRv1Alpha1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoPermissionPolicyGroup,
		Resource: ArangoPermissionPolicyResource,
		Version:  ArangoPermissionPolicyVersionV1Alpha1,
	}
}
