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

	pbImplAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func init() {
	global.MustRegister("authorization-client", func(cmd *cobra.Command) (svc.Handler, bool, error) {
		if p, err := flagAuth.Get(cmd); err != nil {
			return nil, false, err
		} else if p == "" {
			return nil, false, nil
		}

		p, err := flagAuthMode.Get(cmd)
		if err != nil {
			return nil, false, err
		}

		authz, err := pbImplAuthorizationV1.New(cmd.Context(), pbImplAuthorizationV1.NewConfiguration().With(func(in pbImplAuthorizationV1.Configuration) pbImplAuthorizationV1.Configuration {
			in.Type = pbImplAuthorizationV1.ConfigurationType(p)
			return in
		}))
		if err != nil {
			return nil, false, err
		}

		return authz, true, nil
	})
}
