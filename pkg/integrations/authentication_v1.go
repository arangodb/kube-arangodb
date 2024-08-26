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

	pbImplAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1"
	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
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

func (a *authenticationV1) Register(cmd *cobra.Command, arg ArgGen) error {
	f := cmd.Flags()

	f.StringVar(&a.config.Path, arg("path"), "", "Path to the JWT Folder")
	f.BoolVar(&a.config.Enabled, arg("enabled"), true, "Defines if Authentication is enabled")
	f.DurationVar(&a.config.TTL, arg("ttl"), pbImplAuthenticationV1.DefaultTTL, "TTL of the JWT cache")
	f.StringVar(&a.config.Create.DefaultUser, arg("token.user"), pbImplAuthenticationV1.DefaultUser, "Default user of the Token")
	f.DurationVar(&a.config.Create.DefaultTTL, arg("token.ttl.default"), pbImplAuthenticationV1.DefaultTokenDefaultTTL, "Default Token TTL")
	f.DurationVar(&a.config.Create.MinTTL, arg("token.ttl.min"), pbImplAuthenticationV1.DefaultTokenMinTTL, "Min Token TTL")
	f.DurationVar(&a.config.Create.MaxTTL, arg("token.ttl.max"), pbImplAuthenticationV1.DefaultTokenMaxTTL, "Max Token TTL")
	f.Uint16Var(&a.config.Create.MaxSize, arg("token.max-size"), pbImplAuthenticationV1.DefaultMaxTokenSize, "Max Token max size in bytes")
	f.StringSliceVar(&a.config.Create.AllowedUsers, arg("token.allowed"), []string{}, "Allowed users for the Token")

	return nil
}

func (a *authenticationV1) Handler(ctx context.Context) (svc.Handler, error) {
	return pbImplAuthenticationV1.New(ctx, a.config)
}

func (a *authenticationV1) Name() string {
	return pbAuthenticationV1.Name
}

func (a *authenticationV1) Description() string {
	return "Enable AuthenticationV1 Integration Service"
}
