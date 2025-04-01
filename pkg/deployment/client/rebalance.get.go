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
	goHttp "net/http"
)

type RebalanceGetResponse struct {
	Result RebalanceGetResponseResult `json:"result,omitempty"`
}

type RebalanceGetResponseResult struct {
	PendingMoveShards int `json:"pendingMoveShards"`
	TodoMoveShards    int `json:"todoMoveShards"`
}

func (c *client) RebalanceGet(ctx context.Context) (RebalanceGetResponse, error) {
	req, err := c.c.NewRequest(goHttp.MethodGet, "/_admin/cluster/rebalance")
	if err != nil {
		return RebalanceGetResponse{}, err
	}

	resp, err := c.c.Do(ctx, req)
	if err != nil {
		return RebalanceGetResponse{}, err
	}

	if err := resp.CheckStatus(goHttp.StatusOK); err != nil {
		return RebalanceGetResponse{}, err
	}

	var d RebalanceGetResponse

	if err := resp.ParseBody("", &d); err != nil {
		return RebalanceGetResponse{}, err
	}

	return d, nil
}
