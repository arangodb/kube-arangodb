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

package agency

import (
	"context"
	goHttp "net/http"

	adbDriverV2Connection "github.com/arangodb/go-driver/v2/connection"

	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
)

func GetAgencyConfig(ctx context.Context, connection adbDriverV2Connection.Connection) (Config, error) {
	return arangod.GetRequest[Config](ctx, connection, "/_api/agency/config").Do(ctx).AcceptCode(goHttp.StatusOK).Response()
}

type Config struct {
	LeaderId string `json:"leaderId"`

	CommitIndex uint64 `json:"commitIndex"`

	Configuration struct {
		ID string `json:"id"`
	} `json:"configuration"`
}
