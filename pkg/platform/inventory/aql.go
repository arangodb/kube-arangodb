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
	"time"

	adbDriverV2 "github.com/arangodb/go-driver/v2/arangodb"
	adbDriverV2Shared "github.com/arangodb/go-driver/v2/arangodb/shared"
	adbDriverV2Connection "github.com/arangodb/go-driver/v2/connection"

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

// ExecuteBasicAQLIfExists runs the query only when the given collection exists in the database, and is
// otherwise a no-op. It keeps the inventory resilient to optional collections (e.g. _events, which only
// exists when the collector is enabled).
func ExecuteBasicAQLIfExists(db, collection, aql string, bind map[string]any) Executor {
	inner := ExecuteAQL(db, aql, bind, false)

	return func(conn adbDriverV2Connection.Connection, cfg *Configuration, out chan<- *Item) executor.RunFunc {
		return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
			d, err := adbDriverV2.NewClient(conn).GetDatabase(ctx, db, nil)
			if err != nil {
				return err
			}

			exists, err := d.CollectionExists(ctx, collection)
			if err != nil {
				return err
			}

			if !exists {
				log.Str("collection", collection).Info("Collection not present, skipping")
				return nil
			}

			return inner(conn, cfg, out)(ctx, log, t, h)
		}
	}
}

func ExecuteAQL(db string, aql string, bind map[string]any, telemetry bool) Executor {
	return func(conn adbDriverV2Connection.Connection, cfg *Configuration, out chan<- *Item) executor.RunFunc {
		return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
			if telemetry {
				if !cfg.WithTelemetry() {
					log.Info("Telemetry disabled")
					return nil
				}
				log.Info("Collecting Telemetry details")
			}

			c := adbDriverV2.NewClient(adbDriverV2Connection.NewConnectionAsyncWrapper(conn))

			d, err := c.GetDatabase(ctx, db, nil)
			if err != nil {
				return err
			}

			nctx := adbDriverV2Connection.WithAsync(ctx)

			_, err = d.Query(nctx, aql, &adbDriverV2.QueryOptions{
				BindVars: bind,
			})
			if err == nil {
				return errors.Errorf("Async execution of the query should be prepared")
			}

			jobId, ok := adbDriverV2Connection.IsAsyncJobInProgress(err)
			if !ok {
				return errors.Wrapf(err, "Async execution of the query should be prepared")
			}

			var cursor adbDriverV2.Cursor

			for {
				zctx := adbDriverV2Connection.WithAsyncID(ctx, jobId)

				query, err := d.Query(zctx, aql, &adbDriverV2.QueryOptions{
					BindVars: bind,
				})
				if err != nil {
					_, ok := adbDriverV2Connection.IsAsyncJobInProgress(err)
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
					if adbDriverV2Shared.IsNoMoreDocuments(err) {
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
