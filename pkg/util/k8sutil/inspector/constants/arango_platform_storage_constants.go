//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

	"github.com/arangodb/kube-arangodb/pkg/apis/platform"
	platformApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
)

// ArangoPlatformStorage
const (
	ArangoPlatformStorageGroup    = platform.ArangoPlatformGroupName
	ArangoPlatformStorageResource = platform.ArangoPlatformStorageResourcePlural
	ArangoPlatformStorageKind     = platform.ArangoPlatformStorageResourceKind
	// deprecated: Use v1beta1 instead
	ArangoPlatformStorageVersionV1Alpha1 = platformApiv1alpha1.ArangoPlatformVersion
	ArangoPlatformStorageVersionV1Beta1  = platformApi.ArangoPlatformVersion
)

func init() {
	register[*platformApi.ArangoPlatformStorage](ArangoPlatformStorageGKv1Beta1(), ArangoPlatformStorageGRv1Beta1())
}

func ArangoPlatformStorageGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoPlatformStorageGroup,
		Kind:  ArangoPlatformStorageKind,
	}
}

// deprecated: Use v1beta1 instead
func ArangoPlatformStorageGKv1Alpha1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoPlatformStorageGroup,
		Kind:    ArangoPlatformStorageKind,
		Version: ArangoPlatformStorageVersionV1Alpha1,
	}
}

func ArangoPlatformStorageGKv1Beta1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoPlatformStorageGroup,
		Kind:    ArangoPlatformStorageKind,
		Version: ArangoPlatformStorageVersionV1Beta1,
	}
}

func ArangoPlatformStorageGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoPlatformStorageGroup,
		Resource: ArangoPlatformStorageResource,
	}
}

// deprecated: Use v1beta1 instead
func ArangoPlatformStorageGRv1Alpha1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoPlatformStorageGroup,
		Resource: ArangoPlatformStorageResource,
		Version:  ArangoPlatformStorageVersionV1Alpha1,
	}
}

func ArangoPlatformStorageGRv1Beta1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoPlatformStorageGroup,
		Resource: ArangoPlatformStorageResource,
		Version:  ArangoPlatformStorageVersionV1Beta1,
	}
}
