//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package client

import (
	"context"
	"net/http"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type RebalanceExecuteRequest struct {
	Version int `json:"version"`

	Moves RebalanceExecuteRequestMoves `json:"moves,omitempty"`
}

type RebalanceExecuteRequestMoves []RebalanceExecuteRequestMove

type RebalanceExecuteRequestMove struct {
	Database   string `json:"database"`
	Collection string `json:"collection"`
	Shard      string `json:"shard"`

	From string `json:"from"`
	To   string `json:"to"`

	IsLeader bool `json:"isLeader"`
}

func (c *client) RebalanceExecuteMoves(ctx context.Context, moves ...RebalanceExecuteRequestMove) error {
	return c.RebalanceExecute(ctx, &RebalanceExecuteRequest{
		Version: 1,
		Moves:   moves,
	})
}

func (c *client) RebalanceExecute(ctx context.Context, request *RebalanceExecuteRequest) error {
	req, err := c.c.NewRequest(http.MethodPost, "/_admin/cluster/rebalance/execute")
	if err != nil {
		return err
	}

	request = util.InitType(request)

	// Always set to 1
	request.Version = 1

	if r, err := req.SetBody(request); err != nil {
		return err
	} else {
		req = r
	}

	resp, err := c.c.Do(ctx, req)
	if err != nil {
		return err
	}

	if err := resp.CheckStatus(http.StatusAccepted); err != nil {
		return err
	}

	return nil
}
