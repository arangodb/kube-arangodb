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

package platform

import (
	goHttp "net/http"

	"github.com/spf13/cobra"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"

	"github.com/arangodb/kube-arangodb/pkg/util"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
)

func arangoDBConnection(cmd *cobra.Command) (driver.Connection, error) {
	endpoint, err := flagArangoDBEndpoint.Get(cmd)
	if err != nil {
		return nil, err
	}

	var mods []util.Mod[goHttp.Transport]

	transport := operatorHTTP.Transport(mods...)

	stageEndpoint := endpoint

	connConfig := http.ConnectionConfig{
		Transport:          transport,
		DontFollowRedirect: true,
		Endpoints:          []string{stageEndpoint},
	}

	conn, err := http.NewConnection(connConfig)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
