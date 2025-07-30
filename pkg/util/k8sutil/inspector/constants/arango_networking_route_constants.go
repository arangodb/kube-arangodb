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

	"github.com/arangodb/kube-arangodb/pkg/apis/networking"
	networkingApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1alpha1"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1beta1"
)

// ArangoRoute
const (
	ArangoRouteGroup    = networking.ArangoNetworkingGroupName
	ArangoRouteResource = networking.ArangoRouteResourcePlural
	ArangoRouteKind     = networking.ArangoRouteResourceKind
	// deprecated: Use v1beta1 instead
	ArangoRouteVersionV1Alpha1 = networkingApiv1alpha1.ArangoNetworkingVersion
	ArangoRouteVersionV1Beta1  = networkingApi.ArangoNetworkingVersion
)

func init() {
	register[*networkingApi.ArangoRoute](ArangoRouteGKv1Beta1(), ArangoRouteGRv1Beta1())
}

func ArangoRouteGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoRouteGroup,
		Kind:  ArangoRouteKind,
	}
}

// deprecated: Use v1beta1 instead
func ArangoRouteGKv1Alpha1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoRouteGroup,
		Kind:    ArangoRouteKind,
		Version: ArangoRouteVersionV1Alpha1,
	}
}

func ArangoRouteGKv1Beta1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoRouteGroup,
		Kind:    ArangoRouteKind,
		Version: ArangoRouteVersionV1Beta1,
	}
}

func ArangoRouteGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoRouteGroup,
		Resource: ArangoRouteResource,
	}
}

// deprecated: Use v1beta1 instead
func ArangoRouteGRv1Alpha1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoRouteGroup,
		Resource: ArangoRouteResource,
		Version:  ArangoRouteVersionV1Alpha1,
	}
}

func ArangoRouteGRv1Beta1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoRouteGroup,
		Resource: ArangoRouteResource,
		Version:  ArangoRouteVersionV1Beta1,
	}
}
