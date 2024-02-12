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
	"fmt"
	"sort"
	"sync"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

var (
	lock       sync.Mutex
	registered []Factory
)

func register(i Factory) {
	lock.Lock()
	defer lock.Unlock()

	registered = append(registered, i)
}

func Register(cmd *cobra.Command) error {
	var c configuration

	return c.Register(cmd)
}

type configuration struct {
	registered []Integration

	health struct {
		shutdownEnabled bool

		config svc.Configuration
	}

	services struct {
		config svc.Configuration
	}
}

func (c *configuration) Register(cmd *cobra.Command) error {
	lock.Lock()
	defer lock.Unlock()

	c.registered = make([]Integration, len(registered))
	for id := range registered {
		c.registered[id] = registered[id]()
	}

	sort.Slice(c.registered, func(i, j int) bool {
		return c.registered[i].Name() < c.registered[j].Name()
	})

	subCommand := &cobra.Command{
		Use:  "integration",
		RunE: c.run,
	}

	f := subCommand.Flags()

	f.StringVar(&c.health.config.Address, "health.address", "0.0.0.0:9091", "Address to expose health service")
	f.BoolVar(&c.health.shutdownEnabled, "health.shutdown.enabled", true, "Determines if shutdown service should be enabled and exposed")
	f.StringVar(&c.services.config.Address, "services.address", "127.0.0.1:9092", "Address to expose services")

	for _, service := range c.registered {
		prefix := fmt.Sprintf("integration.%s", service.Name())

		f.Bool(prefix, false, service.Description())

		if err := service.Register(subCommand, func(name string) string {
			return fmt.Sprintf("%s.%s", prefix, name)
		}); err != nil {
			return errors.Wrapf(err, "Unable to register service %s", service.Name())
		}
	}

	cmd.AddCommand(subCommand)
	return nil
}

func (c *configuration) run(cmd *cobra.Command, args []string) error {
	handlers := make([]svc.Handler, 0, len(c.registered))

	for _, handler := range c.registered {
		if ok, err := cmd.Flags().GetBool(fmt.Sprintf("integration.%s", handler.Name())); err != nil {
			return err
		} else {
			logger.Str("service", handler.Name()).Bool("enabled", ok).Info("Service discovered")
			if ok {
				if svc, err := handler.Handler(shutdown.Context()); err != nil {
					return err
				} else {
					handlers = append(handlers, svc)
				}
			}
		}
	}

	var healthServices []svc.Handler

	if c.health.shutdownEnabled {
		healthServices = append(healthServices, shutdown.NewGlobalShutdownServer())
	}

	health := svc.NewHealthService(c.health.config, svc.Readiness, healthServices...)

	healthHandler := health.Start(shutdown.Context())

	logger.Str("address", healthHandler.Address()).Info("Health handler started")

	s := svc.NewService(c.services.config, handlers...).StartWithHealth(shutdown.Context(), health)

	logger.Str("address", s.Address()).Info("Service handler started")

	return s.Wait()
}
