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

	adbDriverV2Connection "github.com/arangodb/go-driver/v2/connection"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func GetAgencyPoll[T interface{}](ctx context.Context, connection adbDriverV2Connection.Connection, index uint64, timeout time.Duration) (Response[T], error) {
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
	resp, err := arangod.GetRequest[Response[T]](ctx, connection, url).Do(ctx).AcceptCode(goHttp.StatusOK).Response()
	if err != nil {
		return util.Default[Response[T]](), errors.WithStack(err)
	}
	return resp, nil
}
