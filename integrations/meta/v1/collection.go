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
	"time"

	"github.com/arangodb/go-driver/v2/arangodb"

	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

func withTTLIndex(in cache.Object[arangodb.Collection]) cache.Object[arangodb.Collection] {
	return cache.NewObject(func(ctx context.Context) (arangodb.Collection, time.Duration, error) {
		col, err := in.Get(ctx)
		if err != nil {
			return nil, 0, err
		}

		if _, _, err := col.EnsureTTLIndex(ctx, []string{"ttl"}, 0, &arangodb.CreateTTLIndexOptions{
			Name: "system_meta_store_object_ttl",
		}); err != nil {
			println(err.Error())
			return nil, 0, err
		}

		return col, time.Hour, nil
	})
}
