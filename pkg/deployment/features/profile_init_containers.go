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
	registerFeature(profileAllInitContainers)
}

var profileAllInitContainers = &feature{
	name:             "apply-profile-all-to-init-containers",
	description:      "Also apply the ArangoProfile container.all template to init containers, not only to regular containers",
	enabledByDefault: false,
	hidden:           false,
}

// ProfileAllInitContainers enables applying the ArangoProfile container.all (generic) template to a pod's
// init containers in addition to its regular containers.
func ProfileAllInitContainers() Feature {
	return profileAllInitContainers
}
