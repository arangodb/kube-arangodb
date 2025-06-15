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

package session

import (
	"context"
	"fmt"
	"time"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/arangodb/shared"
	"github.com/arangodb/go-driver/v2/connection"

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
)

func NewAuthClientFetcherObject(configuration pbImplEnvoyAuthV3Shared.Configuration) cache.Object[pbAuthenticationV1.AuthenticationV1Client] {
	return cache.NewObject[pbAuthenticationV1.AuthenticationV1Client](configuration.GetAuthClientFetcher)
}

func NewConnectionObject(configuration pbImplEnvoyAuthV3Shared.Configuration, auth cache.Object[pbAuthenticationV1.AuthenticationV1Client]) cache.Object[arangodb.Collection] {
	return cache.NewObject(func(ctx context.Context) (arangodb.Collection, time.Duration, error) {
		ac, err := auth.Get(ctx)
		if err != nil {
			return nil, 0, err
		}

		client := arangodb.NewClient(connection.NewHttpConnection(connection.HttpConfiguration{
			Authentication: pbAuthenticationV1.NewRootRequestModifier(ac),
			Endpoint: connection.NewRoundRobinEndpoints([]string{
				fmt.Sprintf("%s://%s:%d", configuration.Database.Proto, configuration.Database.Endpoint, configuration.Database.Port),
			}),
			ContentType:    connection.ApplicationJSON,
			ArangoDBConfig: connection.ArangoDBConfiguration{},
			Transport:      operatorHTTP.RoundTripperWithShortTransport(operatorHTTP.WithTransportTLS(operatorHTTP.Insecure)),
		}))

		db, err := client.GetDatabase(ctx, "_system", nil)
		if err != nil {
			return nil, 0, err
		}

		if _, err := db.GetCollection(ctx, "_gateway_session", nil); err != nil {
			if !shared.IsNotFound(err) {
				return nil, 0, err
			}

			if _, err := db.CreateCollectionWithOptions(ctx, "_gateway_session", &arangodb.CreateCollectionProperties{
				IsSystem: true,
			}, nil); err != nil {
				if !shared.IsConflict(err) {
					return nil, 0, err
				}
			}
		}

		col, err := db.GetCollection(ctx, "_gateway_session", nil)
		if err != nil {
			return nil, 0, err
		}

		return col, 24 * time.Hour, nil
	})
}
