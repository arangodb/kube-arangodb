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

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
)

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
	From  string `json:"from"`
	To    string `json:"to"`
	Shard string `json:"shard"`

	Collection intstr.IntOrString `json:"collection"`
}

func (c *client) RebalancePlan(ctx context.Context, request *RebalancePlanRequest) (RebalancePlanResponse, error) {
	request = util.InitType(request)

	// Always set to 1
	request.Version = 1

	return arangod.PostRequest[*RebalancePlanRequest, RebalancePlanResponse](ctx, c.c, request, "_admin", "cluster", "rebalance").Do(ctx).AcceptCode(goHttp.StatusOK).Response()
}
