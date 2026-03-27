//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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
	goHttp "net/http"

	adbDriverV2 "github.com/arangodb/go-driver/v2/arangodb"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// IsServerAvailable returns true when server is available.
// In active fail-over mode one of the server should be available.
func IsServerAvailable(ctx context.Context, c adbDriverV2.Client) (bool, error) {
	resp := GetRequest[any](ctx, c.Connection(), "_admin", "server", "availability").Do(ctx)

	if err := resp.AcceptCode(goHttp.StatusOK, goHttp.StatusServiceUnavailable).Evaluate(); err != nil {
		return false, errors.WithStack(err)
	}

	return resp.Code() == goHttp.StatusOK, nil
}
