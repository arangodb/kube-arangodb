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

import api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"

func init() {
	registerFeature(gateway)
}

var gateway = &feature{
	name:               "gateway",
	description:        "Defines if gateway extension is enabled",
	enterpriseRequired: false,
	enabledByDefault:   true,
	hidden:             false,
}

func Gateway() Feature {
	return gateway
}

func IsGatewayEnabled(spec api.DeploymentSpec) bool {
	return Gateway().Enabled() && spec.IsGatewayEnabled()
}
