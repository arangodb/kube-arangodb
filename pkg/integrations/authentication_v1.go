//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

	pbImplAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1"
	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/integrations/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func init() {
	registerer.Register(pbAuthenticationV1.Name, func() Integration {
		return &authenticationV1{}
	})
}

type authenticationV1 struct {
	config pbImplAuthenticationV1.Configuration
}

func (a *authenticationV1) Register(cmd *cobra.Command, fs FlagEnvHandler) error {
	return errors.Errors(
		fs.StringVar(&a.config.Path, "path", "", "Path to the JWT Folder"),
		fs.BoolVar(&a.config.Enabled, "enabled", true, "Defines if Authentication is enabled"),
		fs.DurationVar(&a.config.TTL, "ttl", pbImplAuthenticationV1.DefaultTTL, "TTL of the JWT cache"),
		fs.StringVar(&a.config.Create.DefaultUser, "token.user", pbImplAuthenticationV1.DefaultUser, "Default user of the Token"),
		fs.DurationVar(&a.config.Create.DefaultTTL, "token.ttl.default", pbImplAuthenticationV1.DefaultTokenDefaultTTL, "Default Token TTL"),
		fs.DurationVar(&a.config.Create.MinTTL, "token.ttl.min", pbImplAuthenticationV1.DefaultTokenMinTTL, "Min Token TTL"),
		fs.DurationVar(&a.config.Create.MaxTTL, "token.ttl.max", pbImplAuthenticationV1.DefaultTokenMaxTTL, "Max Token TTL"),
		fs.Uint16Var(&a.config.Create.MaxSize, "token.max-size", pbImplAuthenticationV1.DefaultMaxTokenSize, "Max Token max size in bytes"),
		fs.StringSliceVar(&a.config.Create.AllowedUsers, "token.allowed", []string{}, "Allowed users for the Token"),
	)
}

func (a *authenticationV1) Handler(ctx context.Context, cmd *cobra.Command) (svc.Handler, error) {
	if err := shared.FillAll(cmd, &a.config.Database); err != nil {
		return nil, err
	}
	return pbImplAuthenticationV1.New(ctx, a.config)
}

func (a *authenticationV1) Name() string {
	return pbAuthenticationV1.Name
}

func (a *authenticationV1) Description() string {
	return "Enable AuthenticationV1 Integration Service"
}

func (*authenticationV1) Init(ctx context.Context, cmd *cobra.Command) error {
	return nil
}
