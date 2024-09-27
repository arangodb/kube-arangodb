//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

import "fmt"

const ProfileGroup = "profiles.arangodb.com"

const ProfilesDeployment = ProfileGroup + "/deployment"
const ProfilesIntegrationPrefix = "integration." + ProfileGroup

const (
	ProfilesIntegrationAuthn = "authn"
	ProfilesIntegrationAuthz = "authz"
	ProfilesIntegrationSched = "sched"
)

const (
	ProfilesIntegrationV0 = "v0"
	ProfilesIntegrationV1 = "v1"
)

func NewProfileIntegration(name, version string) (string, string) {
	return fmt.Sprintf("%s/%s", ProfilesIntegrationPrefix, name), version
}
