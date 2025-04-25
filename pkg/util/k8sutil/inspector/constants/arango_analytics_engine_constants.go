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

	"github.com/arangodb/kube-arangodb/pkg/apis/analytics"
	analyticsApi "github.com/arangodb/kube-arangodb/pkg/apis/analytics/v1alpha1"
)

// ArangoAnalyticsGraphAnalyticsEngine
const (
	ArangoAnalyticsGraphAnalyticsEngineGroup           = analytics.ArangoAnalyticsGroupName
	ArangoAnalyticsGraphAnalyticsEngineResource        = analytics.GraphAnalyticsEngineResourcePlural
	ArangoAnalyticsGraphAnalyticsEngineKind            = analytics.GraphAnalyticsEngineResourceKind
	ArangoAnalyticsGraphAnalyticsEngineVersionV1Alpha1 = analyticsApi.ArangoAnalyticsVersion
)

func init() {
	register[*analyticsApi.GraphAnalyticsEngine](ArangoAnalyticsGraphAnalyticsEngineGKv1Alpha1(), ArangoAnalyticsGraphAnalyticsEngineGRv1Alpha1())
}

func ArangoAnalyticsGraphAnalyticsEngineGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoAnalyticsGraphAnalyticsEngineGroup,
		Kind:  ArangoAnalyticsGraphAnalyticsEngineKind,
	}
}

func ArangoAnalyticsGraphAnalyticsEngineGKv1Alpha1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoAnalyticsGraphAnalyticsEngineGroup,
		Kind:    ArangoAnalyticsGraphAnalyticsEngineKind,
		Version: ArangoAnalyticsGraphAnalyticsEngineVersionV1Alpha1,
	}
}

func ArangoAnalyticsGraphAnalyticsEngineGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoAnalyticsGraphAnalyticsEngineGroup,
		Resource: ArangoAnalyticsGraphAnalyticsEngineResource,
	}
}

func ArangoAnalyticsGraphAnalyticsEngineGRv1Alpha1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoAnalyticsGraphAnalyticsEngineGroup,
		Resource: ArangoAnalyticsGraphAnalyticsEngineResource,
		Version:  ArangoAnalyticsGraphAnalyticsEngineVersionV1Alpha1,
	}
}
