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

package cache

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver/v2/arangodb"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
)

func (r *remoteCache[T]) List(ctx context.Context, size int, prefix string) (util.NextIterator[[]string], error) {
	col, err := r.collection.Get(ctx)
	if err != nil {
		return nil, err
	}

	db := col.Database()

	query := fmt.Sprintf("FOR doc IN %s", col.Name())
	bindVars := map[string]interface{}{}

	if prefix != "" {
		query += " FILTER doc._key LIKE CONCAT(@prefix, '%')"
		bindVars["prefix"] = prefix
	}

	query += " SORT doc._key RETURN doc._key"

	resp, err := db.Query(ctx, query, &arangodb.QueryOptions{
		BatchSize: size,
		BindVars:  bindVars,
	})
	if err != nil {
		return nil, err
	}

	return arangod.QueryV2NextIterator[string](resp, size), nil
}
