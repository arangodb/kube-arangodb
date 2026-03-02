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
	"github.com/spf13/cobra"

	sidecarSvcAuthz "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/db"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func init() {
	global.MustRegister("authorization", registerAuthorization)
}

func registerAuthorization(cmd *cobra.Command) (svc.Handler, bool, error) {
	if p, err := flagAuth.Get(cmd); err != nil {
		return nil, false, err
	} else if p == "" {
		return nil, false, nil
	}

	client := arangoDBDatabaseClient(cmd)

	return sidecarSvcAuthz.NewAuthorizer(db.NewClient(client).Database("_system")), true, nil
}
