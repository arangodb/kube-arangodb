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

package agency

import (
	"context"
	"net/http"

	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func GetAgencyState(ctx context.Context, connection conn.Connection) (state.Root, error) {
	resp, code, err := conn.NewExecutor[ReadRequest, state.Roots](connection).Execute(ctx, http.MethodPost, "/_api/agency/read", GetAgencyReadRequestFields())
	if err != nil {
		return state.Root{}, err
	}

	if code != http.StatusOK {
		return state.Root{}, errors.Newf("Unknown response code %d", code)
	}

	if resp == nil {
		return state.Root{}, errors.Newf("Missing response body")
	}

	if len(*resp) != 1 {
		return state.Root{}, errors.Newf("Invalid response size")
	}

	return (*resp)[0], nil
}
