//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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
	"fmt"
	"net/http"
	"time"
)

const DeleteExpiredJobsURL = "/_api/job/expired"

func (c *client) DeleteExpiredJobs(ctx context.Context, timeout time.Duration) error {
	req, err := c.c.NewRequest(http.MethodDelete, DeleteExpiredJobsURL)
	if err != nil {
		return err
	}

	req.SetQuery("stamp", fmt.Sprintf("%d", time.Now().UTC().Add(-1*timeout).Unix()))

	resp, err := c.c.Do(ctx, req)
	if err != nil {
		return err
	}

	if err := resp.CheckStatus(http.StatusOK); err != nil {
		return err
	}

	return nil
}
