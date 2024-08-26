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
	"github.com/spf13/cobra"
	pbHealth "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

func init() {
	registerer.MustRegister("health/v1", func(cfg *Config) Client {
		return &healthV1{
			cfg: cfg,
		}
	})
}

type healthV1 struct {
	cfg *Config
}

func (s *healthV1) Name() string {
	return "health"
}

func (s *healthV1) Version() string {
	return "v1"
}

func (s *healthV1) Register(cmd *cobra.Command) error {
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, c, err := client(shutdown.Context(), s.cfg, pbHealth.NewHealthClient)
		if err != nil {
			return err
		}
		defer c.Close()

		res, err := client.Check(shutdown.Context(), &pbHealth.HealthCheckRequest{})
		if err != nil {
			return err
		}

		switch s := res.GetStatus(); s {
		case pbHealth.HealthCheckResponse_SERVING:
			println("OK")
			return nil
		default:
			return errors.Errorf("Not healthy: %s", s.String())
		}
	}
	return nil
}
