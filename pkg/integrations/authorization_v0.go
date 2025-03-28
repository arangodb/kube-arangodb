//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package integrations

import (
	"context"

	"github.com/spf13/cobra"

	pbImplAuthorizationV0 "github.com/arangodb/kube-arangodb/integrations/authorization/v0"
	pbAuthorizationV0 "github.com/arangodb/kube-arangodb/integrations/authorization/v0/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func init() {
	registerer.Register(pbAuthorizationV0.Name, func() Integration {
		return &authorizationV0{}
	})
}

type authorizationV0 struct {
}

func (a authorizationV0) Name() string {
	return pbAuthorizationV0.Name
}

func (a authorizationV0) Description() string {
	return "Enable AuthorizationV0 Integration Service"
}

func (a authorizationV0) Register(cmd *cobra.Command, fs FlagEnvHandler) error {
	return nil
}

func (a authorizationV0) Handler(ctx context.Context, cmd *cobra.Command) (svc.Handler, error) {
	return pbImplAuthorizationV0.New(), nil
}

func (a authorizationV0) Init(ctx context.Context, cmd *cobra.Command) error {
	return nil
}
