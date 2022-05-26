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

package inspector

import (
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Pod
const (
	PodGroup     = core.GroupName
	PodResource  = "pods"
	PodKind      = "Pod"
	PodVersionV1 = "v1"
)

func PodGK() schema.GroupKind {
	return schema.GroupKind{
		Group: PodGroup,
		Kind:  PodKind,
	}
}

func PodGKv1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   PodGroup,
		Kind:    PodKind,
		Version: PodVersionV1,
	}
}

func PodGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    PodGroup,
		Resource: PodResource,
	}
}

func PodGRv1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    PodGroup,
		Resource: PodResource,
		Version:  PodVersionV1,
	}
}

func (p *podsInspectorV1) GroupVersionKind() schema.GroupVersionKind {
	return PodGKv1()
}

func (p *podsInspectorV1) GroupVersionResource() schema.GroupVersionResource {
	return PodGRv1()
}

func (p *podsInspector) GroupKind() schema.GroupKind {
	return PodGK()
}

func (p *podsInspector) GroupResource() schema.GroupResource {
	return PodGR()
}
