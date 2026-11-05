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

const (
	HelmLabelBase             = "platform.arangodb.com"
	HelmLabelInstallationBase = "installation." + HelmLabelBase

	HelmLabelArangoDBManaged = HelmLabelInstallationBase + "/managed"

	HelmLabelArangoDBChart = HelmLabelInstallationBase + "/chart"

	HelmLabelArangoDBDeployment = HelmLabelInstallationBase + "/deployment"

	HelmLabelArangoDBService = HelmLabelInstallationBase + "/service"

	// HelmLabelArangoDBType set to platform or service
	HelmLabelArangoDBType = HelmLabelInstallationBase + "/type"

	HelmLabelTag = HelmLabelBase + "/tag"

	// LabelArangoDBDeploymentName tags a Helm release with the owning ArangoDeployment name. It is kept
	// 1:1 with the SchedulerV2 integration (which sets and lists releases by this label) so that releases
	// installed via ArangoPlatformWorkflow remain discoverable by the same selector.
	LabelArangoDBDeploymentName = "deployment.arangodb.com/name"
)

// HelmType is the value of the HelmLabelArangoDBType label, categorizing a managed Helm release.
type HelmType string

const (
	// HelmTypePlatform tags releases that are part of the platform itself.
	HelmTypePlatform HelmType = "platform"
	// HelmTypeService tags releases that provide a platform service (e.g. installed via SchedulerV2).
	HelmTypeService HelmType = "service"
)

// String returns the label value for the type.
func (t HelmType) String() string {
	return string(t)
}
