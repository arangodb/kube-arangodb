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

package v1

import (
	"context"
	"fmt"
	goStrings "strings"
	"testing"
	"time"

	"github.com/dchest/uniuri"

	"github.com/arangodb/go-driver/v2/arangodb"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func GetCacheObjectForKVStore(t *testing.T) cache.Object[arangodb.Collection] {
	client := tests.TestArangoDBConfig(t).Client(t)
	return withTTLIndex(cache.NewObject(func(ctx context.Context) (arangodb.Collection, time.Duration, error) {
		db, err := client.CreateDatabase(t.Context(), fmt.Sprintf("db-%s", goStrings.ToLower(uniuri.NewLen(6))), &arangodb.CreateDatabaseOptions{})
		if err != nil {
			return nil, 0, err
		}

		col, err := db.CreateCollectionV2(t.Context(), "_meta", &arangodb.CreateCollectionPropertiesV2{
			IsSystem: util.NewType(true),
		})
		if err != nil {
			return nil, 0, err
		}

		_, err = col.Properties(t.Context())
		if err != nil {
			return nil, 0, err
		}

		return col, time.Hour, nil
	}))
}

func GetInternalRemoteCache(t *testing.T) cache.RemoteCache[*Object] {
	return cache.NewRemoteCache[*Object](withTTLIndex(GetCacheObjectForKVStore(t)))
}
