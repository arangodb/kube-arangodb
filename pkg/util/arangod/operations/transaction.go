//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package operations

import (
	"context"
	"time"

	"github.com/arangodb/go-driver/v2/arangodb"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type TransactionFunc[T any] func(ctx context.Context, c arangodb.Transaction) (T, error)

func WithTransaction[T any](ctx context.Context, db arangodb.Database, cols arangodb.TransactionCollections, opts *arangodb.BeginTransactionOptions, process TransactionFunc[T]) (T, error) {
	if opts == nil {
		opts = &arangodb.BeginTransactionOptions{}
	}

	// Enforce sync
	opts.WaitForSync = true
	opts.LockTimeoutDuration = 10 * time.Second

	tx, err := db.BeginTransaction(ctx, cols, opts)
	if err != nil {
		return util.Default[T](), err
	}

	out, err := process(ctx, tx)
	if err != nil {
		// New context is required for abort
		actx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		if cerr := tx.Abort(actx, &arangodb.AbortTransactionOptions{}); cerr != nil {
			return util.Default[T](), errors.Errors(err, cerr)
		}

		return util.Default[T](), errors.Wrapf(err, "Processing failed")
	}

	if err := tx.Commit(ctx, &arangodb.CommitTransactionOptions{}); err != nil {
		return util.Default[T](), errors.Wrapf(err, "Committing failed")
	}

	return out, nil
}
