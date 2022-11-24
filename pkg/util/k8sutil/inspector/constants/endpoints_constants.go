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

// Endpoints
const (
	EndpointsGroup     = core.GroupName
	EndpointsResource  = "endpoints"
	EndpointsKind      = "Endpoints"
	EndpointsVersionV1 = "v1"
)

func EndpointsGK() schema.GroupKind {
	return schema.GroupKind{
		Group: EndpointsGroup,
		Kind:  EndpointsKind,
	}
}

func EndpointsGKv1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   EndpointsGroup,
		Kind:    EndpointsKind,
		Version: EndpointsVersionV1,
	}
}

func EndpointsGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    EndpointsGroup,
		Resource: EndpointsResource,
	}
}

func EndpointsGRv1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    EndpointsGroup,
		Resource: EndpointsResource,
		Version:  EndpointsVersionV1,
	}
}
