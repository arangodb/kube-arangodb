//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ServiceAccount
const (
	ServiceAccountGroup     = core.GroupName
	ServiceAccountResource  = "serviceaccounts"
	ServiceAccountKind      = "ServiceAccount"
	ServiceAccountVersionV1 = "v1"
)

func ServiceAccountGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ServiceAccountGroup,
		Kind:  ServiceAccountKind,
	}
}

func ServiceAccountGKv1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ServiceAccountGroup,
		Kind:    ServiceAccountKind,
		Version: ServiceAccountVersionV1,
	}
}

func ServiceAccountGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ServiceAccountGroup,
		Resource: ServiceAccountResource,
	}
}

func ServiceAccountGRv1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ServiceAccountGroup,
		Resource: ServiceAccountResource,
		Version:  ServiceAccountVersionV1,
	}
}
