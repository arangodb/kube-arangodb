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

package clients

import (
	"context"

	"github.com/spf13/cobra"

	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	pbShutdownV1 "github.com/arangodb/kube-arangodb/integrations/shutdown/v1/definition"
)

func init() {
	registerer.MustRegister("shutdown/v1", func(cfg *Config) Client {
		return &shutdownV1{
			cfg: cfg,
		}
	})
}

type shutdownV1 struct {
	cfg *Config
}

func (s *shutdownV1) Name() string {
	return "shutdown"
}

func (s *shutdownV1) Version() string {
	return "v1"
}

func (s *shutdownV1) Register(cmd *cobra.Command) error {
	withCommandRun(cmd, s.cfg, pbShutdownV1.NewShutdownV1Client).
		Register("shutdown", "Runs the Shutdown GRPC Call", func(ctx context.Context, client pbShutdownV1.ShutdownV1Client) error {
			_, err := client.Shutdown(ctx, &pbSharedV1.Empty{})

			return err
		})
	return nil
}
