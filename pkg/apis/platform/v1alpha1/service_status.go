//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package v1alpha1

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
)

type ArangoPlatformServiceStatus struct {
	// Chart keeps the Deployment Reference
	Deployment *sharedApi.Object `json:"deployment,omitempty"`

	// Chart keeps the Chart Reference
	Chart *sharedApi.Object `json:"chart,omitempty"`

	// Conditions specific to the entire service
	// +doc/type: api.Conditions
	Conditions api.ConditionList `json:"conditions,omitempty"`

	// Values keeps the values of the Service
	Values sharedApi.Any `json:"values,omitempty,omitzero"`

	// ChartInfo keeps the info about Chart
	ChartInfo *ChartStatusInfo `json:"chartInfo,omitempty"`

	// Release keeps the release status
	Release *ArangoPlatformServiceStatusRelease `json:"release,omitempty"`
}
