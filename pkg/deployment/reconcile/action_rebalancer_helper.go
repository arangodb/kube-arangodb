//
// DISCLAIMER
//
// Copyright 2023-2026 ArangoDB GmbH, Cologne, Germany
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

	adbDriverV2 "github.com/arangodb/go-driver/v2/arangodb"
)

type RebalanceActions []RebalanceAction

type RebalanceAction struct {
	Database   string `json:"database"`
	Collection string `json:"collection"`

	Shard string `json:"shard"`
	From  string `json:"from"`
	To    string `json:"to"`

	DependsOn []int `json:"depends_on,omitempty"`
}

func runMoveJobs(ctx context.Context, client adbDriverV2.Client, a RebalanceActions) ([]string, []error) {
	var errors []error
	var ids []string
	for _, z := range a {
		id, ok, err := runMoveJob(ctx, client, z)
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

func runMoveJob(ctx context.Context, client adbDriverV2.Client, a RebalanceAction) (string, bool, error) {
	if len(a.DependsOn) != 0 {
		return "", false, nil
	}

	db, err := client.GetDatabase(ctx, a.Database, nil)
	if err != nil {
		return "", false, err
	}

	col, err := db.GetCollection(ctx, a.Collection, nil)
	if err != nil {
		return "", false, err
	}

	var jobID string

	if id, err := client.MoveShard(ctx, col, adbDriverV2.ShardID(a.Shard), adbDriverV2.ServerID(a.From), adbDriverV2.ServerID(a.To)); err != nil {
		return "", false, err
	} else {
		jobID = id
	}

	return jobID, jobID != "", nil
}
