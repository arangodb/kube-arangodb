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

package reconcile

import (
	"context"
	"sort"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/permission"
	permissionApi "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1"
	versioned "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	sidecarSvcAuthzDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/integration"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

// managedObjectPrefix marks these authorization sidecar objects as operator-managed predefined
// roles. Predefined roles are visible to customers but not editable; a policy can be attached
// to one through an ArangoPermissionPolicyRoleBinding that references it by this name.
const managedObjectPrefix = permission.ManagedPredefinedPrefix

// managedRoleName builds the sidecar object name for a predefined role. A role and its policy
// share the same name (they live in separate policy/role collections, so there is no clash).
func managedRoleName(name string) string { return managedObjectPrefix + name }

// allowAllStatements permits every action on every resource (`*` action / `*` resource).
func allowAllStatements() []*sidecarSvcAuthzTypes.PolicyStatement {
	return []*sidecarSvcAuthzTypes.PolicyStatement{
		{
			Effect:    sidecarSvcAuthzTypes.Effect_Allow,
			Actions:   []string{"*"},
			Resources: []string{"*"},
		},
	}
}

// allowAllScope is the Allow-all scope bound to the root user (`*` with scoping `*`).
func allowAllScope() *sidecarSvcAuthzTypes.Policy {
	return &sidecarSvcAuthzTypes.Policy{Statements: allowAllStatements()}
}

// predefinedRole is an operator-managed role synced into the authorization sidecar. Customers
// can assign and scope these roles but cannot edit them. Only the super-admin role ships with
// a policy and is bound to the root user; the remaining roles are created as empty containers
// whose policies are defined in a later iteration.
type predefinedRole struct {
	// Name is the short, hyphenated role identifier, e.g. "coredb-reader".
	Name string

	// Description is the human-readable description surfaced to the UI.
	Description string

	// Statements, when set, are attached to the role through an operator-managed policy.
	Statements []*sidecarSvcAuthzTypes.PolicyStatement

	// BindRootUser binds the deployment root user to this role with an Allow-all scope.
	BindRootUser bool
}

// predefinedRoles is the catalog of operator-managed roles created in every deployment's
// authorization sidecar. Descriptions are taken from the RBAC product definition.
var predefinedRoles = []predefinedRole{
	{
		Name:         "super-admin",
		Description:  "Reserved role providing full access to all functionality. Cannot be assigned by the customer.",
		Statements:   allowAllStatements(),
		BindRootUser: true,
	},
	{
		Name:        "tenant-admin",
		Description: "Manages users and role bindings.",
	},
	{
		Name:        "coredb-reader",
		Description: "Reads scoped resources and executes read-only database operations.",
	},
	{
		Name:        "coredb-developer",
		Description: "Reads and writes scoped resources and executes read and write database operations.",
	},
	{
		Name:        "coredb-admin",
		Description: "Manages scoped resources' structures and lifecycle.",
	},
	{
		Name:        "ai-user",
		Description: "Executes AI workflows and reads resulting outputs within scoped resources.",
	},
	{
		Name:        "ai-developer",
		Description: "Builds, configures, manages, and executes AI workflows and artifacts within scoped resources.",
	},
	{
		Name:        "platform-operator",
		Description: "Operates platform services, manages bundled services, views observability, and starts containers within scoped resources.",
	},
	{
		Name:        "secret-admin",
		Description: "Manages secrets within scoped resources.",
	},
}

// managedRBACClient opens an authorization sidecar client for the deployment. `enabled` is
// false (with a nil client) when the deployment does not run the gateway/authorization
// sidecar. The returned close function must be called when the client is no longer needed.
func managedRBACClient(kube kubernetes.Interface, depl *api.ArangoDeployment) (sidecarSvcAuthzDefinition.AuthorizationAPIClient, func() error, bool, error) {
	conn, enabled, err := integration.NewIntegrationConnectionFromDeployment(kube, depl, utilToken.WithRelativeDuration(time.Minute))
	if err != nil {
		return nil, nil, false, err
	}

	if !enabled {
		return nil, nil, false, nil
	}

	return sidecarSvcAuthzDefinition.NewAuthorizationAPIClient(conn), conn.Close, true, nil
}

// syncRBACPermissions ensures the operator-managed predefined roles exist in the authorization
// sidecar and match their canonical definition, merged with any policies customers have attached
// through ArangoPermissionPolicyRoleBindings. `boundPolicies` returns the sidecar policy names
// attached to a role by such bindings. It is idempotent - missing objects are created and
// drifted ones are repaired - so it is safe to run repeatedly.
func syncRBACPermissions(ctx context.Context, conn sidecarSvcAuthzDefinition.AuthorizationAPIClient, boundPolicies func(roleName string) ([]string, error)) error {
	for _, role := range predefinedRoles {
		extra, err := boundPolicies(managedRoleName(role.Name))
		if err != nil {
			return err
		}

		if err := ensurePredefinedRole(ctx, conn, role, extra); err != nil {
			return err
		}
	}

	return nil
}

func ensurePredefinedRole(ctx context.Context, conn sidecarSvcAuthzDefinition.AuthorizationAPIClient, r predefinedRole, boundPolicies []string) error {
	name := managedRoleName(r.Name)

	var policies []string

	if len(r.Statements) > 0 {
		// The bundled policy shares the role name.
		if err := ensureManagedPolicy(ctx, conn, name, &sidecarSvcAuthzTypes.Policy{
			Description: r.Description,
			Statements:  r.Statements,
		}); err != nil {
			return err
		}

		policies = append(policies, name)
	}

	// Merge in policies attached to this role via ArangoPermissionPolicyRoleBindings.
	policies = append(policies, boundPolicies...)
	policies = util.UniqueList(policies)
	sort.Strings(policies)

	if err := ensureManagedRole(ctx, conn, name, &sidecarSvcAuthzTypes.Role{
		Description: r.Description,
		Policies:    policies,
	}); err != nil {
		return err
	}

	if r.BindRootUser {
		return ensureRootUserRoleBinding(ctx, conn, name)
	}

	return nil
}

// collectBoundPolicies lists the ArangoPermissionPolicyRoleBindings that target the given
// predefined role and resolves their policy CRD references to sidecar policy names, mirroring
// the aggregation the ArangoPermissionRole handler performs for CRD roles.
func collectBoundPolicies(ctx context.Context, arango versioned.Interface, ns, roleName string) ([]string, error) {
	bindings, err := arango.PermissionV1alpha1().ArangoPermissionPolicyRoleBindings(ns).List(ctx, meta.ListOptions{})
	if err != nil {
		return nil, err
	}

	var policies []string
	seen := map[string]struct{}{}

	for i := range bindings.Items {
		b := &bindings.Items[i]

		if b.Spec.Role.GetReference() != roleName {
			continue
		}

		if !b.Status.Conditions.IsTrue(permissionApi.ReadyPolicyCondition) || !b.Status.Conditions.IsTrue(permissionApi.ReadyRoleCondition) {
			continue
		}

		if b.Spec.Policy == nil {
			continue
		}

		policyObj, err := arango.PermissionV1alpha1().ArangoPermissionPolicies(ns).Get(ctx, b.Spec.Policy.GetReference(), meta.GetOptions{})
		if err != nil {
			continue
		}

		if !policyObj.Ready() || policyObj.Status.Policy == nil {
			continue
		}

		name := policyObj.Status.Policy.GetName()
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		policies = append(policies, name)
	}

	sort.Strings(policies)

	return policies, nil
}

func ensureManagedPolicy(ctx context.Context, conn sidecarSvcAuthzDefinition.AuthorizationAPIClient, name string, desired *sidecarSvcAuthzTypes.Policy) error {
	existing, err := conn.APIGetPolicy(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest{Name: name})
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return err
		}

		if _, err := conn.APICreatePolicy(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIPolicyRequest{Name: name, Item: desired}); err != nil {
			if status.Code(err) != codes.AlreadyExists {
				return err
			}
		}

		return nil
	}

	if existing.GetItem().Hash() == desired.Hash() {
		return nil
	}

	_, err = conn.APIUpdatePolicy(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIPolicyRequest{Name: name, Item: desired})
	return err
}

