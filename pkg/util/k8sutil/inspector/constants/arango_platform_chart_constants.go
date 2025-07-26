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
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
)

// ArangoPlatformChart
const (
	ArangoPlatformChartGroup           = platform.ArangoPlatformGroupName
	ArangoPlatformChartResource        = platform.ArangoPlatformChartResourcePlural
	ArangoPlatformChartKind            = platform.ArangoPlatformChartResourceKind
	ArangoPlatformChartVersionV1Alpha1 = platformApi.ArangoPlatformVersion
)

func init() {
	register[*platformApi.ArangoPlatformChart](ArangoPlatformChartGKv1Alpha1(), ArangoPlatformChartGRv1Alpha1())
}

func ArangoPlatformChartGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoPlatformChartGroup,
		Kind:  ArangoPlatformChartKind,
	}
}

func ArangoPlatformChartGKv1Alpha1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoPlatformChartGroup,
		Kind:    ArangoPlatformChartKind,
		Version: ArangoPlatformChartVersionV1Alpha1,
	}
}

func ArangoPlatformChartGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoPlatformChartGroup,
		Resource: ArangoPlatformChartResource,
	}
}

func ArangoPlatformChartGRv1Alpha1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoPlatformChartGroup,
		Resource: ArangoPlatformChartResource,
		Version:  ArangoPlatformChartVersionV1Alpha1,
	}
}
