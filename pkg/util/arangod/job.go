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

package arangod

import (
	"context"
	"net/http"
	"path"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// IsJobFinished checks whether a given job is finished.
// Returns no error if it is finished.
func IsJobFinished(ctx context.Context, cli driver.Client, jobID string) error {
	conn := cli.Connection()

	req, err := conn.NewRequest(http.MethodPut, path.Join("_api/job", jobID))
	if err != nil {
		return errors.WithStack(err)
	}

	resp, err := conn.Do(ctx, req)
	if err != nil {
		return errors.WithStack(err)
	}

	// 200 means that the job is finished.
	//   It is not possible to get 200 twice from the same job ID, because job is removed altogether with this request.
	// 204 means that the job is still in a pending queue or not finished yet.
	// 400 means that the job ID is not provided.
	// 404 means that the job is gone (deleted or the result was fetched beforehand).
	if err := resp.CheckStatus(200); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
