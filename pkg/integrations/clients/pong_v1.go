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
	"fmt"

	"github.com/spf13/cobra"

	pbPongV1 "github.com/arangodb/kube-arangodb/integrations/pong/v1/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

func init() {
	registerer.MustRegister("pong/v1", func(cfg *Config) Client {
		return &pongV1{
			cfg: cfg,
		}
	})
}

type pongV1 struct {
	cfg *Config
}

func (s *pongV1) Name() string {
	return "pong"
}

func (s *pongV1) Version() string {
	return "v1"
}

func (s *pongV1) Register(cmd *cobra.Command) error {
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, c, err := client(shutdown.Context(), s.cfg, pbPongV1.NewPongV1Client)
		if err != nil {
			return err
		}
		defer c.Close()

		_, err = client.Ping(shutdown.Context(), &pbSharedV1.Empty{})
		if err != nil {
			return err
		}

		services, err := client.Services(shutdown.Context(), &pbSharedV1.Empty{})
		if err != nil {
			return err
		}

		for _, svc := range services.GetServices() {
			println(fmt.Sprintf("%s.%s: %s", svc.GetName(), svc.GetVersion(), util.BoolSwitch(svc.GetEnabled(), "Enabled", "Disabled")))
		}

		return nil
	}
	return nil
}
