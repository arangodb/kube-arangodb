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

package sidecar

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/arangodb/go-driver/v2/arangodb"

	"github.com/arangodb/kube-arangodb/pkg/util/arangod/client"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/http"
)

func arangoDBDatabaseAuth(cmd *cobra.Command) cache.Object[client.Authentication] {
	return cache.NewObject(func(ctx context.Context) (client.Authentication, time.Duration, error) {
		path, err := flagAuth.Get(cmd)
		if err != nil {
			return nil, 0, err
		}
		return client.FolderArangoDBAuthentication(path), 30 * time.Second, nil
	})
}

func arangoDBDatabaseClient(cmd *cobra.Command) cache.Object[arangodb.Client] {
	auth := arangoDBDatabaseAuth(cmd)

	return cache.NewObject(func(ctx context.Context) (arangodb.Client, time.Duration, error) {
		auth, err := auth.Get(ctx)
		if err != nil {
			return nil, 0, err
		}

		addr, err := flagArangodb.Get(cmd)
		if err != nil {
			return nil, 0, err
		}

		client, err := client.NewFactory(auth, client.HTTPClientFactory(
			http.ShortTransport(),
			http.WithTransportTLS(http.Insecure),
		)).Client(cmd.Context(), addr)

		if err != nil {
			return nil, 0, err
		}
		return client, time.Hour, nil
	})
}