func ensureManagedRole(ctx context.Context, conn sidecarSvcAuthzDefinition.AuthorizationAPIClient, name string, desired *sidecarSvcAuthzTypes.Role) error {
	existing, err := conn.APIGetRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest{Name: name})
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return err
		}

		if _, err := conn.APICreateRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIRoleRequest{Name: name, Item: desired}); err != nil {
			if status.Code(err) != codes.AlreadyExists {
				return err
			}
		}

		return nil
	}

	if existing.GetItem().Hash() == desired.Hash() {
		return nil
	}

	_, err = conn.APIUpdateRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIRoleRequest{Name: name, Item: desired})
	return err
}

func ensureRootUserRoleBinding(ctx context.Context, conn sidecarSvcAuthzDefinition.AuthorizationAPIClient, roleName string) error {
	desiredScope := allowAllScope()

	bindings, err := conn.APIListUserRoleBindings(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIUserRequest{User: api.UserNameRoot})
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return err
		}
	}

	if bindings != nil {
		for _, b := range bindings.GetBindings() {
			if b.GetRole() != roleName {
				continue
			}

			// Binding exists - repair the scope only when it has drifted.
			var currentScopeHash string
			if s := b.GetScope(); s != nil {
				currentScopeHash = s.Hash()
			}

			if currentScopeHash == desiredScope.Hash() {
				return nil
			}

			_, err := conn.APIReplaceUserRoleScope(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingRequest{
				User:  api.UserNameRoot,
				Role:  roleName,
				Scope: desiredScope,
			})
			return err
		}
	}

	// Binding missing - assign it.
	if _, err := conn.APIAssignUserRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingRequest{
		User:  api.UserNameRoot,
		Role:  roleName,
		Scope: desiredScope,
	}); err != nil {
		if status.Code(err) != codes.AlreadyExists {
			return err
		}
	}

	return nil
}
