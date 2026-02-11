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
	"github.com/arangodb/go-driver/v2/arangodb/shared"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

type CollectionProps func(ctx context.Context, db arangodb.Database) (*arangodb.CreateCollectionPropertiesV2, error)

func StaticProps(props arangodb.CreateCollectionPropertiesV2) CollectionProps {
	return func(ctx context.Context, db arangodb.Database) (*arangodb.CreateCollectionPropertiesV2, error) {
		return &props, nil
	}
}

func SourceCollectionProps(name string) CollectionProps {
	return func(ctx context.Context, db arangodb.Database) (*arangodb.CreateCollectionPropertiesV2, error) {
		col, err := db.GetCollection(ctx, name, &arangodb.GetCollectionOptions{SkipExistCheck: true})
		if err != nil {
			return nil, err
		}

		prop, err := col.Properties(ctx)
		if err != nil {
			return nil, err
		}

		return &arangodb.CreateCollectionPropertiesV2{
			IsSystem:          util.NewType(prop.IsSystem),
			WriteConcern:      util.NewType(prop.WriteConcern),
			ReplicationFactor: util.NewType(prop.ReplicationFactor),
		}, nil
	}
}

type Database interface {
	CreateCollection(name string, props CollectionProps) Collection
	Collection(name string) Collection

	Get() cache.Object[arangodb.Database]
}

type database struct {
	cache cache.Object[arangodb.Database]
}

func (d database) CreateCollection(name string, props CollectionProps) Collection {
	return collection{
		cache: cache.NewObject(func(ctx context.Context) (arangodb.Collection, time.Duration, error) {
			db, err := d.cache.Get(ctx)
			if err != nil {
				return nil, 0, err
			}

			_, err = db.GetCollection(ctx, name, &arangodb.GetCollectionOptions{})
			if err != nil {
				if !shared.IsNotFound(err) {
					return nil, 0, err
				}

				opts, err := props(ctx, db)
				if err != nil {
					return nil, 0, err
				}

				if _, err := db.CreateCollectionV2(ctx, name, opts); err != nil {
					if !shared.IsConflict(err) {
						return nil, 0, err
					}
				}
			}

			col, err := db.GetCollection(ctx, name, &arangodb.GetCollectionOptions{})
			if err != nil {
				return nil, 0, err
			}

			return col, DefaultTTL, nil
		}),
	}
}

func (d database) Collection(name string) Collection {
	return collection{
		cache: cache.NewObject(func(ctx context.Context) (arangodb.Collection, time.Duration, error) {
			db, err := d.cache.Get(ctx)
			if err != nil {
				return nil, 0, err
			}

			col, err := db.GetCollection(ctx, name, &arangodb.GetCollectionOptions{})
			if err != nil {
				return nil, 0, err
			}

			return col, DefaultTTL, nil
		}),
	}
}

func (d database) Get() cache.Object[arangodb.Database] {
	return d.cache
}
