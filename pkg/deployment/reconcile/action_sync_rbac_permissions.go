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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func newSyncRBACPermissionsAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionSyncRBACPermissions{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionSyncRBACPermissions ensures the operator-managed predefined roles are present in the
// authorization sidecar. It connects to the sidecar and upserts the predefined roles (and, for
// super-admin, its policy and the root user binding) directly through the authorization API (no
// intermediate CRs). It is idempotent, so the throttled high plan builder can re-run it to
// recreate or repair drift.
type actionSyncRBACPermissions struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a actionSyncRBACPermissions) Start(ctx context.Context) (bool, error) {
	depl, ok := a.actionCtx.GetAPIObject().(*api.ArangoDeployment)
	if !ok {
		return false, errors.Errorf("Unable to sync RBAC permissions: API object is not an ArangoDeployment")
	}

	client := a.actionCtx.ACS().CurrentClusterCache().Client()
	ns := a.actionCtx.GetNamespace()

	conn, closeConn, enabled, err := managedRBACClient(client.Kubernetes(), depl)
	if err != nil {
		return false, err
	}

	if !enabled {
		// Authorization sidecar is not enabled - nothing to sync. The plan builder gates on
		// the same condition, so this is only reached on a race.
		return true, nil
	}
	defer closeConn()

	if err := syncRBACPermissions(ctx, conn, func(roleName string) ([]string, error) {
		return collectBoundPolicies(ctx, client.Arango(), ns, roleName)
	}); err != nil {
		return false, err
	}

	// Report the sync as successful so the RBACBootstrapped condition (and UpToDate through it)
	// can become true.
	if err := a.actionCtx.UpdateClusterCondition(ctx, api.ConditionTypeRBACBootstrapped, true, "RBAC Bootstrapped", "RBAC permissions have been synced"); err != nil {
		return false, err
	}

	return true, nil
}
