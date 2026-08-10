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
	registerFeature(harden)
}

var harden = &feature{
	name:               "harden",
	description:        "Adds hardening arguments to the ArangoDB server containers",
	enterpriseRequired: false,
	enabledByDefault:   false,
	dependencies:       []Feature{SecuredContainers()},
}

// Harden returns the harden feature. When enabled it appends hardening arguments to the arangod
// containers of all server groups. It is only enabled when the secured-containers feature is enabled.
func Harden() Feature {
	return harden
}
