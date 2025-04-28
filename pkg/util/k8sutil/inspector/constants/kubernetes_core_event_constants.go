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
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Event
const (
	EventGroup     = core.GroupName
	EventResource  = "events"
	EventKind      = "Event"
	EventVersionV1 = "v1"
)

func init() {
	register[*core.Event](EventGKv1(), EventGRv1())
}

func EventGK() schema.GroupKind {
	return schema.GroupKind{
		Group: EventGroup,
		Kind:  EventKind,
	}
}

func EventGKv1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   EventGroup,
		Kind:    EventKind,
		Version: EventVersionV1,
	}
}

func EventGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    EventGroup,
		Resource: EventResource,
	}
}

func EventGRv1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    EventGroup,
		Resource: EventResource,
		Version:  EventVersionV1,
	}
}
