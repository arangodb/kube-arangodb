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

package inventory

import (
	"context"
	goHttp "net/http"

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/executor"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

func init() {
	global.MustRegister("server.mode", func(conn driver.Connection, cfg *Configuration, out chan<- *Item) executor.RunFunc {
		return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
			resp := arangod.GetRequestWithTimeout[driver.ClusterHealth](ctx, globals.GetGlobals().Timeouts().ArangoD().Get(), conn, "_admin", "cluster", "health")

			if resp.Code() == goHttp.StatusOK {
				return errors.Errors(
					Produce[string](out, "ARANGO_DEPLOYMENT", map[string]string{
						"detail": "mode",
					}, "CLUSTER"),
				)
			}

			if resp.Code() == goHttp.StatusForbidden {
				return errors.Errors(
					Produce[string](out, "ARANGO_DEPLOYMENT", map[string]string{
						"detail": "mode",
					}, "SINGLE"),
				)
			}

			return resp.AcceptCode(goHttp.StatusOK).Evaluate()
		}
	})
}
