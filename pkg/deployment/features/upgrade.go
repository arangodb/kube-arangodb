//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	registerFeature(upgradeVersionCheck)
	registerFeature(upgradeVersionCheckV2)
}

var upgradeVersionCheck Feature = &feature{
	name:               "upgrade-version-check",
	description:        "Enable initContainer with pre version check",
	version:            "3.6.0",
	enterpriseRequired: false,
	enabledByDefault:   true,
}

var upgradeVersionCheckV2 Feature = &feature{
	name:               "upgrade-version-check-v2",
	description:        "Enable initContainer with pre version check based by Operator",
	version:            "3.6.0",
	enterpriseRequired: false,
	enabledByDefault:   false,
}

func UpgradeVersionCheck() Feature {
	return upgradeVersionCheck
}

func UpgradeVersionCheckV2() Feature {
	return upgradeVersionCheckV2
}
