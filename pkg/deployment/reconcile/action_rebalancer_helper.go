//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

	"github.com/arangodb/go-driver"
	"github.com/arangodb/rebalancer/pkg/inventory/server"
	"github.com/arangodb/rebalancer/pkg/inventory/shard"
)

type RebalanceActions []RebalanceAction

type RebalanceAction struct {
	Database   string `json:"database"`
	Collection string `json:"collection"`

	Shard shard.ID  `json:"shard"`
	From  server.ID `json:"from"`
	To    server.ID `json:"to"`

	DependsOn []int `json:"depends_on,omitempty"`
}

func runMoveJobs(ctx context.Context, client driver.Client, cluster driver.Cluster, a RebalanceActions) ([]string, []error) {
	var errors []error
	var ids []string
	for _, z := range a {
		id, ok, err := runMoveJob(ctx, client, cluster, z)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		if !ok {
			continue
		}

		ids = append(ids, id)
	}

	return ids, errors
}

func runMoveJob(ctx context.Context, client driver.Client, cluster driver.Cluster, a RebalanceAction) (string, bool, error) {
	if len(a.DependsOn) != 0 {
		return "", false, nil
	}

	db, err := client.Database(ctx, a.Database)
	if err != nil {
		return "", false, err
	}

	col, err := db.Collection(ctx, a.Collection)
	if err != nil {
		return "", false, err
	}

	var jobID string
	jctx := driver.WithJobIDResponse(ctx, &jobID)

	if err := cluster.MoveShard(jctx, col, driver.ShardID(a.Shard), driver.ServerID(a.From), driver.ServerID(a.To)); err != nil {
		return "", false, err
	}

	return jobID, jobID != "", nil
}
