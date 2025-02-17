//
// DISCLAIMER
//
// Copyright 2023-2025 ArangoDB GmbH, Cologne, Germany
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
	registerFeature(enforcedResignLeadership)
	registerFeature(memberReplaceMigration)
}

var enforcedResignLeadership = &feature{
	name:               "enforced-resign-leadership",
	description:        "Enforce ResignLeadership and ensure that Leaders are moved from restarted DBServer",
	enterpriseRequired: false,
	enabledByDefault:   true,
}

var memberReplaceMigration = &feature{
	name:               "replace-migration",
	description:        "During member replacement shards are migrated directly to the new server",
	enterpriseRequired: false,
	enabledByDefault:   true,
}

// EnforcedResignLeadership returns enforced ResignLeadership.
func EnforcedResignLeadership() Feature {
	return enforcedResignLeadership
}

// MemberReplaceMigration returns enforced MemberReplaceMigration.
func MemberReplaceMigration() Feature {
	return memberReplaceMigration
}
