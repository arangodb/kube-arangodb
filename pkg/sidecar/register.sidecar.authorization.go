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

	"github.com/spf13/cobra"

	pbImplAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1"
	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
	sidecarSvcAuthz "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/db"
	utilConstantsContext "github.com/arangodb/kube-arangodb/pkg/util/constants/context"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func newAuthorizationClient(ctx context.Context, cmd *cobra.Command) (svc.Handler, pbImplAuthorizationV1Shared.Evaluator, bool, error) {
	p, err := flagAuth.Get(cmd)
	if err != nil {
		return nil, nil, false, err
	} else if p == "" {
		return nil, pbImplAuthorizationV1Shared.NewAlwaysPlugin(), false, nil
	}

	c, ok := utilConstantsContext.ArangoDBClientCache.Get(ctx)
	if !ok {
		return nil, pbImplAuthorizationV1Shared.NewNeverPlugin(), false, errors.Errorf("Client not defined")
	}

	pm, err := flagAuthMode.Get(cmd)
	if err != nil {
		return nil, nil, false, err
	}

	pz := pbImplAuthorizationV1.ConfigurationType(pm)
	if err := pz.Validate(); err != nil {
		return nil, nil, false, err
	}

	auth := sidecarSvcAuthz.NewAuthorizer(db.NewClient(c).Database("_system"), pz)
	return auth, auth.Plugin(), true, nil
}
