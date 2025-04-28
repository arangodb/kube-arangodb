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
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// StatefulSet
const (
	StatefulSetGroup     = apps.GroupName
	StatefulSetResource  = "statefulsets"
	StatefulSetKind      = "StatefulSet"
	StatefulSetVersionV1 = "v1"
)

func init() {
	register[*apps.StatefulSet](StatefulSetGKv1(), StatefulSetGRv1())
}

func StatefulSetGK() schema.GroupKind {
	return schema.GroupKind{
		Group: StatefulSetGroup,
		Kind:  StatefulSetKind,
	}
}

func StatefulSetGKv1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   StatefulSetGroup,
		Kind:    StatefulSetKind,
		Version: StatefulSetVersionV1,
	}
}

func StatefulSetGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    StatefulSetGroup,
		Resource: StatefulSetResource,
	}
}

func StatefulSetGRv1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    StatefulSetGroup,
		Resource: StatefulSetResource,
		Version:  StatefulSetVersionV1,
	}
}
