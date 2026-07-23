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

const (
	ArangoPermissionTokenCRDName        = ArangoPermissionTokenResourcePlural + "." + ArangoPermissionGroupName
	ArangoPermissionTokenResourceKind   = "ArangoPermissionToken"
	ArangoPermissionTokenResourcePlural = "arangopermissiontokens"

	ArangoPermissionRoleCRDName        = ArangoPermissionRoleResourcePlural + "." + ArangoPermissionGroupName
	ArangoPermissionRoleResourceKind   = "ArangoPermissionRole"
	ArangoPermissionRoleResourcePlural = "arangopermissionroles"

	ArangoPermissionPolicyCRDName        = ArangoPermissionPolicyResourcePlural + "." + ArangoPermissionGroupName
	ArangoPermissionPolicyResourceKind   = "ArangoPermissionPolicy"
	ArangoPermissionPolicyResourcePlural = "arangopermissionpolicies"

	ArangoPermissionPolicyRoleBindingCRDName        = ArangoPermissionPolicyRoleBindingResourcePlural + "." + ArangoPermissionGroupName
	ArangoPermissionPolicyRoleBindingResourceKind   = "ArangoPermissionPolicyRoleBinding"
	ArangoPermissionPolicyRoleBindingResourcePlural = "arangopermissionpolicyrolebindings"

	ArangoPermissionRoleUserBindingCRDName        = ArangoPermissionRoleUserBindingResourcePlural + "." + ArangoPermissionGroupName
	ArangoPermissionRoleUserBindingResourceKind   = "ArangoPermissionRoleUserBinding"
	ArangoPermissionRoleUserBindingResourcePlural = "arangopermissionroleuserbindings"

	ArangoPermissionGroupName = "permission.arangodb.com"

	// LabelPolicyRoleBindingRole is a label set on ArangoPermissionPolicyRoleBinding
	// with the name of the referenced ArangoPermissionRole CRD.
	LabelPolicyRoleBindingRole = ArangoPermissionGroupName + "/role"

	// ManagedPredefinedPrefix is the reserved name prefix of operator-managed predefined roles
	// (and their policies) created directly in the authorization sidecar. An
	// ArangoPermissionPolicyRoleBinding may reference such a role by this direct sidecar name -
	// predefined roles have no ArangoPermissionRole CRD - to attach a policy to it.
	ManagedPredefinedPrefix = "managed:predefined:"
)
