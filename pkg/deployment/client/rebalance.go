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

package client

import (
	"context"
	"net/http"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/arangodb/rebalancer/pkg/inventory/server"
	"github.com/arangodb/rebalancer/pkg/inventory/shard"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type RebalanceClient interface {
	GenerateRebalanceMoves(ctx context.Context, request *RebalancePlanRequest) (RebalancePlanResponse, error)
}

type RebalancePlanRequest struct {
	Version              int   `json:"version"`
	MaximumNumberOfMoves *int  `json:"maximumNumberOfMoves,omitempty"`
	LeaderChanges        *bool `json:"leaderChanges,omitempty"`
	MoveLeaders          *bool `json:"moveLeaders,omitempty"`
	MoveFollowers        *bool `json:"moveFollowers,omitempty"`
}

type RebalancePlanResponse struct {
	Result RebalancePlanResponseResult `json:"result"`
}

type RebalancePlanResponseResult struct {
	Moves RebalancePlanMoves `json:"moves"`
}

type RebalancePlanMoves []RebalancePlanMove

type RebalancePlanMove struct {
	From  server.ID `json:"from"`
	To    server.ID `json:"to"`
	Shard shard.ID  `json:"shard"`

	Collection intstr.IntOrString `json:"collection"`
}

func (c *client) GenerateRebalanceMoves(ctx context.Context, request *RebalancePlanRequest) (RebalancePlanResponse, error) {
	req, err := c.c.NewRequest(http.MethodPost, "/_admin/cluster/rebalance")
	if err != nil {
		return RebalancePlanResponse{}, err
	}

	request = util.InitType(request)

	// Always set to 1
	request.Version = 1

	if r, err := req.SetBody(request); err != nil {
		return RebalancePlanResponse{}, err
	} else {
		req = r
	}

	resp, err := c.c.Do(ctx, req)
	if err != nil {
		return RebalancePlanResponse{}, err
	}

	if err := resp.CheckStatus(http.StatusOK); err != nil {
		return RebalancePlanResponse{}, err
	}

	var d RebalancePlanResponse

	if err := resp.ParseBody("", &d); err != nil {
		return RebalancePlanResponse{}, err
	}

	return d, nil
}
