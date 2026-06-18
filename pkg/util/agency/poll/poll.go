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

package poll

import (
	"context"
	"fmt"
	goHttp "net/http"
	goStrings "strings"
	"time"

	"github.com/arangodb-helper/go-helper/pkg/arangod/conn"
	"github.com/arangodb-helper/go-helper/pkg/errors"
)

func GetAgencyPoll[T interface{}](ctx context.Context, connection conn.Connection, index uint64, timeout time.Duration) (Response[T], error) {
	var def Response[T]
	params := make([]string, 0, 2)
	if index > 0 {
		params = append(params, fmt.Sprintf("index=%d", index))
	}
	if timeout > 0 {
		params = append(params, fmt.Sprintf("timeout=%f", float64(timeout)/float64(time.Second)))
	} else if timeout == 0 {
		params = append(params, "timeout=0")
	}
	url := "/_api/agency/poll"
	if len(params) > 0 {
		url = fmt.Sprintf("%s?%s", url, goStrings.Join(params, "&"))
	}
	resp, code, err := conn.NewExecutor[any, Response[T]](connection).Execute(ctx, goHttp.MethodGet, url, nil)
	if err != nil {
		return def, errors.WithMessage(err, "generic Execute failed")
	}
	if code != goHttp.StatusOK {
		return def, errors.Newf("Unexpected response code %d", code)
	}
	if resp == nil {
		return def, errors.Newf("response object is nil")
	}
	return *resp, nil
}
