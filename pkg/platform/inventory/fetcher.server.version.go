//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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

package inventory

import (
	"context"
	goHttp "net/http"

	adbDriverV2 "github.com/arangodb/go-driver/v2/arangodb"
	adbDriverV2Connection "github.com/arangodb/go-driver/v2/connection"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/executor"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

func init() {
	global.MustRegister("server.info", func(conn adbDriverV2Connection.Connection, cfg *Configuration, out chan<- *Item) executor.RunFunc {
		return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
			resp, err := arangod.GetRequestWithTimeout[adbDriverV2.VersionInfo](ctx, globals.GetGlobals().Timeouts().ArangoD().Get(), conn, "_api", "version").
				AcceptCode(goHttp.StatusOK).
				Response()
			if err != nil {
				return err
			}

			return errors.Errors(
				Produce[string](out, "ARANGO_DEPLOYMENT", map[string]string{
					"detail": "version",
				}, string(resp.Version)),
				Produce(out, "ARANGO_DEPLOYMENT", map[string]string{
					"detail": "license",
				}, resp.License),
			)
		}
	})
}
