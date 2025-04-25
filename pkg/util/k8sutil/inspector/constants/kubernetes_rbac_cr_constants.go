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

// ClusterRole
const (
	ClusterRoleGroup     = rbac.GroupName
	ClusterRoleResource  = "clusterroles"
	ClusterRoleKind      = "ClusterRole"
	ClusterRoleVersionV1 = "v1"
)

func init() {
	register[*rbac.ClusterRole](ClusterRoleGKv1(), ClusterRoleGRv1())
}

func ClusterRoleGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ClusterRoleGroup,
		Kind:  ClusterRoleKind,
	}
}

func ClusterRoleGKv1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ClusterRoleGroup,
		Kind:    ClusterRoleKind,
		Version: ClusterRoleVersionV1,
	}
}

func ClusterRoleGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ClusterRoleGroup,
		Resource: ClusterRoleResource,
	}
}

func ClusterRoleGRv1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ClusterRoleGroup,
		Resource: ClusterRoleResource,
		Version:  ClusterRoleVersionV1,
	}
}
