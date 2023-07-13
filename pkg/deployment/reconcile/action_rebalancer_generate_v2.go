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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func newRebalancerGenerateV2Action(action api.Action, actionCtx ActionContext) Action {
	a := &actionRebalancerGenerateV2{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionRebalancerGenerateV2 struct {
	actionImpl

	actionEmptyCheckProgress
}

func (r actionRebalancerGenerateV2) Start(ctx context.Context) (bool, error) {
	spec := r.actionCtx.GetSpec()

	if spec.Rebalancer == nil {
		if err := r.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
			if s.Rebalancer == nil {
				return false
			}

			s.Rebalancer = nil
			return true
		}); err != nil {
			r.log.Err(err).Warn("Unable to propagate changes")
			return true, nil
		}

		return true, nil
	}

	c, err := r.actionCtx.GetMembersState().State().GetDatabaseClient()
	if err != nil {
		r.log.Err(err).Error("Unable to get client")
		return true, nil
	}

	nctx, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()

	resp, err := client.NewClient(c.Connection(), r.log).GenerateRebalanceMoves(nctx, &client.RebalancePlanRequest{
		MaximumNumberOfMoves: util.NewType(spec.Rebalancer.GetParallelMoves()),
	})
	if err != nil {
		r.log.Err(err).Error("Unable to generate rebalancer moves")
		return true, nil
	}

	if len(resp.Result.Moves) > 0 {
		cache, ok := r.actionCtx.GetAgencyCache()
		if !ok {
			r.log.Debug("AgencyCache is not ready")
			return true, nil
		}

		actions := make(RebalanceActions, len(resp.Result.Moves))

		for id, move := range resp.Result.Moves {
			db, ok := cache.GetCollectionDatabaseByID(move.Collection.String())
			if !ok {
				r.log.Warn("Database not found for Collection %s", move.Collection)
				return true, nil
			}

			actions[id] = RebalanceAction{
				Database:   db,
				Collection: move.Collection.String(),
				Shard:      move.Shard,
				From:       move.From,
				To:         move.To,
			}
		}

		cluster, err := c.Cluster(ctx)
		if err != nil {
			r.log.Err(err).Warn("Unable to get cluster")
			return true, nil
		}

		if err := r.executeActions(ctx, spec.Rebalancer.GetParallelMoves(), c, cluster, actions); err != nil {
			r.log.Err(err).Warn("Unable to execute actions")
			return true, nil
		}

		return true, nil
	}

	if err := r.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		s.Rebalancer = &api.ArangoDeploymentRebalancerStatus{
			LastCheckTime: k8sutil.NewTime(meta.Now()),
		}

		return true
	}); err != nil {
		r.log.Err(err).Warn("Unable to save plan")
		return true, nil
	}

	return true, nil
}

func (r actionRebalancerGenerateV2) executeActions(ctx context.Context, size int, client driver.Client, cluster driver.Cluster, a RebalanceActions) error {
	if len(a) > size {
		a = a[0:size]
	}

	r.actionCtx.Metrics().GetRebalancer().AddMoves(len(a))

	ids, errors := runMoveJobs(ctx, client, cluster, a)

	r.actionCtx.Metrics().GetRebalancer().AddFailures(len(errors))

	for _, err := range errors {
		r.log.Err(err).Warn("MoveShard failed")
	}

	if err := r.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		s.Rebalancer = &api.ArangoDeploymentRebalancerStatus{}

		s.Rebalancer.MoveJobs = append(s.Rebalancer.MoveJobs, ids...)

		return true
	}); err != nil {
		return err
	}

	return nil
}
