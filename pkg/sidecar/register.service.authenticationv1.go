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

	pbImplAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func init() {
	register("authentication-v1", func(ctx context.Context, cmd *cobra.Command) (svc.Handler, bool, error) {
		authPath, err := flagAuth.Get(cmd)
		if err != nil {
			return nil, false, err
		}

		if authPath == "" {
			return nil, false, nil
		}

		handler, err := pbImplAuthenticationV1.New(ctx, pbImplAuthenticationV1.NewConfiguration().With(func(c pbImplAuthenticationV1.Configuration) pbImplAuthenticationV1.Configuration {
			c.Path = authPath
			return c
		}))
		if err != nil {
			return nil, false, err
		}

		return handler, true, nil
	})
}
