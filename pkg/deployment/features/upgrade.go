//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

import (
	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func init() {
	registerFeature(upgradeVersionCheck)
	registerFeature(upgradeVersionCheckV2)
	registerFeature(upgradeAlternativeOrder)
	registerFeature(upgradeIndexOrderIssue)
}

var upgradeVersionCheck Feature = &feature{
	name:               "upgrade-version-check",
	description:        "Enable initContainer with pre version check",
	enterpriseRequired: false,
	enabledByDefault:   true,
}

var upgradeVersionCheckV2 Feature = &feature{
	name:               "upgrade-version-check-v2",
	description:        "Enable initContainer with pre version check based by Operator",
	enterpriseRequired: false,
	enabledByDefault:   false,
}

var upgradeAlternativeOrder Feature = &feature{
	name:               "upgrade-alternative-order",
	description:        "Changes order of the upgrade process - Coordinators are upgraded before DBServers",
	enterpriseRequired: false,
	enabledByDefault:   false,
	hidden:             true,
}

var upgradeIndexOrderIssue Feature = &feature{
	name:               "upgrade-index-order-issue",
	description:        "Changes the default upgrade mode for DBServers while upgrading from 3.12.2/3.12.3 to 3.12.4+",
	enterpriseRequired: false,
	enabledByDefault:   true,
	hidden:             true,
}

func UpgradeVersionCheck() Feature {
	return upgradeVersionCheck
}

func UpgradeVersionCheckV2() Feature {
	return upgradeVersionCheckV2
}

func UpgradeAlternativeOrder() Feature { return upgradeAlternativeOrder }

func UpgradeIndexOrderIssue() Feature { return upgradeIndexOrderIssue }

func IsUpgradeIndexOrderIssueEnabled(group api.ServerGroup, from, to driver.Version) bool {
	if !UpgradeIndexOrderIssue().Enabled() {
		return false
	}

	if group != api.ServerGroupDBServers {
		return false
	}

	if from.CompareTo("3.12.2") < 0 || from.CompareTo("3.12.3") > 0 {
		// Outside of versions
		return false
	}

	if to.CompareTo("3.12.4") < 0 {
		return false
	}

	return true
}
