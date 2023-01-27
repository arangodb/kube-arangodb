//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	"time"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const coordinatorHealthFailedTimeout time.Duration = time.Minute

func (r *Reconciler) createClusterOperationPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	planCtx PlanBuilderContext) api.Plan {

	if spec.GetMode() != api.DeploymentModeCluster {
		return nil
	}

	c, err := planCtx.GetMembersState().State().GetDatabaseClient()
	if err != nil {
		return nil
	}

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	cluster, err := c.Cluster(ctxChild)
	if err != nil {
		r.log.Err(err).Warn("Unable to get Cluster client")
		return nil
	}

	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	health, err := cluster.Health(ctxChild)
	if err != nil {
		r.log.Err(err).Warn("Unable to get Cluster health")
		return nil
	}

	membersHealth := health.Health

	for _, e := range status.Members.AsList() {
		delete(membersHealth, driver.ServerID(e.Member.ID))
	}

	if len(membersHealth) == 0 {
		return nil
	}

	for id, member := range membersHealth {
		switch member.Role {
		case driver.ServerRoleCoordinator:
			if member.Status != driver.ServerStatusFailed {
				continue
			}

			if member.LastHeartbeatAcked.Add(coordinatorHealthFailedTimeout).Before(time.Now()) {
				return api.Plan{
					actions.NewAction(api.ActionTypeClusterMemberCleanup, api.ServerGroupCoordinators, shared.WithPredefinedMember(string(id))),
				}
			}
		case driver.ServerRoleDBServer:
			if member.Status != driver.ServerStatusFailed {
				continue
			}

			if !member.CanBeDeleted {
				continue
			}

			return api.Plan{
				actions.NewAction(api.ActionTypeClusterMemberCleanup, api.ServerGroupDBServers, shared.WithPredefinedMember(string(id))),
			}
		}
	}

	return nil
}
