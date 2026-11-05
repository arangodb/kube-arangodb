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
	registerFeature(rbacCoreDB)
}

var rbacCoreDB = &feature{
	name:             "rbac-coredb",
	description:      "Point the serving member's arangod at the local authorization integration sidecar via --server.external-rbac-service",
	enabledByDefault: false,
	hidden:           true,
	dependencies:     []Feature{RBACEnforced(), CentralServices()},
}

// RBACCoreDB reports whether the serving member should be pointed at the local authorization integration
// sidecar. It is only enabled when both rbac-enforced and central-services are enabled.
func RBACCoreDB() Feature {
	return rbacCoreDB
}
