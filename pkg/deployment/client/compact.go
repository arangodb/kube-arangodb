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

type CompactRequest struct {
	CompactBottomMostLevel *bool `json:"compactBottomMostLevel,omitempty"`
	ChangeLevel            *bool `json:"changeLevel,omitempty"`
}

const CompactUrl = "/_admin/compact"

func (c *client) Compact(ctx context.Context, request *CompactRequest) error {
	req, err := c.c.NewRequest(goHttp.MethodPut, CompactUrl)
	if err != nil {
		return err
	}

	if request == nil {
		request = new(CompactRequest)
	}

	req, err = req.SetBody(request)
	if err != nil {
		return err
	}

	resp, err := c.c.Do(ctx, req)
	if err != nil {
		return err
	}

	if err := resp.CheckStatus(goHttp.StatusOK); err != nil {
		return err
	}

	return nil
}
