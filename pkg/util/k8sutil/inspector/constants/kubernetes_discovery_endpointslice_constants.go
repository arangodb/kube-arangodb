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
	discovery "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Endpoints
const (
	EndpointSlicesGroup     = discovery.GroupName
	EndpointSlicesResource  = "endpointslices"
	EndpointSlicesKind      = "EndpointSlices"
	EndpointSlicesVersionV1 = "v1"
)

func init() {
	register[*discovery.EndpointSlice](EndpointSlicesGKv1(), EndpointSlicesGRv1())
}

func EndpointSlicesGK() schema.GroupKind {
	return schema.GroupKind{
		Group: EndpointSlicesGroup,
		Kind:  EndpointSlicesKind,
	}
}

func EndpointSlicesGKv1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   EndpointSlicesGroup,
		Kind:    EndpointSlicesKind,
		Version: EndpointSlicesVersionV1,
	}
}

func EndpointSlicesGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    EndpointSlicesGroup,
		Resource: EndpointSlicesResource,
	}
}

func EndpointSlicesGRv1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    EndpointSlicesGroup,
		Resource: EndpointSlicesResource,
		Version:  EndpointSlicesVersionV1,
	}
}
