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
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1alpha1"
)

// ArangoRoute
const (
	ArangoRouteGroup           = networking.ArangoNetworkingGroupName
	ArangoRouteResource        = networking.ArangoRouteResourcePlural
	ArangoRouteKind            = networking.ArangoRouteResourceKind
	ArangoRouteVersionV1Alpha1 = networkingApi.ArangoNetworkingVersion
)

func init() {
	register[*networkingApi.ArangoRoute](ArangoRouteGKv1Alpha1(), ArangoRouteGRv1Alpha1())
}

func ArangoRouteGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoRouteGroup,
		Kind:  ArangoRouteKind,
	}
}

func ArangoRouteGKv1Alpha1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoRouteGroup,
		Kind:    ArangoRouteKind,
		Version: ArangoRouteVersionV1Alpha1,
	}
}

func ArangoRouteGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoRouteGroup,
		Resource: ArangoRouteResource,
	}
}

func ArangoRouteGRv1Alpha1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoRouteGroup,
		Resource: ArangoRouteResource,
		Version:  ArangoRouteVersionV1Alpha1,
	}
}
