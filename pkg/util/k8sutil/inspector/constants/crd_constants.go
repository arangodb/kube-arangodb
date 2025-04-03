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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CustomResourceDefinition
const (
	CustomResourceDefinitionGroup     = "apiextensions.k8s.io"
	CustomResourceDefinitionResource  = "customresourcesdefinition"
	CustomResourceDefinitionKind      = "CustomResourceDefinition"
	CustomResourceDefinitionVersionV1 = "v1"
)

func CustomResourceDefinitionGK() schema.GroupKind {
	return schema.GroupKind{
		Group: CustomResourceDefinitionGroup,
		Kind:  CustomResourceDefinitionKind,
	}
}

func CustomResourceDefinitionGKv1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   CustomResourceDefinitionGroup,
		Kind:    CustomResourceDefinitionKind,
		Version: CustomResourceDefinitionVersionV1,
	}
}

func CustomResourceDefinitionGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    CustomResourceDefinitionGroup,
		Resource: CustomResourceDefinitionResource,
	}
}

func CustomResourceDefinitionGRv1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    CustomResourceDefinitionGroup,
		Resource: CustomResourceDefinitionResource,
		Version:  CustomResourceDefinitionVersionV1,
	}
}
