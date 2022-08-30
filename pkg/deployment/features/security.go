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

package features

func init() {
	registerFeature(ephemeralVolumes)
	registerFeature(sensitiveInformationProtection)
}

var ephemeralVolumes = &feature{
	name:               "ephemeral-volumes",
	description:        "Enables ephemeral volumes for apps and tmp directory",
	version:            "3.7.0",
	enterpriseRequired: false,
	enabledByDefault:   false,
}

var sensitiveInformationProtection = &feature{
	name:               "sensitive-information-protection",
	description:        "Hide sensitive information from metrics and logs",
	version:            "3.7.0",
	enterpriseRequired: false,
	enabledByDefault:   false,
}

func EphemeralVolumes() Feature {
	return ephemeralVolumes
}

func SensitiveInformationProtection() Feature {
	return sensitiveInformationProtection
}
