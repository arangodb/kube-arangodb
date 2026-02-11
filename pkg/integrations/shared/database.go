//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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

package shared

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/connection"

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/db"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
)

type Database struct {
	Proto    string
	Endpoint string
	Port     int
	Database string

	Source DatabaseSource
}

type DatabaseSource struct {
	Collection string
}

func (d *Database) Validate() error {
	if d == nil {
		return errors.Errorf("Database Ref is empty")
	}

	if d.Endpoint == "" {
		return errors.Errorf("Database Endpoint is empty")
	}

	if d.Proto == "" {
		return errors.Errorf("Database Proto is empty")
	}

	if d.Database == "" {
		return errors.Errorf("Database Database is empty")
	}
	if d.Source.Collection == "" {
		return errors.Errorf("Database Source Collection is empty")
	}

	return nil
}

func (d *Database) New(cmd *cobra.Command) error {
	if d == nil {
		return errors.Errorf("Database Ref is empty")
	}

	f := cmd.Flags()

	dbE, err := f.GetString("database.endpoint")
	if err != nil {
		return err
	}

	dbP, err := f.GetString("database.proto")
	if err != nil {
		return err
	}

	dbPort, err := f.GetInt("database.port")
	if err != nil {
		return err
	}
	dbName, err := f.GetString("database.name")
	if err != nil {
		return err
	}
	dbSource, err := f.GetString("database.source")
	if err != nil {
		return err
	}

	*d = Database{
		Proto:    dbP,
		Endpoint: dbE,
		Port:     dbPort,
		Database: dbName,
		Source: DatabaseSource{
			Collection: dbSource,
		},
	}

	return d.Validate()
}

func (d *Database) DatabaseClient(endpoint Endpoint) cache.Object[arangodb.Client] {
	auth := endpoint.AuthClient()

	return cache.NewObject(func(ctx context.Context) (arangodb.Client, time.Duration, error) {
		if d == nil {
			return nil, 0, errors.Errorf("Database Ref is empty")
		}

		ac, err := auth.Get(ctx)
		if err != nil {
			return nil, 0, err
		}

		client := arangodb.NewClient(connection.NewHttpConnection(connection.HttpConfiguration{
			Authentication: pbAuthenticationV1.NewRootRequestModifier(ac),
			Endpoint: connection.NewRoundRobinEndpoints([]string{
				fmt.Sprintf("%s://%s:%d", d.Proto, d.Endpoint, d.Port),
			}),
			ContentType:    connection.ApplicationJSON,
			ArangoDBConfig: connection.ArangoDBConfiguration{},
			Transport:      operatorHTTP.RoundTripperWithShortTransport(operatorHTTP.WithTransportTLS(operatorHTTP.Insecure)),
		}))

		return client, time.Hour, nil
	})
}

func (d *Database) WithDatabase(endpoint Endpoint) db.Database {
	return db.NewClient(d.DatabaseClient(endpoint)).Database(d.Database)
}

func (d *Database) SourceCollectionProps() db.CollectionProps {
	return db.SourceCollectionProps(d.Source.Collection)
}
