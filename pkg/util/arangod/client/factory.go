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

package client

import (
	"context"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/connection"
)

func NewFactory(auth Authentication, client HTTPClient) Factory {
	return factory{
		auth:   auth,
		client: client,
	}
}

type Factory interface {
	Connection(ctx context.Context, hosts ...string) (connection.Connection, error)
	Client(ctx context.Context, hosts ...string) (arangodb.Client, error)
}

type factory struct {
	auth   Authentication
	client HTTPClient
}

func (f factory) Connection(ctx context.Context, hosts ...string) (connection.Connection, error) {
	client, err := f.client.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	connConfig := connection.HttpConfiguration{
		Transport: client,
		Endpoint:  connection.NewRoundRobinEndpoints(hosts),
	}

	if auth, ok, err := f.auth.Authentication(ctx); err != nil {
		return nil, err
	} else if ok {
		connConfig.Authentication = auth
	}

	return connection.NewHttpConnection(connConfig), nil
}

func (f factory) Client(ctx context.Context, hosts ...string) (arangodb.Client, error) {
	conn, err := f.Connection(ctx, hosts...)
	if err != nil {
		return nil, err
	}

	return arangodb.NewClient(conn), nil
}
