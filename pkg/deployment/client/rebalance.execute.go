//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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
	goHttp "net/http"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
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
	request = util.InitType(request)

	// Always set to 1
	request.Version = 1

	return arangod.PostRequest[*RebalanceExecuteRequest, any](ctx, c.c, request, "_admin", "cluster", "rebalance", "execute").Do(ctx).AcceptCode(goHttp.StatusAccepted).Evaluate()
}
