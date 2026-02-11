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

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

type Collection interface {
	Database() Database

	WithTTLIndex(name string, ttl time.Duration, path ...string) Collection
	WithUniqueIndex(name string, path ...string) Collection

	Get() cache.Object[arangodb.Collection]
}

type collection struct {
	cache cache.Object[arangodb.Collection]
}

func (c collection) WithUniqueIndex(name string, path ...string) Collection {
	return collection{
		cache: cache.NewObject(func(ctx context.Context) (arangodb.Collection, time.Duration, error) {
			col, err := c.cache.Get(ctx)
			if err != nil {
				return nil, 0, err
			}

			if _, _, err := col.EnsurePersistentIndex(ctx, path, &arangodb.CreatePersistentIndexOptions{
				Name:   name,
				Unique: util.NewType(true),
			}); err != nil {
				return nil, 0, err
			}

			return col, DefaultTTL, nil
		}),
	}
}

func (c collection) WithTTLIndex(name string, ttl time.Duration, path ...string) Collection {
	return collection{
		cache: cache.NewObject(func(ctx context.Context) (arangodb.Collection, time.Duration, error) {
			col, err := c.cache.Get(ctx)
			if err != nil {
				return nil, 0, err
			}

			if _, _, err := col.EnsureTTLIndex(ctx, path, int(ttl/time.Second), &arangodb.CreateTTLIndexOptions{
				Name: name,
			}); err != nil {
				return nil, 0, err
			}

			return col, DefaultTTL, nil
		}),
	}
}

func (c collection) Database() Database {
	return database{
		cache: cache.NewObject(func(ctx context.Context) (arangodb.Database, time.Duration, error) {
			col, err := c.cache.Get(ctx)
			if err != nil {
				return nil, 0, err
			}

			return col.Database(), DefaultTTL, nil
		}),
	}
}

func (c collection) Get() cache.Object[arangodb.Collection] {
	return c.cache
}
