//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
)

// ArangoPlatformWorkflow
const (
	ArangoPlatformWorkflowGroup    = platform.ArangoPlatformGroupName
	ArangoPlatformWorkflowResource = platform.ArangoPlatformWorkflowResourcePlural
	ArangoPlatformWorkflowKind     = platform.ArangoPlatformWorkflowResourceKind

	ArangoPlatformWorkflowVersionV1Beta1 = platformApi.ArangoPlatformVersion
)

func init() {
	register[*platformApi.ArangoPlatformWorkflow](ArangoPlatformWorkflowGKv1Beta1(), ArangoPlatformWorkflowGRv1Beta1())
}

func ArangoPlatformWorkflowGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoPlatformWorkflowGroup,
		Kind:  ArangoPlatformWorkflowKind,
	}
}

func ArangoPlatformWorkflowGKv1Beta1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoPlatformWorkflowGroup,
		Kind:    ArangoPlatformWorkflowKind,
		Version: ArangoPlatformWorkflowVersionV1Beta1,
	}
}

func ArangoPlatformWorkflowGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoPlatformWorkflowGroup,
		Resource: ArangoPlatformWorkflowResource,
	}
}

func ArangoPlatformWorkflowGRv1Beta1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoPlatformWorkflowGroup,
		Resource: ArangoPlatformWorkflowResource,
		Version:  ArangoPlatformWorkflowVersionV1Beta1,
	}
}
