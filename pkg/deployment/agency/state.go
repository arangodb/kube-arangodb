//
// DISCLAIMER
//
// Copyright 2023-2026 ArangoDB GmbH, Cologne, Germany
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
	goHttp "net/http"

	adbDriverV2Connection "github.com/arangodb/go-driver/v2/connection"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func GetAgencyState[T interface{}](ctx context.Context, connection adbDriverV2Connection.Connection) (T, error) {
	resp, err := arangod.PostRequest[ReadRequest, []T](ctx, connection, GetAgencyReadRequestFields(), "/_api/agency/read").Do(ctx).AcceptCode(goHttp.StatusOK).Response()
	if err != nil {
		return util.Default[T](), err
	}

	if len(resp) != 1 {
		return util.Default[T](), errors.Errorf("Invalid response size")
	}

	return (resp)[0], nil
}
