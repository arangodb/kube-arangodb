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
	"time"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/util/connection/wrappers/async"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/executor"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

func ExecuteTelemetryAQL(db string, aql string, bind map[string]any) Executor {
	return ExecuteAQL(db, aql, bind, true)
}

func ExecuteBasicAQL(db string, aql string, bind map[string]any) Executor {
	return ExecuteAQL(db, aql, bind, false)
}

func ExecuteAQL(db string, aql string, bind map[string]any, telemetry bool) Executor {
	return func(conn driver.Connection, cfg *Configuration, out chan<- *Item) executor.RunFunc {
		return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
			if telemetry {
				if !cfg.WithTelemetry() {
					log.Info("Telemetry disabled")
					return nil
				}
				log.Info("Collecting Telemetry details")
			}

			c, err := driver.NewClient(driver.ClientConfig{Connection: async.NewConnectionAsyncWrapper(conn)})
			if err != nil {
				return err
			}

			d, err := c.Database(ctx, db)
			if err != nil {
				return err
			}

			nctx := driver.WithAsync(ctx)

			_, err = d.Query(nctx, aql, bind)
			if err == nil {
				return errors.Errorf("Async execution of the query should be prepared")
			}

			jobId, ok := async.IsAsyncJobInProgress(err)
			if !ok {
				return errors.Wrapf(err, "Async execution of the query should be prepared")
			}

			var cursor driver.Cursor

			for {
				zctx := driver.WithAsyncID(ctx, jobId)

				query, err := d.Query(zctx, aql, bind)
				if err != nil {
					_, ok := async.IsAsyncJobInProgress(err)
					if !ok {
						return errors.Wrapf(err, "Async execution of the query should be prepared")
					}

					t.Wait(125 * time.Millisecond)

					continue
				}

				cursor = query
				break
			}

			for {
				var ret ugrpc.Object[*Item]

				if _, err := cursor.ReadDocument(ctx, &ret); err != nil {
					if driver.IsNoMoreDocuments(err) {
						break
					}

					return err
				}

				if err := ret.Object.Validate(); err != nil {
					return err
				}

				out <- ret.Object
			}

			return nil
		}
	}
}
