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

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	pbImplEnvoyAuthV3 "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func init() {
	registerer.Register(pbImplEnvoyAuthV3.Name, func() Integration {
		return &envoyAuthV3{}
	})
}

type envoyAuthV3 struct {
	config pbImplEnvoyAuthV3.Configuration
}

func (a envoyAuthV3) Name() string {
	return pbImplEnvoyAuthV3.Name
}

func (a *envoyAuthV3) Description() string {
	return "Enable EnvoyAuthV3 Integration Service"
}

func (a *envoyAuthV3) Register(cmd *cobra.Command, fs FlagEnvHandler) error {
	return errors.Errors(
		fs.BoolVar(&a.config.Extensions.JWT, "extensions.jwt", true, "Defines if JWT extension is enabled"),
		fs.BoolVar(&a.config.Extensions.CookieJWT, "extensions.cookie.jwt", true, "Defines if Cookie JWT extension is enabled"),
	)
}

func (a *envoyAuthV3) Handler(ctx context.Context, cmd *cobra.Command) (svc.Handler, error) {
	f := cmd.Flags()

	v, err := f.GetString("services.address")
	if err != nil {
		return nil, err
	}

	c, _, err := ugrpc.NewGRPCClient(ctx, pbAuthenticationV1.NewAuthenticationV1Client, v)
	if err != nil {
		return nil, err
	}

	cfg := a.config

	cfg.AuthClient = c

	return pbImplEnvoyAuthV3.New(cfg), nil
}
