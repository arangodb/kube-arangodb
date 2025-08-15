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

	pbInventoryV1 "github.com/arangodb/kube-arangodb/integrations/inventory/v1/definition"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

func (c *client) Inventory(ctx context.Context) (*pbInventoryV1.Inventory, error) {
	req, err := c.c.NewRequest(goHttp.MethodGet, utilConstants.EnvoyInventoryConfigDestination)
	if err != nil {
		return nil, err
	}

	resp, err := c.c.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := resp.CheckStatus(goHttp.StatusOK); err != nil {
		return nil, err
	}

	var l ugrpc.GRPC[*pbInventoryV1.Inventory]

	if err := resp.ParseBody("", &l); err != nil {
		return nil, err
	}

	return l.Object, nil
}
