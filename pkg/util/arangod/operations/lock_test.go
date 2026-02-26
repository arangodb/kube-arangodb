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
	"fmt"
	goStrings "strings"
	"testing"

	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/require"

	"github.com/arangodb/go-driver/v2/arangodb"

	"github.com/arangodb/kube-arangodb/pkg/util/arangod/db"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_ArangoDB_Lock(t *testing.T) {
	database, err := db.NewClient(cache.NewObject(tests.TestArangoDBConfig(t).ClientCache())).
		CreateDatabase(fmt.Sprintf("db-%s", goStrings.ToLower(uniuri.NewLen(6))), &arangodb.CreateDatabaseOptions{}).
		CreateCollection("testing", db.StaticProps(arangodb.CreateCollectionPropertiesV2{})).Database().Get().Get(t.Context())
	require.NoError(t, err)

	type documentType struct {
		Key  string `json:"_key"`
		Data string `json:"data"`
	}

	documentKey := fmt.Sprintf("document-%s", goStrings.ToLower(uniuri.NewLen(6)))

	getDocument := func(t *testing.T) documentType {
		col, err := database.GetCollection(t.Context(), "testing", &arangodb.GetCollectionOptions{})
		require.NoError(t, err)

		var z documentType

		_, err = col.ReadDocument(t.Context(), documentKey, &z)
		require.NoError(t, err)
		return z
	}

	updateDocument := func(t *testing.T, c arangodb.Transaction, d documentType) {
		col, err := c.GetCollection(t.Context(), "testing", &arangodb.GetCollectionOptions{})
		require.NoError(t, err)

		_, err = col.UpdateDocument(t.Context(), documentKey, &d)
		require.NoError(t, err)
	}

	t.Run("Create", func(t *testing.T) {
		_, err := WithTransaction[documentType](t.Context(), database, arangodb.TransactionCollections{Write: []string{"testing"}}, nil, func(ctx context.Context, c arangodb.Transaction) (documentType, error) {

			col, err := c.GetCollection(t.Context(), "testing", &arangodb.GetCollectionOptions{})
			require.NoError(t, err)

			var z documentType

			z.Key = documentKey

			_, err = col.CreateDocument(t.Context(), &z)
			require.NoError(t, err)

			return z, nil
		})
		require.NoError(t, err)

		d := getDocument(t)
		require.Equal(t, documentKey, d.Key)
		require.Equal(t, "", d.Data)
	})

	t.Run("First", func(t *testing.T) {
		_, err := WithTransaction[documentType](t.Context(), database, arangodb.TransactionCollections{Write: []string{"testing"}}, nil, func(ctx context.Context, c arangodb.Transaction) (documentType, error) {
			var d documentType

			d.Key = documentKey
			d.Data = "1"

			updateDocument(t, c, d)

			return d, nil
		})
		require.NoError(t, err)

		d := getDocument(t)
		require.Equal(t, documentKey, d.Key)
		require.Equal(t, "1", d.Data)
	})

	t.Run("Second", func(t *testing.T) {
		_, err := WithTransaction[documentType](t.Context(), database, arangodb.TransactionCollections{Write: []string{"testing"}}, nil, func(ctx context.Context, c arangodb.Transaction) (documentType, error) {
			var d documentType

			d.Key = documentKey
			d.Data = "2"

			updateDocument(t, c, d)

			return d, nil
		})
		require.NoError(t, err)

		d := getDocument(t)
		require.Equal(t, documentKey, d.Key)
		require.Equal(t, "2", d.Data)
	})

	t.Run("With error", func(t *testing.T) {
		_, err := WithTransaction[documentType](t.Context(), database, arangodb.TransactionCollections{Write: []string{"testing"}}, nil, func(ctx context.Context, c arangodb.Transaction) (documentType, error) {
			var d documentType

			d.Key = documentKey
			d.Data = "3"

			updateDocument(t, c, d)

			return d, errors.Errorf("")
		})
		require.Error(t, err)

		d := getDocument(t)
		require.Equal(t, documentKey, d.Key)
		require.Equal(t, "2", d.Data)
	})

	t.Run("With Init Lock", func(t *testing.T) {
		_, err := WithTransaction[string](t.Context(), database, arangodb.TransactionCollections{Write: []string{"testing"}}, nil, func(ctx context.Context, c arangodb.Transaction) (string, error) {
			_, err := WithLock[string]("testing", func(ctx context.Context, c arangodb.Transaction, lock *LockDocument) (string, error) {
				return "", nil
			})(ctx, c)
			require.NoError(t, err)
			return "", nil
		})
		require.NoError(t, err)
	})

	t.Run("With Lock", func(t *testing.T) {
		_, err := WithTransaction[string](t.Context(), database, arangodb.TransactionCollections{Write: []string{"testing"}}, nil, func(ctx context.Context, c arangodb.Transaction) (string, error) {
			_, err := WithLock[string]("testing", func(ctx context.Context, c arangodb.Transaction, lock *LockDocument) (string, error) {
				return "", nil
			})(ctx, c)
			require.NoError(t, err)

			_, err = WithLock[string]("testing", func(ctx context.Context, c arangodb.Transaction, lock *LockDocument) (string, error) {
				return "", nil
			})(ctx, c)
			require.NoError(t, err)

			_, err = WithLock[string]("testing", func(ctx context.Context, c arangodb.Transaction, lock *LockDocument) (string, error) {
				_, err := WithTransaction[string](t.Context(), database, arangodb.TransactionCollections{Write: []string{"testing"}}, nil, WithLock[string]("testing", func(ctx context.Context, c arangodb.Transaction, lock *LockDocument) (string, error) {
					return "", nil
				}))
				require.Error(t, err)
				require.ErrorIs(t, err, AlreadyLocked{})
				return "", nil
			})(ctx, c)
			require.NoError(t, err)

			return "", nil
		})
		require.NoError(t, err)
	})
}
