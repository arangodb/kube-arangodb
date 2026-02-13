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

package authorization

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/arangodb/shared"

	"github.com/arangodb/kube-arangodb/pkg/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

type LockFunc[T proto.Message] func(ctx context.Context, c arangodb.Transaction, lock *types.LockDocument) (T, error)

func WithLock[T proto.Message](collection string, process LockFunc[T]) TransactionFunc[T] {
	return func(ctx context.Context, c arangodb.Transaction) (T, error) {
		col, err := c.GetCollection(ctx, collection, &arangodb.GetCollectionOptions{SkipExistCheck: true})
		if err != nil {
			return util.Default[T](), errors.Wrapf(err, "Failed to get collection %s", collection)
		}

		var lock ugrpc.Object[*types.LockDocument]

		if _, err := col.ReadDocument(ctx, types.LockDocumentID, &lock); err != nil {
			if !shared.IsNotFound(err) {
				// Unable to fetch document
				return util.Default[T](), errors.Wrapf(err, "Failed to get lock for collection %s", collection)
			}

			lock = ugrpc.NewObject[*types.LockDocument](&types.LockDocument{
				Key: types.LockDocumentID,
			})

			if _, err := col.CreateDocument(ctx, lock); err != nil {
				return util.Default[T](), errors.Wrapf(err, "Failed to create lock for collection %s", collection)
			}

			if _, err := col.ReadDocument(ctx, types.LockDocumentID, &lock); err != nil {
				return util.Default[T](), errors.Wrapf(err, "Failed to get lock for collection %s", collection)
			}
		}

		lock.Object.Lock = string(uuid.NewUUID())

		if _, err := col.UpdateDocument(ctx, types.LockDocumentID, lock); err != nil {
			return util.Default[T](), AlreadyLocked{}
		}

		ret, err := process(ctx, c, lock.Object)
		if err != nil {
			return util.Default[T](), errors.Wrapf(err, "Failed to process")
		}

		if _, err := col.UpdateDocument(ctx, types.LockDocumentID, lock); err != nil {
			return util.Default[T](), err
		}

		return ret, nil
	}
}
