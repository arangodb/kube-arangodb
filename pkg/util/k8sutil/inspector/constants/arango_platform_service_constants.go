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

	"github.com/arangodb/kube-arangodb/pkg/apis/platform"
	platformApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
)

// ArangoPlatformService
const (
	ArangoPlatformServiceGroup    = platform.ArangoPlatformGroupName
	ArangoPlatformServiceResource = platform.ArangoPlatformServiceResourcePlural
	ArangoPlatformServiceKind     = platform.ArangoPlatformServiceResourceKind
	// deprecated: Use v1beta1 instead
	ArangoPlatformServiceVersionV1Alpha1 = platformApiv1alpha1.ArangoPlatformVersion
	ArangoPlatformServiceVersionV1Beta1  = platformApi.ArangoPlatformVersion
)

func init() {
	register[*platformApi.ArangoPlatformService](ArangoPlatformServiceGKv1Beta1(), ArangoPlatformServiceGRv1Beta1())
}

func ArangoPlatformServiceGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoPlatformServiceGroup,
		Kind:  ArangoPlatformServiceKind,
	}
}

// deprecated: Use v1beta1 instead
func ArangoPlatformServiceGKv1Alpha1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoPlatformServiceGroup,
		Kind:    ArangoPlatformServiceKind,
		Version: ArangoPlatformServiceVersionV1Alpha1,
	}
}

func ArangoPlatformServiceGKv1Beta1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoPlatformServiceGroup,
		Kind:    ArangoPlatformServiceKind,
		Version: ArangoPlatformServiceVersionV1Beta1,
	}
}

func ArangoPlatformServiceGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoPlatformServiceGroup,
		Resource: ArangoPlatformServiceResource,
	}
}

// deprecated: Use v1beta1 instead
func ArangoPlatformServiceGRv1Alpha1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoPlatformServiceGroup,
		Resource: ArangoPlatformServiceResource,
		Version:  ArangoPlatformServiceVersionV1Alpha1,
	}
}

func ArangoPlatformServiceGRv1Beta1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoPlatformServiceGroup,
		Resource: ArangoPlatformServiceResource,
		Version:  ArangoPlatformServiceVersionV1Beta1,
	}
}
