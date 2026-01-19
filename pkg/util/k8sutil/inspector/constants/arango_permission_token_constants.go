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
	"github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1"
	platformApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
)

// ArangoPermissionToken
const (
	ArangoPermissionTokenGroup           = permission.ArangoPermissionGroupName
	ArangoPermissionTokenResource        = permission.ArangoPermissionTokenResourcePlural
	ArangoPermissionTokenKind            = permission.ArangoPermissionTokenResourceKind
	ArangoPermissionTokenVersionV1Alpha1 = platformApiv1alpha1.ArangoPlatformVersion
)

func init() {
	register[*v1alpha1.ArangoPermissionToken](ArangoPermissionTokenGKv1Alpha1(), ArangoPermissionTokenGRv1Alpha1())
}

func ArangoPermissionTokenGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoPermissionTokenGroup,
		Kind:  ArangoPermissionTokenKind,
	}
}

func ArangoPermissionTokenGKv1Alpha1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoPermissionTokenGroup,
		Kind:    ArangoPermissionTokenKind,
		Version: ArangoPermissionTokenVersionV1Alpha1,
	}
}

func ArangoPermissionTokenGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoPermissionTokenGroup,
		Resource: ArangoPermissionTokenResource,
	}
}

func ArangoPermissionTokenGRv1Alpha1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoPermissionTokenGroup,
		Resource: ArangoPermissionTokenResource,
		Version:  ArangoPermissionTokenVersionV1Alpha1,
	}
}
