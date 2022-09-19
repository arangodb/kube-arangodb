//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package replication

import (
	"github.com/arangodb/arangosync-client/client"
	"github.com/arangodb/arangosync-client/tasks"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func GetSyncServerClient(clientCache *client.ClientCache, token string, source client.Endpoint) (client.API, error) {
	tlsAuth := tasks.TLSAuthentication{
		TLSClientAuthentication: tasks.TLSClientAuthentication{
			ClientToken: token,
		},
	}
	auth := client.NewAuthentication(tlsAuth, "")
	insecureSkipVerify := true
	c, err := clientCache.GetClient(client.NewExternalEndpoints(source), auth, insecureSkipVerify)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}
