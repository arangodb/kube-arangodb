//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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
	registerFeature(failoverLeadership)
	registerFeature(activeFailover)
}

var activeFailover = &feature{
	name:               "active-failover",
	description:        "Support for ActiveFailover mode",
	version:            newFeatureVersion("", "3.12"),
	enterpriseRequired: false,
	enabledByDefault:   true,
}

var failoverLeadership = &feature{
	name:               "failover-leadership",
	description:        "Support for leadership in fail-over mode",
	version:            newFeatureVersion("", "3.12"),
	enterpriseRequired: false,
	enabledByDefault:   false,
}

func FailoverLeadership() Feature {
	return failoverLeadership
}

func ActiveFailover() Feature {
	return activeFailover
}
