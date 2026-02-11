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

package db

import (
	"context"
	"time"

	"github.com/arangodb/go-driver/v2/arangodb"

	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

func NewClient(c cache.Object[arangodb.Client]) Client {
	return client{c}
}

type Client interface {
	CreateDatabase(name string, options *arangodb.CreateDatabaseOptions) Database
	Database(name string) Database

	Get() cache.Object[arangodb.Client]
}

type client struct {
	cache cache.Object[arangodb.Client]
}

func (c client) CreateDatabase(name string, options *arangodb.CreateDatabaseOptions) Database {
	return database{
		cache: cache.NewObject(func(ctx context.Context) (arangodb.Database, time.Duration, error) {
			client, err := c.cache.Get(ctx)
			if err != nil {
				return nil, 0, err
			}
			db, err := client.CreateDatabase(ctx, name, options)
			if err != nil {
				return nil, 0, err
			}

			return db, DefaultTTL, nil
		}),
	}
}

func (c client) Database(name string) Database {
	return database{
		cache: cache.NewObject(func(ctx context.Context) (arangodb.Database, time.Duration, error) {
			client, err := c.cache.Get(ctx)
			if err != nil {
				return nil, 0, err
			}
			db, err := client.GetDatabase(ctx, name, &arangodb.GetDatabaseOptions{SkipExistCheck: true})
			if err != nil {
				return nil, 0, err
			}

			return db, DefaultTTL, nil
		}),
	}
}

func (c client) Get() cache.Object[arangodb.Client] {
	return c.cache
}
