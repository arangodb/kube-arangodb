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

package permission

// Predefined role short identifiers. These are the operator-managed roles seeded into every
// RBAC-enabled deployment's authorization sidecar under the ManagedPredefinedPrefix. This list is
// the single source of truth for the role names: the reconcile catalog must cover exactly this set,
// and tests reference these identifiers instead of hard-coding strings.
const (
	PredefinedRoleSuperAdmin       = "super-admin"
	PredefinedRoleTenantAdmin      = "tenant-admin"
	PredefinedRoleCoreDBReader     = "coredb-reader"
	PredefinedRoleCoreDBDeveloper  = "coredb-developer"
	PredefinedRoleCoreDBAdmin      = "coredb-admin"
	PredefinedRoleAIUser           = "ai-user"
	PredefinedRoleAIDeveloper      = "ai-developer"
	PredefinedRolePlatformOperator = "platform-operator"
	PredefinedRoleSecretAdmin      = "secret-admin"
)

// PredefinedRoleNames is the ordered catalog of predefined role short identifiers created in every
// RBAC-enabled deployment. The reconcile catalog must cover exactly this set (asserted by unit
// test), and tests iterate it to verify every role exists in the sidecar.
var PredefinedRoleNames = []string{
	PredefinedRoleSuperAdmin,
	PredefinedRoleTenantAdmin,
	PredefinedRoleCoreDBReader,
	PredefinedRoleCoreDBDeveloper,
	PredefinedRoleCoreDBAdmin,
	PredefinedRoleAIUser,
	PredefinedRoleAIDeveloper,
	PredefinedRolePlatformOperator,
	PredefinedRoleSecretAdmin,
}

// ManagedPredefinedRoleName returns the reserved sidecar object name for a predefined role short
// identifier, e.g. "coredb-reader" -> "managed:predefined:coredb-reader".
func ManagedPredefinedRoleName(name string) string {
	return ManagedPredefinedPrefix + name
}

// IsReservedRoleName reports whether an authorization role name is reserved and must not be
// referenced by customer-created bindings. The predefined super-admin role is reserved: it grants
// full access and the operator binds it to the deployment root user automatically, so allowing a
// customer to assign it would be a privilege escalation.
func IsReservedRoleName(name string) bool {
	return name == ManagedPredefinedRoleName(PredefinedRoleSuperAdmin)
}
