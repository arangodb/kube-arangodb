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

package features

func init() {
	registerFeature(arangoDeploymentStatus)
}

var arangoDeploymentStatus = &feature{
	name:             "enable-arango-deployment-status",
	description:      "Ensures the status subresource on the ArangoDeployment v1 CRD when enabled; when disabled the operator leaves it as the chart ships it (neither adds nor removes it)",
	enabledByDefault: false,
}

func EnableArangoDeploymentStatus() Feature {
	return arangoDeploymentStatus
}
