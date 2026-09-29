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

package cache

import (
	"context"
	"net/url"
	"time"

	adbDriverV2 "github.com/arangodb/go-driver/v2/arangodb"
	adbDriverV2Shared "github.com/arangodb/go-driver/v2/arangodb/shared"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// MoveResult describes the outcome of a RemoteCache.Move call.
type MoveResult int

const (
	// MoveResultMoved indicates the object has been moved successfully.
	MoveResultMoved MoveResult = iota
	// MoveResultSourceNotFound indicates the source key does not exist.
	MoveResultSourceNotFound
	// MoveResultRevisionConflict indicates the source revision precondition was not met.
	MoveResultRevisionConflict
	// MoveResultDestinationExists indicates the destination key already exists.
	MoveResultDestinationExists
)

// moveLockTimeout defines the transaction lock timeout for the Move operation.
const moveLockTimeout = 10 * time.Second

func (r *remoteCache[T]) Move(ctx context.Context, from, to, rev string) (MoveResult, error) {
	col, err := r.collection.Get(ctx)
	if err != nil {
		return MoveResultMoved, err
	}

	tx, err := col.Database().BeginTransaction(ctx, adbDriverV2.TransactionCollections{
		Read:  []string{col.Name()},
		Write: []string{col.Name()},
	}, &adbDriverV2.BeginTransactionOptions{
		WaitForSync:         true,
		LockTimeoutDuration: moveLockTimeout,
	})
	if err != nil {
		return MoveResultMoved, err
	}

	result, err := r.moveInTransaction(ctx, tx, col.Name(), from, to, rev)
	if err != nil || result != MoveResultMoved {
		// Nothing must be persisted unless the move fully succeeded.
		// A fresh context is required for the abort as ctx might be cancelled.
		actx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		if aerr := tx.Abort(actx, &adbDriverV2.AbortTransactionOptions{}); aerr != nil {
			return MoveResultMoved, errors.Errors(err, aerr)
		}

		return result, err
	}

	if err := tx.Commit(ctx, &adbDriverV2.CommitTransactionOptions{}); err != nil {
		return MoveResultMoved, err
	}

	// Keep the local cache consistent with the store.
	r.cache.Invalidate(from)
	r.cache.Invalidate(to)

	return MoveResultMoved, nil
}

// moveInTransaction performs the read/create/delete sequence within the given transaction.
// The document is copied verbatim (including the possibly encrypted payload) under the new key,
// so no re-encryption or type-specific handling is required.
func (r *remoteCache[T]) moveInTransaction(ctx context.Context, tx adbDriverV2.Transaction, colName, from, to, rev string) (MoveResult, error) {
	col, err := tx.GetCollection(ctx, colName, &adbDriverV2.GetCollectionOptions{SkipExistCheck: true})
	if err != nil {
		return MoveResultMoved, err
	}

	var raw map[string]interface{}

	readOpts := &adbDriverV2.CollectionDocumentReadOptions{}
	if rev != "" {
		readOpts.IfMatch = rev
	}

	if _, err := col.ReadDocumentWithOptions(ctx, url.QueryEscape(from), &raw, readOpts); err != nil {
		if adbDriverV2Shared.IsPreconditionFailed(err) {
			return MoveResultRevisionConflict, nil
		}

		if adbDriverV2Shared.IsNotFound(err) {
			return MoveResultSourceNotFound, nil
		}

		return MoveResultMoved, err
	}

	// Drop identity fields and assign the destination key.
	delete(raw, "_id")
	delete(raw, "_rev")
	raw["_key"] = to

	if _, err := col.CreateDocument(ctx, raw); err != nil {
		if adbDriverV2Shared.IsConflict(err) {
			return MoveResultDestinationExists, nil
		}

		return MoveResultMoved, err
	}

	if _, err := col.DeleteDocument(ctx, url.QueryEscape(from)); err != nil {
		return MoveResultMoved, err
	}

	return MoveResultMoved, nil
}
